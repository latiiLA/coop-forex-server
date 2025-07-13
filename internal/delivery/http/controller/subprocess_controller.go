package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubprocessController interface {
	GetAllSubprocesses(c *gin.Context)
	GetSubprocessByProcessID(c *gin.Context)
}

type subprocessController struct {
	subprocessUsecase usecase.SubprocessUsecase
}

func NewSubprocessController(subprocessUsecase usecase.SubprocessUsecase) SubprocessController {
	return &subprocessController{
		subprocessUsecase: subprocessUsecase,
	}
}

func (pc *subprocessController) GetAllSubprocesses(c *gin.Context) {
	subprocesses, err := pc.subprocessUsecase.GetAllSubprocesses(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, subprocesses)
}

func (pc *subprocessController) GetSubprocessByProcessID(c *gin.Context) {

	// get process id from param
	processID := c.Param("id")
	if processID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "process ID is required"})
		return
	}

	processObjID, err := primitive.ObjectIDFromHex(processID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	fmt.Println("processID", processObjID)

	subprocesses, err := pc.subprocessUsecase.GetSubprocessByProcessID(c, processObjID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, subprocesses)
}
