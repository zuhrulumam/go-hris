package payslip_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zuhrulumam/go-hris/business/entity"
	uc "github.com/zuhrulumam/go-hris/business/usecase/payslip"
	mockAttendance "github.com/zuhrulumam/go-hris/mocks/domain/attendance"
	mockPayslip "github.com/zuhrulumam/go-hris/mocks/domain/payslip"
	mockReimbursement "github.com/zuhrulumam/go-hris/mocks/domain/reimbursement"
	mockTx "github.com/zuhrulumam/go-hris/mocks/domain/transaction"
	mockUser "github.com/zuhrulumam/go-hris/mocks/domain/user"
	"github.com/zuhrulumam/go-hris/pkg"
	"go.uber.org/mock/gomock"
)

func TestGetPayslip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPayslipDom := mockPayslip.NewMockDomainItf(ctrl)

	usecase := uc.InitPayslipUsecase(uc.Option{
		PayslipDom: mockPayslipDom,
	})

	filter := entity.GetPayslipRequest{
		UserID: pkg.UintPtr(1),
	}

	expectedPayslips := []entity.Payslip{
		{
			ID:     1,
			UserID: 1,
		},
	}

	tests := []struct {
		name           string
		mockReturnData []entity.Payslip
		mockTotalData  int64
		mockTotalPage  int
		mockErr        error
		expectErr      bool
		errMsgContains string
	}{
		{
			name:           "success get payslip",
			mockReturnData: expectedPayslips,
			mockTotalData:  1,
			mockTotalPage:  1,
			mockErr:        nil,
			expectErr:      false,
		},
		{
			name:           "no payslip found",
			mockReturnData: []entity.Payslip{},
			mockTotalData:  0,
			mockTotalPage:  0,
			mockErr:        nil,
			expectErr:      true,
			errMsgContains: "payslip not found",
		},
		{
			name:           "error from domain",
			mockReturnData: nil,
			mockTotalData:  0,
			mockTotalPage:  0,
			mockErr:        errors.New("db error"),
			expectErr:      true,
			errMsgContains: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPayslipDom.EXPECT().
				GetPayslip(gomock.Any(), filter).
				Return(tt.mockReturnData, tt.mockTotalData, tt.mockTotalPage, tt.mockErr)

			result, totalData, totalPage, err := usecase.GetPayslip(context.Background(), filter)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsgContains != "" {
					assert.Contains(t, err.Error(), tt.errMsgContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturnData, result)
				assert.Equal(t, tt.mockTotalData, totalData)
				assert.Equal(t, tt.mockTotalPage, totalPage)
			}
		})
	}
}

// func TestCreatePayroll(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockTransactionDom := mockTx.NewMockDomainItf(ctrl)
// 	mockUserDom := mockUser.NewMockDomainItf(ctrl)
// 	mockPayslipDom := mockPayslip.NewMockDomainItf(ctrl)
// 	mockAsynqClient := mockAsynq.NewMockClient(ctrl)

// 	// override task function
// 	originalCreateTask := task.NewCreatePayrollTask
// 	defer func() {
// 		task.NewCreatePayrollTask = originalCreateTask
// 	}()

// 	usecase := uc.InitPayslipUsecase(uc.Option{
// 		TransactionDom: mockTransactionDom,
// 		UserDom:        mockUserDom,
// 		PayslipDom:     mockPayslipDom,
// 		AsynqClient:    mockAsynqClient,
// 	})

// 	periodID := uint(10)
// 	users := []entity.User{
// 		{ID: 1},
// 		{ID: 2},
// 	}

// 	payrollJobs := []entity.PayrollJob{
// 		{ID: 1001, UserID: 1, AttendancePeriodID: periodID},
// 		{ID: 1002, UserID: 2, AttendancePeriodID: periodID},
// 	}

// 	fakeTask := asynq.NewTask("fake", []byte("payload"))

// 	task.NewCreatePayrollTask = func(periodID, userID, jobID uint) (*asynq.Task, error) {
// 		return fakeTask, nil
// 	}

// 	mockTransactionDom.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
// 		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
// 			return fn(ctx)
// 		})

// 	mockUserDom.EXPECT().
// 		GetUsers(gomock.Any(), entity.GetUserFilter{Role: string(entity.RoleEmployee)}).
// 		Return(users, nil)

// 	for i, user := range users {
// 		mockPayslipDom.EXPECT().
// 			CreatePayrollJob(gomock.Any(), gomock.AssignableToTypeOf(entity.PayrollJob{})).
// 			DoAndReturn(func(_ context.Context, job entity.PayrollJob) (entity.PayrollJob, error) {
// 				return payrollJobs[i], nil
// 			})

// 		mockAsynqClient.EXPECT().
// 			Enqueue(fakeTask).
// 			Return(&asynq.TaskInfo{}, nil)
// 	}

// 	err := usecase.CreatePayroll(context.Background(), periodID)
// 	assert.NoError(t, err)
// }

