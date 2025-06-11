package reimbursement_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zuhrulumam/go-hris/business/entity"
	uc "github.com/zuhrulumam/go-hris/business/usecase/reimbursement"
	mockAttendance "github.com/zuhrulumam/go-hris/mocks/domain/attendance"
	mockReimbursement "github.com/zuhrulumam/go-hris/mocks/domain/reimbursement"
	mockTx "github.com/zuhrulumam/go-hris/mocks/domain/transaction"
	"go.uber.org/mock/gomock"
)

func TestGetReimbursement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDom := mockReimbursement.NewMockDomainItf(ctrl)
	usecase := uc.InitReimbursementUsecase(uc.Option{
		ReimbursementDom: mockDom,
	})

	filter := entity.GetReimbursementFilter{
		UserID:             1,
		AttendancePeriodID: 2024,
	}

	expectedResult := []entity.Reimbursement{
		{ID: 1, UserID: 1, AttendancePeriodID: 2024, Amount: 100000},
	}

	tests := []struct {
		name         string
		mockSetup    func()
		expectedData []entity.Reimbursement
		expectErr    bool
	}{
		{
			name: "success get reimbursement",
			mockSetup: func() {
				mockDom.EXPECT().GetReimbursements(gomock.Any(), filter).
					Return(expectedResult, nil)
			},
			expectedData: expectedResult,
			expectErr:    false,
		},
		{
			name: "domain returns error",
			mockSetup: func() {
				mockDom.EXPECT().GetReimbursements(gomock.Any(), filter).
					Return(nil, errors.New("database error"))
			},
			expectedData: nil,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := usecase.GetReimbursement(context.Background(), filter)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, result)
			}
		})
	}
}

func TestSubmitReimbursement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTx := mockTx.NewMockDomainItf(ctrl)
	mockAttendance := mockAttendance.NewMockDomainItf(ctrl)
	mockReimb := mockReimbursement.NewMockDomainItf(ctrl)

	usecase := uc.InitReimbursementUsecase(uc.Option{
		TransactionDom:   mockTx,
		AttendanceDom:    mockAttendance,
		ReimbursementDom: mockReimb,
	})

	date := time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC)

	input := entity.SubmitReimbursementData{
		UserID: 1,
		Amount: 50000,
		Date:   date,
	}

	attPeriod := []entity.AttendancePeriod{
		{ID: 2024, Status: "open"},
	}

	tests := []struct {
		name      string
		mockSetup func()
		expectErr bool
	}{
		{
			name: "success",
			mockSetup: func() {
				mockTx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						// Simulate what happens inside RunInTx
						return fn(ctx)
					},
				)

				mockAttendance.EXPECT().
					GetAttendancePeriods(gomock.Any(), entity.GetAttendancePeriodFilter{
						ContainsDate: &date,
						Status:       "open",
					}).
					Return(attPeriod, nil)

				mockReimb.EXPECT().
					SubmitReimbursement(gomock.Any(), gomock.AssignableToTypeOf(entity.SubmitReimbursementData{})).
					Return(nil)
			},
			expectErr: false,
		},
		{
			name: "attendance period fetch error",
			mockSetup: func() {
				mockTx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					},
				)

				mockAttendance.EXPECT().
					GetAttendancePeriods(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db error"))
			},
			expectErr: true,
		},
		{
			name: "submit reimbursement error",
			mockSetup: func() {
				mockTx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					},
				)

				mockAttendance.EXPECT().
					GetAttendancePeriods(gomock.Any(), gomock.Any()).
					Return(attPeriod, nil)

				mockReimb.EXPECT().
					SubmitReimbursement(gomock.Any(), gomock.Any()).
					Return(errors.New("insert error"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := usecase.SubmitReimbursement(context.Background(), input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
