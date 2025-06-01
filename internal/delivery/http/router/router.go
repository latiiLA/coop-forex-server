package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/middleware"
	"go.mongodb.org/mongo-driver/mongo"
)

func RouterSetup(timeout time.Duration, db *mongo.Database, router *gin.Engine){
	publicRouter := router.Group("")
	// All public APIS
	NewPublicRouter(timeout, db, publicRouter)

	// All authenticated user routes (any logged-in user)
	authRouter := router.Group("")
	authRouter.Use(middleware.JwtAuthMiddleware(configs.JwtSecret))

	// Admin-only routes
	adminRouter := router.Group("")
	adminRouter.Use(middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRoles("admin"))

	// Branch-only routes
	branchRouter := router.Group("")
	branchRouter.Use(middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRoles("branch"))
	
	// Forex-user only routes
	forexUserRouter := router.Group("")
	forexUserRouter.Use(middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRoles("forexuser"))

	// Manager-only routes
	managerRouter := router.Group("")
	managerRouter.Use(middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRoles("manager"))

	superAdminRouter := router.Group("")
	superAdminRouter.Use(middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRoles("admin"))
	NewRoleRouter(timeout, db, superAdminRouter)
}

