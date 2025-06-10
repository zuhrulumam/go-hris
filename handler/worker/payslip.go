package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
	"github.com/zuhrulumam/go-hris/business/usecase/payslip"
	"github.com/zuhrulumam/go-hris/task"
)

type Handler struct {
	Payslip payslip.UsecaseItf
}

func (h *Handler) HandleCreatePayrollTask(ctx context.Context, t *asynq.Task) error {
	var payload task.CreatePayrollPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	log.Printf("⏳ Processing payroll for user %d in period %d", payload.UserID, payload.PeriodID)
	return h.Payslip.CreatePayslipForUser(ctx, payload.PeriodID, payload.UserID)
}
