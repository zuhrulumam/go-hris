package user

import (
	"context"

	"github.com/zuhrulumam/go-hris/business/entity"
	"gorm.io/gorm"
)

//go:generate mockgen -source=business/domain/user/user.go -destination=mocks/domain/user/mock_user.go -package=mocks
type DomainItf interface {
	Register(ctx context.Context, req entity.RegisterRequest) error
	Login(ctx context.Context, req entity.LoginRequest) (*entity.User, error)

	GetUsers(ctx context.Context, filter entity.GetUserFilter) ([]entity.User, error)
}

type user struct {
	db *gorm.DB
}

type Option struct {
	DB *gorm.DB
}

func InitUserDomain(opt Option) DomainItf {
	p := &user{
		db: opt.DB,
	}

	return p
}
