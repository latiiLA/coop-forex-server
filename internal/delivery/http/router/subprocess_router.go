package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/controller"
	"github.com/latiiLA/coop-forex-server/internal/repository"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewSubprocessRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	subprocessRepo := repository.NewSubprocessRepository(db)
	subprocessUsecase := usecase.NewSubprocessUsecase(subprocessRepo, timeout)
	subprocessController := controller.NewSubprocessController(subprocessUsecase)

	// group.POST("/subprocess", processController.AddProcess)
	group.GET("/subprocesses/:id", subprocessController.GetSubprocessByProcessID)
	group.GET("/subprocesses", subprocessController.GetAllSubprocesses)
}
