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
	db := pkg.GetTransactionFromCtx(ctx, r.db)

	var result []entity.Reimbursement

	err := db.WithContext(ctx).
		Where("user_id = ? AND attendance_period_id = ?", filter.UserID, filter.AttendancePeriodID).
		Order("created_at ASC").
		Find(&result).Error

	if err != nil {
		return nil, x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch reimbursements")
	}

	return result, nil
}
