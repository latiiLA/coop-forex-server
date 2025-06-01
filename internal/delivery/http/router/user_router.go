package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/controller"
	"github.com/latiiLA/coop-forex-server/internal/repository"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewUserRouter(timeout time.Duration, db *mongo.Database, group *gin.RouterGroup) {
	userRepo := repository.NewUserRepository(db, timeout)
	roleRepo := repository.NewRoleRepository(db, timeout)
	userUsecase := usecase.NewUserUsecase(userRepo, roleRepo, timeout)
	userController := controller.NewUserController(userUsecase)

	group.POST("/login", userController.Login)
	group.POST("/register", userController.Register)
}