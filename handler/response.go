package handler

import "github.com/zuhrulumam/go-hris/business/entity"

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
	Data       []entity.Payslip `json:"data"`
	TotalData  int              `json:"total_data"`
	TotalPages int              `json:"total_pages"`
}

type GetPayrollSummaryResponse struct {
	Items      []entity.PayrollSummaryItem `json:"items"`
	GrandTotal float64                     `json:"grand_total"`
}
