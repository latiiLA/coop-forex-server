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

func NewProcessRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	processRepo := repository.NewProcessRepository(db)
	processUsecase := usecase.NewProcessUsecase(processRepo, timeout)
	processController := controller.NewProcessController(processUsecase)

	// group.POST("/process", processController.AddProcess)
	group.GET("/processes", middleware.JwtAuthMiddleware(configs.JwtSecret), processController.GetAllProcesses)
}
