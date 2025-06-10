package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zuhrulumam/go-hris/business/domain/user"
	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestRegister(t *testing.T) {
	tests := []struct {
		name        string
		input       entity.RegisterRequest
		mockSetup   func(mock sqlmock.Sqlmock, input entity.RegisterRequest)
		expectError bool
		errorText   string
	}{
		{
			name: "Successful registration",
			input: entity.RegisterRequest{
				Username: "johndoe",
				Password: "securepass",
				FullName: "John Doe",
				Salary:   5000000,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.RegisterRequest) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT count.*FROM "users"`).
					WithArgs(input.Username).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				mock.ExpectQuery(`INSERT INTO "users"`).
					WithArgs(input.Username, sqlmock.AnyArg(), input.FullName, "employee", input.Salary, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			},
			expectError: false,
		},
		{
			name: "Username already exists",
			input: entity.RegisterRequest{
				Username: "existinguser",
				Password: "123456",
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.RegisterRequest) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT count.*FROM "users"`).
					WithArgs(input.Username).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectError: true,
			errorText:   "username already taken",
		},
		{
			name: "DB error when checking username",
			input: entity.RegisterRequest{
				Username: "erroruser",
				Password: "123456",
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.RegisterRequest) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT count.*FROM "users"`).
					WithArgs(input.Username).
					WillReturnError(errors.New("query failed"))
			},
			expectError: true,
			errorText:   "failed to check existing user",
		},
		{
			name: "Hash password error",
			input: entity.RegisterRequest{
				Username: "badhash",
				Password: string(make([]byte, 73)), // bcrypt limit = 72
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.RegisterRequest) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT count.*FROM "users"`).
					WithArgs(input.Username).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectError: true,
			errorText:   "failed to hash password",
		},
		{
			name: "DB error on insert",
			input: entity.RegisterRequest{
				Username: "dberror",
				Password: "pass123",
				FullName: "DB Error",
				Salary:   3000000,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.RegisterRequest) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT count.*FROM "users"`).
					WithArgs(input.Username).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				mock.ExpectQuery(`INSERT INTO "users"`).
					WithArgs(input.Username, sqlmock.AnyArg(), input.FullName, "employee", input.Salary, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("insert failed"))
			},
			expectError: true,
			errorText:   "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			if tt.mockSetup != nil {
				tt.mockSetup(mock, tt.input)
			}

			u := user.InitUserDomain(user.Option{
				DB: db,
			})

			tx := db.Begin()
			ctx := context.WithValue(context.Background(), pkg.TxCtxValue, tx)

			err := u.Register(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name        string
		input       entity.LoginRequest
		mockSetup   func(mock sqlmock.Sqlmock, input entity.LoginRequest, hashedPassword string)
		expectError bool
		errorText   string
	}{
		{
			name: "Successful login",
			input: entity.LoginRequest{
				Username: "johndoe",
				Password: "securepass",
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.LoginRequest, hashedPassword string) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT.*FROM "users"`).
					WithArgs(input.Username, 1).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "username", "password", "full_name", "role", "salary", "created_at", "updated_at",
					}).AddRow(1, input.Username, hashedPassword, "John Doe", "employee", 5000000, time.Now(), time.Now()))
			},
			expectError: false,
		},
		{
			name: "User not found",
			input: entity.LoginRequest{
				Username: "unknown",
				Password: "any",
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.LoginRequest, _ string) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT.*FROM "users"`).
					WithArgs(input.Username, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorText:   "invalid username or password",
		},
		{
			name: "Wrong password",
			input: entity.LoginRequest{
				Username: "johndoe",
				Password: "wrongpass",
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.LoginRequest, hashedPassword string) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT.*FROM "users"`).
					WithArgs(input.Username, 1).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "username", "password", "full_name", "role", "salary", "created_at", "updated_at",
					}).AddRow(1, input.Username, hashedPassword, "John Doe", "employee", 5000000, time.Now(), time.Now()))
			},
			expectError: true,
			errorText:   "invalid username or password",
		},
		{
			name: "DB error",
			input: entity.LoginRequest{
				Username: "dberror",
				Password: "somepass",
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.LoginRequest, _ string) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT.*FROM "users"`).
					WithArgs(input.Username, 1).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
			errorText:   "failed to query user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			// Buat password hash dari string yang benar
			password := "securepass"
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

			if tt.mockSetup != nil {
				tt.mockSetup(mock, tt.input, string(hashedPassword))
			}

			u := user.InitUserDomain(user.Option{
				DB: db,
			})

			tx := db.Begin()
			ctx := context.WithValue(context.Background(), pkg.TxCtxValue, tx)

			result, err := u.Login(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.Username, result.Username)
				assert.Empty(t, result.Password)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetUsers(t *testing.T) {
	tests := []struct {
		name        string
		filter      entity.GetUserFilter
		mockSetup   func(mock sqlmock.Sqlmock, filter entity.GetUserFilter)
		expectError bool
		errorText   string
		expectLen   int
	}{
		{
			name:   "Get all users",
			filter: entity.GetUserFilter{},
			mockSetup: func(mock sqlmock.Sqlmock, _ entity.GetUserFilter) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT.*FROM "users"`).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "username", "password", "full_name", "role", "salary", "created_at", "updated_at",
					}).AddRow(1, "johndoe", "hashedpass", "John Doe", "employee", 5000000, time.Now(), time.Now()))
			},
			expectError: false,
			expectLen:   1,
		},
		{
			name: "Get by ID",
			filter: entity.GetUserFilter{
				ID: 1,
			},
			mockSetup: func(mock sqlmock.Sqlmock, f entity.GetUserFilter) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT.*FROM "users" WHERE id = ?`).
					WithArgs(f.ID).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "username", "password", "full_name", "role", "salary", "created_at", "updated_at",
					}).AddRow(f.ID, "johndoe", "hashedpass", "John Doe", "employee", 5000000, time.Now(), time.Now()))
			},
			expectError: false,
			expectLen:   1,
		},
		{
			name: "Get by Role",
			filter: entity.GetUserFilter{
				Role: "admin",
			},
			mockSetup: func(mock sqlmock.Sqlmock, f entity.GetUserFilter) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT.*FROM "users" WHERE role = ?`).
					WithArgs(f.Role).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "username", "password", "full_name", "role", "salary", "created_at", "updated_at",
					}).AddRow(2, "adminuser", "hashedpass", "Admin", "admin", 10000000, time.Now(), time.Now()))
			},
			expectError: false,
			expectLen:   1,
		},
		{
			name: "Get by Email",
			filter: entity.GetUserFilter{
				Email: "user@example.com",
			},
			mockSetup: func(mock sqlmock.Sqlmock, f entity.GetUserFilter) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT.*FROM "users" WHERE email = ?`).
					WithArgs(f.Email).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "username", "password", "full_name", "role", "email", "salary", "created_at", "updated_at",
					}).AddRow(3, "emailuser", "hashedpass", "Email User", "employee", f.Email, 7000000, time.Now(), time.Now()))
			},
			expectError: false,
			expectLen:   1,
		},
		{
			name:   "DB error",
			filter: entity.GetUserFilter{},
			mockSetup: func(mock sqlmock.Sqlmock, _ entity.GetUserFilter) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT.*FROM "users"`).
					WillReturnError(errors.New("db failure"))
			},
			expectError: true,
			errorText:   "failed to get users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			if tt.mockSetup != nil {
				tt.mockSetup(mock, tt.filter)
			}

			r := user.InitUserDomain(user.Option{
				DB: db,
			})

			tx := db.Begin()
			ctx := context.WithValue(context.Background(), pkg.TxCtxValue, tx)

			users, err := r.GetUsers(ctx, tt.filter)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, users, tt.expectLen)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
