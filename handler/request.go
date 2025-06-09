package handler

import "time"

type CheckInRequest struct {
	EmployeeID uint `json:"employee_id" validate:"required"`
}

type CheckOutRequest struct {
	EmployeeID uint `json:"employee_id" validate:"required"`
}

type OvertimeRequest struct {
	Hours       float64 `json:"hours" validate:"required,max=3"`
	Description string  `json:"description"`
}

type ReimbursementRequest struct {
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Description string    `json:"description" validate:"required"`
	Date        time.Time `json:"date" validate:"required"`
}

type CreatePayrollRequest struct {
	PeriodID uint `json:"period_id" example:"1" binding:"required"`
}
