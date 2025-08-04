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
	requestUsecase := usecase.NewRequestUsecase(requestRepo, userRepo, timeout)
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

	group.GET("/authorizedrequests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"request:view-authorized"}), requestController.GetAuthorizedRequests)
	group.GET("/newrequests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{""}, []string{"request:view-new"}), requestController.GetNewOrgRequests)
	group.GET("/rejectedrequests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:view-rejected"}), requestController.GetRejectedRequests)
	group.GET("/validatedrequests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:view-validated"}), requestController.GetValidatedRequests)
	group.GET("/approvedrequests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:view-approved"}), requestController.GetApprovedRequests)
	group.GET("/acceptedrequests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:view-accepted"}), requestController.GetAcceptedRequests)
	group.GET("/declinedrequests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:view-declined"}), requestController.GetDeclinedRequests)
	group.POST("/lockrequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:lock"}), requestController.LockRequest)
	group.POST("/unlockrequest/:id", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:unlock"}), requestController.UnLockRequest)

	group.GET("/orgdraftedrequests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:view-orgdrafted"}), requestController.GetDraftedOrgRequests)
	group.GET("/orgauthorizedrequests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:view-orgauthorized"}), requestController.GetAuthorizedOrgRequests)
	group.GET("/orgapprovedrequests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:view-orgapproved"}), requestController.GetApprovedOrgRequests)
	group.GET("/orgacceptedrequests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:view-orgaccepted"}), requestController.GetAcceptedOrgRequests)
	group.GET("/orgdeclinedrequests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:view-orgdeclined"}), requestController.GetDeclinedOrgRequests)
	group.GET("/orgrejectedrequests", middleware.JwtAuthMiddleware(configs.JwtSecret), middleware.AuthorizeRolesOrPermissions([]string{}, []string{"request:view-orgrejected"}), requestController.GetRejectedOrgRequests)

}
