package attendance_test

// func TestPark(t *testing.T) {

// 	tests := []struct {
// 		name        string
// 		setupMocks  func(p mockAttendance.MockDomainItf, t mockTx.MockDomainItf)
// 		expectedErr bool
// 	}{
// 		{
// 			name: "success attendance",
// 			setupMocks: func(p mockAttendance.MockDomainItf, t mockTx.MockDomainItf) {
// 				t.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {

// 					p.EXPECT().GetAvailableAttendanceSpot(gomock.Any(), gomock.Any()).
// 						Return([]entity.AttendanceSpot{{ID: 1, Floor: 1, Row: 1, Col: 1}}, nil)

// 					p.EXPECT().InsertVehicle(gomock.Any(), gomock.Any()).
// 						Return(nil)

// 					p.EXPECT().UpdateAttendanceSpot(gomock.Any(), gomock.Any()).
// 						Return(nil)

// 					return fn(ctx)
// 				})

// 			},
// 			expectedErr: false,
// 		},
// 		{
// 			name: "no available spot",
// 			setupMocks: func(p mockAttendance.MockDomainItf, t mockTx.MockDomainItf) {
// 				t.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {

// 					p.EXPECT().GetAvailableAttendanceSpot(gomock.Any(), gomock.Any()).
// 						Return([]entity.AttendanceSpot{}, nil)

// 					return fn(ctx)
// 				})

// 			},
// 			expectedErr: true,
// 		},
// 		{
// 			name: "insert vehicle failed",
// 			setupMocks: func(p mockAttendance.MockDomainItf, t mockTx.MockDomainItf) {
// 				t.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {

// 					p.EXPECT().GetAvailableAttendanceSpot(gomock.Any(), gomock.Any()).
// 						Return([]entity.AttendanceSpot{{ID: 1, Floor: 1, Row: 1, Col: 1}}, nil)

// 					p.EXPECT().InsertVehicle(gomock.Any(), gomock.Any()).
// 						Return(errors.New("insert failed"))

// 					return fn(ctx)
// 				})

// 			},
// 			expectedErr: true,
// 		},
// 		{
// 			name: "update spot failed",
// 			setupMocks: func(p mockAttendance.MockDomainItf, t mockTx.MockDomainItf) {
// 				t.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {

// 					p.EXPECT().GetAvailableAttendanceSpot(gomock.Any(), gomock.Any()).
// 						Return([]entity.AttendanceSpot{{ID: 1, Floor: 1, Row: 1, Col: 1}}, nil)

// 					p.EXPECT().InsertVehicle(gomock.Any(), gomock.Any()).
// 						Return(nil)

// 					p.EXPECT().UpdateAttendanceSpot(gomock.Any(), gomock.Any()).
// 						Return(errors.New("update failed"))

// 					return fn(ctx)
// 				})
// 			},
// 			expectedErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()
// 			mocktx := mockTx.NewMockDomainItf(ctrl)
// 			mockpark := mockAttendance.NewMockDomainItf(ctrl)

// 			tt.setupMocks(*mockpark, *mocktx)

// 			usecase := uc.InitAttendanceUsecase(uc.Option{
// 				AttendanceDom:  mockpark,
// 				TransactionDom: mocktx,
// 			})

// 			err := usecase.Park(context.Background(), entity.Park{
// 				VehicleNumber: "B1234XYZ",
// 				VehicleType:   "car",
// 			})

// 			if tt.expectedErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}

// 		})
// 	}
// }

// func TestUnpark(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		setupMocks  func(p *mockAttendance.MockDomainItf, t *mockTx.MockDomainItf)
// 		expectedErr bool
// 	}{
// 		{
// 			name: "success unpark",
// 			setupMocks: func(p *mockAttendance.MockDomainItf, t *mockTx.MockDomainItf) {
// 				t.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
// 					p.EXPECT().GetVehicle(gomock.Any(), gomock.Any()).Return(entity.Vehicle{
// 						ID:            1,
// 						VehicleNumber: "B1234XYZ",
// 						SpotID:        "1-2-3",
// 						UnparkedAt:    nil,
// 					}, nil)

// 					p.EXPECT().UpdateVehicle(gomock.Any(), gomock.Any()).Return(nil)

// 					p.EXPECT().UpdateAttendanceSpot(gomock.Any(), gomock.Any()).Return(nil)

// 					return fn(ctx)
// 				})
// 			},
// 			expectedErr: false,
// 		},
// 		{
// 			name: "already unparked",
// 			setupMocks: func(p *mockAttendance.MockDomainItf, t *mockTx.MockDomainItf) {
// 				t.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
// 					now := time.Now()
// 					p.EXPECT().GetVehicle(gomock.Any(), gomock.Any()).Return(entity.Vehicle{
// 						ID:            1,
// 						VehicleNumber: "B1234XYZ",
// 						SpotID:        "1-2-3",
// 						UnparkedAt:    &now,
// 					}, nil)
// 					return fn(ctx)
// 				})
// 			},
// 			expectedErr: true,
// 		},
// 		{
// 			name: "get vehicle failed",
// 			setupMocks: func(p *mockAttendance.MockDomainItf, t *mockTx.MockDomainItf) {
// 				t.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
// 					p.EXPECT().GetVehicle(gomock.Any(), gomock.Any()).Return(entity.Vehicle{}, errors.New("not found"))
// 					return fn(ctx)
// 				})
// 			},
// 			expectedErr: true,
// 		},
// 		{
// 			name: "update vehicle failed",
// 			setupMocks: func(p *mockAttendance.MockDomainItf, t *mockTx.MockDomainItf) {
// 				t.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
// 					p.EXPECT().GetVehicle(gomock.Any(), gomock.Any()).Return(entity.Vehicle{
// 						ID:            1,
// 						VehicleNumber: "B1234XYZ",
// 						SpotID:        "1-2-3",
// 						UnparkedAt:    nil,
// 					}, nil)

