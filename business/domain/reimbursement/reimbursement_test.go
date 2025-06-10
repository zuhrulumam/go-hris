package reimbursement_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zuhrulumam/go-hris/business/domain/reimbursement"
	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
)

func TestSubmitReimbursement(t *testing.T) {

	tests := []struct {
		name        string
		input       entity.SubmitReimbursementData
		mockSetup   func(mock sqlmock.Sqlmock, input entity.SubmitReimbursementData)
		expectError bool
		errorText   string
	}{
		{
			name: "Success submit reimbursement",
			input: entity.SubmitReimbursementData{
				UserID:             1,
				AttendancePeriodID: 10,
				Amount:             100000,
				Description:        "Transport reimbursement",
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.SubmitReimbursementData) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "reimbursements"`).
					WithArgs(
						input.UserID,
						input.AttendancePeriodID,
						input.Amount,
						input.Description,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectError: false,
		},
		{
			name: "DB error on insert",
			input: entity.SubmitReimbursementData{
				UserID:             2,
				AttendancePeriodID: 11,
				Amount:             50000,
				Description:        "Meal reimbursement",
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.SubmitReimbursementData) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "reimbursements"`).
					WithArgs(
						input.UserID,
						input.AttendancePeriodID,
						input.Amount,
						input.Description,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnError(errors.New("insert failed"))

			},
			expectError: true,
			errorText:   "failed to submit reimbursement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			if tt.mockSetup != nil {
				tt.mockSetup(mock, tt.input)
			}

			r := reimbursement.InitReimbursementDomain(reimbursement.Option{
				DB: db,
			})

			tx := db.Begin()
			ctx := context.WithValue(context.Background(), pkg.TxCtxValue, tx)

			err := r.SubmitReimbursement(ctx, tt.input)

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

func TestGetReimbursements(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		filter      entity.GetReimbursementFilter
		mockSetup   func(mock sqlmock.Sqlmock, filter entity.GetReimbursementFilter)
		expectError bool
		errorText   string
	}{
		{
			name: "Success get reimbursements with user filter",
			filter: entity.GetReimbursementFilter{
				UserID: 1,
			},
			mockSetup: func(mock sqlmock.Sqlmock, filter entity.GetReimbursementFilter) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "attendance_period_id", "amount", "description", "status", "created_at",
				}).AddRow(
					1, filter.UserID, 10, 100000, "Transport", "PENDING", now,
				)

				mock.ExpectQuery(`SELECT .* FROM "reimbursements"`).
					WithArgs(filter.UserID).
					WillReturnRows(rows)
			},
			expectError: false,
		},
		{
			name: "Success get reimbursements with full filter",
			filter: entity.GetReimbursementFilter{
				UserID:             1,
				AttendancePeriodID: 2,
				Status:             "APPROVED",
				StartDate:          now.Add(-24 * time.Hour),
				EndDate:            now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, filter entity.GetReimbursementFilter) {
				mock.ExpectQuery(`SELECT .* FROM "reimbursements"`).
					WithArgs(
						filter.UserID,
						filter.AttendancePeriodID,
						filter.Status,
						filter.StartDate,
						filter.EndDate,
					).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "user_id", "attendance_period_id", "amount", "description", "status", "created_at",
					}).AddRow(
						1, filter.UserID, filter.AttendancePeriodID, 50000, "Meal", "APPROVED", now,
					))
			},
			expectError: false,
		},
		{
			name: "DB error on query",
			filter: entity.GetReimbursementFilter{
				UserID: 2,
			},
			mockSetup: func(mock sqlmock.Sqlmock, filter entity.GetReimbursementFilter) {
				mock.ExpectQuery(`SELECT .* FROM "reimbursements"`).
					WithArgs(filter.UserID).
					WillReturnError(errors.New("query failed"))
			},
			expectError: true,
			errorText:   "failed to fetch reimbursements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			if tt.mockSetup != nil {
				tt.mockSetup(mock, tt.filter)
			}

			r := reimbursement.InitReimbursementDomain(reimbursement.Option{
				DB: db,
			})

			_, err := r.GetReimbursements(context.Background(), tt.filter)

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
