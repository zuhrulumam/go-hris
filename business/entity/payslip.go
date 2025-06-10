package entity

import "time"

type Payslip struct {
	ID                 uint
	UserID             uint
	AttendancePeriodID uint
	BaseSalary         float64
	WorkingDays        int
	AttendedDays       int
	AttendanceAmount   float64
	OvertimeHours      float64
	OvertimePay        float64
	ReimbursementTotal float64
	TotalPay           float64
	CreatedAt          time.Time
}

type CreatePayrollData struct {
	AttendancePeriodID uint
}

type GetPayslipRequest struct {
	UserID             *uint
	AttendancePeriodID *uint
	Status             *string
	Limit              int
	Page               int
}

type GetPayrollSummaryRequest struct {
	AttendancePeriodIDs []uint
}

type PayrollSummaryItem struct {
	UserID   uint
	Username string
	TotalPay float64
}

type GetPayrollSummaryResponse struct {
	Items      []PayrollSummaryItem
	GrandTotal float64
}

type PayrollJob struct {
	ID                 uint
	AttendancePeriodID uint
	UserID             uint
	Status             string // pending, processing, done, failed
	Attempts           int
	LastError          *string
	NextRunAt          time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type UpdatePayslipJob struct {
	ID           uint
	Status       string
	StartedAt    *time.Time
	FailedReason *string
}

type CreatePayslipForUserData struct {
	UserID   uint
	PeriodID uint
	JobID    uint
}
