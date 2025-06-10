package payslip_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zuhrulumam/go-hris/business/domain/payslip"
	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
)

func TestGetPayslip(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		filter        entity.GetPayslipRequest
		mockQuery     string
		mockCount     *sqlmock.Rows
		mockData      *sqlmock.Rows
		expectError   bool
		expectedData  []entity.Payslip
		expectedTotal int64
		expectedPages int
	}{
		{
			name: "Success with pagination and filters",
			filter: entity.GetPayslipRequest{
				UserID:             pkg.UintPtr(1),
				AttendancePeriodID: pkg.UintPtr(2),
				Status:             pkg.StringPtr("approved"),
				Limit:              5,
				Page:               1,
			},
			mockQuery: `SELECT .* FROM "payslips"`,
			mockCount: sqlmock.NewRows([]string{"count"}).AddRow(1),
			mockData: sqlmock.NewRows([]string{
				"id", "user_id", "attendance_period_id", "status", "total_amount", "created_at",
			}).AddRow(
				1, 1, 2, "approved", 1000000, now,
			),
			expectError: false,
			expectedData: []entity.Payslip{
				{
					ID:                 1,
					UserID:             1,
					AttendancePeriodID: 2,
					CreatedAt:          now,
				},
			},
			expectedTotal: 1,
			expectedPages: 1,
		},
		{
			name: "Empty result",
			filter: entity.GetPayslipRequest{
				UserID: pkg.UintPtr(99),
				Limit:  10,
				Page:   1,
			},
			mockQuery:     `SELECT .* FROM "payslips"`,
			mockCount:     sqlmock.NewRows([]string{"count"}).AddRow(0),
			mockData:      sqlmock.NewRows([]string{"id", "user_id", "attendance_period_id", "status", "total_amount", "created_at"}),
			expectError:   false,
			expectedData:  []entity.Payslip{},
			expectedTotal: 0,
			expectedPages: 0,
		},
		{
			name: "DB count error",
			filter: entity.GetPayslipRequest{
				Status: pkg.StringPtr("failed"),
			},
			mockQuery:   `SELECT .* FROM "payslips"`,
			mockCount:   nil,
			mockData:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			// Expect count query
			countQuery := mock.ExpectQuery(`SELECT count\(\*\) FROM "payslips"`)
			if tt.mockCount != nil {
				countQuery.WillReturnRows(tt.mockCount)
			} else {
				countQuery.WillReturnError(errors.New("count error"))
			}

			// Expect data query if count succeeded
			if tt.mockCount != nil && tt.mockData != nil {
				dataQuery := mock.ExpectQuery(tt.mockQuery)
				dataQuery.WillReturnRows(tt.mockData)
			}

			p := payslip.InitPayslipDomain(payslip.Option{DB: db})
			result, total, pages, err := p.GetPayslip(context.Background(), tt.filter)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, result)
				assert.Equal(t, tt.expectedTotal, total)
				assert.Equal(t, tt.expectedPages, pages)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetPayrollSummary(t *testing.T) {
	tests := []struct {
		name         string
		request      entity.GetPayrollSummaryRequest
		mockQuery    string
		mockRows     *sqlmock.Rows
		expectError  bool
		expectedData *entity.GetPayrollSummaryResponse
	}{
		{
			name: "Success - summary with multiple users",
			request: entity.GetPayrollSummaryRequest{
				AttendancePeriodIDs: []uint{1, 2},
			},
			mockQuery: `SELECT payslips\.user_id, users\.username, SUM\(payslips\.total_pay\) AS total_pay FROM "payslips"`,
			mockRows: sqlmock.NewRows([]string{"user_id", "username", "total_pay"}).
				AddRow(1, "user1", 500000).
				AddRow(2, "user2", 750000),
			expectError: false,
			expectedData: &entity.GetPayrollSummaryResponse{
				Items: []entity.PayrollSummaryItem{
					{UserID: 1, Username: "user1", TotalPay: 500000},
					{UserID: 2, Username: "user2", TotalPay: 750000},
				},
				GrandTotal: 1250000,
			},
		},
		{
			name: "Error - no attendance period IDs",
			request: entity.GetPayrollSummaryRequest{
				AttendancePeriodIDs: []uint{},
			},
			expectError:  true,
			expectedData: nil,
		},
		{
			name: "Error - DB query fails",
			request: entity.GetPayrollSummaryRequest{
				AttendancePeriodIDs: []uint{1},
			},
			mockQuery:   `SELECT payslips\.user_id, users\.username, SUM\(payslips\.total_pay\) AS total_pay FROM "payslips"`,
			mockRows:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			if len(tt.request.AttendancePeriodIDs) > 0 {
				query := mock.ExpectQuery(tt.mockQuery) // using AnyArg to allow slice
				if tt.mockRows != nil {
					query.WillReturnRows(tt.mockRows)
				} else {
					query.WillReturnError(errors.New("query failed"))
				}
			}

			p := payslip.InitPayslipDomain(payslip.Option{DB: db})
			result, err := p.GetPayrollSummary(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreatePayslip(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		input       []entity.Payslip
		mockSetup   func(mock sqlmock.Sqlmock, input []entity.Payslip)
		expectError bool
		errorText   string
	}{
		{
			name: "Success create multiple payslips",
			input: []entity.Payslip{
				{UserID: 1, AttendancePeriodID: 10, TotalPay: 1000000, CreatedAt: now},
				{UserID: 2, AttendancePeriodID: 10, TotalPay: 1200000, CreatedAt: now},
			},
			mockSetup: func(mock sqlmock.Sqlmock, input []entity.Payslip) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "payslips"`).
					WithArgs(
						input[0].UserID, input[0].AttendancePeriodID, input[0].BaseSalary,
						sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						input[0].TotalPay,
						sqlmock.AnyArg(),
						input[1].UserID, input[1].AttendancePeriodID, input[1].BaseSalary,
						sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						input[1].TotalPay,
						sqlmock.AnyArg(),
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
			},
			expectError: false,
		},
		{
			name:        "Error - no payslip data provided",
			input:       []entity.Payslip{},
			mockSetup:   nil,
			expectError: true,
			errorText:   "no payslip data provided",
		},
		{
			name: "Error - DB insert fails",
			input: []entity.Payslip{
				{UserID: 3, AttendancePeriodID: 11, TotalPay: 1100000, CreatedAt: now},
			},
			mockSetup: func(mock sqlmock.Sqlmock, input []entity.Payslip) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "payslips"`).
					WithArgs(
						input[0].UserID, input[0].AttendancePeriodID, input[0].BaseSalary,
						sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						input[0].TotalPay,
						sqlmock.AnyArg(),
					).
					WillReturnError(errors.New("insert error"))
				// mock.ExpectRollback()
			},
			expectError: true,
			errorText:   "failed to create payslips",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			if tt.mockSetup != nil {
				tt.mockSetup(mock, tt.input)
			}

			p := payslip.InitPayslipDomain(payslip.Option{DB: db})
			tx := db.Begin()
			ctx := context.WithValue(context.Background(), pkg.TxCtxValue, tx)

			err := p.CreatePayslip(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreatePayrollJob(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		input       entity.PayrollJob
		mockSetup   func(mock sqlmock.Sqlmock, input entity.PayrollJob)
		expectError bool
		errorText   string
	}{
		{
			name: "Success create payroll job",
			input: entity.PayrollJob{
				AttendancePeriodID: 1,
				Status:             "queued",
				CreatedAt:          now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.PayrollJob) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "payroll_jobs"`).
					WithArgs(
						input.AttendancePeriodID,
						sqlmock.AnyArg(),
						input.Status,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						input.CreatedAt,
						sqlmock.AnyArg(),
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectError: false,
		},
		{
			name: "Error - DB insert fails",
			input: entity.PayrollJob{
				AttendancePeriodID: 2,
				Status:             "queued",
				CreatedAt:          now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.PayrollJob) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "payroll_jobs"`).
					WithArgs(
						input.AttendancePeriodID,
						sqlmock.AnyArg(),
						input.Status,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						input.CreatedAt,
						sqlmock.AnyArg(),
					).
					WillReturnError(errors.New("insert error"))
			},
			expectError: true,
			errorText:   "failed to create payroll job",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			if tt.mockSetup != nil {
				tt.mockSetup(mock, tt.input)
			}

			p := payslip.InitPayslipDomain(payslip.Option{DB: db})
			tx := db.Begin()
			ctx := context.WithValue(context.Background(), pkg.TxCtxValue, tx)

			result, err := p.CreatePayrollJob(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.AttendancePeriodID, result.AttendancePeriodID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdatePayslipJob(t *testing.T) {
	now := time.Now()
	failReason := "some failure"

	tests := []struct {
		name        string
		input       entity.UpdatePayslipJob
		mockSetup   func(mock sqlmock.Sqlmock, input entity.UpdatePayslipJob)
		expectError bool
		errorText   string
	}{
		{
			name: "Success update status and started_at",
			input: entity.UpdatePayslipJob{
				ID:        1,
				Status:    "processing",
				StartedAt: &now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.UpdatePayslipJob) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "payroll_jobs"`).
					WithArgs(input.StartedAt, input.Status, sqlmock.AnyArg(), input.ID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name: "Success update failed reason only",
			input: entity.UpdatePayslipJob{
				ID:           2,
				FailedReason: &failReason,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.UpdatePayslipJob) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "payroll_jobs"`).
					WithArgs(input.FailedReason, sqlmock.AnyArg(), input.ID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name: "Missing payslip job ID",
			input: entity.UpdatePayslipJob{
				Status: "done",
			},
			mockSetup:   nil,
			expectError: true,
			errorText:   "payslip job ID is required",
		},
		{
			name: "No update fields provided",
			input: entity.UpdatePayslipJob{
				ID: 3,
			},
			mockSetup:   nil,
			expectError: true,
			errorText:   "no updates provided",
		},
		{
			name: "No rows affected (conflict)",
			input: entity.UpdatePayslipJob{
				ID:        4,
				Status:    "done",
				StartedAt: &now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.UpdatePayslipJob) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "payroll_jobs"`).
					WithArgs(input.StartedAt, input.Status, sqlmock.AnyArg(), input.ID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // No rows affected
			},
			expectError: true,
			errorText:   "payslip job was updated by someone else",
		},
		{
			name: "DB error on update",
			input: entity.UpdatePayslipJob{
				ID:        5,
				Status:    "done",
				StartedAt: &now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.UpdatePayslipJob) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "payroll_jobs"`).
					WithArgs(input.StartedAt, input.Status, sqlmock.AnyArg(), input.ID).
					WillReturnError(errors.New("db update failed"))
			},
			expectError: true,
			errorText:   "failed to update payslip job",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			if tt.mockSetup != nil {
				tt.mockSetup(mock, tt.input)
			}

			p := payslip.InitPayslipDomain(payslip.Option{DB: db})

			tx := db.Begin()
			ctx := context.WithValue(context.Background(), pkg.TxCtxValue, tx)

			err := p.UpdatePayslipJob(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
