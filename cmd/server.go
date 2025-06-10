package cmd

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/spf13/cobra"
	"github.com/zuhrulumam/go-hris/business/domain"
	"github.com/zuhrulumam/go-hris/business/usecase"
	"github.com/zuhrulumam/go-hris/handler"
	"github.com/zuhrulumam/go-hris/pkg/logger"
	"github.com/zuhrulumam/go-hris/pkg/middlewares"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var serverCommand = &cobra.Command{
	Use:   "start",
	Short: "start server",
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

var (
	dom *domain.Domain
	uc  *usecase.Usecase
	db  *gorm.DB
	lg  *zap.Logger
)

func run() {

	lg = logger.NewZapLogger()

	app := gin.Default()
	app.Use(middlewares.RequestContextMiddleware(lg))

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

	// init asynq client
	aClient := NewAsynqClient()

	// init usecase
	uc = usecase.Init(dom, usecase.Option{
		AsynqClient: aClient,
	})

	// init rest
	handler.Init(handler.Option{
		Uc:  uc,
		App: app,
		Log: lg,
	})

	log.Println(app.Run(":8080"))
}

func NewAsynqClient() *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr: os.Getenv("REDIS_HOST"),
	})
}

// TODO: Gracefull shutdown
