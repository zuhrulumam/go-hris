package attendance

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
	"gorm.io/gorm"

	x "github.com/zuhrulumam/go-hris/pkg/errors"
)

func (p *attendance) CreateAttendance(ctx context.Context, data entity.CreateAttendance) error {
	db := pkg.GetTransactionFromCtx(ctx, p.db)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Insert new attendance
	attendance := entity.Attendance{
		UserID:             data.UserID,
		Date:               today,
		AttendancePeriodID: data.AttendancePeriodID,
		CheckedInAt:        &now,
		CreatedAt:          now,
	}

	if err := db.WithContext(ctx).Create(&attendance).Error; err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to check in")
	}

	return nil
}

func (p *attendance) UpdateAttendance(ctx context.Context, data entity.UpdateAttendance) error {
	db := pkg.GetTransactionFromCtx(ctx, p.db)

	if data.AttendanceID == 0 {
		return x.NewWithCode(http.StatusBadRequest, "attendance ID is required")
	}

	updates := map[string]interface{}{}

	if data.CheckInAt != nil {
		updates["check_in_at"] = data.CheckInAt
	}

	if data.CheckOutAt != nil {
		updates["checked_out_at"] = data.CheckOutAt
	}

	if len(updates) == 0 {
		return x.NewWithCode(http.StatusBadRequest, "no updates provided")
	}

	// Add version update
	updates["version"] = data.Version + 1

	// Optimistic update
	tx := db.WithContext(ctx).
		Model(&entity.Attendance{}).
		Where("id = ? AND version = ?", data.AttendanceID, data.Version).
		Updates(updates)

	if tx.Error != nil {
		return x.WrapWithCode(tx.Error, http.StatusInternalServerError, "failed to update attendance")
	}

	if tx.RowsAffected == 0 {
		return x.NewWithCode(http.StatusConflict, "attendance was updated by someone else, please retry")
	}

	return nil
}

func (p *attendance) CreateOvertime(ctx context.Context, data entity.CreateOvertimeData) error {
	db := pkg.GetTransactionFromCtx(ctx, p.db)
	now := time.Now()

	// Insert overtime
	overtime := entity.Overtime{
		UserID:             data.UserID,
		Date:               data.Date,
		Hours:              data.Hours,
		Description:        data.Description,
		AttendancePeriodID: data.AttendancePeriodID,
		CreatedAt:          now,
	}

	if err := db.WithContext(ctx).Create(&overtime).Error; err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to submit overtime")
	}

	return nil
}

func (p *attendance) GetOvertime(ctx context.Context, filter entity.GetOvertimeFilter) ([]entity.Overtime, error) {
	var (
		result []entity.Overtime
		db     = pkg.GetTransactionFromCtx(ctx, p.db)
	)

	db = db.Model(&entity.Overtime{}).WithContext(ctx)

	// Dynamic filters
	if filter.UserID > 0 {
		db = db.Where("user_id = ?", filter.UserID)
	}

	if filter.AttendancePeriodID > 0 {
		db = db.Where("attendance_period_id = ?", filter.AttendancePeriodID)
	}

	if !filter.Date.IsZero() {
		db = db.Where("date = ?", filter.Date)
	}

	// Order and execute
	err := db.Order("date ASC").Find(&result).Error
	if err != nil {
		return result, x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch overtime records")
	}

	return result, nil
}

func (r *attendance) GetAttendance(ctx context.Context, filter entity.GetAttendance) ([]entity.Attendance, error) {
	var (
		att []entity.Attendance
		db  = r.db.WithContext(ctx).Model(&entity.Attendance{})
	)

	// Dynamic filters
	if filter.UserID > 0 {
		db = db.Where("user_id = ?", filter.UserID)
	}

	if !filter.Date.IsZero() {
		db = db.Where("DATE(checked_in_at) = ?", filter.Date.Format("2006-01-02"))
	}

	if filter.AttendancePeriodID > 0 {
		db = db.Where("attendance_period_id = ?", filter.AttendancePeriodID)
	}

	// Query execution
	err := db.Order("checked_in_at ASC").Find(&att).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, x.NewWithCode(http.StatusNotFound, "attendance not found")
		}
		return nil, x.WrapWithCode(err, http.StatusInternalServerError, "failed to get attendance")
	}

	return att, nil
}

func (r *attendance) CreateAttendancePeriod(ctx context.Context, data entity.AttendancePeriod) error {
	db := pkg.GetTransactionFromCtx(ctx, r.db)
	if err := db.WithContext(ctx).Create(&data).Error; err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to create attendance period")
	}
	return nil
}

func (r *attendance) UpdateAttendancePeriod(ctx context.Context, data entity.UpdateAttendancePeriod) error {
	db := pkg.GetTransactionFromCtx(ctx, r.db)

	if data.ID < 1 {
		return x.NewWithCode(http.StatusBadRequest, "attendance period ID is required")
	}

	updates := map[string]interface{}{}

	if data.Status != nil {
		updates["status"] = *data.Status
	}
	if data.StartDate != nil {
		updates["start_date"] = *data.StartDate
	}
	if data.EndDate != nil {
		updates["end_date"] = *data.EndDate
	}
	if data.ClosedAt != nil {
		updates["closed_at"] = *data.ClosedAt
	}

	if len(updates) == 0 {
		return x.NewWithCode(http.StatusBadRequest, "no updates provided")
	}

	if err := db.WithContext(ctx).
		Model(&entity.AttendancePeriod{}).
		Where("id = ?", data.ID).
		Updates(updates).Error; err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to update attendance period")
	}

	return nil

}

func (r *attendance) GetAttendancePeriods(ctx context.Context, filter entity.GetAttendancePeriodFilter) ([]entity.AttendancePeriod, error) {
	var result []entity.AttendancePeriod
	db := pkg.GetTransactionFromCtx(ctx, r.db).WithContext(ctx).Model(&entity.AttendancePeriod{})

	// Dynamic filters
	if filter.ID != "" {
		db = db.Where("id = ?", filter.ID)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}

	if filter.StartDate != nil {
		db = db.Where("start_date >= ?", filter.StartDate)
	}

	if filter.EndDate != nil {
		db = db.Where("end_date <= ?", filter.EndDate)
	}

	if filter.UserID != "" {
		db = db.Where("user_id = ?", filter.UserID)
	}

	if filter.ContainsDate != nil {
		db = db.Where("start_date <= ? AND end_date >= ?", filter.ContainsDate, filter.ContainsDate)
	}

	err := db.Order("start_date DESC").Find(&result).Error
	if err != nil {
		return nil, x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch attendance periods")
	}

	return result, nil
}
