package reimbursement

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"

	x "github.com/zuhrulumam/go-hris/pkg/errors"
)

func (r *reimbursement) SubmitReimbursement(ctx context.Context, data entity.SubmitReimbursementData) error {
	db := pkg.GetTransactionFromCtx(ctx, r.db)

	// Validate
	if data.Amount <= 0 {
		return x.NewWithCode(http.StatusBadRequest, "reimbursement amount must be greater than 0")
	}
	if strings.TrimSpace(data.Description) == "" {
		return x.NewWithCode(http.StatusBadRequest, "description is required")
	}

	reim := entity.Reimbursement{
		UserID:             data.UserID,
		AttendancePeriodID: data.AttendancePeriodID,
		Amount:             data.Amount,
		Description:        data.Description,
		CreatedAt:          time.Now(),
	}

	if err := db.WithContext(ctx).Create(&reim).Error; err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to submit reimbursement")
	}

	return nil
}

func (r *reimbursement) GetReimbursements(ctx context.Context, filter entity.GetReimbursementFilter) ([]entity.Reimbursement, error) {
	var (
		result []entity.Reimbursement
		db     = pkg.GetTransactionFromCtx(ctx, r.db).WithContext(ctx).Model(&entity.Reimbursement{})
	)

	// Dynamic filters
	if filter.UserID > 0 {
		db = db.Where("user_id = ?", filter.UserID)
	}

	if filter.AttendancePeriodID > 0 {
		db = db.Where("attendance_period_id = ?", filter.AttendancePeriodID)
	}

	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}

	if !filter.StartDate.IsZero() {
		db = db.Where("created_at >= ?", filter.StartDate)
	}

	if !filter.EndDate.IsZero() {
		db = db.Where("created_at <= ?", filter.EndDate)
	}

	// Query execution
	err := db.Order("created_at ASC").Find(&result).Error
	if err != nil {
		return nil, x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch reimbursements")
	}

	return result, nil
}
