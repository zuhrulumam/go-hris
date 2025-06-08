package attendance

import (
	"context"
	"net/http"
	"time"

	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
	x "github.com/zuhrulumam/go-hris/pkg/errors"
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
