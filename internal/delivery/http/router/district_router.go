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

func NewDistrictRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	districtRepo := repository.NewDistrictRepository(db)
	districtUsecase := usecase.NewDistrictUsecase(districtRepo, timeout)
	districtController := controller.NewDistrictController(districtUsecase)

	group.GET("/districts", middleware.JwtAuthMiddleware(configs.JwtSecret), districtController.GetAllDistricts)
}
