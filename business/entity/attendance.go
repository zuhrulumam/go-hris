package entity

import "time"

type Attendance struct {
	ID                 uint
	UserID             uint
	Date               time.Time
	AttendancePeriodID uint
	CheckedInAt        *time.Time
	CheckedOutAt       *time.Time
	CreatedAt          time.Time
	UpdatedAt          *time.Time
	Version            uint
}

type CreateAttendance struct {
	UserID             uint
	AttendancePeriodID uint
	CheckInAt          time.Time
}

type UpdateAttendance struct {
	AttendanceID       uint
	UserID             uint
	AttendancePeriodID uint
	CheckOutAt         *time.Time
	CheckInAt          *time.Time
	Version            uint
}

type CreateOvertimeData struct {
	UserID             uint
	AttendancePeriodID uint
	Date               time.Time // or assume it's "today"
	Hours              float64   // max 3
	Description        string
}

type GetOvertimeFilter struct {
	UserID             uint
	AttendancePeriodID uint       // optional
	StartDate          *time.Time // optional
	EndDate            *time.Time // optional
	Date               time.Time  // optional, for single day query
}

type Overtime struct {
	ID                 uint
	UserID             uint
	Date               time.Time
	Hours              float64
	AttendancePeriodID uint
	Description        string
	CreatedAt          time.Time
}

type CheckIn struct {
	UserID uint
	Date   time.Time // Date of check-in
}

type CheckOut struct {
	UserID uint
	Date   time.Time // Date of check-out
}

type GetAttendance struct {
	UserID             uint
	AttendancePeriodID uint
	Date               time.Time
}

type GetAttendancePeriodFilter struct {
	ID           string
	Status       string
	UserID       string
	StartDate    *time.Time
	EndDate      *time.Time
	ContainsDate *time.Time
}

type AttendancePeriod struct {
	ID        uint
	StartDate time.Time
	EndDate   time.Time
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UpdateAttendancePeriod struct {
	ID        uint
	Status    *string
	StartDate *time.Time
	EndDate   *time.Time
	ClosedAt  *time.Time
}

type CreateAttendancePeriodRequest struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}
