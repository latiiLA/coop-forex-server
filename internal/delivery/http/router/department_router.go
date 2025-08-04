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

func NewDepartmentRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	departmentRepo := repository.NewDepartmentRepository(db)
	departmentUsecase := usecase.NewDepartmentUsecase(departmentRepo, timeout)
	departmentController := controller.NewDepartmentController(departmentUsecase)

	// group.POST("/department", departmentController.AddDepartment)
	group.GET("/departments/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), departmentController.GetDepartmentsByProcessID)
	group.GET("/departments", middleware.JwtAuthMiddleware(configs.JwtSecret), departmentController.GetAllDepartments)
}
