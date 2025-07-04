package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/zuhrulumam/go-hris/business/usecase"
	_ "github.com/zuhrulumam/go-hris/docs"
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
	r.app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.app.POST("/login", r.Login)
	r.app.POST("/register", r.Register)

	api := r.app.Group("/api")
	api.Use(middlewares.JWTMiddleware())

	api.POST("/attendance/checkin", r.CheckIn)
	api.POST("/attendance/checkout", r.CheckOut)
	api.POST("/attendance/overtime", r.CreateOvertime)

	api.POST("/reimbursement/submit", r.SubmitReimbursement)

	api.POST("/payroll/create", r.CreatePayroll)
	api.GET("/payslip", r.GetPayslip)

	api.GET("/payroll/summary", r.GetPayrollSummary)

	api.POST("/attendance/period", r.CreateAttendancePeriod)
}
