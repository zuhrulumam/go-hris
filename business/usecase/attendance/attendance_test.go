package attendance_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zuhrulumam/go-hris/business/entity"
	uc "github.com/zuhrulumam/go-hris/business/usecase/attendance"
	mockAttendance "github.com/zuhrulumam/go-hris/mocks/domain/attendance"
	mockTx "github.com/zuhrulumam/go-hris/mocks/domain/transaction"
	"github.com/zuhrulumam/go-hris/pkg"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestCheckIn(t *testing.T) {
	tests := []struct {
		name        string
		input       entity.CheckIn
		setupMocks  func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf)
		expectErr   bool
		errorString string
	}{
		{
			name: "success check-in",
			input: entity.CheckIn{
				UserID: 1,
				Date:   time.Date(2025, 6, 10, 9, 0, 0, 0, time.UTC), // Selasa
			},
			setupMocks: func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf) {
				tx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						a.EXPECT().GetAttendancePeriods(gomock.Any(), gomock.Any()).
							Return([]entity.AttendancePeriod{{ID: 10}}, nil)

						a.EXPECT().CreateAttendance(gomock.Any(), entity.CreateAttendance{
							UserID:             1,
							AttendancePeriodID: 10,
							CheckInAt:          time.Date(2025, 6, 10, 9, 0, 0, 0, time.UTC),
						}).Return(nil)

						return fn(ctx)
					})
			},
			expectErr: false,
		},
		{
			name: "check-in on weekend",
			input: entity.CheckIn{
				UserID: 1,
				Date:   time.Date(2025, 6, 8, 9, 0, 0, 0, time.UTC), // Minggu
			},
			setupMocks:  func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf) {}, // no call
			expectErr:   true,
			errorString: "cannot check in on weekends",
		},
		{
			name: "failed to get attendance period",
			input: entity.CheckIn{
				UserID: 1,
				Date:   time.Date(2025, 6, 10, 9, 0, 0, 0, time.UTC),
			},
			setupMocks: func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf) {
				tx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						a.EXPECT().GetAttendancePeriods(gomock.Any(), gomock.Any()).
							Return(nil, errors.New("get period failed"))
						return fn(ctx)
					})
			},
			expectErr:   true,
			errorString: "get period failed",
		},
		{
			name: "failed to create attendance",
			input: entity.CheckIn{
				UserID: 1,
				Date:   time.Date(2025, 6, 10, 9, 0, 0, 0, time.UTC),
			},
			setupMocks: func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf) {
				tx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						a.EXPECT().GetAttendancePeriods(gomock.Any(), gomock.Any()).
							Return([]entity.AttendancePeriod{{ID: 11}}, nil)

						a.EXPECT().CreateAttendance(gomock.Any(), gomock.Any()).
							Return(errors.New("insert failed"))

						return fn(ctx)
					})
			},
			expectErr:   true,
			errorString: "insert failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTx := mockTx.NewMockDomainItf(ctrl)
			mockAtt := mockAttendance.NewMockDomainItf(ctrl)

			tt.setupMocks(*mockAtt, *mockTx)

			usecase := uc.InitAttendanceUsecase(uc.Option{
				AttendanceDom:  mockAtt,
				TransactionDom: mockTx,
			})

			err := usecase.CheckIn(context.Background(), tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorString != "" {
					assert.Contains(t, err.Error(), tt.errorString)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckOut(t *testing.T) {
	tests := []struct {
		name        string
		input       entity.CheckOut
		setupMocks  func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf)
		expectErr   bool
		errorString string
	}{
		{
			name: "success checkout",
			input: entity.CheckOut{
				UserID: 1,
				Date:   time.Date(2025, 6, 10, 17, 0, 0, 0, time.UTC),
			},
			setupMocks: func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf) {
				tx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						a.EXPECT().GetAttendance(gomock.Any(), entity.GetAttendance{
							UserID: 1,
							Date:   time.Date(2025, 6, 10, 17, 0, 0, 0, time.UTC),
						}).Return([]entity.Attendance{{
							ID:      100,
							Version: 1,
						}}, nil)

						a.EXPECT().UpdateAttendance(gomock.Any(), entity.UpdateAttendance{
							AttendanceID: 100,
							CheckOutAt:   pkg.TimePtr(time.Date(2025, 6, 10, 17, 0, 0, 0, time.UTC)),
							Version:      1,
						}).Return(nil)

						return fn(ctx)
					})
			},
			expectErr: false,
		},
		{
			name: "get attendance failed",
			input: entity.CheckOut{
				UserID: 1,
				Date:   time.Date(2025, 6, 10, 17, 0, 0, 0, time.UTC),
			},
			setupMocks: func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf) {
				tx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						a.EXPECT().GetAttendance(gomock.Any(), gomock.Any()).
							Return(nil, errors.New("get attendance failed"))
						return fn(ctx)
					})
			},
			expectErr:   true,
			errorString: "get attendance failed",
		},
		{
			name: "update attendance failed",
			input: entity.CheckOut{
				UserID: 1,
				Date:   time.Date(2025, 6, 10, 17, 0, 0, 0, time.UTC),
			},
			setupMocks: func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf) {
				tx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						a.EXPECT().GetAttendance(gomock.Any(), gomock.Any()).
							Return([]entity.Attendance{{
								ID:      100,
								Version: 1,
							}}, nil)

						a.EXPECT().UpdateAttendance(gomock.Any(), gomock.Any()).
							Return(errors.New("update failed"))
						return fn(ctx)
					})
			},
			expectErr:   true,
			errorString: "update failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTx := mockTx.NewMockDomainItf(ctrl)
			mockAtt := mockAttendance.NewMockDomainItf(ctrl)

			tt.setupMocks(*mockAtt, *mockTx)

			usecase := uc.InitAttendanceUsecase(uc.Option{
				AttendanceDom:  mockAtt,
				TransactionDom: mockTx,
			})

			err := usecase.CheckOut(context.Background(), tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorString != "" {
					assert.Contains(t, err.Error(), tt.errorString)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateOvertime(t *testing.T) {
	tests := []struct {
		name        string
		input       entity.CreateOvertimeData
		setupMocks  func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf)
		expectErr   bool
		errorString string
	}{
		{
			name: "success create overtime",
			input: entity.CreateOvertimeData{
				UserID: 1,
				Date:   time.Date(2025, 6, 10, 18, 0, 0, 0, time.UTC),
				Hours:  2,
			},
			setupMocks: func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf) {
				tx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						a.EXPECT().GetAttendance(gomock.Any(), gomock.Any()).
							Return([]entity.Attendance{{
								ID:                 1,
								AttendancePeriodID: 10,
								CheckedOutAt:       pkg.TimePtr(time.Date(2025, 6, 10, 17, 0, 0, 0, time.UTC)),
							}}, nil)

						a.EXPECT().GetOvertime(gomock.Any(), gomock.Any()).
							Return(nil, gorm.ErrRecordNotFound)

						a.EXPECT().CreateOvertime(gomock.Any(), gomock.Any()).
							Return(nil)

						return fn(ctx)
					})
			},
			expectErr: false,
		},
		{
			name: "hours exceed max limit",
			input: entity.CreateOvertimeData{
				UserID: 1,
				Date:   time.Now(),
				Hours:  4,
			},
			setupMocks:  func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf) {},
			expectErr:   true,
			errorString: "overtime cannot be more than 3 hours per day",
		},
		{
			name: "attendance not found",
			input: entity.CreateOvertimeData{
				UserID: 2,
				Date:   time.Date(2025, 6, 10, 18, 0, 0, 0, time.UTC),
				Hours:  2,
			},
			setupMocks: func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf) {
				tx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						a.EXPECT().GetAttendance(gomock.Any(), gomock.Any()).
							Return(nil, errors.New("attendance not found for the date"))
						return fn(ctx)
					})
			},
			expectErr:   true,
			errorString: "attendance not found for the date",
		},
		{
			name: "not checked out yet",
			input: entity.CreateOvertimeData{
				UserID: 3,
				Date:   time.Date(2025, 6, 10, 18, 0, 0, 0, time.UTC),
				Hours:  2,
			},
			setupMocks: func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf) {
				tx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						a.EXPECT().GetAttendance(gomock.Any(), gomock.Any()).
							Return([]entity.Attendance{{
								ID:           1,
								CheckedOutAt: nil,
							}}, nil)
						return fn(ctx)
					})
			},
			expectErr:   true,
			errorString: "must check out before submitting overtime",
		},
		{
			name: "overtime already submitted",
			input: entity.CreateOvertimeData{
				UserID: 4,
				Date:   time.Date(2025, 6, 10, 18, 0, 0, 0, time.UTC),
				Hours:  2,
			},
			setupMocks: func(a mockAttendance.MockDomainItf, tx mockTx.MockDomainItf) {
				tx.EXPECT().RunInTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						a.EXPECT().GetAttendance(gomock.Any(), gomock.Any()).
							Return([]entity.Attendance{{
								ID:                 1,
								AttendancePeriodID: 10,
								CheckedOutAt:       pkg.TimePtr(time.Date(2025, 6, 10, 17, 0, 0, 0, time.UTC)),
							}}, nil)

						a.EXPECT().GetOvertime(gomock.Any(), gomock.Any()).
							Return([]entity.Overtime{{ID: 123}}, nil)

						return fn(ctx)
					})
			},
			expectErr:   true,
			errorString: "overtime already submitted for this date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTx := mockTx.NewMockDomainItf(ctrl)
			mockAtt := mockAttendance.NewMockDomainItf(ctrl)

			tt.setupMocks(*mockAtt, *mockTx)

			usecase := uc.InitAttendanceUsecase(uc.Option{
				AttendanceDom:  mockAtt,
				TransactionDom: mockTx,
			})

			err := usecase.CreateOvertime(context.Background(), tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorString != "" {
					assert.Contains(t, err.Error(), tt.errorString)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetOvertime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAtt := mockAttendance.NewMockDomainItf(ctrl)

	usecase := uc.InitAttendanceUsecase(uc.Option{
		AttendanceDom: mockAtt,
	})

	filter := entity.GetOvertimeFilter{
		UserID: 1,
		Date:   time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC),
	}

	t.Run("success get overtime", func(t *testing.T) {
		expected := []entity.Overtime{
			{ID: 1, UserID: 1, Date: filter.Date, Hours: 2},
		}

		mockAtt.EXPECT().GetOvertime(gomock.Any(), filter).Return(expected, nil)

		result, err := usecase.GetOvertime(context.Background(), filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("failed to get overtime", func(t *testing.T) {
		mockAtt.EXPECT().GetOvertime(gomock.Any(), filter).
			Return(nil, errors.New("database error"))

		result, err := usecase.GetOvertime(context.Background(), filter)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database error")
	})
}

func TestCreateAttendancePeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAtt := mockAttendance.NewMockDomainItf(ctrl)

	usecase := uc.InitAttendanceUsecase(uc.Option{
		AttendanceDom: mockAtt,
	})

	req := entity.CreateAttendancePeriodRequest{
		StartDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC),
	}

	t.Run("success create attendance period", func(t *testing.T) {
		mockAtt.EXPECT().
			CreateAttendancePeriod(gomock.Any(), gomock.AssignableToTypeOf(entity.AttendancePeriod{})).
			DoAndReturn(func(_ context.Context, p entity.AttendancePeriod) error {
				assert.Equal(t, req.StartDate, p.StartDate)
				assert.Equal(t, req.EndDate, p.EndDate)
				assert.Equal(t, "open", p.Status)
				return nil
			})

		err := usecase.CreateAttendancePeriod(context.Background(), req)
		assert.NoError(t, err)
	})

	t.Run("failed create attendance period", func(t *testing.T) {
		mockAtt.EXPECT().
			CreateAttendancePeriod(gomock.Any(), gomock.Any()).
			Return(errors.New("insert failed"))

		err := usecase.CreateAttendancePeriod(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create attendance period")
	})
}
