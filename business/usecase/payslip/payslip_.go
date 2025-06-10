package payslip

import (
	"context"
	"net/http"
	"time"

	"github.com/zuhrulumam/go-hris/business/entity"
	x "github.com/zuhrulumam/go-hris/pkg/errors"
	"github.com/zuhrulumam/go-hris/task"
)

func (p *payslip) GetPayslip(ctx context.Context, filter entity.GetPayslipRequest) ([]entity.Payslip, int64, int, error) {

	payslips, totalData, totalPage, err := p.PayslipDom.GetPayslip(ctx, filter)
	if err != nil {
		return nil, 0, 0, err
	}

	if len(payslips) < 1 {
		return nil, totalData, totalPage, x.NewWithCode(http.StatusNotFound, "payslip not found")
	}

	return payslips, totalData, totalPage, nil
}

func (p *payslip) CreatePayroll(ctx context.Context, periodID uint) error {

	return p.TransactionDom.RunInTx(ctx, func(newCtx context.Context) error {
		// Get All Users
		users, err := p.UserDom.GetUsers(ctx, entity.GetUserFilter{
			Role: string(entity.RoleEmployee),
		})
		if err != nil {
			return err
		}

		// queue task to asynq
		for _, user := range users {
			// Insert PayrollJob
			job := entity.PayrollJob{
				AttendancePeriodID: periodID,
				UserID:             user.ID,
				Status:             "pending",
				NextRunAt:          time.Now(),
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			}

			j, err := p.PayslipDom.CreatePayrollJob(newCtx, job)
			if err != nil {
				return err
			}

			task, err := task.NewCreatePayrollTask(periodID, user.ID, j.ID)
			if err != nil {
				return err
			}

			_, err = p.AsynqClient.Enqueue(task)
			if err != nil {
				return err
			}
		}
		return nil
	})

}

func (p *payslip) CreatePayslipForUser(ctx context.Context, data entity.CreatePayslipForUserData) error {
	return p.TransactionDom.RunInTx(ctx, func(newCtx context.Context) error {

		var (
			salary float64
		)

		user, err := p.UserDom.GetUsers(newCtx, entity.GetUserFilter{
			ID: data.UserID,
		})
		if err != nil {
			return err
		}

		if len(user) < 1 {
			return x.NewWithCode(http.StatusNotFound, "user not found")
		}

		salary = user[0].Salary

		// Get all attendance, overtime, and reimbursement data for the period
		attendances, err := p.AttendanceDom.GetAttendance(newCtx, entity.GetAttendance{
			AttendancePeriodID: data.PeriodID,
			UserID:             data.UserID,
		})
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch attendance data")
		}

		overtimes, err := p.AttendanceDom.GetOvertime(newCtx, entity.GetOvertimeFilter{
			AttendancePeriodID: data.PeriodID,
			UserID:             data.UserID,
		})
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch overtime data")
		}

		reimbursements, err := p.ReimbursementDom.GetReimbursements(newCtx, entity.GetReimbursementFilter{
			AttendancePeriodID: data.PeriodID,
			UserID:             data.UserID,
		})
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch reimbursement data")
		}

		// Create payslips
		var payslip entity.Payslip
		userAttendances := attendances
		userOvertimes := overtimes
		userReimbursements := reimbursements

		attendedDays := len(userAttendances)
		workingDays := 22 // default value, or fetch from period

		overtimeHours := float64(0)
		for _, ot := range userOvertimes {
			overtimeHours += ot.Hours
		}

		reimbursementTotal := float64(0)
		for _, rb := range userReimbursements {
			reimbursementTotal += rb.Amount
		}

		attendanceAmount := (float64(attendedDays) / float64(workingDays)) * salary
		overtimeAmount := overtimeHours * (salary / float64(workingDays)) * 1.5
		totalPay := attendanceAmount + overtimeAmount + reimbursementTotal

		payslip = entity.Payslip{
			UserID:             data.UserID,
			AttendancePeriodID: data.PeriodID,
			BaseSalary:         salary,
			WorkingDays:        workingDays,
			AttendedDays:       attendedDays,
			AttendanceAmount:   attendanceAmount,
			OvertimeHours:      overtimeHours,
			OvertimePay:        overtimeAmount,
			ReimbursementTotal: reimbursementTotal,
			TotalPay:           totalPay,
			CreatedAt:          time.Now(),
		}

		// Save payslip
		err = p.PayslipDom.CreatePayslip(newCtx, []entity.Payslip{payslip})
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to save payslips")
		}

		// update job
		err = p.PayslipDom.UpdatePayslipJob(newCtx, entity.UpdatePayslipJob{
			ID:     data.JobID,
			Status: "completed",
		})
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to save payslip job")
		}

		return nil
	})
}

func (p *payslip) GetPayrollSummary(ctx context.Context, filter entity.GetPayrollSummaryRequest) (*entity.GetPayrollSummaryResponse, error) {
	summary, err := p.PayslipDom.GetPayrollSummary(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(summary.Items) < 1 {
		return nil, x.NewWithCode(http.StatusNotFound, "payroll summary not found")
	}

	return summary, nil
}
