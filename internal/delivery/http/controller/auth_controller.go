package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	log "github.com/sirupsen/logrus"
)

type AuthController interface {
	Login(c *gin.Context)
	Register(c *gin.Context)
}

type authController struct {
	authUsecase usecase.AuthUsecase
	userUsecase usecase.UserUsecase
}

func NewAuthController(u usecase.AuthUsecase, ldap usecase.UserUsecase) AuthController {
	return &authController{
		authUsecase: u,
		userUsecase: ldap,
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
	user, err := a.authUsecase.Authenticate(c, req.Username, req.Password, c.ClientIP())
	if err != nil {
		log.WithFields(log.Fields{
			"trace_id": traceID,
			"username": req.Username,
			"ip":       clientIP,
			"error":    err.Error(),
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

func (a *authController) Register(c *gin.Context) {
	logEntry := utils.GetLogger(c)
	authUserID, _ := utils.GetUserID(c)

	var registerReq model.RegisterRequestDTO
	if err := c.ShouldBindJSON(&registerReq); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	// --- STEP 1: Call AuthUsecase (LDAP) ---
	// We get the official data from Active Directory
	adUser, err := a.authUsecase.GetUserDetails(c.Request.Context(), registerReq.Username)
	if err != nil {
		logEntry.Warn("Get user detail from AD error: ", err)
		c.JSON(http.StatusForbidden, response.ErrorResponse{Message: "User must exist in AD"})
		return
	}

	RegisterUsecaseReq := model.RegisterUsecaseRequestDTO{
		Username:     registerReq.Username,
		FirstName:    adUser.Profile.FirstName,
		MiddleName:   adUser.Profile.MiddleName,
		LastName:     registerReq.LastName,
		DisplayName:  adUser.Profile.DisplayName,
		Email:        adUser.Profile.Email,
		BranchID:     registerReq.BranchID,
		DepartmentID: registerReq.DepartmentID,
		Role:         registerReq.Role,
	}

	// --- STEP 2: Call UserUsecase (MongoDB) ---
	// We pass the "enriched" request to your working MongoDB usecase
	err = a.userUsecase.Register(c.Request.Context(), authUserID, &RegisterUsecaseReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "User registered successfully"})
}
