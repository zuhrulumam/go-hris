package attendance

import (
	"context"

	"github.com/zuhrulumam/go-hris/business/entity"
	"gorm.io/gorm"
)

//go:generate mockgen -source=business/domain/attendance/attendance.go -destination=mocks/mock_attendance.go -package=mocks
type DomainItf interface {
	GetAvailableAttendanceSpot(ctx context.Context, data entity.GetAvailableAttendanceSpot) ([]entity.AttendanceSpot, error)
	InsertVehicle(ctx context.Context, data entity.InsertVehicle) error
	UpdateAttendanceSpot(ctx context.Context, data entity.UpdateAttendanceSpot) error
	UpdateVehicle(ctx context.Context, data entity.UpdateVehicle) error
	GetVehicle(ctx context.Context, data entity.SearchVehicle) (entity.Vehicle, error)
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
