package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
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

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Call the usecase to perform authentication
	if err := a.usecase.Authenticate(req.Username, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Login Successful"})
}
