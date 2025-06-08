package attendance

import (
	"context"

	attendanceDom "github.com/zuhrulumam/go-hris/business/domain/attendance"
	transactionDom "github.com/zuhrulumam/go-hris/business/domain/transaction"
	"github.com/zuhrulumam/go-hris/business/entity"
)

type UsecaseItf interface {
	Unpark(ctx context.Context, data entity.UnPark) error
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
