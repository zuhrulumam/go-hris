package payslip

import (
	"context"

	"github.com/zuhrulumam/go-hris/business/entity"
	"gorm.io/gorm"
)

//go:generate mockgen -source=business/domain/payslip/payslip.go -destination=mocks/mock_payslip.go -package=mocks
type DomainItf interface {
	GetPayslip(ctx context.Context, req entity.GetPayslipRequest) (*entity.Payslip, error)
	GetPayrollSummary(ctx context.Context, req entity.GetPayrollSummaryRequest) (*entity.GetPayrollSummaryResponse, error)
	IsPayrollExists(ctx context.Context, periodID uint) (bool, error)

	CreatePayslip(ctx context.Context, payslips []entity.Payslip) error
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
