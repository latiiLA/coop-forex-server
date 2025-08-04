package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BranchController interface {
	GetAllBranches(c *gin.Context)
	GetBranchesByDistrictID(c *gin.Context)
}

type branchController struct {
	branchUsecase usecase.BranchUsecase
}

func NewBranchController(branchUsecase usecase.BranchUsecase) BranchController {
	return &branchController{
		branchUsecase: branchUsecase,
	}
}

func (bc *branchController) GetAllBranches(c *gin.Context) {
	branches, err := bc.branchUsecase.GetAllBranches(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, branches)
}

func (bc *branchController) GetBranchesByDistrictID(c *gin.Context) {

	// get process id from param
	districtIDStr := c.Param("id")
	if districtIDStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "process ID is required"})
		return
	}

	districtID, err := primitive.ObjectIDFromHex(districtIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	fmt.Println("districtID", districtID)

	branches, err := bc.branchUsecase.GetBranchesByDistrictID(c, districtID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, branches)
}
