package entity

import "time"

type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleEmployee UserRole = "employee"
)

type User struct {
	ID        uint
	Username  string
	Password  string
	FullName  string
	Role      UserRole
	Salary    float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RegisterRequest struct {
	Username string
	Password string
	FullName string
	Role     string // "admin" or "employee"
	Salary   float64
}

type LoginRequest struct {
	Username string
	Password string
}

type GetUserFilter struct {
	ID    uint
	Role  string
	Email string
}
