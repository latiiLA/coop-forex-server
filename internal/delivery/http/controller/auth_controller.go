package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	log "github.com/sirupsen/logrus"
)

type AuthController interface {
	Login(c *gin.Context)
}

type authController struct {
	usecase usecase.AuthUsecase
}

func NewAuthController(u usecase.AuthUsecase) AuthController {
	return &authController{
		usecase: u,
	}
}

func (a *authController) Login(c *gin.Context) {
	var req model.LoginRequestDTO
	traceID := c.GetString("TraceID")

	// Log metadata early
	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithFields(log.Fields{
			"ip":         clientIP,
			"user_agent": userAgent,
			"error":      err.Error(),
		}).Warn("Invalid login request payload")

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	log.WithFields(log.Fields{
		"trace_id":   traceID,
		"username":   req.Username,
		"ip":         clientIP,
		"user_agent": userAgent,
	}).Info("Login attempt")

	// Call the usecase to perform authentication
	user, err := a.usecase.Authenticate(c, req.Username, req.Password, c.ClientIP())
	if err != nil {
		log.WithFields(log.Fields{
			"trace_id": traceID,
			"username": req.Username,
			"ip":       clientIP,
		}).Warn("Login failed")
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: "invalid credential or internal server error"})
		return
	}

	log.WithFields(log.Fields{
		"trace_id": traceID,
		"username": req.Username,
		"ip":       clientIP,
	}).Info("Login successful")

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Login Successful", Data: user})
}
