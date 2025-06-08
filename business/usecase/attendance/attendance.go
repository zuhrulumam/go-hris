package attendance

import (
	"context"

	attendanceDom "github.com/zuhrulumam/go-hris/business/domain/attendance"
	transactionDom "github.com/zuhrulumam/go-hris/business/domain/transaction"
	"github.com/zuhrulumam/go-hris/business/entity"
)

type UsecaseItf interface {
	CheckIn(ctx context.Context, data entity.CheckIn) error
	CheckOut(ctx context.Context, data entity.CheckOut) error
	CreateOvertime(ctx context.Context, data entity.CreateOvertimeData) error
	GetOvertime(ctx context.Context, filter entity.GetOvertimeFilter) ([]entity.Overtime, error)
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
