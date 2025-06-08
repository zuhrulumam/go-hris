package user

import (
	"context"

	transactionDom "github.com/zuhrulumam/go-hris/business/domain/transaction"
	userDom "github.com/zuhrulumam/go-hris/business/domain/user"
	"github.com/zuhrulumam/go-hris/business/entity"
)

type UsecaseItf interface {
	Register(ctx context.Context, input entity.RegisterRequest) error
	Login(ctx context.Context, input entity.LoginRequest) (*entity.User, error)
}

type Option struct {
	UserDom        userDom.DomainItf
	TransactionDom transactionDom.DomainItf
}

type user struct {
	UserDom        userDom.DomainItf
	TransactionDom transactionDom.DomainItf
}

func InitUserUsecase(opt Option) UsecaseItf {
	p := &user{
		UserDom:        opt.UserDom,
		TransactionDom: opt.TransactionDom,
	}

	return p
}
