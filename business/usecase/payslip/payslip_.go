package payslip

import (
	"context"
	"net/http"
	"time"

	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
	x "github.com/zuhrulumam/go-hris/pkg/errors"
	"github.com/zuhrulumam/go-hris/task"
)

func (p *payslip) GetPayslip(ctx context.Context, userID, periodID uint) (*entity.Payslip, error) {
	// 1. Get user with salary info
	users, err := p.UserDom.GetUsers(ctx, entity.GetUserFilter{ID: userID})
	if err != nil || len(users) == 0 {
		return nil, x.NewWithCode(http.StatusNotFound, "user not found")
	}
	user := users[0]

	// 2. Get attendance
	attendances, err := p.AttendanceDom.GetAttendance(ctx, entity.GetAttendance{
		AttendancePeriodID: periodID,
	})
	if err != nil {
		return nil, x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch attendance")
	}

	var userAttendances []entity.Attendance
	for _, a := range attendances {
		if a.UserID == userID {
			userAttendances = append(userAttendances, a)
		}
	}
	attendedDays := len(userAttendances)
	workingDays := 22 // or fetch dynamically

	// 3. Get overtime
	overtimes, err := p.AttendanceDom.GetOvertime(ctx, entity.GetOvertimeFilter{
		AttendancePeriodID: periodID,
	})
	if err != nil {
		return nil, x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch overtime")
	}

	var userOT []entity.Overtime
	var totalHours float64
	for _, o := range overtimes {
		if o.UserID == userID {
			userOT = append(userOT, o)
			totalHours += o.Hours
		}
	}

	// 4. Get reimbursements
	reimbursements, err := p.ReimbursementDom.GetReimbursements(ctx, entity.GetReimbursementFilter{
		AttendancePeriodID: periodID,
	})
	if err != nil {
		return nil, x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch reimbursements")
	}

	var userReimbursements []entity.Reimbursement
	var totalReimbursement float64
	for _, r := range reimbursements {
		if r.UserID == userID {
			userReimbursements = append(userReimbursements, r)
			totalReimbursement += r.Amount
		}
	}

	// 5. Calculate components
	attendanceAmount := (float64(attendedDays) / float64(workingDays)) * user.Salary
	overtimeAmount := totalHours * (user.Salary / float64(workingDays)) * 1.5
	totalPay := attendanceAmount + overtimeAmount + totalReimbursement

	// 6. Return full payslip
	payslip := &entity.Payslip{
		UserID:             userID,
		AttendancePeriodID: periodID,
		BaseSalary:         user.Salary,
		WorkingDays:        workingDays,
		AttendedDays:       attendedDays,
		AttendanceAmount:   attendanceAmount,
		OvertimeHours:      totalHours,
		OvertimePay:        overtimeAmount,
		ReimbursementTotal: totalReimbursement,
		TotalPay:           totalPay,
		CreatedAt:          time.Now(), // or fetch from DB if already generated
	}

	return payslip, nil
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
			// Step 1: Insert PayrollJob
			job := entity.PayrollJob{
				PeriodID:  periodID,
				UserID:    user.ID,
				Status:    "pending",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
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
		// Check if payroll already exists
		exists, err := p.PayslipDom.IsPayrollExists(newCtx, data.PeriodID)
		if err != nil {
			return err
		}
		if exists {
			return x.NewWithCode(http.StatusBadRequest, "payroll already run for this period")
		}

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
			ID:          data.JobID,
			Status:      "completed",
			CompletedAt: pkg.TimePtr(time.Now()),
		})
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to save payslip job")
		}

		return nil
	})
}
