package payslip

import (
	"context"
	"net/http"
	"time"

	"github.com/zuhrulumam/go-hris/business/entity"
	x "github.com/zuhrulumam/go-hris/pkg/errors"
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
		OvertimeAmount:     overtimeAmount,
		ReimbursementTotal: totalReimbursement,
		TotalPay:           totalPay,
		CreatedAt:          time.Now(), // or fetch from DB if already generated
	}

	return payslip, nil
}

func (p *payslip) CreatePayroll(ctx context.Context, periodID uint) error {
	return p.TransactionDom.RunInTx(ctx, func(newCtx context.Context) error {
		// 1. Check if payroll already exists
		exists, err := p.PayslipDom.IsPayrollExists(newCtx, periodID)
		if err != nil {
			return err
		}
		if exists {
			return x.NewWithCode(http.StatusBadRequest, "payroll already run for this period")
		}

		// 2. Get all users (can be filtered by department, role, etc. if needed)
		users, err := p.UserDom.GetUsers(newCtx, entity.GetUserFilter{})
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch users")
		}

		// 3. Get all attendance, overtime, and reimbursement data for the period
		attendances, err := p.AttendanceDom.GetAttendance(newCtx, entity.GetAttendance{
			AttendancePeriodID: periodID,
		})
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch attendance data")
		}

		overtimes, err := p.AttendanceDom.GetOvertime(newCtx, entity.GetOvertimeFilter{
			AttendancePeriodID: periodID,
		})
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch overtime data")
		}

		reimbursements, err := p.ReimbursementDom.GetReimbursements(newCtx, entity.GetReimbursementFilter{
			AttendancePeriodID: periodID,
		})
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to fetch reimbursement data")
		}

		// 4. Map data for faster access
		attendanceMap := groupAttendanceByUser(attendances)
		overtimeMap := groupOvertimeByUser(overtimes)
		reimbursementMap := groupReimbursementsByUser(reimbursements)

		// 5. Create payslips
		var payslips []entity.Payslip
		for _, user := range users {
			userAttendances := attendanceMap[user.ID]
			userOvertimes := overtimeMap[user.ID]
			userReimbursements := reimbursementMap[user.ID]

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

			attendanceAmount := (float64(attendedDays) / float64(workingDays)) * user.Salary
			overtimeAmount := overtimeHours * (user.Salary / float64(workingDays)) * 1.5
			totalPay := attendanceAmount + overtimeAmount + reimbursementTotal

			payslips = append(payslips, entity.Payslip{
				UserID:             user.ID,
				AttendancePeriodID: periodID,
				BaseSalary:         user.Salary,
				WorkingDays:        workingDays,
				AttendedDays:       attendedDays,
				AttendanceAmount:   attendanceAmount,
				OvertimeHours:      overtimeHours,
				OvertimeAmount:     overtimeAmount,
				ReimbursementTotal: reimbursementTotal,
				TotalPay:           totalPay,
				CreatedAt:          time.Now(),
			})
		}

		// 6. Save all payslips
		err = p.PayslipDom.CreatePayslip(newCtx, payslips)
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to save payslips")
		}

		// 7. Close the attendance period
		err = p.AttendanceDom.UpdateAttendancePeriod(newCtx, entity.UpdateAttendancePeriod{
			ID: periodID,
			// Status: ptr.String("closed"),
		})
		if err != nil {
			return x.WrapWithCode(err, http.StatusInternalServerError, "failed to close attendance period")
		}

		return nil
	})
}

func groupAttendanceByUser(data []entity.Attendance) map[uint][]entity.Attendance {
	result := make(map[uint][]entity.Attendance)
	for _, a := range data {
		result[a.UserID] = append(result[a.UserID], a)
	}
	return result
}

func groupOvertimeByUser(data []entity.Overtime) map[uint][]entity.Overtime {
	result := make(map[uint][]entity.Overtime)
	for _, o := range data {
		result[o.UserID] = append(result[o.UserID], o)
	}
	return result
}

func groupReimbursementsByUser(data []entity.Reimbursement) map[uint][]entity.Reimbursement {
	result := make(map[uint][]entity.Reimbursement)
	for _, r := range data {
		result[r.UserID] = append(result[r.UserID], r)
	}
	return result
}
