package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/controller"
	"github.com/latiiLA/coop-forex-server/internal/repository"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewTravelPurposeRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	travelPurposeRepository := repository.NewTravelPurposeRepository(db)
	travelPurposeUsecase := usecase.NewTravelPurposeUsecase(travelPurposeRepository, timeout)
	travelController := controller.NewTravelPurposeController(travelPurposeUsecase)

	group.GET("/travelpurpose", travelController.GetAllTravelPurposes)
}
