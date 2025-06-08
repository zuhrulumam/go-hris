package entity

import "time"

type Reimbursement struct {
	ID                 uint
	UserID             uint
	AttendancePeriodID uint
	Amount             float64
	Description        string
	CreatedAt          time.Time
}

type SubmitReimbursementData struct {
	UserID             uint
	AttendancePeriodID uint
	Amount             float64
	Description        string
}

type GetReimbursementFilter struct {
	UserID             uint
	AttendancePeriodID uint
}
