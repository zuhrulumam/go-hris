package payslip

import (
	"context"

	"github.com/hibiken/asynq"
	attendanceDom "github.com/zuhrulumam/go-hris/business/domain/attendance"
	payslipDom "github.com/zuhrulumam/go-hris/business/domain/payslip"
	reimbursementDom "github.com/zuhrulumam/go-hris/business/domain/reimbursement"
	transactionDom "github.com/zuhrulumam/go-hris/business/domain/transaction"
	userDom "github.com/zuhrulumam/go-hris/business/domain/user"
	"github.com/zuhrulumam/go-hris/business/entity"
)

type UsecaseItf interface {
	CreatePayroll(ctx context.Context, periodID uint) error
	GetPayslip(ctx context.Context, userID, periodID uint) (*entity.Payslip, error)

	CreatePayslipForUser(ctx context.Context, data entity.CreatePayslipForUserData) error
}

type Option struct {
	PayslipDom       payslipDom.DomainItf
	TransactionDom   transactionDom.DomainItf
	AttendanceDom    attendanceDom.DomainItf
	ReimbursementDom reimbursementDom.DomainItf
	UserDom          userDom.DomainItf
	AsynqClient      *asynq.Client
}

type payslip struct {
	TransactionDom   transactionDom.DomainItf
	PayslipDom       payslipDom.DomainItf
	AttendanceDom    attendanceDom.DomainItf
	ReimbursementDom reimbursementDom.DomainItf
	UserDom          userDom.DomainItf
	AsynqClient      *asynq.Client
}

func InitPayslipUsecase(opt Option) UsecaseItf {
	p := &payslip{
		TransactionDom:   opt.TransactionDom,
		PayslipDom:       opt.PayslipDom,
		AttendanceDom:    opt.AttendanceDom,
		ReimbursementDom: opt.ReimbursementDom,
		UserDom:          opt.UserDom,
		AsynqClient:      opt.AsynqClient,
	}

	return p
}
