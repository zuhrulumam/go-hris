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
}

type CreateAttendance struct {
	UserID             uint
	AttendancePeriodID uint
	CheckInAt          time.Time
}

type UpdateAttendance struct {
	UserID             uint
	AttendancePeriodID uint
	CheckOutAt         *time.Time
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
	AttendancePeriodID uint
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
