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
	OvertimeAmount     float64
	ReimbursementTotal float64
	TotalPay           float64
	CreatedAt          time.Time
}

type CreatePayrollData struct {
	AttendancePeriodID uint
}

type GetPayslipRequest struct {
	UserID             uint
	AttendancePeriodID uint
}

type GetPayrollSummaryRequest struct {
	AttendancePeriodID uint
}

type PayrollSummaryItem struct {
	UserID   uint
	FullName string
	TotalPay float64
}

type GetPayrollSummaryResponse struct {
	Items      []PayrollSummaryItem
	GrandTotal float64
}
