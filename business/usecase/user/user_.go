package user

import (
	"context"

	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
	"github.com/zuhrulumam/go-hris/pkg/tracer"
)

func (p *user) Register(ctx context.Context, input entity.RegisterRequest) error {

	return p.UserDom.Register(ctx, input)
}

func (p *user) Login(ctx context.Context, input entity.LoginRequest) (string, error) {
	ctx, done := tracer.Start(ctx, "useruc.login")
	defer done()

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
