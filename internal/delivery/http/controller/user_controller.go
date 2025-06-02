package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	GetAllUsers(c *gin.Context)
	UpdateUser(c *gin.Context)
}

type userController struct {
	userUsecase usecase.UserUsecase
}

func NewUserController(userUsecase usecase.UserUsecase) UserController {
	return &userController{
		userUsecase: userUsecase,
	}
}

func (uc *userController) Register(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	var registerReq model.RegisterRequestDTO

	err = c.ShouldBindJSON(&registerReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = uc.userUsecase.Register(c, authUserID, &registerReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "User created successfully"})
}

func (uc *userController) Login(c *gin.Context) {
	var userReq model.LoginRequestDTO

	err := c.ShouldBindJSON(&userReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	user, err := uc.userUsecase.Login(c, userReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Login successful", Data: user})
}

func (uc *userController) GetAllUsers(c *gin.Context) {
	users, err := uc.userUsecase.GetAllUsers(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (uc *userController) UpdateUser(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)

	if err != nil {
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
			return
		}
	}

	// get user id from param
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "user ID is required"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	var updateData model.UpdateRequestDTO
	err = c.ShouldBindJSON(&updateData)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	user, err := uc.userUsecase.UpdateUserByID(c, userObjID, authUserID, &updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "user updated successful", Data: user})
}
