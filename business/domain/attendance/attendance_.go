package attendance

import (
	"context"
	"net/http"
	"time"

	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"

	x "github.com/zuhrulumam/go-hris/pkg/errors"
)

// func (p *attendance) GetAvailableAttendanceSpot(ctx context.Context, data entity.GetAvailableAttendanceSpot) ([]entity.AttendanceSpot, error) {

// 	var (
// 		result []entity.AttendanceSpot
// 		db     = pkg.GetTransactionFromCtx(ctx, p.db)
// 	)

// 	db = db.Model(&entity.AttendanceSpot{})

// 	// Filter by type
// 	if data.VehicleType != "" {
// 		db = db.Where("type = ?", data.VehicleType)
// 	}

// 	// Filter by active status
// 	if data.Active != nil {
// 		db = db.Where("active = ?", *data.Active)
// 	}

// 	// Filter by occupied status
// 	if data.Occupied != nil {
// 		db = db.Where("occupied = ?", *data.Occupied)
// 	}

// 	// if use lock
// 	if data.UseLock {
// 		db.Clauses(clause.Locking{Strength: "UPDATE"})
// 	}

// 	// Only get the first match
// 	err := db.Find(&result).Error
// 	if err != nil {
// 		return result, x.WrapWithCode(err, http.StatusInternalServerError, "error get available attendance spot")
// 	}

// 	return result, nil
// }

// func (p *attendance) InsertVehicle(ctx context.Context, data entity.InsertVehicle) error {
// 	db := pkg.GetTransactionFromCtx(ctx, p.db)

// 	vehicle := entity.Vehicle{
// 		VehicleNumber: data.VehicleNumber,
// 		VehicleType:   data.VehicleType,
// 		SpotID:        data.SpotID,
// 		ParkedAt:      time.Now(),
// 	}

// 	if err := db.WithContext(ctx).Create(&vehicle).Error; err != nil {
// 		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to insert vehicle")
// 	}

// 	return nil
// }

// func (p *attendance) UpdateAttendanceSpot(ctx context.Context, data entity.UpdateAttendanceSpot) error {

// 	db := pkg.GetTransactionFromCtx(ctx, p.db)

// 	tx := db.WithContext(ctx).Model(&entity.AttendanceSpot{})

// 	// Build conditional WHERE clause
// 	if data.ID > 0 {
// 		tx = tx.Where("id = ?", data.ID)
// 	} else if data.Floor > 0 && data.Row > 0 && data.Col > 0 {
// 		tx = tx.Where("floor = ? AND row = ? AND col = ?", data.Floor, data.Row, data.Col)
// 	} else {
// 		return x.NewWithCode(http.StatusBadRequest, "must provide either spot_id or (floor, row, col)")
// 	}

// 	updates := map[string]interface{}{}
// 	if data.Occupied != nil {
// 		updates["occupied"] = data.Occupied
// 	}

// 	if len(updates) == 0 {
// 		return x.NewWithCode(http.StatusBadRequest, "no updates provided")
// 	}

// 	if err := tx.Updates(updates).Error; err != nil {
// 		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to update attendance spot")
// 	}

// 	return nil
// }

// func (p *attendance) UpdateVehicle(ctx context.Context, data entity.UpdateVehicle) error {
// 	db := pkg.GetTransactionFromCtx(ctx, p.db)

// 	if data.ID < 1 {
// 		return x.NewWithCode(http.StatusBadRequest, "vehicle id is required")
// 	}

// 	updates := map[string]interface{}{}
// 	if data.UnparkedAt != nil {
// 		updates["unparked_at"] = data.UnparkedAt
// 	}

// 	if len(updates) == 0 {
// 		return x.NewWithCode(http.StatusBadRequest, "no updates provided")
// 	}

// 	if err := db.WithContext(ctx).
// 		Model(&entity.Vehicle{}).
// 		Where("id = ?", data.ID).
// 		Updates(updates).Error; err != nil {
// 		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to update vehicle")
// 	}

