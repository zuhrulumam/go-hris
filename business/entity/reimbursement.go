package entity

import "time"

type Reimbursement struct {
	ID                 uint
	UserID             uint
	AttendancePeriodID uint
	Amount             float64
	Description        string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type SubmitReimbursementData struct {
	UserID             uint
	AttendancePeriodID uint
	Amount             float64
	Description        string
	Date               time.Time
}

type GetReimbursementFilter struct {
	UserID             uint
	AttendancePeriodID uint
	Status             string
	StartDate          time.Time
	EndDate            time.Time
}
