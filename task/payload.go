package task

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const TypeCreatePayroll = "payroll:create"

type CreatePayrollPayload struct {
	PeriodID uint
	UserID   uint
}

func NewCreatePayrollTask(periodID, userID uint) (*asynq.Task, error) {
	payload, err := json.Marshal(CreatePayrollPayload{
		PeriodID: periodID,
		UserID:   userID,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeCreatePayroll, payload), nil
}
