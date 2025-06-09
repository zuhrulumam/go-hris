package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/zuhrulumam/go-hris/business/entity"

	x "github.com/zuhrulumam/go-hris/pkg/errors"
)

var (
	validate = validator.New()
)

// CheckIn godoc
// @Summary      Employee check-in
// @Description  Records employee check-in attendance
// @Tags         Attendance
// @Accept       json
// @Produce      json
// @Success      200 {object} handler.CheckInResponse
// @Failure      400 {object} handler.ErrorResponse
// @Router       /api/attendance/checkin [post]
func (e *rest) CheckIn(c *gin.Context) {
	ctx := c.Request.Context()

	userID, ok := c.Get("userID")
	if !ok {
		e.compileError(c, x.NewWithCode(http.StatusUnauthorized, "missing user context"))
		return
	}

	err := e.uc.Attendance.CheckIn(ctx, entity.CheckIn{
		UserID: userID.(uint),
		Date:   time.Now(),
	})
	if err != nil {
		e.compileError(c, err)
		return
	}

	c.JSON(http.StatusOK, CheckInResponse{
		Success: true,
		Message: "Check-in successful",
	})
}

// CheckOut godoc
// @Summary      Employee check-out
// @Description  Records employee check-out attendance
// @Tags         Attendance
// @Accept       json
// @Produce      json
// @Success      200 {object} handler.CheckOutResponse
// @Failure      400 {object} handler.ErrorResponse
// @Router       /api/attendance/checkout [post]
func (e *rest) CheckOut(c *gin.Context) {
	ctx := c.Request.Context()

	userID, ok := c.Get("userID")
	if !ok {
		e.compileError(c, x.NewWithCode(http.StatusUnauthorized, "missing user context"))
		return
	}

	err := e.uc.Attendance.CheckOut(ctx, entity.CheckOut{
		UserID: userID.(uint),
		Date:   time.Now(),
	})
	if err != nil {
		e.compileError(c, err)
		return
	}

	c.JSON(http.StatusOK, CheckOutResponse{
		Success: true,
		Message: "Check-out successful",
	})
}

// CreateOvertime godoc
// @Summary      Submit overtime request
// @Description  Allows an employee to submit an overtime record
// @Tags         Overtime
// @Accept       json
// @Produce      json
// @Param        body body handler.OvertimeRequest true "Overtime Info"
// @Success      200 {object} handler.GenericResponse
// @Failure      400 {object} handler.ErrorResponse
// @Router       /api/attendance/overtime [post]
func (e *rest) CreateOvertime(c *gin.Context) {
	var (
		input OvertimeRequest
		ctx   = c.Request.Context()
	)

	if err := c.ShouldBindJSON(&input); err != nil {
		e.compileError(c, x.WrapWithCode(err, http.StatusBadRequest, "invalid input"))
		return
	}

	if err := validate.Struct(input); err != nil {
		e.compileError(c, x.WrapWithCode(err, http.StatusBadRequest, "failed validation"))
		return
	}

	userID, ok := c.Get("userID")
	if !ok {
		e.compileError(c, x.NewWithCode(http.StatusUnauthorized, "missing user context"))
		return
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		e.compileError(c, err)
		return
	}

	err = e.uc.Attendance.CreateOvertime(ctx, entity.CreateOvertimeData{
		UserID:      userID.(uint),
		Hours:       input.Hours,
		Description: input.Description,
		Date:        date,
	})
	if err != nil {
		e.compileError(c, err)
		return
	}

	c.JSON(http.StatusOK, GenericResponse{
		Success: true,
		Message: "Overtime submitted successfully!",
	})
}
