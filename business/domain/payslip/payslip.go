package payslip

import (
	"context"

	"github.com/zuhrulumam/go-hris/business/entity"
	"gorm.io/gorm"
)

//go:generate mockgen -source=business/domain/payslip/payslip.go -destination=mocks/domain/payslip/mock_payslip.go -package=mocks
type DomainItf interface {
	GetPayslip(ctx context.Context, filter entity.GetPayslipRequest) ([]entity.Payslip, int64, int, error)
	GetPayrollSummary(ctx context.Context, req entity.GetPayrollSummaryRequest) (*entity.GetPayrollSummaryResponse, error)

	CreatePayslip(ctx context.Context, payslips []entity.Payslip) error
	CreatePayrollJob(ctx context.Context, data entity.PayrollJob) (*entity.PayrollJob, error)
	UpdatePayslipJob(ctx context.Context, data entity.UpdatePayslipJob) error
}

type payslip struct {
	db *gorm.DB
}

type Option struct {
	DB *gorm.DB
}

func InitPayslipDomain(opt Option) DomainItf {
	p := &payslip{
		db: opt.DB,
	}

	return p
}
