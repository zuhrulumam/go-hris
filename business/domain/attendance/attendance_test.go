package attendance_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zuhrulumam/go-hris/business/domain/attendance"
	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
)

func TestCreateAttendance(t *testing.T) {
	tests := []struct {
		name         string
		input        entity.CreateAttendance
		existingData bool // Simulate attendance already exists
		expectError  bool
		errorType    error // Optional, to check for conflict error
	}{
		{
			name: "Success create attendance",
			input: entity.CreateAttendance{
				UserID:             1,
				AttendancePeriodID: 100,
			},
			existingData: false,
			expectError:  false,
		},
		{
			name: "DB insert error",
			input: entity.CreateAttendance{
				UserID:             3,
				AttendancePeriodID: 300,
			},
			existingData: false,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			now := time.Now()
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

			// Expect SELECT to check if already attended
			if tt.existingData {
				rows := sqlmock.NewRows([]string{"id", "user_id", "date"}).
					AddRow(1, tt.input.UserID, today)
				mock.ExpectQuery(`SELECT .* FROM "attendances"`).
					WithArgs(tt.input.UserID, today).
					WillReturnRows(rows)
			} else {
				if tt.expectError {
					// Simulate insert error
					mock.ExpectBegin()
					mock.ExpectExec(`INSERT INTO "attendances"`).
						WillReturnError(errors.New("insert error"))
					mock.ExpectRollback()
				} else {
					mock.ExpectBegin()
					mock.ExpectQuery(`INSERT INTO "attendances"`).
						WithArgs(tt.input.UserID, today, tt.input.AttendancePeriodID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
						WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
					mock.ExpectCommit()
				}
			}

			a := attendance.InitAttendanceDomain(attendance.Option{
				DB: db,
			})

			tx := db.Begin()
			ctx := context.WithValue(context.Background(), pkg.TxCtxValue, tx)

			err := a.CreateAttendance(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.EqualError(t, err, tt.errorType.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateAttendance(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		input       entity.UpdateAttendance
		mockSetup   func(mock sqlmock.Sqlmock, input entity.UpdateAttendance)
		expectError bool
		errorText   string
	}{
		{
			name: "Success update attendance",
			input: entity.UpdateAttendance{
				AttendanceID: 1,
				Version:      1,
				CheckOutAt:   &now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.UpdateAttendance) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "attendances"`).
					WithArgs(sqlmock.AnyArg(), input.Version+1, sqlmock.AnyArg(), input.AttendanceID, input.Version).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "Missing attendance ID",
			input: entity.UpdateAttendance{
				Version:   1,
				CheckInAt: &now,
			},
			mockSetup:   nil,
			expectError: true,
			errorText:   "attendance ID is required",
		},
		{
			name: "No update fields provided",
			input: entity.UpdateAttendance{
				AttendanceID: 2,
				Version:      1,
			},
			mockSetup:   nil,
			expectError: true,
			errorText:   "no updates provided",
		},
		{
			name: "Version conflict",
			input: entity.UpdateAttendance{
				AttendanceID: 3,
				Version:      1,
				CheckOutAt:   &now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.UpdateAttendance) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "attendances"`).
					WithArgs(sqlmock.AnyArg(), input.Version+1, sqlmock.AnyArg(), input.AttendanceID, input.Version).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 row affected
				mock.ExpectCommit()
			},
			expectError: true,
			errorText:   "attendance was updated by someone else, please retry",
		},
		{
			name: "DB error on update",
			input: entity.UpdateAttendance{
				AttendanceID: 4,
				Version:      1,
				CheckOutAt:   &now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.UpdateAttendance) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "attendances"`).
					WithArgs(sqlmock.AnyArg(), input.AttendanceID, input.Version).
					WillReturnError(errors.New("update failed"))
				mock.ExpectRollback()
			},
			expectError: true,
			errorText:   "failed to update attendance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			if tt.mockSetup != nil {
				tt.mockSetup(mock, tt.input)
			}

			a := attendance.InitAttendanceDomain(attendance.Option{
				DB: db,
			})

			tx := db.Begin()
			ctx := context.WithValue(context.Background(), pkg.TxCtxValue, tx)

			err := a.UpdateAttendance(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateAttendancePeriod(t *testing.T) {
	now := time.Now()
	status := "closed"

	tests := []struct {
		name        string
		input       entity.UpdateAttendancePeriod
		mockSetup   func(mock sqlmock.Sqlmock, input entity.UpdateAttendancePeriod)
		expectError bool
		errorText   string
	}{
		{
			name: "Success update attendance period",
			input: entity.UpdateAttendancePeriod{
				ID:        1,
				Status:    &status,
				StartDate: &now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.UpdateAttendancePeriod) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "attendance_periods"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "Missing attendance period ID",
			input: entity.UpdateAttendancePeriod{
				Status: &status,
			},
			mockSetup:   nil,
			expectError: true,
			errorText:   "attendance period ID is required",
		},
		{
			name: "No update fields provided",
			input: entity.UpdateAttendancePeriod{
				ID: 2,
			},
			mockSetup:   nil,
			expectError: true,
			errorText:   "no updates provided",
		},
		{
			name: "DB error on update",
			input: entity.UpdateAttendancePeriod{
				ID:      3,
				EndDate: &now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.UpdateAttendancePeriod) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "attendance_periods"`).
					WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			expectError: true,
			errorText:   "failed to update attendance period",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			if tt.mockSetup != nil {
				tt.mockSetup(mock, tt.input)
			}

			a := attendance.InitAttendanceDomain(attendance.Option{DB: db})
			tx := db.Begin()
			ctx := context.WithValue(context.Background(), pkg.TxCtxValue, tx)

			err := a.UpdateAttendancePeriod(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateOvertime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		input       entity.CreateOvertimeData
		mockSetup   func(mock sqlmock.Sqlmock, input entity.CreateOvertimeData)
		expectError bool
		errorText   string
	}{
		{
			name: "Success create overtime",
			input: entity.CreateOvertimeData{
				UserID:             1,
				Date:               now,
				Hours:              2,
				Description:        "Lembur proyek A",
				AttendancePeriodID: 10,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.CreateOvertimeData) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "overtimes"`).
					WithArgs(
						input.UserID,
						input.Date,
						input.Hours,
						input.AttendancePeriodID,
						input.Description,
						sqlmock.AnyArg(), // CreatedAt (now)
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "DB error on insert",
			input: entity.CreateOvertimeData{
				UserID:             2,
				Date:               now,
				Hours:              3,
				Description:        "Lembur B",
				AttendancePeriodID: 20,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.CreateOvertimeData) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "overtimes"`).
					WithArgs(
						input.UserID,
						input.Date,
						input.Hours,
						input.Description,
						input.AttendancePeriodID,
						sqlmock.AnyArg(),
					).
					WillReturnError(errors.New("insert failed"))
				mock.ExpectRollback()
			},
			expectError: true,
			errorText:   "failed to submit overtime",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			if tt.mockSetup != nil {
				tt.mockSetup(mock, tt.input)
			}

			a := attendance.InitAttendanceDomain(attendance.Option{DB: db})
			tx := db.Begin()
			ctx := context.WithValue(context.Background(), pkg.TxCtxValue, tx)

			err := a.CreateOvertime(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateAttendancePeriod(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		input       entity.AttendancePeriod
		mockSetup   func(mock sqlmock.Sqlmock, input entity.AttendancePeriod)
		expectError bool
		errorText   string
	}{
		{
			name: "Success create attendance period",
			input: entity.AttendancePeriod{
				StartDate: now,
				EndDate:   now.AddDate(0, 0, 30),
				Status:    "open",
				CreatedAt: now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.AttendancePeriod) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "attendance_periods"`).
					WithArgs(
						input.StartDate,
						input.EndDate,
						input.Status,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "DB error on insert",
			input: entity.AttendancePeriod{
				StartDate: now,
				EndDate:   now.AddDate(0, 0, 28),
				Status:    "open",
				CreatedAt: now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, input entity.AttendancePeriod) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "attendance_periods"`).
					WithArgs(
						input.StartDate,
						input.EndDate,
						input.Status,
						sqlmock.AnyArg(),
						input.CreatedAt,
						sqlmock.AnyArg(),
					).
					WillReturnError(errors.New("insert error"))
				mock.ExpectRollback()
			},
			expectError: true,
			errorText:   "failed to create attendance period",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			if tt.mockSetup != nil {
				tt.mockSetup(mock, tt.input)
			}

			a := attendance.InitAttendanceDomain(attendance.Option{DB: db})
			tx := db.Begin()
			ctx := context.WithValue(context.Background(), pkg.TxCtxValue, tx)

			err := a.CreateAttendancePeriod(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetOvertime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		filter       entity.GetOvertimeFilter
		mockQuery    string
		mockRows     *sqlmock.Rows
		expectError  bool
		expectedData []entity.Overtime
	}{
		{
			name: "Success with UserID and AttendancePeriodID",
			filter: entity.GetOvertimeFilter{
				UserID:             1,
				AttendancePeriodID: 2,
			},
			mockQuery: `SELECT \* FROM "overtimes"`,
			mockRows: sqlmock.NewRows([]string{"id", "user_id", "date", "hours", "description", "attendance_period_id", "created_at"}).
				AddRow(1, 1, now, 2.5, "Extra work", 2, now),
			expectError: false,
			expectedData: []entity.Overtime{
				{
					ID:                 1,
					UserID:             1,
					Date:               now,
					Hours:              2.5,
					Description:        "Extra work",
					AttendancePeriodID: 2,
					CreatedAt:          now,
				},
			},
		},
		{
			name: "No results found",
			filter: entity.GetOvertimeFilter{
				UserID: 999,
			},
			mockQuery:    `SELECT \* FROM "overtimes"`,
			mockRows:     sqlmock.NewRows([]string{"id", "user_id", "date", "hours", "description", "attendance_period_id", "created_at"}),
			expectError:  false,
			expectedData: []entity.Overtime{},
		},
		{
			name: "DB error",
			filter: entity.GetOvertimeFilter{
				UserID: 1,
			},
			mockQuery:   `SELECT \* FROM "overtimes"`,
			mockRows:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			query := mock.ExpectQuery(tt.mockQuery)

			if tt.mockRows != nil {
				query.WillReturnRows(tt.mockRows)
			} else {
				query.WillReturnError(errors.New("db error"))
			}

			a := attendance.InitAttendanceDomain(attendance.Option{DB: db})

			result, err := a.GetOvertime(context.Background(), tt.filter)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetAttendance(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		filter       entity.GetAttendance
		mockQuery    string
		mockRows     *sqlmock.Rows
		expectError  bool
		expectedData []entity.Attendance
	}{
		{
			name: "Success with all filters",
			filter: entity.GetAttendance{
				UserID:             1,
				Date:               now,
				AttendancePeriodID: 2,
			},
			mockQuery: `SELECT \* FROM "attendances"`,
			mockRows: sqlmock.NewRows([]string{
				"id", "user_id", "checked_in_at", "checked_out_at", "attendance_period_id", "created_at",
			}).AddRow(
				1, 1, now, now.Add(8*time.Hour), 2, now,
			),
			expectError: false,
			expectedData: []entity.Attendance{
				{
					ID:                 1,
					UserID:             1,
					CheckedInAt:        pkg.TimePtr(now),
					CheckedOutAt:       pkg.TimePtr(now.Add(8 * time.Hour)),
					AttendancePeriodID: 2,
					CreatedAt:          now,
				},
			},
		},
		{
			name: "No attendance found",
			filter: entity.GetAttendance{
				UserID: 1234,
			},
			mockQuery: `SELECT \* FROM "attendances"`,
			mockRows: sqlmock.NewRows([]string{
				"id", "user_id", "checked_in_at", "checked_out_at", "attendance_period_id", "created_at",
			}),
			expectError:  false,
			expectedData: []entity.Attendance{},
		},
		{
			name: "Database error",
			filter: entity.GetAttendance{
				UserID: 1,
			},
			mockQuery:   `SELECT \* FROM "attendances"`,
			mockRows:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			query := mock.ExpectQuery(tt.mockQuery)

			if tt.mockRows != nil {
				query.WillReturnRows(tt.mockRows)
			} else {
				query.WillReturnError(errors.New("db error"))
			}

			a := attendance.InitAttendanceDomain(attendance.Option{DB: db})
			result, err := a.GetAttendance(context.Background(), tt.filter)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetAttendancePeriods(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		filter       entity.GetAttendancePeriodFilter
		mockQuery    string
		mockRows     *sqlmock.Rows
		expectError  bool
		expectedData []entity.AttendancePeriod
	}{
		{
			name: "Success with Status and UserID filters",
			filter: entity.GetAttendancePeriodFilter{
				Status: "approved",
				UserID: "user-123",
			},
			mockQuery: `SELECT \* FROM "attendance_periods"`,
			mockRows: sqlmock.NewRows([]string{
				"id", "status", "user_id", "start_date", "end_date", "created_at",
			}).AddRow(
				1, "approved", "user-123", now.AddDate(0, 0, -7), now, now,
			),
			expectError: false,
			expectedData: []entity.AttendancePeriod{
				{
					ID:        uint(1),
					Status:    "approved",
					StartDate: now.AddDate(0, 0, -7),
					EndDate:   now,
					CreatedAt: now,
				},
			},
		},
		{
			name: "No records match",
			filter: entity.GetAttendancePeriodFilter{
				Status: "rejected",
			},
			mockQuery:    `SELECT \* FROM "attendance_periods"`,
			mockRows:     sqlmock.NewRows([]string{"id", "status", "user_id", "start_date", "end_date", "created_at"}),
			expectError:  false,
			expectedData: []entity.AttendancePeriod{},
		},
		{
			name: "Database error",
			filter: entity.GetAttendancePeriodFilter{
				ID: "error-case",
			},
			mockQuery:   `SELECT \* FROM "attendance_periods"`,
			mockRows:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			query := mock.ExpectQuery(tt.mockQuery)

			if tt.mockRows != nil {
				query.WillReturnRows(tt.mockRows)
			} else {
				query.WillReturnError(errors.New("db error"))
			}

			a := attendance.InitAttendanceDomain(attendance.Option{DB: db})
			result, err := a.GetAttendancePeriods(context.Background(), tt.filter)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
