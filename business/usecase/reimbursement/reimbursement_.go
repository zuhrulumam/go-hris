package reimbursement

import (
	"context"
	"net/http"

	"github.com/zuhrulumam/go-hris/business/entity"
	x "github.com/zuhrulumam/go-hris/pkg/errors"
)

func (p *reimbursement) SubmitReimbursement(ctx context.Context, data entity.SubmitReimbursementData) error {
	// Simple validation
	if data.Amount <= 0 {
		return x.NewWithCode(http.StatusBadRequest, "reimbursement amount must be greater than zero")
	}

	return p.ReimbursementDom.SubmitReimbursement(ctx, data)

}

func (p *reimbursement) GetReimbursement(ctx context.Context, filter entity.GetReimbursementFilter) ([]entity.Reimbursement, error) {
	return p.ReimbursementDom.GetReimbursements(ctx, filter)
}
