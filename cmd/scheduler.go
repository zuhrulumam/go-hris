package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var schedulerCommand = &cobra.Command{
	Use:   "start-scheduler",
	Short: "start scheduler",
	Run: func(cmd *cobra.Command, args []string) {
		runScheduler()
	},
}

func runScheduler() {
	// init sql
	g, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}

	db = g

	// init asynq client
	aClient = NewAsynqClient()

	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			if err := processPendingPayrollJobs(); err != nil {
				log.Println("Scheduler error:", err)
			}
			if err := closeFinishedAttendancePeriods(); err != nil {
				log.Println("Scheduler error (closure):", err)
			}
		}
	}
}

const maxJobsPerTick = 100 // Limit per scheduler cycle

func processPendingPayrollJobs() error {
	var (
		jobs []entity.PayrollJob
		ctx  = context.Background()
	)

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock rows for processing
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("status = ? AND next_run_at <= ?", "pending", time.Now()).
			Order("next_run_at ASC").
			Limit(maxJobsPerTick).
			Find(&jobs).Error; err != nil {
			return fmt.Errorf("failed to load pending jobs: %w", err)
		}

		// Mark jobs as processing
		for _, job := range jobs {
			if err := tx.Model(&entity.PayrollJob{}).
				Where("id = ?", job.ID).
				Updates(map[string]interface{}{
					"status":     "processing",
					"updated_at": time.Now(),
				}).Error; err != nil {
				return fmt.Errorf("failed to mark job %d as processing: %w", job.ID, err)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Enqueue outside transaction
	for _, job := range jobs {
		t, err := task.NewCreatePayrollTask(job.AttendancePeriodID, job.UserID, job.ID)
		if err != nil {
			log.Printf("failed to create task for job %d: %v", job.ID, err)
			continue
		}

		if _, err := aClient.Enqueue(t); err != nil {
			log.Printf("failed to enqueue task for job %d: %v", job.ID, err)
			continue
		}

		log.Printf("Enqueued payroll job: ID=%d UserID=%d PeriodID=%d", job.ID, job.UserID, job.AttendancePeriodID)
	}

	return nil
}

func closeFinishedAttendancePeriods() error {
	ctx := context.Background()
	var periods []entity.AttendancePeriod

	if err := db.WithContext(ctx).
		Where("status = ?", "open").
		Find(&periods).Error; err != nil {
		return err
	}

	for _, period := range periods {
		var total, completed int64

		if err := db.Model(&entity.PayrollJob{}).
			Where("attendance_period_id = ?", period.ID).
			Count(&total).Error; err != nil {
			continue
		}

		if err := db.Model(&entity.PayrollJob{}).
			Where("attendance_period_id = ? AND status = ?", period.ID, "completed").
			Count(&completed).Error; err != nil {
			continue
		}

		if total > 0 && total == completed {
			log.Printf("Closing period ID %d", period.ID)
			if err := db.WithContext(ctx).Model(&entity.AttendancePeriod{}).
				Where("id = ?", period.ID).
				Updates(map[string]interface{}{
					"status":    "closed",
					"closed_at": time.Now(),
				}).Error; err != nil {
				log.Printf("Failed to close period %d: %v", period.ID, err)
			}
		}
	}

	return nil
}
