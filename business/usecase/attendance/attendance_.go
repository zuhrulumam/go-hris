package attendance

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
	x "github.com/zuhrulumam/go-hris/pkg/errors"
	"gorm.io/gorm"
)

func (p *attendance) CheckIn(ctx context.Context, data entity.CheckIn) error {
	if data.Date.Weekday() == time.Saturday || data.Date.Weekday() == time.Sunday {
		return x.NewWithCode(http.StatusBadRequest, "cannot check in on weekends")
	}

	// Create attendance record
	return p.AttendanceDom.CreateAttendance(ctx, entity.CreateAttendance{
		UserID:    data.UserID,
		CheckInAt: data.Date,
	})

}

func (p *attendance) CheckOut(ctx context.Context, data entity.CheckOut) error {
	return p.TransactionDom.RunInTx(ctx, func(newCtx context.Context) error {

		// Update with check-out time
		return p.AttendanceDom.UpdateAttendance(newCtx, entity.UpdateAttendance{
			// ID:         att.ID,
			CheckOutAt: pkg.TimePtr(data.Date),
		})
	})
}

func (p *attendance) CreateOvertime(ctx context.Context, data entity.CreateOvertimeData) error {
	return p.TransactionDom.RunInTx(ctx, func(newCtx context.Context) error {
		// Check: max 3 hours
		if data.Hours > 3 {
			return x.NewWithCode(http.StatusBadRequest, "overtime cannot be more than 3 hours per day")
		}

		// Check: already checked out for the day
		att, err := p.AttendanceDom.GetAttendance(newCtx, entity.GetAttendance{
			UserID: data.UserID,
			Date:   data.Date,
		})
		if err != nil {
			return x.WrapWithCode(err, http.StatusNotFound, "attendance not found for the date")
		}
		if att[0].CheckedOutAt == nil {
			return x.NewWithCode(http.StatusBadRequest, "must check out before submitting overtime")
		}

		// Check: already submitted overtime?
		existing, err := p.AttendanceDom.GetOvertime(newCtx, entity.GetOvertimeFilter{
			UserID: data.UserID,
			Date:   data.Date,
		})
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if len(existing) > 0 {
			return x.NewWithCode(http.StatusBadRequest, "overtime already submitted for this date")
		}

		return p.AttendanceDom.CreateOvertime(newCtx, data)
	})
}

func (p *attendance) GetOvertime(ctx context.Context, filter entity.GetOvertimeFilter) ([]entity.Overtime, error) {
	return p.AttendanceDom.GetOvertime(ctx, filter)
}
