package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zuhrulumam/go-hris/business/entity"
	"github.com/zuhrulumam/go-hris/pkg"
	x "github.com/zuhrulumam/go-hris/pkg/errors"
)

// GetPayslip godoc
// @Summary      Get user's payslip
// @Description  Retrieve the payslip for the currently logged-in user for a specific attendance period
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        period_id query int true "Attendance Period ID"
// @Param        page query int false "Page number"
// @Param        limit query int false "Page limit"
// @Success      200 {object} handler.PayslipListResponse
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

	payslip, _, _, err := e.uc.Payslip.GetPayslip(ctx, entity.GetPayslipRequest{
		UserID:             userID.(*uint),
		AttendancePeriodID: pkg.UintPtr(uint(periodID)),
	})
	if err != nil {
		e.compileError(c, err)
		return
	}

	c.JSON(http.StatusOK, payslip[0])
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

	isAdmin, ok := c.Get("isAdmin")
	if !ok {
		e.compileError(c, x.NewWithCode(http.StatusUnauthorized, "missing user context"))
		return
	}

	if !isAdmin.(bool) {
		e.compileError(c, x.NewWithCode(http.StatusUnauthorized, "only admin"))
		return
	}

	err := e.uc.Payslip.CreatePayroll(c.Request.Context(), req.PeriodID)
	if err != nil {
		e.compileError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Payroll successfully created"})
}

// GetPayrollSummary godoc
// @Summary      Get payroll summary
// @Description  Retrieve payroll summary for multiple attendance periods, grouped by user
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        period_ids query string true "Comma-separated Attendance Period IDs (e.g., 1,2,3)"
// @Success      200 {object} GetPayrollSummaryResponse
// @Failure      400 {object} handler.ErrorResponse
// @Failure      401 {object} handler.ErrorResponse
// @Failure      500 {object} handler.ErrorResponse
// @Router       /api/payroll/summary [get]
func (e *rest) GetPayrollSummary(c *gin.Context) {
	ctx := c.Request.Context()

	isAdmin, ok := c.Get("isAdmin")
	if !ok {
		e.compileError(c, x.NewWithCode(http.StatusUnauthorized, "missing user context"))
		return
	}

	if !isAdmin.(bool) {
		e.compileError(c, x.NewWithCode(http.StatusUnauthorized, "only admin"))
		return
	}

	periodIDsParam := c.Query("period_ids")
	if periodIDsParam == "" {
		e.compileError(c, x.NewWithCode(http.StatusBadRequest, "missing period_ids"))
		return
	}

	// Parse comma-separated values into []int64
	strIDs := strings.Split(periodIDsParam, ",")
	var periodIDs []uint
	for _, str := range strIDs {
		id, err := strconv.ParseInt(strings.TrimSpace(str), 10, 64)
		if err != nil || id <= 0 {
			e.compileError(c, x.NewWithCode(http.StatusBadRequest, fmt.Sprintf("invalid period_id: %s", str)))
			return
		}
		periodIDs = append(periodIDs, uint(id))
	}

	summary, err := e.uc.Payslip.GetPayrollSummary(ctx, entity.GetPayrollSummaryRequest{
		AttendancePeriodIDs: periodIDs,
	})
	if err != nil {
		e.compileError(c, err)
		return
	}

	c.JSON(http.StatusOK, GetPayrollSummaryResponse{
		Items:      summary.Items,
		GrandTotal: summary.GrandTotal,
	})
}
