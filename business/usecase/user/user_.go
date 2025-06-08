package user

import (
	"context"

	"github.com/zuhrulumam/go-hris/business/entity"
)

func (p *user) Register(ctx context.Context, input entity.RegisterRequest) error {

	return p.UserDom.Register(ctx, input)
}

func (p *user) Login(ctx context.Context, input entity.LoginRequest) (*entity.User, error) {

	return p.UserDom.Login(ctx, input)
}
