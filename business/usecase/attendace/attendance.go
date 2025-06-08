package attendance

import (
	"context"

	attendanceDom "github.com/zuhrulumam/go-hris/business/domain/attendance"
	transactionDom "github.com/zuhrulumam/go-hris/business/domain/transaction"
	"github.com/zuhrulumam/go-hris/business/entity"
)

type UsecaseItf interface {
	Park(ctx context.Context, data entity.Park) error
	Unpark(ctx context.Context, data entity.UnPark) error
	AvailableSpot(ctx context.Context, data entity.GetAvailablePark) ([]entity.AttendanceSpot, error)
	SearchVehicle(ctx context.Context, data entity.SearchVehicle) (entity.Vehicle, error)
}

type Option struct {
	AttendanceDom  attendanceDom.DomainItf
	TransactionDom transactionDom.DomainItf
}

type attendance struct {
	AttendanceDom  attendanceDom.DomainItf
	TransactionDom transactionDom.DomainItf
}

func InitAttendanceUsecase(opt Option) UsecaseItf {
	p := &attendance{
		AttendanceDom:  opt.AttendanceDom,
		TransactionDom: opt.TransactionDom,
	}

	return p
}