func TestGetPayrollSummary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPayslipDom := mockPayslip.NewMockDomainItf(ctrl)

	usecase := uc.InitPayslipUsecase(uc.Option{
		PayslipDom: mockPayslipDom,
	})

	filter := entity.GetPayrollSummaryRequest{
		AttendancePeriodIDs: []uint{1},
	}

	expectedSummary := &entity.GetPayrollSummaryResponse{
		Items: []entity.PayrollSummaryItem{
			{UserID: 123, TotalPay: 5000000},
		},
	}

	tests := []struct {
		name           string
		mockSetup      func()
		expectedResult *entity.GetPayrollSummaryResponse
		expectError    bool
		errorContains  string
	}{
		{
			name: "success get summary",
			mockSetup: func() {
				mockPayslipDom.EXPECT().
					GetPayrollSummary(gomock.Any(), filter).
					Return(expectedSummary, nil)
			},
			expectedResult: expectedSummary,
			expectError:    false,
		},
		{
			name: "summary not found (empty items)",
			mockSetup: func() {
				mockPayslipDom.EXPECT().
					GetPayrollSummary(gomock.Any(), filter).
					Return(&entity.GetPayrollSummaryResponse{Items: []entity.PayrollSummaryItem{}}, nil)
			},
			expectError:   true,
			errorContains: "payroll summary not found",
		},
		{
			name: "error from domain",
			mockSetup: func() {
				mockPayslipDom.EXPECT().
					GetPayrollSummary(gomock.Any(), filter).
					Return(nil, errors.New("db error"))
			},
			expectError:   true,
			errorContains: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := usecase.GetPayrollSummary(context.Background(), filter)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestCreatePayslipForUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTx := mockTx.NewMockDomainItf(ctrl)
	mockUserDom := mockUser.NewMockDomainItf(ctrl)
	mockAttendanceDom := mockAttendance.NewMockDomainItf(ctrl)
	mockReimbursementDom := mockReimbursement.NewMockDomainItf(ctrl)
	mockPayslipDom := mockPayslip.NewMockDomainItf(ctrl)

	usecase := uc.InitPayslipUsecase(uc.Option{
		TransactionDom:   mockTx,
		UserDom:          mockUserDom,
		AttendanceDom:    mockAttendanceDom,
		ReimbursementDom: mockReimbursementDom,
		PayslipDom:       mockPayslipDom,
	})

	userID := uint(1)
	periodID := uint(100)
	jobID := uint(500)

	tests := []struct {
		name         string
		mockSetup    func()
		expectErr    bool
		errorMessage string
	}{
		{
			name: "success create payslip",
			mockSetup: func() {
				mockTx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				mockUserDom.EXPECT().GetUsers(gomock.Any(), entity.GetUserFilter{ID: userID}).
					Return([]entity.User{{ID: userID, Salary: 2200000}}, nil)

				mockAttendanceDom.EXPECT().GetAttendance(gomock.Any(), entity.GetAttendance{
					UserID: userID, AttendancePeriodID: periodID,
				}).Return([]entity.Attendance{{ID: 1}}, nil)

				mockAttendanceDom.EXPECT().GetOvertime(gomock.Any(), entity.GetOvertimeFilter{
					UserID: userID, AttendancePeriodID: periodID,
				}).Return([]entity.Overtime{{Hours: 2}}, nil)

				mockReimbursementDom.EXPECT().GetReimbursements(gomock.Any(), entity.GetReimbursementFilter{
					UserID: userID, AttendancePeriodID: periodID,
				}).Return([]entity.Reimbursement{{Amount: 100000}}, nil)

				mockPayslipDom.EXPECT().CreatePayslip(gomock.Any(), gomock.Any()).Return(nil)
				mockPayslipDom.EXPECT().UpdatePayslipJob(gomock.Any(), entity.UpdatePayslipJob{
					ID: jobID, Status: "completed",
				}).Return(nil)
			},
			expectErr: false,
		},
		{
			name: "user not found",
			mockSetup: func() {
				mockTx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				mockUserDom.EXPECT().GetUsers(gomock.Any(), entity.GetUserFilter{ID: userID}).
					Return([]entity.User{}, nil)
			},
			expectErr:    true,
			errorMessage: "user not found",
		},
		{
			name: "attendance fetch error",
			mockSetup: func() {
				mockTx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				mockUserDom.EXPECT().GetUsers(gomock.Any(), entity.GetUserFilter{ID: userID}).
					Return([]entity.User{{ID: userID, Salary: 2000000}}, nil)

				mockAttendanceDom.EXPECT().GetAttendance(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("attendance error"))
			},
			expectErr:    true,
			errorMessage: "attendance error",
		},
		{
			name: "payslip creation error",
			mockSetup: func() {
				mockTx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				mockUserDom.EXPECT().GetUsers(gomock.Any(), entity.GetUserFilter{ID: userID}).
					Return([]entity.User{{ID: userID, Salary: 2000000}}, nil)

				mockAttendanceDom.EXPECT().GetAttendance(gomock.Any(), gomock.Any()).
					Return([]entity.Attendance{{ID: 1}}, nil)

				mockAttendanceDom.EXPECT().GetOvertime(gomock.Any(), gomock.Any()).
					Return([]entity.Overtime{{Hours: 1}}, nil)

				mockReimbursementDom.EXPECT().GetReimbursements(gomock.Any(), gomock.Any()).
					Return([]entity.Reimbursement{}, nil)

				mockPayslipDom.EXPECT().CreatePayslip(gomock.Any(), gomock.Any()).
					Return(errors.New("db error"))
			},
			expectErr:    true,
			errorMessage: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := usecase.CreatePayslipForUser(context.Background(), entity.CreatePayslipForUserData{
				UserID:   userID,
				PeriodID: periodID,
				JobID:    jobID,
			})
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
