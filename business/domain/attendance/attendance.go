package attendance

import (
	"context"

	"github.com/zuhrulumam/go-hris/business/entity"
	"gorm.io/gorm"
)

//go:generate mockgen -source=business/domain/attendance/attendance.go -destination=mocks/mock_attendance.go -package=mocks
type DomainItf interface {
	CreateAttendance(ctx context.Context, data entity.CreateAttendance) error
	UpdateAttendance(ctx context.Context, data entity.UpdateAttendance) error

	GetAttendance(ctx context.Context, filter entity.GetAttendance) ([]entity.Attendance, error)

	CreateOvertime(ctx context.Context, data entity.CreateOvertimeData) error
	GetOvertime(ctx context.Context, filter entity.GetOvertimeFilter) ([]entity.Overtime, error)

	CreateAttendancePeriod(ctx context.Context, data entity.AttendancePeriod) error
	UpdateAttendancePeriod(ctx context.Context, data entity.UpdateAttendancePeriod) error
	GetAttendancePeriods(ctx context.Context, filter entity.GetAttendancePeriodFilter) ([]entity.AttendancePeriod, error)
}

type attendance struct {
	db *gorm.DB
}

type Option struct {
	DB *gorm.DB
}

func InitAttendanceDomain(opt Option) DomainItf {
	p := &attendance{
		db: opt.DB,
	}

	return p
}