// 					p.EXPECT().UpdateVehicle(gomock.Any(), gomock.Any()).Return(errors.New("update failed"))

// 					return fn(ctx)
// 				})
// 			},
// 			expectedErr: true,
// 		},
// 		{
// 			name: "update attendance spot failed",
// 			setupMocks: func(p *mockAttendance.MockDomainItf, t *mockTx.MockDomainItf) {
// 				t.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
// 					p.EXPECT().GetVehicle(gomock.Any(), gomock.Any()).Return(entity.Vehicle{
// 						ID:            1,
// 						VehicleNumber: "B1234XYZ",
// 						SpotID:        "1-2-3",
// 						UnparkedAt:    nil,
// 					}, nil)

// 					p.EXPECT().UpdateVehicle(gomock.Any(), gomock.Any()).Return(nil)

// 					p.EXPECT().UpdateAttendanceSpot(gomock.Any(), gomock.Any()).Return(errors.New("update spot failed"))

// 					return fn(ctx)
// 				})
// 			},
// 			expectedErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockTx := mockTx.NewMockDomainItf(ctrl)
// 			mockPark := mockAttendance.NewMockDomainItf(ctrl)

// 			tt.setupMocks(mockPark, mockTx)

// 			usecase := uc.InitAttendanceUsecase(uc.Option{
// 				AttendanceDom:  mockPark,
// 				TransactionDom: mockTx,
// 			})

// 			err := usecase.Unpark(context.Background(), entity.UnPark{
// 				VehicleNumber: "B1234XYZ",
// 			})

// 			if tt.expectedErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func TestAvailableSpot(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockPark := mockAttendance.NewMockDomainItf(ctrl)

// 	usecase := uc.InitAttendanceUsecase(uc.Option{
// 		AttendanceDom:  mockPark,
// 		TransactionDom: nil,
// 	})

// 	tests := []struct {
// 		name        string
// 		mockReturn  []entity.AttendanceSpot
// 		input       entity.GetAvailablePark
// 		mockError   error
// 		expectError bool
// 	}{
// 		{
// 			name: "success",
// 			mockReturn: []entity.AttendanceSpot{
// 				{ID: 1, Floor: 1, Row: 1, Col: 1},
// 			},
// 			input: entity.GetAvailablePark{
// 				VehicleType: "M",
// 			},
// 			mockError:   nil,
// 			expectError: false,
// 		},
// 		{
// 			name:       "error from domain",
// 			mockReturn: nil,
// 			input: entity.GetAvailablePark{
// 				VehicleType: "M",
// 			},
// 			mockError:   errors.New("db error"),
// 			expectError: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockPark.EXPECT().
// 				GetAvailableAttendanceSpot(gomock.Any(), gomock.Any()).
// 				Return(tt.mockReturn, tt.mockError)

// 			spots, err := usecase.AvailableSpot(context.Background(), tt.input)
// 			if tt.expectError {
// 				assert.Error(t, err)
// 				assert.Nil(t, spots)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tt.mockReturn, spots)
// 			}
// 		})
// 	}
// }

// func TestSearchVehicle(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockPark := mockAttendance.NewMockDomainItf(ctrl)

// 	usecase := uc.InitAttendanceUsecase(uc.Option{
// 		AttendanceDom:  mockPark,
// 		TransactionDom: nil,
// 	})

// 	tests := []struct {
// 		name        string
// 		inputNumber entity.SearchVehicle
// 		mockReturn  entity.Vehicle
// 		mockError   error
// 		expectError bool
// 	}{
// 		{
// 			name: "success",
// 			inputNumber: entity.SearchVehicle{
// 				VehicleNumber: "B1234XYZ",
// 			},
// 			mockReturn:  entity.Vehicle{VehicleNumber: "B1234XYZ"},
// 			mockError:   nil,
// 			expectError: false,
// 		},
// 		{
// 			name: "vehicle not found",
// 			inputNumber: entity.SearchVehicle{
// 				VehicleNumber: "B1234XYZ",
// 			},
// 			mockReturn:  entity.Vehicle{},
// 			mockError:   errors.New("not found"),
// 			expectError: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockPark.EXPECT().
// 				GetVehicle(gomock.Any(), tt.inputNumber).
// 				Return(tt.mockReturn, tt.mockError)

// 			vehicle, err := usecase.SearchVehicle(context.Background(), tt.inputNumber)
// 			if tt.expectError {
// 				assert.Error(t, err)
// 				assert.Empty(t, vehicle.VehicleNumber)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tt.mockReturn.VehicleNumber, vehicle.VehicleNumber)
// 			}
// 		})
// 	}
// }
