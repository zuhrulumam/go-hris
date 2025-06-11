package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zuhrulumam/go-hris/business/entity"
	uc "github.com/zuhrulumam/go-hris/business/usecase/user"
	mockUser "github.com/zuhrulumam/go-hris/mocks/domain/user"
	"go.uber.org/mock/gomock"
)

func TestUser_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDom := mockUser.NewMockDomainItf(ctrl)

	usecase := uc.InitUserUsecase(uc.Option{
		UserDom: mockUserDom,
	})

	tests := []struct {
		name      string
		input     entity.RegisterRequest
		mockSetup func()
		expectErr bool
	}{
		{
			name: "success register",
			input: entity.RegisterRequest{
				Username: "Umam",
				Password: "secure123",
			},
			mockSetup: func() {
				mockUserDom.EXPECT().
					Register(gomock.Any(), gomock.AssignableToTypeOf(entity.RegisterRequest{})).
					Return(nil)
			},
			expectErr: false,
		},
		{
			name: "domain error",
			input: entity.RegisterRequest{
				Username: "Umam",
				Password: "wrong",
			},
			mockSetup: func() {
				mockUserDom.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(errors.New("duplicate email"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := usecase.Register(context.Background(), tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDom := mockUser.NewMockDomainItf(ctrl)
	usecase := uc.InitUserUsecase(uc.Option{
		UserDom: mockUserDom,
	})

	tests := []struct {
		name       string
		input      entity.LoginRequest
		mockUser   entity.User
		mockErr    error
		expectErr  bool
		expectRole entity.UserRole
	}{
		{
			name: "login as employee",
			input: entity.LoginRequest{
				Username: "emp@example.com",
				Password: "123456",
			},
			mockUser: entity.User{
				ID:       1,
				Username: "employee",
				Role:     entity.RoleEmployee,
			},
			mockErr:    nil,
			expectErr:  false,
			expectRole: entity.RoleEmployee,
		},
		{
			name: "login as admin",
			input: entity.LoginRequest{
				Username: "admin@example.com",
				Password: "adminpass",
			},
			mockUser: entity.User{
				ID:       2,
				Username: "admin",
				Role:     entity.RoleAdmin,
			},
			mockErr:    nil,
			expectErr:  false,
			expectRole: entity.RoleAdmin,
		},
		{
			name: "invalid login",
			input: entity.LoginRequest{
				Username: "wrong@example.com",
				Password: "wrongpass",
			},
			mockUser:  entity.User{},
			mockErr:   errors.New("invalid credentials"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserDom.EXPECT().
				Login(gomock.Any(), tt.input).
				Return(&tt.mockUser, tt.mockErr)

			token, err := usecase.Login(context.Background(), tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				assert.NoError(t, err)
			}
		})
	}
}
