package user

import (
	"context"

	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
)

func (p *user) Register(ctx context.Context, input entity.RegisterRequest) error {

	return p.UserDom.Register(ctx, input)
}

func (p *user) Login(ctx context.Context, input entity.LoginRequest) (string, error) {

	var (
		isAdmin bool
	)

	user, err := p.UserDom.Login(ctx, input)
	if err != nil {
		return "", err
	}

	if user.Role == entity.RoleAdmin {
		isAdmin = true
	}

	return pkg.GenerateJWT(user.ID, user.Username, isAdmin)
}
