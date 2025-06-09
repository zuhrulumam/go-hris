package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zuhrulumam/go-hris/pkg/errors"
	"github.com/zuhrulumam/go-hris/pkg/logger"
)

func (e *rest) compileError(c *gin.Context, err error) {
	var (
		httpStatus int
		he         string
		code       = errors.ErrCode(err)
	)

	switch code {
	case 400:
		httpStatus = http.StatusBadRequest
		he = errors.EM.Message("EN", "badrequest")
	case 404:
		httpStatus = http.StatusNotFound
		he = errors.EM.Message("EN", "notfound")
	default:
		httpStatus = http.StatusInternalServerError
		he = errors.EM.Message("EN", "internal")
	}

	logger.LogWithCtx(c, e.log, err.Error())

	c.JSON(httpStatus, ErrorResponse{
		HumanError: he,
		DebugError: err.Error(),
		Success:    false,
	})
}
