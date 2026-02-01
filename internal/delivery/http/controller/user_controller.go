package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	GetAllUsers(c *gin.Context)
	UpdateUser(c *gin.Context)
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

// func (uc *userController) Register(c *gin.Context) {
// 	logEntry := utils.GetLogger(c)
// 	authUserID, err := utils.GetUserID(c)
// 	if err != nil {
// 		logEntry.WithField("error", err.Error()).Warn("user id fetch failed")
// 		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
// 		return
// 	}

// 	var registerReq model.RegisterRequestDTO

// 	err = c.ShouldBindJSON(&registerReq)
// 	if err != nil {
// 		logEntry.WithField("error", err.Error()).Warn("register data binding failed")
// 		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
// 		return
// 	}

// 	err = uc.userUsecase.Register(c, authUserID, &registerReq)
// 	if err != nil {
// 		logEntry.WithField("error", err.Error()).Warn("register failed")
// 		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
// 		return
// 	}

// 	logEntry.Info("user created successfully")
// 	c.JSON(http.StatusOK, response.SuccessResponse{Message: "User created successfully"})
// }

func (uc *userController) Login(c *gin.Context) {
	var userReq model.LoginRequestDTO

	logEntry := utils.GetLogger(c)

	err := c.ShouldBindJSON(&userReq)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid login payload")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	logEntry.WithField("username", userReq.Username).Debug("Login attempt")

	clientIP := c.ClientIP()

	user, err := uc.userUsecase.Login(c, userReq, clientIP)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Login failed")
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	logEntry.Info("Login Successful")
	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Login successful", Data: user})
}

func (uc *userController) GetAllUsers(c *gin.Context) {
	logEntry := utils.GetLogger(c)
	users, err := uc.userUsecase.GetAllUsers(c)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("user fetch failed")
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	logEntry.Info("user fetch successful")
	c.JSON(http.StatusOK, users)
}

func (uc *userController) UpdateUser(c *gin.Context) {
	logEntry := utils.GetLogger(c)
	authUserID, err := utils.GetUserID(c)

	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("user update failed")
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	// get user id from param
	userID := c.Param("id")
	if userID == "" {
		logEntry.Warn("user id not found")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "user ID is required"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("user id not correct id")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	var updateData model.UpdateUserRequestDTO
	err = c.ShouldBindJSON(&updateData)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("update data binding failed")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	user, err := uc.userUsecase.UpdateUserByID(c, userObjID, authUserID, &updateData)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("user update failed")
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	logEntry.Info("user update successful")
	c.JSON(http.StatusOK, response.SuccessResponse{Message: "user update successful", Data: user})
}

func (uc *userController) IP(c *gin.Context) {
	clientIP := c.ClientIP()

	logrus.Info("request coming from", clientIP)
	c.JSON(http.StatusOK, response.SuccessResponse{Data: map[string]string{"ip": clientIP}, Message: "client ip"})
}

func (ac *userController) RefreshToken(ctx *gin.Context) {
	var refreshInput model.RefreshTokenDTO
	if err := ctx.ShouldBindJSON(&refreshInput); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	clientIP := ctx.ClientIP() // get client IP

	access_token, refresh_token, err := ac.userUsecase.RefreshToken(ctx, refreshInput, clientIP)
	if err != nil {
		logrus.Error("invalid token", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	response := model.TokenResponseDTO{
		Token:        access_token,
		RefreshToken: refresh_token,
	}

	ctx.JSON(http.StatusOK, response)
}
