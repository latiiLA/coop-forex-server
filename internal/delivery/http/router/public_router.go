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
	tokenRepo := repository.NewTokenBlacklistRepository(db)
	activityLogRepo := repository.NewActivityLogRepository(db)
	activityLogUsecase := usecase.NewActivityLogUsecase(activityLogRepo, timeout)

	middleware.SetAccessTokenBlacklist(tokenRepo)

	userUsecase := usecase.NewUserUsecase(userRepo, roleRepo, profileRepo, tokenRepo, activityLogUsecase, timeout, db.Client())
	userController := controller.NewUserController(userUsecase)
	activityLogController := controller.NewActivityLogController(activityLogUsecase)

	group.POST("/login", userController.Login)
	group.POST("/logout", middleware.JwtAuthMiddleware(configs.JwtSecret), userController.Logout)
	group.GET("/users", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{"superadmin"}, []string{"user:view"}), userController.GetAllUsers)
	group.GET("/users/summary", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"user:view"}), userController.GetUserStatusSummary)
	group.GET("/users/approvers", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:approve"}), userController.GetEligibleApprovers)
	group.GET("/users/branch-authorizers", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:add", "request:send", "request:authorize", "request:update"}), userController.GetEligibleBranchAuthorizers)
	group.PUT("/users/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{"admin"}, []string{"user:update"}), userController.UpdateUser)
	group.POST("/users/:id/suspend", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"user:update"}), userController.SuspendUser)
	group.POST("/users/:id/unsuspend", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"user:update"}), userController.UnsuspendUser)
	group.GET("/ip", middleware.JwtAuthMiddleware(configs.JwtSecret), userController.IP)
	group.POST("/refreshtoken", userController.RefreshToken)
	group.GET("/activity-logs", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{"superadmin"}, []string{"user:view", "request:view", "request:status", "analytics:view"}), activityLogController.GetRecent)

	// LDAP configuration
	authUsecase := usecase.NewLDAPAuthUsecase(userRepo, configs.LDAPHost, configs.LDAPPort, configs.LDAPBaseDN, configs.LDAPBindUser, configs.LDAPBindPassword, "sAMAccountName", configs.LDAPUseTLS, configs.LDAPTLSInsecure, timeout) // "uid" for testing using docker test setup - for correct AD setup sAMAccountName
	authController := controller.NewAuthController(authUsecase, userUsecase, activityLogUsecase)
	group.POST("/login-ldap", authController.Login)
	group.POST("/register", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"user:add"}), authController.Register)
}
