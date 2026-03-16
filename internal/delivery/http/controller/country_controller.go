package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/common"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
)

type CountryController interface {
	GetAllCountry(c *gin.Context)
}

type countryController struct {
	countryUsecase usecase.CountryUsecase
}

func NewCountryController(countryUsecase usecase.CountryUsecase) CountryController {
	return &countryController{
		countryUsecase: countryUsecase,
	}
}

func (cc *countryController) GetAllCountry(c *gin.Context) {
	countries, err := cc.countryUsecase.GetAllCountry(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Countries fetched successfully", Data: countries})
}
