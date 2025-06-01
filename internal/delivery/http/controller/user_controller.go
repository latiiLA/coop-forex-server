package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
)

type UserController interface{
	Register(c *gin.Context)
	Login(c *gin.Context)
}

type userController struct {
	userUsecase usecase.UserUsecase
}

func NewUserController(userUsecase usecase.UserUsecase) UserController{
	return &userController{
		userUsecase: userUsecase,
	}
}

func (uc *userController) Register(c *gin.Context){
	var user model.User

	err := c.ShouldBindJSON(&user)
	if err != nil{
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = uc.userUsecase.Register(c, &user)
	if err != nil{
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "User created successfully"})
}

func (uc *userController) Login(c *gin.Context){
	var userReq model.LoginRequestDTO

	err := c.ShouldBindJSON(&userReq)
	if err != nil{
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}
	
	user, accessToken, err := uc.userUsecase.Login(c, userReq)
	if err != nil{
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": accessToken, "user": user})
}