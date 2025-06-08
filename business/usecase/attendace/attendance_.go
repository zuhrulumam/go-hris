package attendance

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
	x "github.com/zuhrulumam/go-hris/pkg/errors"
)

func (p *attendance) Park(ctx context.Context, data entity.Park) error {

	err := p.TransactionDom.RunInTx(ctx, func(newCtx context.Context) error {

		// check attendance_spot by vehicle type, active, and not occupied
		pSpots, err := p.AttendanceDom.GetAvailableAttendanceSpot(newCtx, entity.GetAvailableAttendanceSpot{
			VehicleType: data.VehicleType,
			Active:      pkg.BoolPtr(true),
			Occupied:    pkg.BoolPtr(false),
			UseLock:     true,
		})
		if err != nil {
			return err
		}

		if len(pSpots) < 1 {
			return errors.New("no available attendance")
		}

		spot := pSpots[0]
		spotID := fmt.Sprintf("%d-%d-%d", spot.Floor, spot.Row, spot.Col)

		// update attendance_spot occupied = true as floor row col
		err = p.AttendanceDom.UpdateAttendanceSpot(newCtx, entity.UpdateAttendanceSpot{
			ID:       spot.ID,
			Occupied: pkg.BoolPtr(true),
		})
		if err != nil {
			return err
		}

		// insert vehicle
		err = p.AttendanceDom.InsertVehicle(newCtx, entity.InsertVehicle{
			VehicleNumber: data.VehicleNumber,
			VehicleType:   string(data.VehicleType),
			SpotID:        spotID,
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *attendance) Unpark(ctx context.Context, data entity.UnPark) error {

	return p.TransactionDom.RunInTx(ctx, func(newCtx context.Context) error {

		// get vehicle by spotID, vehicle number, and UnparkedAt null
		vec, err := p.AttendanceDom.GetVehicle(newCtx, entity.SearchVehicle{
			VehicleNumber: data.VehicleNumber,
		})
		if err != nil {
			return err
		}

		if vec.UnparkedAt != nil {
			return x.NewWithCode(http.StatusBadRequest, "already unparked")
		}

		// update vehicle
		err = p.AttendanceDom.UpdateVehicle(newCtx, entity.UpdateVehicle{
			ID:         vec.ID,
			UnparkedAt: pkg.TimePtr(time.Now()),
		})
		if err != nil {
			return err
		}

		sp, err := pkg.ParseSpotID(vec.SpotID)
		if err != nil {
			return err
		}

		// update attendance_spot to occupied = false
		err = p.AttendanceDom.UpdateAttendanceSpot(newCtx, entity.UpdateAttendanceSpot{
			Floor:    sp.Floor,
			Row:      sp.Row,
			Col:      sp.Col,
			Occupied: pkg.BoolPtr(false),
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func (p *attendance) AvailableSpot(ctx context.Context, data entity.GetAvailablePark) ([]entity.AttendanceSpot, error) {
	// check attendance_spot by vehicle type, active, and not occupied
	return p.AttendanceDom.GetAvailableAttendanceSpot(ctx, entity.GetAvailableAttendanceSpot{
		VehicleType: data.VehicleType,
		Active:      pkg.BoolPtr(true),
		Occupied:    pkg.BoolPtr(false),
	})

}

func (p *attendance) SearchVehicle(ctx context.Context, data entity.SearchVehicle) (entity.Vehicle, error) {
	return p.AttendanceDom.GetVehicle(ctx, data)
}
