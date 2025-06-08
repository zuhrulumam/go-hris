package cmd

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleEmployee UserRole = "employee"
)

type User struct {
	ID        uint     `gorm:"primaryKey"`
	Username  string   `gorm:"unique;not null"` // Unique index
	Password  string   `gorm:"not null"`
	Role      UserRole `gorm:"type:varchar(20);index"` // Optional index
	Salary    float64  `gorm:"default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AttendancePeriod struct {
	ID        uint      `gorm:"primaryKey"`
	StartDate time.Time `gorm:"index"` // Optional index
	EndDate   time.Time `gorm:"index"` // Optional index
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Attendance struct {
	ID                 uint `gorm:"primaryKey"`
	UserID             uint `gorm:"index"` // For per-user lookup
	User               User
	Date               time.Time `gorm:"index"` // For filtering by date
	AttendancePeriodID uint      `gorm:"index"` // For payroll filtering
	AttendancePeriod   AttendancePeriod
	CreatedAt          time.Time

	// Ensure uniqueness: one submission per user per day
	// This also improves lookup speed for validation
	// Named composite index
	_  struct{} `gorm:"uniqueIndex:idx_attendance_user_date,priority:1"`
	_2 struct{} `gorm:"uniqueIndex:idx_attendance_user_date,priority:2"`
}

type Overtime struct {
	ID                 uint `gorm:"primaryKey"`
	UserID             uint `gorm:"index"` // For per-user filtering
	User               User
	Date               time.Time `gorm:"index"`    // For date-based filtering
	Hours              float64   `gorm:"not null"` // Max 3 hrs per day
	AttendancePeriodID uint      `gorm:"index"`    // For payroll run
	AttendancePeriod   AttendancePeriod
	CreatedAt          time.Time

	// Optional: prevent duplicate overtime per user per date
	_  struct{} `gorm:"uniqueIndex:idx_overtime_user_date,priority:1"`
	_2 struct{} `gorm:"uniqueIndex:idx_overtime_user_date,priority:2"`
}

type Reimbursement struct {
	ID                 uint `gorm:"primaryKey"`
	UserID             uint `gorm:"index"` // For filtering
	User               User
	Amount             float64 `gorm:"not null"`
	Description        string
	AttendancePeriodID uint `gorm:"index"` // For payroll run
	AttendancePeriod   AttendancePeriod
	CreatedAt          time.Time
}

type Payslip struct {
	ID                 uint `gorm:"primaryKey"`
	UserID             uint `gorm:"index"` // Employee can query their payslip
	User               User
	AttendancePeriodID uint `gorm:"index"` // For period filtering
	AttendancePeriod   AttendancePeriod

	TotalWorkDays      int
	TotalOvertimeHours float64
	TotalReimburse     float64
	BaseSalary         float64
	ProratedSalary     float64
	OvertimePay        float64
	TakeHomePay        float64
	CreatedAt          time.Time

	// Ensure only one payslip per user per period
	_  struct{} `gorm:"uniqueIndex:idx_payslip_user_period,priority:1"`
	_2 struct{} `gorm:"uniqueIndex:idx_payslip_user_period,priority:2"`
}

var seedCommand = &cobra.Command{
	Use:  "seed [floors] [rows] [cols]",
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {

		db, _ := connectDB()

		seed(db)
	},
}

func seed(db *gorm.DB) {
	// migrate db
	if err := db.AutoMigrate(
		&User{},
		&AttendancePeriod{},
		&Attendance{},
		&Overtime{},
		&Reimbursement{},
		&Payslip{},
	); err != nil {
		log.Fatalf("failed to migrate tables: %v", err)
	}

	// // add constraint
	// err := db.Exec(`
	// 	CREATE UNIQUE INDEX IF NOT EXISTS unique_active_spot
	// 	ON vehicles(spot_id)
	// 	WHERE unparked_at IS NULL
	// `).Error
	// if err != nil {
	// 	log.Fatalf("failed to add index table: %v", err)
	// }

	seedAdmin(db)
	seedEmployees(db, 100)
}

func connectDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// Enable debug mode if not production
	if os.Getenv("ENV") != "production" {
		db = db.Debug()
	}
	return db, nil
}

func seedAdmin(db *gorm.DB) {
	hashedPassword := hashPassword("admin123") // Implement a real hash function!
	admin := User{
		Username: "admin",
		Password: hashedPassword,
		Role:     RoleAdmin,
		Salary:   0, // Admin doesn't need salary
	}

	if err := db.FirstOrCreate(&admin, User{Username: "admin"}).Error; err != nil {
		log.Fatalf("failed to seed admin: %v", err)
	}

	log.Println("✅ Admin user created")
}

func seedEmployees(db *gorm.DB, count int) {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < count; i++ {
		username := fmt.Sprintf("employee%d", i+1)
		hashedPassword := hashPassword("password123")   // Use real hashing
		salary := float64(rand.Intn(5000000) + 3000000) // 3M–8M range

		employee := User{
			Username: username,
			Password: hashedPassword,
			Role:     RoleEmployee,
			Salary:   salary,
		}

		if err := db.Create(&employee).Error; err != nil {
			log.Printf("⚠️  Failed to insert employee %s: %v", username, err)
		}
	}

	log.Printf("✅ Seeded %d employees\n", count)
}

func hashPassword(pw string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}
	return string(hashed)
}
