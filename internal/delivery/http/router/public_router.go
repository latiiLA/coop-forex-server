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

func NewPublicRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo, roleRepo, profileRepo, timeout, db.Client())
	userController := controller.NewUserController(userUsecase)

	// middleware.AuthorizeRoles("admin")

	//  middleware.JwtAuthMiddleware(configs.JwtSecret)

	group.POST("/login", userController.Login)
	group.POST("/register", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"user:add"}), userController.Register)
	group.GET("/users", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{"superadmin"}, []string{"user:view"}), userController.GetAllUsers)
	group.PUT("/users/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{"admin"}, []string{"user:update"}), userController.UpdateUser)
	group.GET("/ip", middleware.JwtAuthMiddleware(configs.JwtSecret), userController.IP)

	// group.PATCH("/users", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRoles("admin"), userController.DeleteUser)

	// LDAP configuration

	// authUsecase := usecase.NewLDAPAuthUsecase(configs.LDAPHost, configs.LDAPPort, configs.LDAPBaseDN, configs.LDAPBindUser, configs.LDAPBindPassword)
	// authController := controller.NewAuthController(authUsecase)
	// group.POST("/login", authController.Login)
	// group.POST("/register", userController.Register)
}
