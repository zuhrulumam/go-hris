package reimbursement

import (
	"context"

	"github.com/zuhrulumam/go-hris/business/entity"
	"gorm.io/gorm"
)

//go:generate mockgen -source=business/domain/reimbursement/reimbursement.go -destination=mocks/domain/reimbursement/mock_reimbursement.go -package=mocks
type DomainItf interface {
	// submit
	SubmitReimbursement(ctx context.Context, data entity.SubmitReimbursementData) error

	// get
	GetReimbursements(ctx context.Context, filter entity.GetReimbursementFilter) ([]entity.Reimbursement, error)
}

type reimbursement struct {
	db *gorm.DB
}

type Option struct {
	DB *gorm.DB
}

func InitReimbursementDomain(opt Option) DomainItf {
	p := &reimbursement{
		db: opt.DB,
	}

	return p
}
