package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/controller"
	"github.com/latiiLA/coop-forex-server/internal/repository"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewCountryRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	countryRepository := repository.NewCountryRepository(db)
	countryUsecase := usecase.NewCountryUsecase(countryRepository, timeout)
	countryController := controller.NewCountryController(countryUsecase)

	group.GET("/countries", countryController.GetAllCountry)
}
