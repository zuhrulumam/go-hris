package reimbursement

import (
	transactionDom "github.com/zuhrulumam/go-hris/business/domain/transaction"
)

type UsecaseItf interface {
}

type Option struct {
	TransactionDom transactionDom.DomainItf
}

type reimbursement struct {
	TransactionDom transactionDom.DomainItf
}

func InitReimbursementUsecase(opt Option) UsecaseItf {
	p := &reimbursement{
		TransactionDom: opt.TransactionDom,
	}

	return p
}
