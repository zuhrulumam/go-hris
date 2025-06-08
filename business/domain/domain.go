package domain

import (
	"github.com/zuhrulumam/go-hris/business/domain/attendance"
	"github.com/zuhrulumam/go-hris/business/domain/transaction"
	"gorm.io/gorm"
)

type Domain struct {
	Attendance  attendance.DomainItf
	Transaction transaction.DomainItf
}

type Option struct {
	DB *gorm.DB
}

func Init(opt Option) *Domain {
	d := &Domain{
		Attendance: attendance.InitAttendanceDomain(attendance.Option{
			DB: opt.DB,
		}),
		Transaction: transaction.Init(transaction.Option{
			DB: opt.DB,
		}),
	}

	return d
}
