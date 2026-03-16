package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/common"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DepartmentController interface {
	GetAllDepartments(c *gin.Context)
	GetDepartmentsByProcessID(c *gin.Context)
}

type departmentController struct {
	departmentUsecase usecase.DepartmentUsecase
}

func NewDepartmentController(departmentUsecase usecase.DepartmentUsecase) DepartmentController {
	return &departmentController{
		departmentUsecase: departmentUsecase,
	}
}

func (dc *departmentController) GetAllDepartments(c *gin.Context) {
	departments, err := dc.departmentUsecase.GetAllDepartments(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "departments fetched successfully", Data: departments})
}

func (dc *departmentController) GetDepartmentsByProcessID(c *gin.Context) {
	// get subprocess id from param
	subprocessID := c.Param("id")
	if subprocessID == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: "Subprocess is required", Error: "subprocess ID is required"})
		return
	}

	subprocessObjID, err := primitive.ObjectIDFromHex(subprocessID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: "Invalid subprocess", Error: err.Error()})
		return
	}

	subprocesses, err := dc.departmentUsecase.GetDepartmentBySubprocessID(c, subprocessObjID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Subprocess fetched successfully", Data: subprocesses})
}
