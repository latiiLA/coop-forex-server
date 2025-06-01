package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
)

type RoleController interface {
	AddRole(c *gin.Context)
	// GetRoleByID(c *gin.Context)
	GetAllRoles(c *gin.Context)
}

type roleController struct{
	roleUsecase usecase.RoleUsecase
}

func NewRoleController(roleUsecase usecase.RoleUsecase) RoleController{
	return &roleController{
		roleUsecase: roleUsecase,
	}
}

func (rc *roleController) AddRole(c *gin.Context){
	userID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	var role model.Role

	err = c.ShouldBindJSON(&role)
	if err != nil{
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	if len(role.Type) < 3 {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Role name too short"})
		return
	}

	err = rc.roleUsecase.AddRole(c, userID, &role)
	if err != nil{
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Role created successfully"})
}

func (rc *roleController) GetAllRoles(c *gin.Context){
	roles, err := rc.roleUsecase.GetAllRoles(c)
	if err != nil{
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, roles)
}