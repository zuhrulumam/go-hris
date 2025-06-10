package handler

import (
	"time"

	"github.com/zuhrulumam/go-hris/business/entity"
)

type CheckInResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CheckOutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Success    bool   `json:"success"`
	HumanError string `json:"human_error"`
	DebugError string `json:"debug_error"`
}

type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type PayslipListResponse struct {
	Data       []PayslipDataResp `json:"data"`
	TotalData  int               `json:"total_data"`
	TotalPages int               `json:"total_pages"`
}

type GetPayrollSummaryResponse struct {
	Items      []entity.PayrollSummaryItem `json:"data"`
	GrandTotal float64                     `json:"grand_total"`
}

type PayslipDataResp struct {
	AttendancePeriodID uint      `json:"attendance_period_id"`
	BaseSalary         string    `json:"base_salary"`
	WorkingDays        int       `json:"working_days"`
	AttendedDays       int       `json:"attended_days"`
	AttendanceAmount   string    `json:"attendance_amount"`
	OvertimeHours      float64   `json:"overtime_hours"`
	OvertimePay        string    `json:"overtime_pay"`
	ReimbursementTotal string    `json:"reimbursement_total"`
	TotalPay           string    `json:"total_pay"`
	CreatedAt          time.Time `json:"created_at"`
}
