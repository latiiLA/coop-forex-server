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

func NewRequestRouter(db *mongo.Database, timeout time.Duration, group *gin.RouterGroup) {
	requestRepo := repository.NewRequestRepository(db)
	userRepo := repository.NewUserRepository(db)
	receiptRepo := repository.NewReceiptRepository(db)
	currencyRepo := repository.NewCurrencyRepository(db)
	branchRepo := repository.NewBranchRepository(db)
	activityLogRepo := repository.NewActivityLogRepository(db)
	activityLogUsecase := usecase.NewActivityLogUsecase(activityLogRepo, timeout)
	requestUsecase := usecase.NewRequestUsecase(requestRepo, userRepo, receiptRepo, currencyRepo, branchRepo, activityLogUsecase, timeout)
	fileRepo := repository.NewFileRepository(db)
	fileUsecase := usecase.NewFileUsecase(fileRepo, timeout)
	requestController := controller.NewRequestController(requestUsecase, fileUsecase)

	group.POST("/request", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:add"}), requestController.AddRequest)
	group.GET("/requests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:view"}), requestController.GetAllRequests)
	group.GET("/orgrequests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:status"}), requestController.GetAllOrgRequests)
	group.POST("/validaterequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"request:validate"}), requestController.ValidateRequest)
	group.POST("/approverequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"request:approve"}), requestController.ApproveRequest)
	group.POST("/orgauthorizerequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"request:authorize"}), requestController.AuthorizeOrgRequest)

	group.POST("/rejectrequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"request:reject"}), requestController.RejectRequest)
	group.PUT("/updaterequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"request:update"}), requestController.UpdateRequest)
	group.PATCH("/deleterequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"request:delete"}), requestController.DeleteRequest)
	group.POST("/acceptrequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"request:process"}), requestController.AcceptRequest)
	group.POST("/orgsendrequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"request:send"}), requestController.SendRequest)
	group.POST("/orgdeclinerequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"request:decline"}), requestController.DeclineOrgRequest)

	// Assign/unassign: New (branch authorizer), Authorized (validator), Validated (forex authorizer).
	group.POST("/assignrequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:send", "request:authorize", "request:validate", "request:approve"}), requestController.LockRequest)
	group.POST("/unassignrequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:send", "request:authorize", "request:validate", "request:approve"}), requestController.UnLockRequest)

	group.GET("/requests/queue", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{
		"request:view-authorized",
		"request:view-orgauthorized",
		"request:view-validated",
		"request:view-approved",
		"request:view-orgapproved",
		"request:view-accepted",
		"request:view-orgaccepted",
		"request:view-declined",
		"request:view-orgdeclined",
		"request:view-rejected",
		"request:view-orgrejected",
		"request:view-new",
		"request:view-orgdrafted",
	}), requestController.GetRequestQueue)
	group.GET("/requests/summary", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:view", "request:status"}), requestController.GetRequestStatusSummary)

	group.GET("/request/:id/report", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:generate-report"}), requestController.GenerateRequestReport)
	group.GET("/receipts/validate/:receiptNumber", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:generate-report", "request:view", "request:view-accepted"}), requestController.ValidateReceipt)
}
