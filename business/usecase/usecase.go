package usecase

import (
	"github.com/zuhrulumam/go-hris/business/domain"
	"github.com/zuhrulumam/go-hris/business/usecase/attendance"
)

type Usecase struct {
	Attendance attendance.UsecaseItf
}

type Option struct {
}

func Init(dom *domain.Domain, opt Option) *Usecase {
	u := &Usecase{
		Attendance: attendance.InitAttendanceUsecase(attendance.Option{
			AttendanceDom:  dom.Attendance,
			TransactionDom: dom.Transaction,
		}),
	}

	return u
}