// 	return nil
// }

// func (p *attendance) GetVehicle(ctx context.Context, data entity.SearchVehicle) (entity.Vehicle, error) {
// 	var (
// 		result entity.Vehicle
// 	)

// 	db := p.db.WithContext(ctx).Model(&entity.Vehicle{})

// 	// Filter by type
// 	if data.VehicleNumber != "" {
// 		db = db.Where("vehicle_number = ?", data.VehicleNumber)
// 	}

// 	// Only get the first match
// 	err := db.Order("id DESC").First(&result).Error
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return result, x.WrapWithCode(err, http.StatusNotFound, "vehicle not found")
// 		}
// 		return result, x.WrapWithCode(err, http.StatusInternalServerError, "failed get vehicle")
// 	}

// 	return result, nil
// }

func (p *attendance) CreateAttendance(ctx context.Context, data entity.CreateAttendance) error {
	db := pkg.GetTransactionFromCtx(ctx, p.db)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Prevent weekend check-ins
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return x.NewWithCode(http.StatusBadRequest, "cannot check in on weekends")
	}

	// Check if attendance already exists
	var existing entity.Attendance
	err := db.WithContext(ctx).Where("user_id = ? AND date = ?", data.UserID, today).First(&existing).Error
	if err == nil {
		return x.NewWithCode(http.StatusConflict, "already checked in today")
	}

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
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Fetch existing attendance
	var att entity.Attendance
	if err := db.WithContext(ctx).Where("user_id = ? AND date = ?", data.UserID, today).First(&att).Error; err != nil {
		return x.WrapWithCode(err, http.StatusNotFound, "attendance not found for today")
	}

	// Prevent double checkout
	if att.CheckedOutAt != nil {
		return x.NewWithCode(http.StatusConflict, "already checked out")
	}

	// Update checkout time
	if err := db.WithContext(ctx).Model(&att).Update("checked_out_at", now).Error; err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to check out")
	}

	return nil
}

func (p *attendance) CreateOvertime(ctx context.Context, data entity.CreateOvertimeData) error {
	db := pkg.GetTransactionFromCtx(ctx, p.db)
	now := time.Now()

	// Enforce max 3 hours
	if data.Hours > 3 {
		return x.NewWithCode(http.StatusBadRequest, "overtime cannot be more than 3 hours")
	}
	if data.Hours <= 0 {
		return x.NewWithCode(http.StatusBadRequest, "overtime hours must be greater than 0")
	}

	// Ensure the user has checked out already (if required by logic)
	var att entity.Attendance
	today := time.Date(data.Date.Year(), data.Date.Month(), data.Date.Day(), 0, 0, 0, 0, data.Date.Location())
	err := db.WithContext(ctx).Where("user_id = ? AND date = ?", data.UserID, today).First(&att).Error
	if err != nil {
		return x.NewWithCode(http.StatusBadRequest, "cannot submit overtime without checking in today")
	}
	if att.CheckedOutAt == nil {
		return x.NewWithCode(http.StatusBadRequest, "must check out before submitting overtime")
	}

	// Prevent duplicate for the same day
	var existing entity.Overtime
	err = db.WithContext(ctx).Where("user_id = ? AND date = ?", data.UserID, today).First(&existing).Error
	if err == nil {
		return x.NewWithCode(http.StatusConflict, "overtime already submitted for today")
	}

	// Insert overtime
	overtime := entity.Overtime{
		UserID:             data.UserID,
		Date:               today,
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
	db := pkg.GetTransactionFromCtx(ctx, p.db)

	var result []entity.Overtime
	query := db.WithContext(ctx).Where("user_id = ? AND attendance_period_id = ?", filter.UserID, filter.AttendancePeriodID)

	if err := query.Order("date ASC").Find(&result).Error; err != nil {
		return nil, x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch overtime records")
	}

	return result, nil
}
