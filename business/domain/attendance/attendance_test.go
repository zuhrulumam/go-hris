package attendance_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zuhrulumam/go-hris/business/domain/attendance"
	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
)

func TestGetAvailableAttendanceSpot(t *testing.T) {
	tests := []struct {
		name         string
		input        entity.GetAvailableAttendanceSpot
		mockQuery    string
		mockRows     *sqlmock.Rows
		expectError  bool
		expectedData []entity.AttendanceSpot
	}{
		{
			name: "Success with car type, active=true, occupied=false",
			input: entity.GetAvailableAttendanceSpot{
				VehicleType: "car",
				Active:      pkg.BoolPtr(true),
				Occupied:    pkg.BoolPtr(false),
			},
			mockQuery: `SELECT \* FROM "attendance_spots"`,
			mockRows: sqlmock.NewRows([]string{"id", "floor", "row", "col", "type", "occupied", "active"}).
				AddRow(1, 1, 1, 1, "car", false, true),
			expectError: false,
			expectedData: []entity.AttendanceSpot{
				{ID: 1, Floor: 1, Row: 1, Col: 1, Type: "car", Occupied: false, Active: true},
			},
		},
		{
			name: "No results found",
			input: entity.GetAvailableAttendanceSpot{
				VehicleType: "motor",
				Active:      pkg.BoolPtr(true),
			},
			mockQuery:    `SELECT \* FROM "attendance_spots"`,
			mockRows:     sqlmock.NewRows([]string{"id", "floor", "row", "col", "type", "occupied", "active"}),
			expectError:  false,
			expectedData: []entity.AttendanceSpot{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			mock.ExpectQuery(tt.mockQuery).WillReturnRows(tt.mockRows)

			d := attendance.InitAttendanceDomain(attendance.Option{DB: db})
			result, err := d.GetAvailableAttendanceSpot(context.Background(), tt.input)

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

func TestInsertVehicle(t *testing.T) {
	tests := []struct {
		name        string
		input       entity.InsertVehicle
		expectError bool
	}{
		{
			name: "Success insert vehicle",
			input: entity.InsertVehicle{
				VehicleNumber: "B1234XYZ",
				VehicleType:   "car",
				SpotID:        "1-1-1",
			},
			expectError: false,
		},
		{
			name: "Error on insert",
			input: entity.InsertVehicle{
				VehicleNumber: "INVALID",
				VehicleType:   "motor",
				SpotID:        "1-1-2",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			// Simulate insert behavior
			if tt.expectError {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "vehicles"`).
					WillReturnError(errors.New("insert failed"))
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "vehicles"`).
					WithArgs(tt.input.VehicleNumber, tt.input.VehicleType, tt.input.SpotID, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			}

			d := attendance.InitAttendanceDomain(attendance.Option{DB: db})

			tx := db.Begin()
			ctx := context.WithValue(context.Background(), pkg.TxCtxValue, tx)
			err := d.InsertVehicle(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateAttendanceSpot(t *testing.T) {
	tests := []struct {
		name        string
		input       entity.UpdateAttendanceSpot
		expectError bool
	}{
		{
			name: "Success by ID",
			input: entity.UpdateAttendanceSpot{
				ID:       1,
				Occupied: pkg.BoolPtr(true),
			},
			expectError: false,
		},
		{
			name: "Success by coordinates",
			input: entity.UpdateAttendanceSpot{
				Floor:    1,
				Row:      2,
				Col:      3,
				Occupied: pkg.BoolPtr(false),
			},
			expectError: false,
		},
		{
			name: "Missing identifier",
			input: entity.UpdateAttendanceSpot{
				Occupied: pkg.BoolPtr(true),
			},
			expectError: true,
		},
		{
			name: "No update values",
			input: entity.UpdateAttendanceSpot{
				ID: 1,
			},
			expectError: true,
		},
		{
			name: "Database error",
			input: entity.UpdateAttendanceSpot{
				ID:       99,
				Occupied: pkg.BoolPtr(true),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			if !tt.expectError {
				mock.ExpectBegin()

				mock.ExpectExec(`UPDATE "attendance_spots"`).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			} else {
				if tt.input.ID == 0 && (tt.input.Floor == 0 || tt.input.Row == 0 || tt.input.Col == 0) {
					// No DB interaction if input is invalid
				} else if tt.input.Occupied == nil {
					// No DB interaction if no update values
				} else {
					mock.ExpectBegin()
					mock.ExpectExec(`UPDATE "attendance_spots"`).
						WillReturnError(errors.New("update failed"))
					mock.ExpectRollback()
				}
			}

			d := attendance.InitAttendanceDomain(attendance.Option{DB: db})
			err := d.UpdateAttendanceSpot(context.Background(), tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateVehicle(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		input       entity.UpdateVehicle
		expectError bool
		mockQuery   bool
	}{
		{
			name: "Success case",
			input: entity.UpdateVehicle{
				ID:         1,
				UnparkedAt: &now,
			},
			expectError: false,
			mockQuery:   true,
		},
		{
			name: "Missing ID",
			input: entity.UpdateVehicle{
				ID:         0,
				UnparkedAt: &now,
			},
			expectError: true,
			mockQuery:   false,
		},
		{
			name: "No update fields",
			input: entity.UpdateVehicle{
				ID: 2,
			},
			expectError: true,
			mockQuery:   false,
		},
		{
			name: "Database error",
			input: entity.UpdateVehicle{
				ID:         3,
				UnparkedAt: &now,
			},
			expectError: true,
			mockQuery:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			if tt.mockQuery {
				mock.ExpectBegin()

				exec := mock.ExpectExec(`UPDATE "vehicles"`).
					WithArgs(tt.input.UnparkedAt, tt.input.ID)

				if tt.expectError {
					exec.WillReturnError(errors.New("update failed"))
					mock.ExpectRollback()
				} else {
					exec.WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectCommit()
				}
			}

			d := attendance.InitAttendanceDomain(attendance.Option{DB: db})
			err := d.UpdateVehicle(context.Background(), tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestGetVehicle(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		input        entity.SearchVehicle
		expectError  bool
		mockResponse *entity.Vehicle
		mockError    error
	}{
		{
			name: "Success",
			input: entity.SearchVehicle{
				VehicleNumber: "B123XYZ",
			},
			expectError: false,
			mockResponse: &entity.Vehicle{
				ID:            1,
				VehicleNumber: "B123XYZ",
				VehicleType:   "car",
				SpotID:        "1-2-3",
				ParkedAt:      now,
			},
		},
		{
			name: "DB Error",
			input: entity.SearchVehicle{
				VehicleNumber: "ERR123",
			},
			expectError:  true,
			mockResponse: nil,
			mockError:    errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := pkg.SetupMockDB(t)
			defer cleanup()

			query := `SELECT * FROM "vehicles" WHERE vehicle_number = $1 ORDER BY id DESC,"vehicles"."id" LIMIT $2`
			if tt.mockResponse != nil {
				rows := sqlmock.NewRows([]string{"id", "vehicle_number", "vehicle_type", "spot_id", "parked_at"}).
					AddRow(tt.mockResponse.ID, tt.mockResponse.VehicleNumber, tt.mockResponse.VehicleType, tt.mockResponse.SpotID, tt.mockResponse.ParkedAt)
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(tt.input.VehicleNumber, 1).
					WillReturnRows(rows)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(tt.input.VehicleNumber, 1).
					WillReturnError(tt.mockError)
			}

			d := attendance.InitAttendanceDomain(attendance.Option{DB: db})
			_, err := d.GetVehicle(context.Background(), tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
