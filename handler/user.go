package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zuhrulumam/go-hris/business/entity"
	x "github.com/zuhrulumam/go-hris/pkg/errors"
)

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account with email and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        registerRequest  body      RegisterRequest  true  "Register payload"
// @Success      201              {object}  map[string]string "User registered successfully"
// @Failure      400              {object}  map[string]string "Invalid input"
// @Failure      500              {object}  map[string]string "Internal server error"
// @Router       /auth/register [post]
func (r *rest) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.compileError(c, x.WrapWithCode(err, http.StatusBadRequest, "invalid input"))
		return
	}

	err := r.uc.User.Register(c.Request.Context(), entity.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		FullName: req.Fullname,
		Salary:   req.Salary,
	})
	if err != nil {
		r.compileError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Login godoc
// @Summary      Login user and get JWT token
// @Description  Authenticate user and return access token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        loginRequest  body      LoginRequest  true  "Login payload"
// @Success      200           {object}  AuthResponse
// @Failure      400           {object}  map[string]string "Invalid input"
// @Failure      401           {object}  map[string]string "Unauthorized"
// @Failure      500           {object}  map[string]string "Internal server error"
// @Router       /auth/login [post]
func (r *rest) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.compileError(c, x.WrapWithCode(err, http.StatusBadRequest, "invalid input"))
		return
	}

	token, err := r.uc.User.Login(c.Request.Context(), entity.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		r.compileError(c, err)
		return
	}

	c.JSON(http.StatusOK, AuthResponse{Token: token})
}
