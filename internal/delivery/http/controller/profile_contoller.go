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

type ProfileController interface {
	GetProfileByID(c *gin.Context)
	UpdateProfileByID(c *gin.Context)
}

type profileController struct {
	profileUsecase usecase.ProfileUsecase
	userUsecase    usecase.UserUsecase
}

func NewProfileController(profileUsecase usecase.ProfileUsecase, userUsecase usecase.UserUsecase) ProfileController {
	return &profileController{
		profileUsecase: profileUsecase,
		userUsecase:    userUsecase,
	}
}

func (pc *profileController) GetProfileByID(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	profileID := c.Param("id")
	if profileID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "profile id not found in param"})
		return
	}

	profileObjID, err := primitive.ObjectIDFromHex(profileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	profile, err := pc.profileUsecase.GetProfileByID(c, authUserID, profileObjID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "profile fetched successfully", Data: profile})
}

func (pc *profileController) UpdateProfileByID(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	profileID := c.Param("id")
	if profileID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "profile id not found in param"})
		return
	}

	profileObjID, err := primitive.ObjectIDFromHex(profileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	var updatedData model.Profile

	err = c.ShouldBindJSON(&updatedData)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	profile, err := pc.profileUsecase.UpdateProfileByUserID(c, authUserID, profileObjID, &updatedData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "profile updated successfully", Data: profile})
}
