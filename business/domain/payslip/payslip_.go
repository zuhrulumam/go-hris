package payslip

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
	x "github.com/zuhrulumam/go-hris/pkg/errors"
	"gorm.io/gorm"
)

func (p *payslip) CreatePayroll(ctx context.Context, data entity.CreatePayrollData) error {
	db := pkg.GetTransactionFromCtx(ctx, p.db)

	// Step 1: Check if payroll already exists for this period
	var count int64
	err := db.WithContext(ctx).Model(&entity.Payslip{}).
		Where("attendance_period_id = ?", data.AttendancePeriodID).
		Count(&count).Error
	if err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to check existing payroll")
	}
	if count > 0 {
		return x.NewWithCode(http.StatusConflict, "payroll already processed for this period")
	}

	// Step 2: Get all employees
	var employees []entity.User
	err = db.WithContext(ctx).Where("role = ?", "employee").Find(&employees).Error
	if err != nil {
		return x.WrapWithCode(err, http.StatusInternalServerError, "failed to get employees")
	}

	// Step 3: For each employee, calculate payslip
	for _, emp := range employees {
		// 3.1 Get attendance days
		var attendedDays int64
		err = db.WithContext(ctx).Model(&entity.Attendance{}).
			Where("user_id = ? AND attendance_period_id = ?", emp.ID, data.AttendancePeriodID).
			Count(&attendedDays).Error
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to count attendance")
		}

		// 3.2 Get overtime hours
		var totalOvertime float64
		err = db.WithContext(ctx).Model(&entity.Overtime{}).
			Select("COALESCE(SUM(hours), 0)").
			Where("user_id = ? AND attendance_period_id = ?", emp.ID, data.AttendancePeriodID).
			Scan(&totalOvertime).Error
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to calculate overtime")
		}

		// 3.3 Get reimbursement total
		var reimbursementTotal float64
		err = db.WithContext(ctx).Model(&entity.Reimbursement{}).
			Select("COALESCE(SUM(amount), 0)").
			Where("user_id = ? AND attendance_period_id = ?", emp.ID, data.AttendancePeriodID).
			Scan(&reimbursementTotal).Error
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to calculate reimbursement")
		}

		// 3.4 Calculate final values
		const workingDays = 20 // assuming fixed working days/month
		proratedSalary := (float64(attendedDays) / float64(workingDays)) * emp.Salary
		overtimeAmount := totalOvertime * (emp.Salary / float64(workingDays) / 8.0 * 2)
		totalPay := proratedSalary + overtimeAmount + reimbursementTotal

		// 3.5 Create payslip record
		payslip := entity.Payslip{
			UserID:             emp.ID,
			AttendancePeriodID: data.AttendancePeriodID,
			BaseSalary:         emp.Salary,
			WorkingDays:        workingDays,
			AttendedDays:       int(attendedDays),
			AttendanceAmount:   proratedSalary,
			OvertimeHours:      totalOvertime,
			OvertimeAmount:     overtimeAmount,
			ReimbursementTotal: reimbursementTotal,
			TotalPay:           totalPay,
			CreatedAt:          time.Now(),
		}

		if err := db.WithContext(ctx).Create(&payslip).Error; err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to create payslip")
		}
	}

	return nil
}

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
