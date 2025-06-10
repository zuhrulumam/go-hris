package usecase

import (
	"github.com/hibiken/asynq"
	"github.com/zuhrulumam/go-hris/business/domain"
	"github.com/zuhrulumam/go-hris/business/usecase/attendance"
	"github.com/zuhrulumam/go-hris/business/usecase/payslip"
	"github.com/zuhrulumam/go-hris/business/usecase/reimbursement"
	"github.com/zuhrulumam/go-hris/business/usecase/user"
)

type Usecase struct {
	Attendance    attendance.UsecaseItf
	Reimbursement reimbursement.UsecaseItf
	Payslip       payslip.UsecaseItf
	User          user.UsecaseItf
}

type Option struct {
	AsynqClient *asynq.Client
}

func Init(dom *domain.Domain, opt Option) *Usecase {
	u := &Usecase{
		Attendance: attendance.InitAttendanceUsecase(attendance.Option{
			AttendanceDom:  dom.Attendance,
			TransactionDom: dom.Transaction,
		}),
		Reimbursement: reimbursement.InitReimbursementUsecase(reimbursement.Option{
			ReimbursementDom: dom.Reimbursement,
			TransactionDom:   dom.Transaction,
			AttendanceDom:    dom.Attendance,
		}),
		Payslip: payslip.InitPayslipUsecase(payslip.Option{
			TransactionDom:   dom.Transaction,
			PayslipDom:       dom.Payslip,
			AttendanceDom:    dom.Attendance,
			ReimbursementDom: dom.Reimbursement,
			UserDom:          dom.User,
			AsynqClient:      opt.AsynqClient,
		}),
		User: user.InitUserUsecase(user.Option{
			UserDom:        dom.User,
			TransactionDom: dom.Transaction,
		}),
	}

	return u
}
