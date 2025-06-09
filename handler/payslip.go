package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	x "github.com/zuhrulumam/go-hris/pkg/errors"
)

// GetPayslip godoc
// @Summary      Get user's payslip
// @Description  Retrieve the payslip for the currently logged-in user for a specific attendance period
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        period_id query int true "Attendance Period ID"
// @Success      200 {object} entity.Payslip
// @Failure      400 {object} handler.ErrorResponse
// @Failure      401 {object} handler.ErrorResponse
// @Failure      500 {object} handler.ErrorResponse
// @Router       /api/payslip [get]
func (e *rest) GetPayslip(c *gin.Context) {
	ctx := c.Request.Context()

	userID, ok := c.Get("userID")
	if !ok {
		e.compileError(c, x.NewWithCode(http.StatusUnauthorized, "missing user context"))
		return
	}

	periodIDStr := c.Query("period_id")
	if periodIDStr == "" {
		e.compileError(c, x.NewWithCode(http.StatusBadRequest, "missing period_id"))
		return
	}

	periodID, err := strconv.Atoi(periodIDStr)
	if err != nil || periodID <= 0 {
		e.compileError(c, x.NewWithCode(http.StatusBadRequest, "invalid period_id"))
		return
	}

	payslip, err := e.uc.Payslip.GetPayslip(ctx, uint(userID.(uint)), uint(periodID))
	if err != nil {
		e.compileError(c, err)
		return
	}

	c.JSON(http.StatusOK, payslip)
}

// CreatePayroll godoc
// @Summary      Create payroll for an attendance period
// @Description  This endpoint processes payroll based on attendance, overtime, and reimbursement records.
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        payload  body      CreatePayrollRequest  true  "Period ID Payload"
// @Success      201      {object}  map[string]string      "message"
// @Failure      400      {object}  map[string]string      "Bad Request"
// @Failure      409      {object}  map[string]string      "Payroll already exists"
// @Failure      500      {object}  map[string]string      "Internal Server Error"
// @Router       /api/payroll/create [post]
func (e *rest) CreatePayroll(c *gin.Context) {
	var req CreatePayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		e.compileError(c, x.WrapWithCode(err, http.StatusBadRequest, "invalid input"))
		return
	}

	err := e.uc.Payslip.CreatePayroll(c.Request.Context(), req.PeriodID)
	if err != nil {
		e.compileError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Payroll successfully created"})
}
