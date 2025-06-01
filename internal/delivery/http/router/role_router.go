package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/controller"
	"github.com/latiiLA/coop-forex-server/internal/repository"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRoleRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	roleRepo := repository.NewRoleRepository(db, timeout)
	roleUsecase := usecase.NewRoleUsecase(roleRepo, timeout)
	roleController := controller.NewRoleController(roleUsecase)

	group.POST("/role", roleController.AddRole)
	group.GET("/role", roleController.GetAllRoles)
}
