package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/controller"
	"github.com/latiiLA/coop-forex-server/internal/repository"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewPublicRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	userRepo := repository.NewUserRepository(db, timeout)
	roleRepo := repository.NewRoleRepository(db, timeout)
	profileRepo := repository.NewProfileRepository(db, timeout)
	userUsecase := usecase.NewUserUsecase(userRepo, roleRepo, profileRepo, timeout)
	userController := controller.NewUserController(userUsecase)

	group.POST("/login", userController.Login)
	group.POST("/register", userController.Register)
}