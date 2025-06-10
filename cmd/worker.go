package cmd

import (
	"log"
	"os"

	"github.com/hibiken/asynq"
	"github.com/spf13/cobra"
	"github.com/zuhrulumam/go-hris/business/domain"
	"github.com/zuhrulumam/go-hris/business/usecase"
	"github.com/zuhrulumam/go-hris/handler/worker"
	"github.com/zuhrulumam/go-hris/task"
)

var workerCommand = &cobra.Command{
	Use:   "start-worker",
	Short: "start worker",
	Run: func(cmd *cobra.Command, args []string) {
		runWorker()
	},
}

func runWorker() {
	r := asynq.RedisClientOpt{Addr: os.Getenv("REDIS_HOST")}

	srv := asynq.NewServer(r, asynq.Config{
		Concurrency: 10, // adjust based on load
	})

	mux := asynq.NewServeMux()

	// init sql
	g, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}

	db = g

	// init domain
	dom = domain.Init(domain.Option{
		DB: db,
	})

	// init usecase
	uc = usecase.Init(dom, usecase.Option{})

	handler := &worker.Handler{
		Payslip: uc.Payslip,
	}

	mux.HandleFunc(task.TypeCreatePayroll, handler.HandleCreatePayrollTask)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("😢 Could not run Asynq worker: %v", err)
	}
}
