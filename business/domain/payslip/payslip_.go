package payslip

import (
	"context"
	"errors"
	"net/http"

	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
	x "github.com/zuhrulumam/go-hris/pkg/errors"
	"gorm.io/gorm"
)

func (p *payslip) GetPayslip(ctx context.Context, req entity.GetPayslipRequest) (*entity.Payslip, error) {
	db := pkg.GetTransactionFromCtx(ctx, p.db)

	var payslip entity.Payslip
	err := db.WithContext(ctx).
		Where("user_id = ? AND attendance_period_id = ?", req.UserID, req.AttendancePeriodID).
		First(&payslip).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, x.NewWithCode(http.StatusNotFound, "payslip not found")
	} else if err != nil {
		return nil, x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch payslip")
	}

	return &payslip, nil
}

func (p *payslip) GetPayrollSummary(ctx context.Context, req entity.GetPayrollSummaryRequest) (*entity.GetPayrollSummaryResponse, error) {
	db := pkg.GetTransactionFromCtx(ctx, p.db)

	var results []entity.PayrollSummaryItem
	err := db.WithContext(ctx).
		Table("payslips").
		Select("payslips.user_id, users.full_name, payslips.total_pay").
		Joins("JOIN users ON payslips.user_id = users.id").
		Where("payslips.attendance_period_id = ?", req.AttendancePeriodID).
		Scan(&results).Error

	if err != nil {
		return nil, x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch payroll summary")
	}

	var grandTotal float64
	for _, item := range results {
		grandTotal += item.TotalPay
	}

	return &entity.GetPayrollSummaryResponse{
		Items:      results,
		GrandTotal: grandTotal,
	}, nil
}

func (p *payslip) IsPayrollExists(ctx context.Context, periodID uint) (bool, error) {
	var count int64
	err := p.db.WithContext(ctx).
		Model(&entity.Payslip{}).
		Where("attendance_period_id = ?", periodID).
		Count(&count).Error
	if err != nil {
		return false, x.WrapWithCode(err, http.StatusInternalServerError, "failed to check payroll existence")
	}
	return count > 0, nil
}

func (p *payslip) CreatePayslip(ctx context.Context, payslips []entity.Payslip) error {
	db := pkg.GetTransactionFromCtx(ctx, p.db)

	if len(payslips) == 0 {
		return x.NewWithCode(http.StatusBadRequest, "no payslip data provided")
	}

	if err := db.WithContext(ctx).Create(&payslips).Error; err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to create payslips")
	}

	return nil
}

func (p *payslip) CreatePayrollJob(ctx context.Context, job entity.PayrollJob) (*entity.PayrollJob, error) {
	db := pkg.GetTransactionFromCtx(ctx, p.db)

	if err := db.WithContext(ctx).Create(&job).Error; err != nil {
		return nil, x.WrapWithCode(err, http.StatusInternalServerError, "failed to create payroll job")
	}

	return &job, nil
}

func (p *payslip) UpdatePayslipJob(ctx context.Context, data entity.UpdatePayslipJob) error {
	db := pkg.GetTransactionFromCtx(ctx, p.db)

	if data.ID == 0 {
		return x.NewWithCode(http.StatusBadRequest, "payslip job ID is required")
	}

	updates := map[string]interface{}{}

	if data.Status != "" {
		updates["status"] = data.Status
	}
	if data.StartedAt != nil {
		updates["started_at"] = data.StartedAt
	}
	if data.CompletedAt != nil {
		updates["completed_at"] = data.CompletedAt
	}
	if data.FailedReason != nil {
		updates["failed_reason"] = data.FailedReason
	}

	if len(updates) == 0 {
		return x.NewWithCode(http.StatusBadRequest, "no updates provided")
	}

	// Optimistic concurrency control: increment version
	updates["version"] = data.Version + 1

	tx := db.WithContext(ctx).
		Model(&entity.PayrollJob{}).
		Where("id = ? AND version = ?", data.ID, data.Version).
		Updates(updates)

	if tx.RowsAffected == 0 {
		return x.NewWithCode(http.StatusConflict, "payslip job was updated by someone else, please retry")
	}

	if tx.Error != nil {
		return x.WrapWithCode(tx.Error, http.StatusInternalServerError, "failed to update payslip job")
	}

	return nil
}
