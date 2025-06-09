package handler

type OvertimeRequest struct {
	Date        string  `json:"date" validate:"required"`
	Hours       float64 `json:"hours" validate:"required,max=3"`
	Description string  `json:"description"`
}

type ReimbursementRequest struct {
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Description string  `json:"description" validate:"required"`
	Date        string  `json:"date" validate:"required"`
}

type CreatePayrollRequest struct {
	PeriodID uint `json:"period_id" example:"1" binding:"required"`
}

type RegisterRequest struct {
	Username string  `json:"username" binding:"required"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=6"`
	Fullname string  `json:"fullname" binding:"required"`
	Salary   float64 `json:"salary" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateAttendancePeriodRequest struct {
	StartDate string `json:"start_date" binding:"required" example:"2025-06-01T00:00:00Z"`
	EndDate   string `json:"end_date" binding:"required" example:"2025-06-15T23:59:59Z"`
}
