package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoleController interface {
	AddRole(c *gin.Context)
	// GetRoleByID(c *gin.Context)
	GetDeletedRoles(c *gin.Context)
	GetAllRoles(c *gin.Context)
	UpdateRole(c *gin.Context)
	DeleteRole(c *gin.Context)
}

type roleController struct {
	roleUsecase usecase.RoleUsecase
}

func NewRoleController(roleUsecase usecase.RoleUsecase) RoleController {
	return &roleController{
		roleUsecase: roleUsecase,
	}
}

func (rc *roleController) AddRole(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	var role model.Role

	err = c.ShouldBindJSON(&role)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	if len(role.Name) < 3 {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Role name too short"})
		return
	}

	err = rc.roleUsecase.AddRole(c, userID, &role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Role created successfully"})
}

func (rc *roleController) GetAllRoles(c *gin.Context) {
	roles, err := rc.roleUsecase.GetAllRoles(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, roles)
}

func (rc *roleController) UpdateRole(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	// get role id from param
	roleIDStr := c.Param("id")
	if roleIDStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "user ID is required"})
		return
	}

	roleID, err := primitive.ObjectIDFromHex(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	var roleUpdate model.UpdateRoleDTO

	err = c.ShouldBindJSON(&roleUpdate)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = rc.roleUsecase.UpdateRole(c, authUserID, roleID, roleUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "role updated successfully"})
}

func (rc *roleController) DeleteRole(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	// get role id from param
	roleIDStr := c.Param("id")
	if roleIDStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "user ID is required"})
		return
	}

	roleID, err := primitive.ObjectIDFromHex(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = rc.roleUsecase.DeleteRole(c, authUserID, roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "role deleted successfully"})
}

func (rc *roleController) GetDeletedRoles(c *gin.Context) {
	roles, err := rc.roleUsecase.GetDeletedRoles(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, roles)
}
