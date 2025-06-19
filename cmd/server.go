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
	"github.com/zuhrulumam/go-hris/pkg/metrics"
	"github.com/zuhrulumam/go-hris/pkg/middlewares"
	"github.com/zuhrulumam/go-hris/pkg/tracer"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
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
	dom     *domain.Domain
	uc      *usecase.Usecase
	db      *gorm.DB
	lg      *zap.Logger
	aClient *asynq.Client
)

func run() {

	trace := tracer.InitTracer(tracer.Option{
		JaegerHost: os.Getenv("JAEGER_HOST"),
	})
	defer trace()

	lg = logger.NewZapLogger()

	app := gin.Default()
	app.Use(middlewares.RequestContextMiddleware(lg))
	app.Use(otelgin.Middleware("go-hris"))
	app.Use(middlewares.TracerLogger())

	// init metrics
	metrics.Init(app, []metrics.SkipHandler{}, "go-hris")

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
	aClient = NewAsynqClient()

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
