package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zuhrulumam/go-hris/business/usecase"
	_ "github.com/zuhrulumam/go-hris/docs" // replace with your module
	"github.com/zuhrulumam/go-hris/pkg/middlewares"
	"go.uber.org/zap"
)

type Rest interface {
}

type Option struct {
	Uc  *usecase.Usecase
	App *gin.Engine
	Log *zap.Logger
}

type rest struct {
	uc  *usecase.Usecase
	app *gin.Engine
	log *zap.Logger
}

func Init(opt Option) Rest {
	e := &rest{
		uc:  opt.Uc,
		app: opt.App,
		log: opt.Log,
	}

	e.Serve()

	return e
}

func (r rest) Serve() {
	// swagger
	// r.app.GET("/swagger/*", swagger.HandlerDefault)

	api := r.app.Group("/api")
	api.Use(middlewares.JWTMiddleware())

	api.POST("/attendace/checkin", r.CheckIn)

	api.POST("/attendace/checkout", r.CheckOut)

	api.POST("/attendace/overtime", r.CreateOvertime)
}
