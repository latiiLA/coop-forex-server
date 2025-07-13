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

func NewRequestRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	requestRepo := repository.NewRequestRepository(db)
	userRepo := repository.NewUserRepository(db)
	requestUsecase := usecase.NewRequestUsecase(requestRepo, userRepo, timeout)
	fileRepo := repository.NewFileRepository(db)
	fileUsecase := usecase.NewFileUsecase(fileRepo, timeout)
	requestController := controller.NewRequestController(requestUsecase, fileUsecase)

	group.POST("/request", middleware.JwtAuthMiddleware(configs.JwtSecret), requestController.AddRequest)
	group.GET("/requests", requestController.GetAllRequests)
	group.POST("/validaterequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), requestController.ValidateRequest)
	group.POST("/approverequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), requestController.ApproveRequest)
}
