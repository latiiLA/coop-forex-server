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

func NewProfileRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	profileRepo := repository.NewProfileRepository(db)
	userRepo := repository.NewUserRepository(db)

	profileUsecase := usecase.NewProfileUsecase(profileRepo, userRepo, timeout)

	roleRepo := repository.NewRoleRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo, roleRepo, profileRepo, timeout, db.Client())
	profileController := controller.NewProfileController(profileUsecase, userUsecase)

	group.GET("/profile/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), profileController.GetProfileByID)
	group.PUT("/profile/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), profileController.UpdateProfileByID)
}
