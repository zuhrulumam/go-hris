package usecase

import (
	"github.com/zuhrulumam/go-hris/business/domain"
	"github.com/zuhrulumam/go-hris/business/usecase/attendance"
	"github.com/zuhrulumam/go-hris/business/usecase/payslip"
	"github.com/zuhrulumam/go-hris/business/usecase/reimbursement"
)

type Usecase struct {
	Attendance    attendance.UsecaseItf
	Reimbursement reimbursement.UsecaseItf
	Payslip       payslip.UsecaseItf
}

type Option struct {
}

func Init(dom *domain.Domain, opt Option) *Usecase {
	u := &Usecase{
		Attendance: attendance.InitAttendanceUsecase(attendance.Option{
			AttendanceDom:  dom.Attendance,
			TransactionDom: dom.Transaction,
		}),
		Reimbursement: reimbursement.InitReimbursementUsecase(reimbursement.Option{
			// ReimbursementDom: dom.,
			TransactionDom: dom.Transaction,
		}),
		Payslip: payslip.InitPayslipUsecase(payslip.Option{
			TransactionDom: dom.Transaction,
		}),
	}

	return u
}
