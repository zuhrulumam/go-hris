package user

import (
	transactionDom "github.com/zuhrulumam/go-hris/business/domain/transaction"
)

type UsecaseItf interface {
}

type Option struct {
	TransactionDom transactionDom.DomainItf
}

type user struct {
	TransactionDom transactionDom.DomainItf
}

func InitUserUsecase(opt Option) UsecaseItf {
	p := &user{
		TransactionDom: opt.TransactionDom,
	}

	return p
}
