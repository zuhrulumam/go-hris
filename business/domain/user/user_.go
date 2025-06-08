package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/zuhrulumam/go-hris/pkg"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/zuhrulumam/go-hris/business/entity"
	x "github.com/zuhrulumam/go-hris/pkg/errors"
)

func (u *user) Register(ctx context.Context, req entity.RegisterRequest) error {
	db := pkg.GetTransactionFromCtx(ctx, u.db)

	// Check if username already exists
	var count int64
	if err := db.WithContext(ctx).
		Model(&entity.User{}).
		Where("username = ?", req.Username).
		Count(&count).Error; err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to check existing user")
	}
	if count > 0 {
		return x.NewWithCode(http.StatusBadRequest, "username already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to hash password")
	}

	user := entity.User{
		Username: req.Username,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Role:     entity.RoleEmployee,
		Salary:   req.Salary,
	}

	if err := db.WithContext(ctx).Create(&user).Error; err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to create user")
	}

	return nil
}

func (u *user) Login(ctx context.Context, req entity.LoginRequest) (*entity.User, error) {
	db := pkg.GetTransactionFromCtx(ctx, u.db)

	var user entity.User
	err := db.WithContext(ctx).
		Where("username = ?", req.Username).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, x.NewWithCode(http.StatusUnauthorized, "invalid username or password")
	} else if err != nil {
		return nil, x.WrapWithCode(err, http.StatusInternalServerError, "failed to query user")
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, x.NewWithCode(http.StatusUnauthorized, "invalid username or password")
	}

	// Remove password before returning (optional)
	user.Password = ""

	return &user, nil
}
