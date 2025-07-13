package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
)

type ProcessController interface {
	GetAllProcesses(c *gin.Context)
}

type processController struct {
	processUsecase usecase.ProcessUsecase
}

func NewProcessController(processUsecase usecase.ProcessUsecase) ProcessController {
	return &processController{
		processUsecase: processUsecase,
	}
}

func (pc *processController) GetAllProcesses(c *gin.Context) {
	processes, err := pc.processUsecase.GetAllProcesses(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, processes)
}
