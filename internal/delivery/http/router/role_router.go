package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/controller"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/middleware"
	"github.com/latiiLA/coop-forex-server/internal/repository"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRoleRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	roleRepo := repository.NewRoleRepository(db)
	roleUsecase := usecase.NewRoleUsecase(roleRepo, timeout)
	roleController := controller.NewRoleController(roleUsecase)

	group.POST("/role", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"role:add"}), roleController.AddRole)
	group.GET("/roles", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"role:view"}), roleController.GetAllRoles)
	group.GET("/roles/deleted", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"role:view-deleted"}), roleController.GetDeletedRoles)
	group.PUT("/role/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"role:update"}), roleController.UpdateRole)
	group.PATCH("/role/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"role:delete"}), roleController.DeleteRole)
}
