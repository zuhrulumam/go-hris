package payslip

import (
	transactionDom "github.com/zuhrulumam/go-hris/business/domain/transaction"
)

type UsecaseItf interface {
}

type Option struct {
	TransactionDom transactionDom.DomainItf
}

type payslip struct {
	TransactionDom transactionDom.DomainItf
}

func InitPayslipUsecase(opt Option) UsecaseItf {
	p := &payslip{
		TransactionDom: opt.TransactionDom,
	}

	return p
}
