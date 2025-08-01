package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
)

type DistrictController interface {
	GetAllDistricts(c *gin.Context)
}

type districtController struct {
	districtUsecase usecase.DistrictUsecase
}

func NewDistrictController(districtUsecase usecase.DistrictUsecase) DistrictController {
	return &districtController{
		districtUsecase: districtUsecase,
	}
}

func (dc *districtController) GetAllDistricts(c *gin.Context) {
	districts, err := dc.districtUsecase.GetAllDistricts(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, districts)
}
