package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func RouterSetup(timeout time.Duration, db *mongo.Database, router *gin.Engine) {
	publicRouter := router.Group("")
	// All public APIS
	NewPublicRouter(db, timeout, publicRouter)

	roleRouter := router.Group("")
	NewRoleRouter(db, timeout, roleRouter)
}
