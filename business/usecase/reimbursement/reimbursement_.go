package reimbursement

import (
	"context"

	"github.com/zuhrulumam/go-hris/business/entity"
)

func (p *reimbursement) SubmitReimbursement(ctx context.Context, data entity.SubmitReimbursementData) error {

	return p.TransactionDom.RunInTx(ctx, func(newCtx context.Context) error {

		attPeriod, err := p.AttendanceDom.GetAttendancePeriods(newCtx, entity.GetAttendancePeriodFilter{
			ContainsDate: &data.Date,
			Status:       "open",
		})
		if err != nil {
			return err
		}

		data.AttendancePeriodID = attPeriod[0].ID

		err = p.ReimbursementDom.SubmitReimbursement(ctx, data)
		if err != nil {
			return err
		}

		return nil
	})

}

func (p *reimbursement) GetReimbursement(ctx context.Context, filter entity.GetReimbursementFilter) ([]entity.Reimbursement, error) {
	return p.ReimbursementDom.GetReimbursements(ctx, filter)
}
