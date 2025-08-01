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

func NewCurrencyRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	currencyRepository := repository.NewCurrencyRepository(db)
	currencyUsecase := usecase.NewCurrencyUsecase(currencyRepository, timeout)
	currencyController := controller.NewCurrencyController(currencyUsecase)

	group.GET("/currencies", middleware.JwtAuthMiddleware(configs.JwtSecret), currencyController.GetAllCurrency)
}
