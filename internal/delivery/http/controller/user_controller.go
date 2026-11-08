package controller

import (
	"errors"
	"fmt"
	"html"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/latiiLA/coop-forex-server/internal/common"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserController interface {
	// Register(c *gin.Context)
	Login(c *gin.Context)
	Logout(c *gin.Context)
	GetAllUsers(c *gin.Context)
	UpdateUser(c *gin.Context)
	SuspendUser(c *gin.Context)
	UnsuspendUser(c *gin.Context)
	GetUserStatusSummary(c *gin.Context)
	GetEligibleApprovers(c *gin.Context)
	GetEligibleBranchAuthorizers(c *gin.Context)
	IP(c *gin.Context)
	RefreshToken(c *gin.Context)
}

type userController struct {
	userUsecase usecase.UserUsecase
}

func NewUserController(userUsecase usecase.UserUsecase) UserController {
	return &userController{
		userUsecase: userUsecase,
	}
}

func (uc *userController) Login(c *gin.Context) {
	var userReq model.LoginRequestDTO

	logEntry := utils.GetLogger(c)

	err := c.ShouldBindJSON(&userReq)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid login payload")
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

	logEntry.WithField("username", userReq.Username).Debug("Login attempt")

	clientIP := c.ClientIP()

	// Sanitize
	userReq.Username = html.EscapeString(userReq.Username)
	userReq.Password = html.EscapeString(userReq.Password)

	user, err := uc.userUsecase.Login(c, userReq, clientIP)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Login failed")

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
			message = "Account has been disabled or deleted. Contact administrator"

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	logEntry.Info("Login Successful")
	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Login successful", Data: user})
}

func (uc *userController) Logout(c *gin.Context) {
	logEntry := utils.GetLogger(c)
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&body)

	accessToken := utils.GetAccessToken(c)
	if accessToken == "" {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > 7 {
			accessToken = authHeader[7:]
		}
	}

	if err := uc.userUsecase.Logout(c, authUserID, accessToken, body.RefreshToken, c.ClientIP()); err != nil {
		logEntry.WithField("error", err.Error()).Warn("logout failed")
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Logged out successfully"})
}

func (uc *userController) GetAllUsers(c *gin.Context) {
	logEntry := utils.GetLogger(c)
	users, err := uc.userUsecase.GetAllUsers(c)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("user fetch failed")
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	logEntry.Info("user fetch successful")
	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Users fetched successfully", Data: users})
}

func (uc *userController) UpdateUser(c *gin.Context) {
	logEntry := utils.GetLogger(c)
	authUserID, err := utils.GetUserID(c)

	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("user update failed")
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	// get user id from param
	userID := c.Param("id")
	if userID == "" {
		logEntry.Warn("user id not found")
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "user ID param is required"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("user id not correct id")
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	var updateData model.UpdateUserRequestDTO
	err = c.ShouldBindJSON(&updateData)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("update data binding failed")

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

	user, err := uc.userUsecase.UpdateUserByID(c, userObjID, authUserID, &updateData)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("user update failed")
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	logEntry.Info("user update successful")
	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "User updated successfully", Data: user})
}

func (uc *userController) SuspendUser(c *gin.Context) {
	logEntry := utils.GetLogger(c)
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	if err := uc.userUsecase.SuspendUser(c, authUserID, userObjID); err != nil {
		status := http.StatusInternalServerError
		message := common.MessInternalServerError
		switch {
		case errors.Is(err, common.ErrSelfSuspension):
			status = http.StatusBadRequest
			message = err.Error()
		case errors.Is(err, common.ErrUserNotFound):
			status = http.StatusNotFound
			message = err.Error()
		}
		logEntry.WithField("error", err.Error()).Warn("suspend failed")
		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "User suspended successfully"})
}

func (uc *userController) UnsuspendUser(c *gin.Context) {
	logEntry := utils.GetLogger(c)
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	if err := uc.userUsecase.UnsuspendUser(c, authUserID, userObjID); err != nil {
		status := http.StatusInternalServerError
		message := common.MessInternalServerError
		switch {
		case errors.Is(err, common.ErrInvalidUserStatus):
			status = http.StatusConflict
			message = err.Error()
		case errors.Is(err, common.ErrUserNotFound):
			status = http.StatusNotFound
			message = err.Error()
		}
		logEntry.WithField("error", err.Error()).Warn("unsuspend failed")
		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "User unsuspended successfully"})
}

func (uc *userController) GetUserStatusSummary(c *gin.Context) {
	summary, err := uc.userUsecase.GetUserStatusSummary(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "User status summary fetched successfully", Data: summary})
}

func (uc *userController) GetEligibleApprovers(c *gin.Context) {
	approvers, err := uc.userUsecase.GetEligibleApprovers(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Eligible approvers fetched successfully", Data: approvers})
}

func (uc *userController) GetEligibleBranchAuthorizers(c *gin.Context) {
	// Prefer live profile branch over JWT claim (claim can be missing/stale after profile edits).
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}
	actor, err := uc.userUsecase.GetUserByID(c, authUserID)
	if err != nil || actor == nil || actor.Profile == nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "user profile not found"})
		return
	}

	var branchID primitive.ObjectID
	if actor.Profile.BranchID != nil && !actor.Profile.BranchID.IsZero() {
		branchID = *actor.Profile.BranchID
	} else {
		branchID, err = utils.GetBranchID(c)
		if err != nil || branchID.IsZero() {
			c.JSON(http.StatusBadRequest, response.Status{
				Message: common.MessInvalidRequest,
				Error:   "your profile has no branch; assign a branch before selecting an authorizer",
			})
			return
		}
	}

	authorizers, err := uc.userUsecase.GetEligibleBranchAuthorizers(c, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Eligible branch authorizers fetched successfully", Data: authorizers})
}

func (uc *userController) IP(c *gin.Context) {
	clientIP := c.ClientIP()

	logrus.Info("request coming from", clientIP)
	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Data: map[string]string{"ip": clientIP}, Message: "client ip"})
}

func (ac *userController) RefreshToken(c *gin.Context) {
	var refreshInput model.RefreshTokenDTO
	if err := c.ShouldBindJSON(&refreshInput); err != nil {
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

	clientIP := c.ClientIP() // get client IP

	access_token, refresh_token, err := ac.userUsecase.RefreshToken(c, refreshInput, clientIP)
	if err != nil {
		logrus.Error("invalid token", err)
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: "Invalid or expired token"})
		return
	}

	tokenResponse := model.TokenResponseDTO{
		Token:        access_token,
		RefreshToken: refresh_token,
	}

	c.JSON(http.StatusOK, response.Status{Message: "Refresh token refreshed successfully", Data: tokenResponse})
}
