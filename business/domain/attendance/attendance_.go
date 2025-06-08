package attendance

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	x "github.com/zuhrulumam/go-hris/pkg/errors"
)

func (p *attendance) GetAvailableAttendanceSpot(ctx context.Context, data entity.GetAvailableAttendanceSpot) ([]entity.AttendanceSpot, error) {

	var (
		result []entity.AttendanceSpot
		db     = pkg.GetTransactionFromCtx(ctx, p.db)
	)

	db = db.Model(&entity.AttendanceSpot{})

	// Filter by type
	if data.VehicleType != "" {
		db = db.Where("type = ?", data.VehicleType)
	}

	// Filter by active status
	if data.Active != nil {
		db = db.Where("active = ?", *data.Active)
	}

	// Filter by occupied status
	if data.Occupied != nil {
		db = db.Where("occupied = ?", *data.Occupied)
	}

	// if use lock
	if data.UseLock {
		db.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	// Only get the first match
	err := db.Find(&result).Error
	if err != nil {
		return result, x.WrapWithCode(err, http.StatusInternalServerError, "error get available attendance spot")
	}

	return result, nil
}

func (p *attendance) InsertVehicle(ctx context.Context, data entity.InsertVehicle) error {
	db := pkg.GetTransactionFromCtx(ctx, p.db)

	vehicle := entity.Vehicle{
		VehicleNumber: data.VehicleNumber,
		VehicleType:   data.VehicleType,
		SpotID:        data.SpotID,
		ParkedAt:      time.Now(),
	}

	if err := db.WithContext(ctx).Create(&vehicle).Error; err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to insert vehicle")
	}

	return nil
}

func (p *attendance) UpdateAttendanceSpot(ctx context.Context, data entity.UpdateAttendanceSpot) error {

	db := pkg.GetTransactionFromCtx(ctx, p.db)

	tx := db.WithContext(ctx).Model(&entity.AttendanceSpot{})

	// Build conditional WHERE clause
	if data.ID > 0 {
		tx = tx.Where("id = ?", data.ID)
	} else if data.Floor > 0 && data.Row > 0 && data.Col > 0 {
		tx = tx.Where("floor = ? AND row = ? AND col = ?", data.Floor, data.Row, data.Col)
	} else {
		return x.NewWithCode(http.StatusBadRequest, "must provide either spot_id or (floor, row, col)")
	}

	updates := map[string]interface{}{}
	if data.Occupied != nil {
		updates["occupied"] = data.Occupied
	}

	if len(updates) == 0 {
		return x.NewWithCode(http.StatusBadRequest, "no updates provided")
	}

	if err := tx.Updates(updates).Error; err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to update attendance spot")
	}

	return nil
}

func (p *attendance) UpdateVehicle(ctx context.Context, data entity.UpdateVehicle) error {
	db := pkg.GetTransactionFromCtx(ctx, p.db)

	if data.ID < 1 {
		return x.NewWithCode(http.StatusBadRequest, "vehicle id is required")
	}

	updates := map[string]interface{}{}
	if data.UnparkedAt != nil {
		updates["unparked_at"] = data.UnparkedAt
	}

	if len(updates) == 0 {
		return x.NewWithCode(http.StatusBadRequest, "no updates provided")
	}

	if err := db.WithContext(ctx).
		Model(&entity.Vehicle{}).
		Where("id = ?", data.ID).
		Updates(updates).Error; err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to update vehicle")
	}

	return nil
}

func (p *attendance) GetVehicle(ctx context.Context, data entity.SearchVehicle) (entity.Vehicle, error) {
	var (
		result entity.Vehicle
	)

	db := p.db.WithContext(ctx).Model(&entity.Vehicle{})

	// Filter by type
	if data.VehicleNumber != "" {
		db = db.Where("vehicle_number = ?", data.VehicleNumber)
	}

	// Only get the first match
	err := db.Order("id DESC").First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return result, x.WrapWithCode(err, http.StatusNotFound, "vehicle not found")
		}
		return result, x.WrapWithCode(err, http.StatusInternalServerError, "failed get vehicle")
	}

	return result, nil
}
