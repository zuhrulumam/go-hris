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
	ID        uint `gorm:"primaryKey"`
	StartDate time.Time
	EndDate   time.Time
	Status    string
	ClosedAt  *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Attendance struct {
	ID                 uint `gorm:"primaryKey"`
	UserID             uint `gorm:"index"` // For per-user lookup
	User               User
	AttendancePeriodID uint `gorm:"index"` // For payroll filtering
	AttendancePeriod   AttendancePeriod
	CheckedInAt        time.Time
	CheckedOutAt       time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Date               time.Time `gorm:"index"`     // For filtering by date
	Version            uint      `gorm:"default:1"` // 👈 For optimistic locking
}

type Overtime struct {
	ID                 uint `gorm:"primaryKey"`
	UserID             uint `gorm:"index"` // For per-user filtering
	User               User
	Date               time.Time `gorm:"index"`    // For date-based filtering
	Hours              float64   `gorm:"not null"` // Max 3 hrs per day
	AttendancePeriodID uint      `gorm:"index"`    // For payroll run
	AttendancePeriod   AttendancePeriod
	Description        string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Reimbursement struct {
	ID                 uint `gorm:"primaryKey"`
	UserID             uint `gorm:"index"` // For filtering
	User               User
	Amount             float64 `gorm:"not null"`
	Description        string
	AttendancePeriodID uint `gorm:"index"` // For payroll run
	AttendancePeriod   AttendancePeriod
	Date               time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Payslip struct {
	ID                 uint `gorm:"primaryKey"`
	UserID             uint `gorm:"index"` // Employee can query their payslip
	User               User
	AttendancePeriodID uint `gorm:"index"` // For period filtering
	AttendancePeriod   AttendancePeriod
	WorkingDays        int
	OvertimeHours      float64
	ReimbursementTotal float64
	BaseSalary         float64
	AttendedDays       int
	AttendanceAmount   float64
	ProratedSalary     float64
	OvertimePay        float64
	TotalPay           float64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type PayrollJob struct {
	ID        uint
	PeriodID  uint
	UserID    uint
	Status    string // pending, processing, done, failed
	Attempts  int
	LastError *string
	NextRunAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

var seedCommand = &cobra.Command{
	Use: "seed",
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
		&PayrollJob{},
	); err != nil {
		log.Fatalf("failed to migrate tables: %v", err)
	}

	err := db.Exec(`
    CREATE INDEX IF NOT EXISTS idx_attendance_period_user 
    ON attendances (attendance_period_id, user_id);
	`).Error
	if err != nil {
		log.Fatalln(err)
	}

	err = db.Exec(`
    CREATE INDEX IF NOT EXISTS idx_reimbursement_period_user 
    ON reimbursements (attendance_period_id, user_id);
	`).Error
	if err != nil {
		log.Fatalln(err)
	}

	err = db.Exec(`
    CREATE INDEX IF NOT EXISTS idx_overtime_period_user 
    ON overtimes (attendance_period_id, user_id);
	`).Error
	if err != nil {
		log.Fatalln(err)
	}

	err = db.Exec(`
    CREATE INDEX IF NOT EXISTS idx_payslip_period_user 
    ON payslips (attendance_period_id, user_id);
	`).Error
	if err != nil {
		log.Fatalln(err)
	}

	seedAdmin(db)
	seedEmployees(db, 100)
	seedAttendancePeriods(db)
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

func seedAttendancePeriods(db *gorm.DB) {
	periods := []AttendancePeriod{
		{
			StartDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2025, 6, 15, 23, 59, 59, 0, time.UTC),
			Status:    "open",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			StartDate: time.Date(2025, 6, 16, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC),
			Status:    "open",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, p := range periods {
		var exists AttendancePeriod
		err := db.Where("start_date = ? AND end_date = ?", p.StartDate, p.EndDate).First(&exists).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&p).Error; err != nil {
				log.Printf("⚠️  Failed to insert attendance period: %v", err)
			}
		}
	}

	log.Println("✅ attendance period created")
}
