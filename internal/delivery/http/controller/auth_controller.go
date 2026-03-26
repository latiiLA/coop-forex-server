package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/latiiLA/coop-forex-server/internal/common"
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

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			e := validationErrors[0]
			message := fmt.Sprintf("%s failed on %s validation", e.Field(), e.Tag())

			c.JSON(http.StatusBadRequest, response.Status{
				Message: message,
				Error:   err.Error(),
			})

			return
		}

		c.JSON(http.StatusBadRequest, response.Status{
			Message: common.MessInvalidRequest,
			Error:   err.Error(),
		})
		return
	}

	log.WithFields(log.Fields{
		"trace_id":   traceID,
		"username":   strings.ToLower(req.Username),
		"ip":         clientIP,
		"user_agent": userAgent,
	}).Info("Login attempt")

	// Call the usecase to perform authentication
	user, err := a.authUsecase.Authenticate(c, strings.ToLower(req.Username), req.Password, c.ClientIP())
	if err != nil {
		log.WithFields(log.Fields{
			"trace_id": traceID,
			"username": strings.ToLower(req.Username),
			"ip":       clientIP,
			"error":    err.Error(),
		}).Warn("Login failed")

		var (
			status  int
			message string
		)

		switch {
		case errors.Is(err, common.ErrInvalidCredentials):
			status = http.StatusUnauthorized
			message = "Invalid credentials"

		case errors.Is(err, common.ErrUserAccessRevoked):
			status = http.StatusUnauthorized
			message = "Account account status has been disabled"

		case errors.Is(err, common.ErrADUserNotFound):
			status = http.StatusForbidden
			message = "User don't have AD account"

		case errors.Is(err, common.ErrUserNotFound):
			status = http.StatusForbidden
			message = "User isn't registered for the system"

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	log.WithFields(log.Fields{
		"trace_id": traceID,
		"username": strings.ToLower(req.Username),
		"ip":       clientIP,
	}).Info("Login successful")

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Logged In Successfully", Data: user})
}

func (a *authController) Register(c *gin.Context) {
	var (
		status  int
		message string
	)

	logEntry := utils.GetLogger(c)
	authUserID, _ := utils.GetUserID(c)

	var registerReq model.RegisterRequestDTO
	if err := c.ShouldBindJSON(&registerReq); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			e := validationErrors[0]
			message := fmt.Sprintf("%s failed on %s validation", e.Field(), e.Tag())

			c.JSON(http.StatusBadRequest, response.Status{
				Message: message,
				Error:   err.Error(),
			})

			return
		}

		c.JSON(http.StatusBadRequest, response.Status{
			Message: common.MessInvalidRequest,
			Error:   err.Error(),
		})
		return
	}

	// Call AuthUsecase (LDAP) ---
	adUser, err := a.authUsecase.GetUserDetails(c.Request.Context(), strings.ToLower(registerReq.Username))
	if err != nil {
		logEntry.Warn("Get user detail from AD error: ", err)

		switch {
		case errors.Is(err, common.ErrADUserNotFound):
			status = http.StatusForbidden
			message = "User don't have AD account"

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	RegisterUsecaseReq := model.RegisterUsecaseRequestDTO{
		Username:     strings.ToLower(registerReq.Username),
		FirstName:    adUser.Profile.FirstName,
		MiddleName:   adUser.Profile.MiddleName,
		LastName:     registerReq.LastName,
		DisplayName:  adUser.Profile.DisplayName,
		Email:        adUser.Profile.Email,
		BranchID:     registerReq.BranchID,
		DepartmentID: registerReq.DepartmentID,
		Role:         registerReq.Role,
	}

	if len(registerReq.Permissions) != 0 {
		RegisterUsecaseReq.Permissions = registerReq.Permissions
	}

	err = a.userUsecase.Register(c.Request.Context(), authUserID, &RegisterUsecaseReq)
	if err != nil {
		switch {
		case errors.Is(err, common.ErrBranchOrDepartmentNotFound):
			status = http.StatusBadRequest
			message = "Branch or Department is required"

		case errors.Is(err, common.ErrUsernameAlreadyExists):

			status = http.StatusConflict
			message = "Username already has been registered"

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "User registered successfully"})
}
