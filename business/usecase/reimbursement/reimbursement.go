package reimbursement

import (
	"context"

	attendanceDom "github.com/zuhrulumam/go-hris/business/domain/attendance"
	reimbursementDom "github.com/zuhrulumam/go-hris/business/domain/reimbursement"
	transactionDom "github.com/zuhrulumam/go-hris/business/domain/transaction"
	"github.com/zuhrulumam/go-hris/business/entity"
)

type UsecaseItf interface {
	SubmitReimbursement(ctx context.Context, data entity.SubmitReimbursementData) error
	GetReimbursement(ctx context.Context, filter entity.GetReimbursementFilter) ([]entity.Reimbursement, error)
}

type Option struct {
	ReimbursementDom reimbursementDom.DomainItf
	TransactionDom   transactionDom.DomainItf
	AttendanceDom    attendanceDom.DomainItf
}

type reimbursement struct {
	ReimbursementDom reimbursementDom.DomainItf
	TransactionDom   transactionDom.DomainItf
	AttendanceDom    attendanceDom.DomainItf
}

func InitReimbursementUsecase(opt Option) UsecaseItf {
	p := &reimbursement{
		ReimbursementDom: opt.ReimbursementDom,
		TransactionDom:   opt.TransactionDom,
		AttendanceDom:    opt.AttendanceDom,
	}

	return p
}
