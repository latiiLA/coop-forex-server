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

func NewBranchRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	branchRepo := repository.NewBranchRepository(db)
	branchUsecase := usecase.NewBranchUsecase(branchRepo, timeout)
	branchController := controller.NewBranchController(branchUsecase)

	group.GET("/branches/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), branchController.GetBranchesByDistrictID)
	group.GET("/branches", middleware.JwtAuthMiddleware(configs.JwtSecret), branchController.GetAllBranches)
}
