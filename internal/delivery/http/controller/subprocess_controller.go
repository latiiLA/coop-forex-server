package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/common"
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
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Subprocesses fetched successfully", Data: subprocesses})
}

func (pc *subprocessController) GetSubprocessByProcessID(c *gin.Context) {

	// get process id from param
	processID := c.Param("id")
	if processID == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "process ID is required"})
		return
	}

	processObjID, err := primitive.ObjectIDFromHex(processID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	subprocesses, err := pc.subprocessUsecase.GetSubprocessByProcessID(c, processObjID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Subprocesses fetched successfully", Data: subprocesses})
}
