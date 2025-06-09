package domain

import (
	"github.com/zuhrulumam/go-hris/business/domain/attendance"
	"github.com/zuhrulumam/go-hris/business/domain/payslip"
	"github.com/zuhrulumam/go-hris/business/domain/reimbursement"
	"github.com/zuhrulumam/go-hris/business/domain/transaction"
	"github.com/zuhrulumam/go-hris/business/domain/user"
	"gorm.io/gorm"
)

type Domain struct {
	Attendance    attendance.DomainItf
	Transaction   transaction.DomainItf
	Reimbursement reimbursement.DomainItf
	Payslip       payslip.DomainItf
	User          user.DomainItf
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
		Reimbursement: reimbursement.InitReimbursementDomain(reimbursement.Option{
			DB: opt.DB,
		}),
		Payslip: payslip.InitPayslipDomain(payslip.Option{
			DB: opt.DB,
		}),
		User: user.InitUserDomain(user.Option{
			DB: opt.DB,
		}),
	}

	return d
}
