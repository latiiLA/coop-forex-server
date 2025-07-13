package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
)

type CurrencyController interface {
	GetAllCurrency(c *gin.Context)
}

type currencyController struct {
	currencyUsecase usecase.CurrencyUsecase
}

func NewCurrencyController(currencyUsecase usecase.CurrencyUsecase) CurrencyController {
	return &currencyController{
		currencyUsecase: currencyUsecase,
	}
}

func (cc *currencyController) GetAllCurrency(c *gin.Context) {
	currencies, err := cc.currencyUsecase.GetAllCurrency(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, currencies)
}
