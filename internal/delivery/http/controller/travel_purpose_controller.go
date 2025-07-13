package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
)

type TravelPurposeController interface {
	GetAllTravelPurposes(c *gin.Context)
}

type travelPurposeController struct {
	travelPurposeUsecase usecase.TravelPurposeUsecase
}

func NewTravelPurposeController(travelPurposeUsecase usecase.TravelPurposeUsecase) TravelPurposeController {
	return &travelPurposeController{
		travelPurposeUsecase: travelPurposeUsecase,
	}
}

func (trc *travelPurposeController) GetAllTravelPurposes(c *gin.Context) {
	travel_purposes, err := trc.travelPurposeUsecase.GetAllTravelPurposes(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, travel_purposes)
}
