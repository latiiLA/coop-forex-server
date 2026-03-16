package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/latiiLA/coop-forex-server/internal/common"
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
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnathorized, Error: err.Error()})
		return
	}

	var role model.CreateRoleDTO

	err = c.ShouldBindJSON(&role)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			e := validationErrors[0]
			message := fmt.Sprintf("%s failed on %s validation", e.Field(), e.Tag())

			c.JSON(http.StatusBadRequest, response.Status{
				Message: message,
				Error:   err.Error(),
			})

			return
		}

		c.JSON(http.StatusBadRequest, response.Status{
			Message: common.MessInvalidRequest,
			Error:   err.Error(),
		})
		return
	}

	if len(role.Name) < 3 {
		c.JSON(http.StatusBadRequest, response.Status{Message: "Role name too short", Error: "Role name failed on character length"})
		return
	}

	err = rc.roleUsecase.AddRole(c, userID, &role)
	if err != nil {
		var (
			status  int
			message string
		)

		switch {
		case errors.Is(err, common.ErrRoleNameNotAllowed):
			status = http.StatusForbidden
			message = "Role name is not allowed"

		case errors.Is(err, common.ErrRoleNameAlreadyExists):
			status = http.StatusConflict
			message = "Role name already exists"

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Role created successfully"})
}

func (rc *roleController) GetAllRoles(c *gin.Context) {
	roles, err := rc.roleUsecase.GetAllRoles(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Roles fetched successfully", Data: roles})
}

func (rc *roleController) UpdateRole(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	// get role id from param
	roleIDStr := c.Param("id")
	if roleIDStr == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "role ID is missing from the request"})
		return
	}

	roleID, err := primitive.ObjectIDFromHex(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	var roleUpdate model.UpdateRoleDTO

	err = c.ShouldBindJSON(&roleUpdate)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			e := validationErrors[0]
			message := fmt.Sprintf("%s failed on %s validation", e.Field(), e.Tag())

			c.JSON(http.StatusBadRequest, response.Status{
				Message: message,
				Error:   err.Error(),
			})

			return
		}

		c.JSON(http.StatusBadRequest, response.Status{
			Message: common.MessInvalidRequest,
			Error:   err.Error(),
		})
		return
	}

	err = rc.roleUsecase.UpdateRole(c, authUserID, roleID, roleUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Role updated successfully"})
}

func (rc *roleController) DeleteRole(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnathorized, Error: err.Error()})
		return
	}

	// get role id from param
	roleIDStr := c.Param("id")
	if roleIDStr == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "Role ID is required"})
		return
	}

	roleID, err := primitive.ObjectIDFromHex(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	err = rc.roleUsecase.DeleteRole(c, authUserID, roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Role deleted successfully"})
}

func (rc *roleController) GetDeletedRoles(c *gin.Context) {
	roles, err := rc.roleUsecase.GetDeletedRoles(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Deleted roles fetched successfully", Data: roles})
}
