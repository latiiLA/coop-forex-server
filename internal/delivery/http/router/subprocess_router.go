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

func NewSubprocessRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	subprocessRepo := repository.NewSubprocessRepository(db)
	subprocessUsecase := usecase.NewSubprocessUsecase(subprocessRepo, timeout)
	subprocessController := controller.NewSubprocessController(subprocessUsecase)

	// group.POST("/subprocess", processController.AddProcess)
	group.GET("/subprocesses/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), subprocessController.GetSubprocessByProcessID)
	group.GET("/subprocesses", middleware.JwtAuthMiddleware(configs.JwtSecret), subprocessController.GetAllSubprocesses)
}
