package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zuhrulumam/go-hris/business/entity"
	x "github.com/zuhrulumam/go-hris/pkg/errors"
)

// SubmitReimbursement godoc
// @Summary      Submit a reimbursement request
// @Description  Allows a user to submit a reimbursement claim
// @Tags         Reimbursement
// @Accept       json
// @Produce      json
// @Param        body body handler.ReimbursementRequest true "Reimbursement Info"
// @Success      200 {object} handler.GenericResponse
// @Failure      400 {object} handler.ErrorResponse
// @Failure      401 {object} handler.ErrorResponse
// @Router       /api/reimbursement [post]
func (e *rest) SubmitReimbursement(c *gin.Context) {
	var input ReimbursementRequest
	ctx := c.Request.Context()

	userID, ok := c.Get("userID")
	if !ok {
		e.compileError(c, x.NewWithCode(http.StatusUnauthorized, "missing user context"))
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		e.compileError(c, x.WrapWithCode(err, http.StatusBadRequest, "invalid input"))
		return
	}

	if err := validate.Struct(input); err != nil {
		e.compileError(c, x.WrapWithCode(err, http.StatusBadRequest, "validation failed"))
		return
	}

	err := e.uc.Reimbursement.SubmitReimbursement(ctx, entity.SubmitReimbursementData{
		UserID:      userID.(uint),
		Amount:      input.Amount,
		Description: input.Description,
		// Date:        input.Date,
	})
	if err != nil {
		e.compileError(c, err)
		return
	}

	c.JSON(http.StatusOK, GenericResponse{
		Success: true,
		Message: "Reimbursement submitted successfully!",
	})
}
