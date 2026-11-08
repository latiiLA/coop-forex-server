package controller

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/common"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RejectRequestPayload struct {
	RejectionReason string `json:"rejection_reason"`
}

type RequestController interface {
	AddRequest(c *gin.Context)
	GetAllRequests(c *gin.Context)
	ValidateRequest(c *gin.Context)
	UpdateRequest(c *gin.Context)
	ApproveRequest(c *gin.Context)
	GetAllOrgRequests(c *gin.Context)
	GetAuthorizedRequests(c *gin.Context)
	GetNewOrgRequests(c *gin.Context)
	AuthorizeOrgRequest(c *gin.Context)
	DeleteRequest(c *gin.Context)
	RejectRequest(c *gin.Context)
	AcceptRequest(c *gin.Context)
	SendRequest(c *gin.Context)
	DeclineOrgRequest(c *gin.Context)
	LockRequest(c *gin.Context)
	UnLockRequest(c *gin.Context)
	GetRejectedOrgRequests(c *gin.Context)
	GetRejectedRequests(c *gin.Context)
	GetValidatedRequests(c *gin.Context)
	GetApprovedRequests(c *gin.Context)
	GetAcceptedRequests(c *gin.Context)
	GetDeclinedRequests(c *gin.Context)
	GetApprovedOrgRequests(c *gin.Context)
	GetAcceptedOrgRequests(c *gin.Context)
	GetDeclinedOrgRequests(c *gin.Context)
	GetAuthorizedOrgRequests(c *gin.Context)
	GetDraftedOrgRequests(c *gin.Context)
	GetRequestQueue(c *gin.Context)
	GetRequestStatusSummary(c *gin.Context)
	GenerateRequestReport(c *gin.Context)
	ValidateReceipt(c *gin.Context)
}

type requestController struct {
	requestUsecase usecase.RequestUsecase
	fileUsecase    usecase.FileUsecase
}

func NewRequestController(requestUsecase usecase.RequestUsecase, fileUsecase usecase.FileUsecase) RequestController {
	return &requestController{
		requestUsecase: requestUsecase,
		fileUsecase:    fileUsecase,
	}
}

// queueBankPerm / queueOrgPerm map status → required view permission.
var queueBankPerm = map[string]string{
	"Authorized": "request:view-authorized",
	"Validated":  "request:view-validated",
	"Approved":   "request:view-approved",
	"Accepted":   "request:view-accepted",
	"Declined":   "request:view-declined",
	"Rejected":   "request:view-rejected",
}

var queueOrgPerm = map[string]string{
	"Drafted":    "request:view-orgdrafted",
	"New":        "request:view-new",
	"Authorized": "request:view-orgauthorized",
	"Approved":   "request:view-orgapproved",
	"Accepted":   "request:view-orgaccepted",
	"Declined":   "request:view-orgdeclined",
	"Rejected":   "request:view-orgrejected",
}

func hasAnyPermission(actorPerms []string, required string) bool {
	for _, p := range actorPerms {
		if p == required {
			return true
		}
	}
	return false
}

// resolveQueueScopeFromPerms picks bank if allowed, else org. Empty = no access.
func resolveQueueScopeFromPerms(status string, actorPerms []string) string {
	if bankPerm, ok := queueBankPerm[status]; ok && hasAnyPermission(actorPerms, bankPerm) {
		return "bank"
	}
	if orgPerm, ok := queueOrgPerm[status]; ok && hasAnyPermission(actorPerms, orgPerm) {
		return "org"
	}
	return ""
}

func resolveOrgFromContext(c *gin.Context) (orgKey string, orgID primitive.ObjectID) {
	departmentID, _ := utils.GetDepartmentID(c)
	branchID, _ := utils.GetBranchID(c)
	if !branchID.IsZero() {
		return "branch_id", branchID
	}
	if !departmentID.IsZero() {
		return "department_id", departmentID
	}
	return "", primitive.NilObjectID
}

func (rc *requestController) GetRequestQueue(c *gin.Context) {
	status := c.Query("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, response.Status{
			Message: common.MessInvalidRequest,
			Error:   "status is required",
		})
		return
	}

	actorPerms := utils.GetPermissions(c)
	scope := resolveQueueScopeFromPerms(status, actorPerms)
	if scope == "" {
		// Unknown status, or caller lacks both bank and org view perms for it.
		if _, bankOK := queueBankPerm[status]; !bankOK {
			if _, orgOK := queueOrgPerm[status]; !orgOK {
				c.JSON(http.StatusBadRequest, response.Status{
					Message: common.MessInvalidRequest,
					Error:   "invalid status",
				})
				return
			}
		}
		c.JSON(http.StatusForbidden, response.Status{Message: common.MessUnauthorized, Error: "permission denied"})
		return
	}

	viewerID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	orgKey, orgID := resolveOrgFromContext(c)
	pageQuery := parsePageQuery(c, 25)

	result, err := rc.requestUsecase.GetRequestQueue(c, status, scope, orgKey, orgID, viewerID, pageQuery)
	if err != nil {
		if errors.Is(err, common.ErrInvalidRequest) {
			c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	if result == nil {
		result = &model.PaginatedRequests{Items: []model.Request{}}
	}
	c.JSON(http.StatusOK, response.Status{
		IsSuccessful: true,
		Message:      "Request queue fetched successfully",
		Data:         result.Items,
		Meta: response.PaginationMeta{
			Total: result.Total,
			Page:  result.Page,
			Limit: result.Limit,
			Scope: scope,
		},
	})
}

func parsePageQuery(c *gin.Context, defaultLimit int) model.PageQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || c.Query("limit") == "" {
		limit = defaultLimit
	}
	pq := model.NormalizePageQuery(page, limit)

	if from := c.Query("date_from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			pq.DateFrom = &t
		}
	}
	if to := c.Query("date_to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			pq.DateTo = &t
		}
	}

	return pq
}

func writePaginatedRequests(c *gin.Context, message string, result *model.PaginatedRequests) {
	if result == nil {
		result = &model.PaginatedRequests{Items: []model.Request{}}
	}
	c.JSON(http.StatusOK, response.Status{
		IsSuccessful: true,
		Message:      message,
		Data:         result.Items,
		Meta: response.PaginationMeta{
			Total: result.Total,
			Page:  result.Page,
			Limit: result.Limit,
		},
	})
}

func (rc *requestController) AddRequest(c *gin.Context) {
	logEntry := utils.GetLogger(c)
	userID, err := utils.GetUserID(c)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid user ID")
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	var request model.RequestDTO

	err = c.ShouldBind(&request)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid request payload")

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			e := validationErrors[0]
			message := fmt.Sprintf("%s failed on %s validation", e.Field(), e.Tag())

			c.JSON(http.StatusBadRequest, response.Status{
				Message: message,
				Error:   err.Error(),
			})

			return
		}

		c.JSON(http.StatusBadRequest, response.Status{
			Message: common.MessInvalidRequest,
			Error:   err.Error(),
		})
		return
	}

	logEntry.WithField("request", request).Debug("Request payload received")

	// Convert string fields to ObjectID
	travelPurposeObjID, err := primitive.ObjectIDFromHex(request.TravelPurposeID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid travel purpose ID")
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: "Invalid travel purpose ID"})
		return
	}

	travelCountryObjID, err := primitive.ObjectIDFromHex(request.TravelCountryID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid travel country ID")
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: "Invalid travel country ID"})
		return
	}

	accountCurrencyObjID, err := primitive.ObjectIDFromHex(request.AccountCurrencyID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid currency account ID")
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: "Invalid currency account ID"})
		return
	}

	fcyRequestedObjID, err := primitive.ObjectIDFromHex(request.FcyRequestedID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid currency ID")
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: "Invalid Currency ID"})
		return
	}

	deposit, err := strconv.ParseFloat(request.AverageDeposit, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid average deposit value")
		c.JSON(http.StatusBadRequest, response.Status{Message: "Invalid average deposit value", Error: "Invalid average deposit value"})
		return
	}

	totalFcy, err := strconv.ParseFloat(request.TotalFcyGenerated, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid total fcy performance value")
		c.JSON(http.StatusBadRequest, response.Status{Message: "Invalid total fcy performance", Error: "Invalid total fcy performance value"})
		return
	}

	currentFcy, err := strconv.ParseFloat(request.CurrentFcyPerformance, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid current fcy performance value")
		c.JSON(http.StatusBadRequest, response.Status{Message: "Invalid current fcy performance value", Error: "Invalid current fcy performance value"})
		return
	}

	fcyAmount, err := strconv.ParseFloat(request.FcyRequestedAmount, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid fcy amount value")
		c.JSON(http.StatusBadRequest, response.Status{Message: "Invalid fcy amount value", Error: "Invalid fcy amount value"})
		return
	}

	uploadPath := configs.FileUploadPath
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		logEntry.WithField("error", err.Error()).Warn("Unable to create a file upload folder")
		os.MkdirAll(uploadPath, os.ModePerm)
	}

	attachments, uploadErr := rc.uploadRequestAttachments(c)
	if uploadErr != nil {
		logEntry.WithField("error", uploadErr.Error()).Warn("Failed to upload request attachment")
		status := http.StatusInternalServerError
		message := common.MessInternalServerError
		if errors.Is(uploadErr, common.ErrInvalidUpload) {
			status = http.StatusBadRequest
			message = "Invalid upload file (max 5MB; pdf/jpg/png only)"
		}
		c.JSON(status, response.Status{
			Message: message,
			Error:   uploadErr.Error(),
		})
		return
	}

	if len(request.AccountsToDeduct) == 0 {
		logEntry.Warn("Account to deduct is required")
		c.JSON(http.StatusBadRequest, response.Status{Message: "Accounts to deduct is required", Error: "accounts_to_deduct is required"})
		return
	}

	if fcyAmount <= 0 {
		c.JSON(http.StatusBadRequest, response.Status{
			Message: "FCY requested amount must be greater than 0",
			Error:   "fcy_requested_amount must be greater than 0",
		})
		return
	}

	authorizerObjID, err := primitive.ObjectIDFromHex(request.AuthorizerID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid authorizer ID")
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: "Invalid authorizer ID"})
		return
	}

	fcyRequest := model.Request{
		ApplicantName:          request.ApplicantName,
		ApplicantAccountNumber: request.ApplicantAccountNumber,
		AverageDeposit:         deposit,
		TotalFcyGenerated:      totalFcy,
		CurrentFcyPerformance:  currentFcy,
		FcyRequestedAmount:     fcyAmount,
		AccountsToDeduct:       request.AccountsToDeduct,
		FcyAcceptanceMode:      request.FcyAcceptanceMode,
		CardAssociatedAccount:  &request.CardAssociatedAccount,
		BranchRecommendation:   &request.BranchRecommendation,
		TravelPurposeID:        travelPurposeObjID,
		TravelCountryID:        travelCountryObjID,
		RequestingAs:           request.RequestingAs,
		AccountCurrencyID:      accountCurrencyObjID,
		FcyRequestedID:         fcyRequestedObjID,
		AssignedTo:             &authorizerObjID,

		PassportAttachment:           attachments.PassportID,
		TicketAttachment:             attachments.TicketID,
		VisaAttachment:               attachments.VisaID,
		BusinessLicenseAttachment:    attachments.BusinessLicenseID,
		EducationLoaAttachment:       attachments.EducationID,
		HealthLetterAttachment:       attachments.HealthLetterID,
		BusinessSupportingAttachment: attachments.SupportingLetterID,
	}

	err = rc.requestUsecase.AddRequest(c, userID, &fcyRequest)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Failed to add the request")

		var (
			status  int
			message string
		)

		switch {
		case errors.Is(err, common.ErrUnauthorized):
			status = http.StatusUnauthorized
			message = common.MessUnauthorized

		case errors.Is(err, common.ErrAccountCurrencyMustBeETB):
			status = http.StatusBadRequest
			message = "Account currency must be ETB"

		case errors.Is(err, common.ErrLocalCurrencyNotAllowed):
			status = http.StatusBadRequest
			message = "Requested currency cannot be ETB"

		case errors.Is(err, common.ErrInvalidCurrency):
			status = http.StatusBadRequest
			message = "Invalid currency"

		case errors.Is(err, common.ErrAssigneeNotEligible):
			status = http.StatusBadRequest
			message = "Selected authorizer is not eligible for this branch"

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError

		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	logEntry.Info("Request created successfully")
	c.JSON(http.StatusOK, response.Status{
		IsSuccessful: true,
		Message:      "Fcy Request drafted successfully",
		Data: gin.H{
			"_id":          fcyRequest.ID.Hex(),
			"request_code": fcyRequest.RequestCode,
			"status":       string(fcyRequest.RequestStatus),
		},
	})
}

func (rc *requestController) GetAllRequests(c *gin.Context) {
	pageQuery := parsePageQuery(c, 25)
	result, err := rc.requestUsecase.GetAllRequests(c, pageQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	writePaginatedRequests(c, "Requests fetched Successfully", result)
}

func (rc *requestController) ValidateRequest(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	var request model.RequestValidationDTO

	err = c.ShouldBind(&request)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			e := validationErrors[0]
			message := fmt.Sprintf("%s failed on %s validation", e.Field(), e.Tag())

			c.JSON(http.StatusBadRequest, response.Status{
				Message: message,
				Error:   err.Error(),
			})

			return
		}

		c.JSON(http.StatusBadRequest, response.Status{
			Message: common.MessInvalidRequest,
			Error:   err.Error(),
		})
		return
	}

	// get user id from param
	requestID := c.Param("id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "user ID is required"})
		return
	}

	requestObjID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	validatedCurrencyObjID, err := primitive.ObjectIDFromHex(request.ValidatedAccountCurrencyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	err = rc.requestUsecase.ValidateRequest(c, userID, requestObjID, validatedCurrencyObjID, &request)
	if err != nil {
		var (
			status  int
			message string
		)

		switch {
		case errors.Is(err, common.ErrRequestNotFound):
			status = http.StatusNotFound
			message = common.MessRequestNotFound

		case errors.Is(err, common.ErrRequestConflict):
			status = http.StatusConflict
			message = common.MessRequestConflict

		case errors.Is(err, common.ErrRequestIsLocked):
			status = http.StatusConflict
			message = common.MessRequestLocked

		case errors.Is(err, common.ErrRequestNotAssigned):
			status = http.StatusConflict
			message = common.MessRequestNotAssigned

		case errors.Is(err, common.ErrInvalidRequestTransition):
			status = http.StatusConflict
			message = "Invalid request status for this action"

		case errors.Is(err, common.ErrCannotActOnOwnRequest):
			status = http.StatusForbidden
			message = common.MessCannotActOnOwnRequest

		case errors.Is(err, common.ErrAccountCurrencyMustBeETB):
			status = http.StatusBadRequest
			message = "Validated account currency must be ETB"

		case errors.Is(err, common.ErrInvalidCurrency):
			status = http.StatusBadRequest
			message = "Invalid currency"

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Request validated successfully"})
}

func (rc *requestController) UpdateRequest(c *gin.Context) {
	logEntry := utils.GetLogger(c)
	userID, err := utils.GetUserID(c)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid user ID")
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	// get user id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	var request model.UpdateRequestDTO

	err = c.ShouldBind(&request)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid request payload")

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			e := validationErrors[0]
			message := fmt.Sprintf("%s failed on %s validation", e.Field(), e.Tag())

			c.JSON(http.StatusBadRequest, response.Status{
				Message: message,
				Error:   err.Error(),
			})

			return
		}

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			e := validationErrors[0]
			message := fmt.Sprintf("%s failed on %s validation", e.Field(), e.Tag())

			c.JSON(http.StatusBadRequest, response.Status{
				Message: message,
				Error:   err.Error(),
			})

			return
		}

		c.JSON(http.StatusBadRequest, response.Status{
			Message: common.MessInvalidRequest,
			Error:   err.Error(),
		})
		return
	}

	logEntry.WithField("request", request).Debug("Request payload received")

	form, err := c.MultipartForm()
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid multipart form data")
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "Invalid multipart form data"})
		return
	}

	request.AccountsToDeduct = form.Value["accounts_to_deduct"]

	// Convert string fields to ObjectID
	travelPurposeObjID, err := primitive.ObjectIDFromHex(request.TravelPurposeID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid travel purpose ID")
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: "Invalid travel purpose ID"})
		return
	}

	travelCountryObjID, err := primitive.ObjectIDFromHex(request.TravelCountryID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid travel country ID")
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: "Invalid travel country ID"})
		return
	}

	accountCurrencyObjID, err := primitive.ObjectIDFromHex(request.AccountCurrencyID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid currency account ID")
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: "Invalid currency account ID"})
		return
	}

	fcyRequestedObjID, err := primitive.ObjectIDFromHex(request.FcyRequestedID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid currency ID")
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: "Invalid Currency ID"})
		return
	}

	deposit, err := strconv.ParseFloat(request.AverageDeposit, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid average deposit value")
		c.JSON(http.StatusBadRequest, response.Status{Message: "Invalid average deposit value", Error: "Invalid average deposit value"})
		return
	}

	totalFcy, err := strconv.ParseFloat(request.TotalFcyGenerated, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid total fcy performance value")
		c.JSON(http.StatusBadRequest, response.Status{Message: "Invalid total fcy performance value", Error: "Invalid total fcy performance value"})
		return
	}

	currentFcy, err := strconv.ParseFloat(request.CurrentFcyPerformance, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid current fcy performance value")
		c.JSON(http.StatusBadRequest, response.Status{Message: "Invalid current fcy performance value", Error: "Invalid current fcy performance value"})
		return
	}

	fcyAmount, err := strconv.ParseFloat(request.FcyRequestedAmount, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid fcy amount value")
		c.JSON(http.StatusBadRequest, response.Status{Message: "Invalid fcy amount value", Error: "Invalid fcy amount value"})
		return
	}

	uploadPath := configs.FileUploadPath
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		logEntry.WithField("error", err.Error()).Warn("Unable to create a file upload folder")
		os.MkdirAll(uploadPath, os.ModePerm)
	}

	attachments, uploadErr := rc.uploadRequestAttachments(c)
	if uploadErr != nil {
		logEntry.WithField("error", uploadErr.Error()).Warn("Failed to upload request attachment")
		status := http.StatusInternalServerError
		message := common.MessInternalServerError
		if errors.Is(uploadErr, common.ErrInvalidUpload) {
			status = http.StatusBadRequest
			message = "Invalid upload file (max 5MB; pdf/jpg/png only)"
		}
		c.JSON(status, response.Status{
			Message: message,
			Error:   uploadErr.Error(),
		})
		return
	}

	if len(request.AccountsToDeduct) == 0 {
		logEntry.Warn("Account to deduct is required")
		c.JSON(http.StatusBadRequest, response.Status{Message: "accounts_to_deduct is required", Error: "accounts_to_deduct is required"})
		return
	}

	if fcyAmount <= 0 {
		c.JSON(http.StatusBadRequest, response.Status{
			Message: "FCY requested amount must be greater than 0",
			Error:   "fcy_requested_amount must be greater than 0",
		})
		return
	}

	authorizerObjID, err := primitive.ObjectIDFromHex(request.AuthorizerID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid authorizer ID")
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: "Invalid authorizer ID"})
		return
	}

	fcyRequest := model.Request{
		ApplicantName:          request.ApplicantName,
		ApplicantAccountNumber: request.ApplicantAccountNumber,
		AverageDeposit:         deposit,
		TotalFcyGenerated:      totalFcy,
		CurrentFcyPerformance:  currentFcy,
		FcyRequestedAmount:     fcyAmount,
		AccountsToDeduct:       request.AccountsToDeduct,
		FcyAcceptanceMode:      request.FcyAcceptanceMode,
		CardAssociatedAccount:  &request.CardAssociatedAccount,
		BranchRecommendation:   &request.BranchRecommendation,
		TravelPurposeID:        travelPurposeObjID,
		TravelCountryID:        travelCountryObjID,
		RequestingAs:           request.RequestingAs,
		AccountCurrencyID:      accountCurrencyObjID,
		FcyRequestedID:         fcyRequestedObjID,
		AssignedTo:             &authorizerObjID,
	}

	if attachments.PassportID != nil {
		fcyRequest.PassportAttachment = attachments.PassportID
	}
	if attachments.TicketID != nil {
		fcyRequest.TicketAttachment = attachments.TicketID
	}
	if attachments.VisaID != nil {
		fcyRequest.VisaAttachment = attachments.VisaID
	}
	if attachments.EducationID != nil {
		fcyRequest.EducationLoaAttachment = attachments.EducationID
	}
	if attachments.BusinessLicenseID != nil {
		fcyRequest.BusinessLicenseAttachment = attachments.BusinessLicenseID
	}
	if attachments.HealthLetterID != nil {
		fcyRequest.HealthLetterAttachment = attachments.HealthLetterID
	}
	if attachments.SupportingLetterID != nil {
		fcyRequest.BusinessSupportingAttachment = attachments.SupportingLetterID
	}

	err = rc.requestUsecase.UpdateRequest(c, userID, requestID, &fcyRequest)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Failed to update the request")

		var (
			status  int
			message string
		)

		switch {
		case errors.Is(err, common.ErrUnauthorized):
			status = http.StatusUnauthorized
			message = common.MessUnauthorized

		case errors.Is(err, common.ErrRequestNotFound):
			status = http.StatusNotFound
			message = common.MessRequestNotFound

		case errors.Is(err, common.ErrRequestConflict):
			status = http.StatusConflict
			message = common.MessRequestConflict

		case errors.Is(err, common.ErrRequestIsLocked):
			status = http.StatusConflict
			message = common.MessRequestLocked

		case errors.Is(err, common.ErrInvalidRequestTransition):
			status = http.StatusConflict
			message = "Invalid request status for this action"

		case errors.Is(err, common.ErrAccountCurrencyMustBeETB):
			status = http.StatusBadRequest
			message = "Account currency must be ETB"

		case errors.Is(err, common.ErrLocalCurrencyNotAllowed):
			status = http.StatusBadRequest
			message = "Requested currency cannot be ETB"

		case errors.Is(err, common.ErrInvalidCurrency):
			status = http.StatusBadRequest
			message = "Invalid currency"

		case errors.Is(err, common.ErrAssigneeNotEligible):
			status = http.StatusBadRequest
			message = "Selected authorizer is not eligible for this branch"

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	logEntry.Info("Request updated successfully")
	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Fcy Request updated successfully"})
}

type requestAttachmentIDs struct {
	PassportID         *primitive.ObjectID
	TicketID           *primitive.ObjectID
	VisaID             *primitive.ObjectID
	EducationID        *primitive.ObjectID
	BusinessLicenseID  *primitive.ObjectID
	HealthLetterID     *primitive.ObjectID
	SupportingLetterID *primitive.ObjectID
}

// uploadOptionalFormFile uploads a multipart field when present.
func (rc *requestController) uploadOptionalFormFile(c *gin.Context, field, category string) (*primitive.ObjectID, error) {
	file, err := c.FormFile(field)
	if err != nil {
		return nil, nil
	}
	id, err := rc.fileUsecase.AddFile(c, file, category)
	if err != nil {
		return nil, fmt.Errorf("failed to upload %s: %w", field, err)
	}
	return id, nil
}

func (rc *requestController) uploadRequestAttachments(c *gin.Context) (*requestAttachmentIDs, error) {
	out := &requestAttachmentIDs{}
	var err error

	if out.PassportID, err = rc.uploadOptionalFormFile(c, "passport_attachment", "passport"); err != nil {
		return nil, err
	}
	if out.TicketID, err = rc.uploadOptionalFormFile(c, "ticket_attachment", "ticket"); err != nil {
		return nil, err
	}
	if out.VisaID, err = rc.uploadOptionalFormFile(c, "visa_attachment", "visa"); err != nil {
		return nil, err
	}
	if out.EducationID, err = rc.uploadOptionalFormFile(c, "education_loa_attachment", "education"); err != nil {
		return nil, err
	}
	if out.BusinessLicenseID, err = rc.uploadOptionalFormFile(c, "business_license_attachment", "business"); err != nil {
		return nil, err
	}
	if out.HealthLetterID, err = rc.uploadOptionalFormFile(c, "health_letter_attachment", "health"); err != nil {
		return nil, err
	}
	// Prefer canonical field; accept legacy name for older clients.
	if out.SupportingLetterID, err = rc.uploadOptionalFormFile(c, "business_supporting_attachment", "support"); err != nil {
		return nil, err
	}
	if out.SupportingLetterID == nil {
		if out.SupportingLetterID, err = rc.uploadOptionalFormFile(c, "business_supporting_letter", "support"); err != nil {
			return nil, err
		}
	}

	return out, nil
}

// helper function for approve request
func getAmount(slice []float64, index int) float64 {
	if index < len(slice) {
		return slice[index]
	}
	return 0
}

func (rc *requestController) ApproveRequest(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	var request model.RequestApprovalDTO
	if err := c.ShouldBind(&request); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			hasAmountError := false

			for _, fe := range ve {
				if fe.Tag() == "gt" &&
					(strings.HasPrefix(fe.Field(), "ApprovedAmounts") ||
						strings.HasPrefix(fe.Field(), "ApprovedAmountInCash") ||
						strings.HasPrefix(fe.Field(), "ApprovedAmountInCard")) {
					hasAmountError = true
					break
				}
			}

			if hasAmountError {
				c.JSON(http.StatusBadRequest, response.Status{
					Message: "Approved amount must be greater than 0",
					Error:   "Approved amount must be greater than 0",
				})
				return
			}

			// fallback for other validation errors
			c.JSON(http.StatusBadRequest, response.Status{
				Message: "Validation failed",
				Error:   "validation failed",
			})
			return
		}

		// fallback for non-validator errors
		c.JSON(http.StatusBadRequest, response.Status{
			Message: common.MessInvalidRequest,
			Error:   err.Error(),
		})
		return

	}

	length := len(request.ApprovedAmounts)

	if length != len(request.ApprovedAmountInCash) ||
		length != len(request.ApprovedAmountInCard) ||
		length != len(request.ApprovedCurrencyIDs) {

		c.JSON(http.StatusBadRequest, response.Status{
			Message: "All approved arrays must have the same length",
			Error:   "All approved arrays must have the same length",
		})
		return
	}

	// Check no approved amount is 0 or negative
	for i, amt := range request.ApprovedAmounts {
		if amt <= 0 {
			c.JSON(http.StatusBadRequest, response.Status{
				Message: fmt.Sprintf("Total approved amount at index %d must be greater than 0", i),
				Error:   "Total approved amount must be greater than 0",
			})
			return
		}
	}

	for i := range request.ApprovedAmounts {
		cash := 0.0
		card := 0.0

		if i < len(request.ApprovedAmountInCash) {
			cash = request.ApprovedAmountInCash[i]
		}

		if i < len(request.ApprovedAmountInCard) {
			card = request.ApprovedAmountInCard[i]
		}

		total := cash + card

		if total != request.ApprovedAmounts[i] {
			c.JSON(http.StatusBadRequest, response.Status{
				Message: fmt.Sprintf(
					"Approved amount at index %d must equal cash + card (%.2f + %.2f = %.2f)",
					i, cash, card, total,
				),
			})
			return
		}
	}

	// get user id from param
	requestID := c.Param("id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "request ID is required"})
		return
	}

	requestObjID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	err = rc.requestUsecase.ApproveRequest(c, userID, requestObjID, &request)
	if err != nil {
		var (
			status  int
			message string
		)

		switch {
		case errors.Is(err, common.ErrRequestNotFound):
			status = http.StatusNotFound
			message = common.MessRequestNotFound

		case errors.Is(err, common.ErrRequestConflict):
			status = http.StatusConflict
			message = common.MessRequestConflict

		case errors.Is(err, common.ErrRequestIsLocked):
			status = http.StatusConflict
			message = common.MessRequestLocked

		case errors.Is(err, common.ErrRequestNotAssigned):
			status = http.StatusConflict
			message = common.MessRequestNotAssigned

		case errors.Is(err, common.ErrInvalidRequestTransition):
			status = http.StatusConflict
			message = "Invalid request status for this action"

		case errors.Is(err, common.ErrCannotActOnOwnRequest):
			status = http.StatusForbidden
			message = common.MessCannotActOnOwnRequest

		case errors.Is(err, common.ErrLocalCurrencyNotAllowed):
			status = http.StatusBadRequest
			message = "Approved currency cannot be ETB"

		case errors.Is(err, common.ErrInvalidCurrency):
			status = http.StatusBadRequest
			message = "Invalid approved currency"

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Request approved successfully"})
}

func (rc *requestController) GetAllOrgRequests(c *gin.Context) {
	orgKey, orgID := resolveOrgFromContext(c)
	pageQuery := parsePageQuery(c, 25)

	// Head-office tokens have no branch/department: return empty (do not 400).
	if orgKey == "" || orgID.IsZero() {
		writePaginatedRequests(c, "All Org requests fetched successfully", &model.PaginatedRequests{
			Items: []model.Request{},
			Total: 0,
			Page:  pageQuery.Page,
			Limit: pageQuery.Limit,
		})
		return
	}

	result, err := rc.requestUsecase.GetAllOrgRequests(c, orgKey, orgID, pageQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	writePaginatedRequests(c, "All Org requests fetched successfully", result)
}

func (rc *requestController) GetAuthorizedRequests(c *gin.Context) {
	_, err := utils.GetDepartmentID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	_, err = utils.GetBranchID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	result, err := rc.requestUsecase.GetAuthorizedRequests(c, model.PageQuery{Page: 1, Limit: 0})
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Authorized requests fetched successfully", Data: result.Items})
}

func (rc *requestController) GetNewOrgRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	var (
		orgKey string
		orgID  primitive.ObjectID
	)

	if !branchID.IsZero() {
		orgKey = "branch_id"
		orgID = branchID
	} else if !departmentID.IsZero() {
		orgKey = "department_id"
		orgID = departmentID
	} else {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "Invalid organization context"})
		return
	}

	result, err := rc.requestUsecase.GetNewOrgRequests(c, orgKey, orgID, model.PageQuery{Page: 1, Limit: 0})
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "New Org requests fetched successfully", Data: result.Items})
}

func (rc *requestController) GetRejectedOrgRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	var (
		orgKey string
		orgID  primitive.ObjectID
	)

	if !branchID.IsZero() {
		orgKey = "branch_id"
		orgID = branchID
	} else if !departmentID.IsZero() {
		orgKey = "department_id"
		orgID = departmentID
	} else {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "Invalid organization context"})
		return
	}

	result, err := rc.requestUsecase.GetRejectedOrgRequests(c, orgKey, orgID, parsePageQuery(c, 25))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	writePaginatedRequests(c, "Rejected requests fetched successfully", result)
}

func (rc *requestController) GetRejectedRequests(c *gin.Context) {
	result, err := rc.requestUsecase.GetRejectedRequests(c, parsePageQuery(c, 25))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	writePaginatedRequests(c, "Rejected Org requests fetched successfully", result)
}

func (rc *requestController) GetApprovedOrgRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	var (
		orgKey string
		orgID  primitive.ObjectID
	)

	if !branchID.IsZero() {
		orgKey = "branch_id"
		orgID = branchID
	} else if !departmentID.IsZero() {
		orgKey = "department_id"
		orgID = departmentID
	} else {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "Invalid organization context"})
		return
	}

	result, err := rc.requestUsecase.GetApprovedOrgRequests(c, orgKey, orgID, parsePageQuery(c, 25))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	writePaginatedRequests(c, "Approved Org requests fetched successfully", result)
}

func (rc *requestController) GetAcceptedOrgRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	var (
		orgKey string
		orgID  primitive.ObjectID
	)

	if !branchID.IsZero() {
		orgKey = "branch_id"
		orgID = branchID
	} else if !departmentID.IsZero() {
		orgKey = "department_id"
		orgID = departmentID
	} else {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "Invalid organization context"})
		return
	}

	result, err := rc.requestUsecase.GetAcceptedOrgRequests(c, orgKey, orgID, parsePageQuery(c, 25))
	if err != nil {

		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	writePaginatedRequests(c, "Accepted Org requests fetched successfully", result)
}

func (rc *requestController) GetDeclinedOrgRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	var (
		orgKey string
		orgID  primitive.ObjectID
	)

	if !branchID.IsZero() {
		orgKey = "branch_id"
		orgID = branchID
	} else if !departmentID.IsZero() {
		orgKey = "department_id"
		orgID = departmentID
	} else {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "Invalid organization context"})
		return
	}

	result, err := rc.requestUsecase.GetDeclinedOrgRequests(c, orgKey, orgID, parsePageQuery(c, 25))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	writePaginatedRequests(c, "Declined Org requests fetched successfully", result)
}

func (rc *requestController) GetAuthorizedOrgRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	var (
		orgKey string
		orgID  primitive.ObjectID
	)

	if !branchID.IsZero() {
		orgKey = "branch_id"
		orgID = branchID
	} else if !departmentID.IsZero() {
		orgKey = "department_id"
		orgID = departmentID
	} else {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "Invalid organization context"})
		return
	}

	result, err := rc.requestUsecase.GetAuthorizedOrgRequests(c, orgKey, orgID, model.PageQuery{Page: 1, Limit: 0})
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Authorized Org requests", Data: result.Items})
}

func (rc *requestController) GetDraftedOrgRequests(c *gin.Context) {
	viewerID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	orgKey, orgID := resolveOrgFromContext(c)
	result, err := rc.requestUsecase.GetRequestQueue(
		c, string(model.ReqStatusDrafted), "org", orgKey, orgID, viewerID, model.PageQuery{Page: 1, Limit: 0},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Drafted requests fetched successfully", Data: result.Items})
}

func (rc *requestController) AuthorizeOrgRequest(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	// get user id from param
	requestID := c.Param("id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessUnauthorized, Error: "request ID is required"})
		return
	}

	requestObjID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	err = rc.requestUsecase.AuthorizeOrgRequest(c, authUserID, requestObjID)
	if err != nil {
		var (
			status  int
			message string
		)

		switch {
		case errors.Is(err, common.ErrRequestNotFound):
			status = http.StatusNotFound
			message = common.MessRequestNotFound

		case errors.Is(err, common.ErrUnauthorized):
			status = http.StatusForbidden
			message = common.MessUnauthorized

		case errors.Is(err, common.ErrRequestNotAssigned):
			status = http.StatusConflict
			message = common.MessRequestNotAssigned

		case errors.Is(err, common.ErrRequestIsLocked):
			status = http.StatusConflict
			message = common.MessRequestLocked

		case errors.Is(err, common.ErrRequestConflict):
			status = http.StatusConflict
			message = common.MessRequestConflict

		case errors.Is(err, common.ErrInvalidRequestTransition):
			status = http.StatusConflict
			message = "Invalid request status for this action"

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Request authorized successfully"})
}

func (rc *requestController) RejectRequest(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	// get user id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessUnauthorized, Error: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	var payload RejectRequestPayload
	err = c.ShouldBind(&payload)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			e := validationErrors[0]
			message := fmt.Sprintf("%s failed on %s validation", e.Field(), e.Tag())

			c.JSON(http.StatusBadRequest, response.Status{
				Message: message,
				Error:   err.Error(),
			})

			return
		}

		c.JSON(http.StatusBadRequest, response.Status{
			Message: common.MessInvalidRequest,
			Error:   err.Error(),
		})
		return
	}

	err = rc.requestUsecase.RejectRequest(c, authUserID, requestID, payload.RejectionReason)
	if err != nil {
		var (
			status  int
			message string
		)

		switch {
		case errors.Is(err, common.ErrUnauthorized):
			status = http.StatusUnauthorized
			message = common.MessUnauthorized

		case errors.Is(err, common.ErrRequestNotFound):
			status = http.StatusNotFound
			message = common.MessRequestNotFound

		case errors.Is(err, common.ErrRequestConflict):
			status = http.StatusConflict
			message = common.MessRequestConflict

		case errors.Is(err, common.ErrRequestIsLocked):
			status = http.StatusConflict
			message = common.MessRequestLocked

		case errors.Is(err, common.ErrRequestNotAssigned):
			status = http.StatusConflict
			message = common.MessRequestNotAssigned

		case errors.Is(err, common.ErrInvalidRequestTransition):
			status = http.StatusConflict
			message = "Invalid request status for this action"

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Request rejected successfully"})
}

func (rc *requestController) DeleteRequest(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	// get user id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	err = rc.requestUsecase.DeleteRequest(c, authUserID, requestID)
	if err != nil {
		var (
			status  int
			message string
		)

		switch {
		case errors.Is(err, common.ErrRequestNotFound):
			status = http.StatusNotFound
			message = common.MessRequestNotFound

		case errors.Is(err, common.ErrRequestCannotBeDeleted):
			status = http.StatusConflict
			message = "Invalid request status for this action"

		case errors.Is(err, common.ErrUnauthorized):
			status = http.StatusForbidden
			message = common.MessUnauthorized

		case errors.Is(err, common.ErrInvalidRequestTransition):
			status = http.StatusConflict
			message = "Invalid request status for this action"

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Request deleted successfully"})
}

func (rc *requestController) AcceptRequest(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	var request model.RequestAcceptanceDTO
	if err := c.ShouldBind(&request); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			hasAmountError := false

			for _, fe := range ve {
				if fe.Tag() == "gt" &&
					(strings.HasPrefix(fe.Field(), "AcceptedAmounts") ||
						strings.HasPrefix(fe.Field(), "AcceptedAmountInCash") ||
						strings.HasPrefix(fe.Field(), "AcceptedAmountInCard")) {
					hasAmountError = true
					break
				}
			}

			if hasAmountError {
				c.JSON(http.StatusBadRequest, response.Status{
					Message: "Accepted amount must be greater than 0",
					Error:   "Accepted amount must be greater than 0",
				})
				return
			}

			// fallback for other validation errors
			c.JSON(http.StatusBadRequest, response.Status{
				Message: "Validation failed",
				Error:   "validation failed",
			})
			return
		}

		// fallback for non-validator errors
		c.JSON(http.StatusBadRequest, response.Status{
			Message: common.MessInvalidRequest,
			Error:   err.Error(),
		})
		return

	}

	length := len(request.AcceptedAmounts)

	if length != len(request.AcceptedAmountInCash) ||
		length != len(request.AcceptedAmountInCard) ||
		length != len(request.AcceptedCurrencyIDs) {

		c.JSON(http.StatusBadRequest, response.Status{
			Message: "All accepted arrays must have the same length",
			Error:   "All accepted arrays must have the same length",
		})
		return
	}

	// Check no accepted amount is 0 or negative
	for i, amt := range request.AcceptedAmounts {
		if amt <= 0 {
			c.JSON(http.StatusBadRequest, response.Status{
				Message: fmt.Sprintf("Total accepted amount at index %d must be greater than 0", i),
				Error:   "Total accepted amount must be greater than 0",
			})
			return
		}
	}

	for i := range request.AcceptedAmounts {
		cash := 0.0
		card := 0.0

		if i < len(request.AcceptedAmountInCash) {
			cash = request.AcceptedAmountInCash[i]
		}

		if i < len(request.AcceptedAmountInCard) {
			card = request.AcceptedAmountInCard[i]
		}

		total := cash + card

		if total != request.AcceptedAmounts[i] {
			c.JSON(http.StatusBadRequest, response.Status{
				Message: fmt.Sprintf(
					"Accepted amount at index %d must equal cash + card (%.2f + %.2f = %.2f)",
					i, cash, card, total,
				),
			})
			return
		}
	}

	// get user id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	processingBranchID, _ := utils.GetBranchID(c)
	err = rc.requestUsecase.AcceptRequest(c, authUserID, requestID, &request, processingBranchID)
	if err != nil {
		var (
			status  int
			message string
		)

		switch {
		case errors.Is(err, common.ErrRequestNotFound):
			status = http.StatusNotFound
			message = common.MessRequestNotFound

		case errors.Is(err, common.ErrUnauthorized):
			status = http.StatusForbidden
			message = common.MessUnauthorized

		case errors.Is(err, common.ErrRequestConflict):
			status = http.StatusConflict
			message = common.MessRequestConflict

		case errors.Is(err, common.ErrRequestIsLocked):
			status = http.StatusConflict
			message = common.MessRequestLocked

		case errors.Is(err, common.ErrInvalidRequestTransition):
			status = http.StatusConflict
			message = "Invalid request status for this action"

		case errors.Is(err, common.ErrInvalidAcceptedCurrency):
			status = http.StatusBadRequest
			message = "Accepted currency must be the same as approved currency"

		case errors.Is(err, common.ErrInvalidAcceptedAmount):
			status = http.StatusBadRequest
			message = "Accepted amount cannot be greater than approved amount"

		case errors.Is(err, common.ErrLocalCurrencyNotAllowed):
			status = http.StatusBadRequest
			message = "Accepted currency cannot be ETB"

		case errors.Is(err, common.ErrInvalidCurrency):
			status = http.StatusBadRequest
			message = "Invalid accepted currency"

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Request accepted successfully"})
}

func (rc *requestController) SendRequest(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Error: err.Error()})
		return
	}

	// get user id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	err = rc.requestUsecase.SendRequest(c, authUserID, requestID)
	if err != nil {
		var (
			status  int
			message string
		)

		switch {
		case errors.Is(err, common.ErrRequestNotFound):
			status = http.StatusNotFound
			message = common.MessRequestNotFound

		case errors.Is(err, common.ErrUnauthorized):
			status = http.StatusForbidden
			message = common.MessUnauthorized

		case errors.Is(err, common.ErrInvalidRequestTransition):
			status = http.StatusConflict
			message = "Invalid request status for this action"

		case errors.Is(err, common.ErrRequestConflict):
			status = http.StatusConflict
			message = common.MessRequestConflict

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Request sent successfully"})
}

func (rc *requestController) DeclineOrgRequest(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	// get user id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	err = rc.requestUsecase.DeclineOrgRequest(c, authUserID, requestID)
	if err != nil {
		var (
			status  int
			message string
		)

		switch {
		case errors.Is(err, common.ErrRequestNotFound):
			status = http.StatusNotFound
			message = common.MessRequestNotFound

		case errors.Is(err, common.ErrUnauthorized):
			status = http.StatusForbidden
			message = common.MessUnauthorized

		case errors.Is(err, common.ErrInvalidRequestTransition):
			status = http.StatusConflict
			message = "Invalid request status for this action"

		case errors.Is(err, common.ErrRequestConflict):
			status = http.StatusConflict
			message = common.MessRequestConflict

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Request declined successfully"})
}

func (rc *requestController) GetRequestStatusSummary(c *gin.Context) {
	departmentID, _ := utils.GetDepartmentID(c)
	branchID, _ := utils.GetBranchID(c)

	orgKey := ""
	orgID := primitive.NilObjectID
	scope := c.DefaultQuery("scope", "bank")
	if scope == "org" {
		if !branchID.IsZero() {
			orgKey = "branch_id"
			orgID = branchID
		} else if !departmentID.IsZero() {
			orgKey = "department_id"
			orgID = departmentID
		}
		// No branch/department on token (typical for head-office admins):
		// fall back to bank-wide summary instead of 400.
	}

	summary, err := rc.requestUsecase.GetRequestStatusSummary(c, orgKey, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Request summary fetched successfully", Data: summary})
}

func (rc *requestController) GetValidatedRequests(c *gin.Context) {
	result, err := rc.requestUsecase.GetValidatedRequests(c, model.PageQuery{Page: 1, Limit: 0})
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Validated requests fetched successfully", Data: result.Items})
}

func (rc *requestController) GetApprovedRequests(c *gin.Context) {
	result, err := rc.requestUsecase.GetApprovedRequests(c, parsePageQuery(c, 25))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	writePaginatedRequests(c, "Approved requests fetched successfully", result)
}

func (rc *requestController) GetAcceptedRequests(c *gin.Context) {
	result, err := rc.requestUsecase.GetAcceptedRequests(c, parsePageQuery(c, 25))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Message: common.MessInternalServerError, Error: err.Error()})
		return
	}

	writePaginatedRequests(c, "Accepted requests fetched successfully", result)
}

func (rc *requestController) GetDeclinedRequests(c *gin.Context) {
	result, err := rc.requestUsecase.GetDeclinedRequests(c, parsePageQuery(c, 25))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{Error: err.Error()})
		return
	}

	writePaginatedRequests(c, "Declined requests fetched successfully", result)
}

func (rc *requestController) LockRequest(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	// get request id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	var dto model.AssignRequestDTO
	_ = c.ShouldBindJSON(&dto)

	var assigneeID *primitive.ObjectID
	if dto.AssigneeID != nil && *dto.AssigneeID != "" {
		oid, parseErr := primitive.ObjectIDFromHex(*dto.AssigneeID)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: parseErr.Error()})
			return
		}
		assigneeID = &oid
	}

	actorPerms := utils.GetPermissions(c)
	result, err := rc.requestUsecase.AssignRequest(c, authUserID, requestID, assigneeID, actorPerms)
	if err != nil {
		var (
			status  int
			message string
		)

		switch {
		case errors.Is(err, common.ErrRequestNotFound):
			status = http.StatusNotFound
			message = common.MessRequestNotFound

		case errors.Is(err, common.ErrRequestConflict):
			status = http.StatusConflict
			message = common.MessRequestConflict

		case errors.Is(err, common.ErrRequestIsLocked):
			status = http.StatusConflict
			message = common.MessRequestLocked

		case errors.Is(err, common.ErrRequestNotAssigned):
			status = http.StatusConflict
			message = common.MessRequestNotAssigned

		case errors.Is(err, common.ErrInvalidRequestTransition):
			status = http.StatusConflict
			message = "Invalid request status for this action"

		case errors.Is(err, common.ErrUnauthorized):
			status = http.StatusForbidden
			message = common.MessUnauthorized

		case errors.Is(err, common.ErrSelfAssignOnly):
			status = http.StatusForbidden
			message = err.Error()

		case errors.Is(err, common.ErrAssigneeNotEligible):
			status = http.StatusBadRequest
			message = err.Error()

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Request assigned successfully", Data: result})
}

func (rc *requestController) UnLockRequest(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	// get request id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	actorPerms := utils.GetPermissions(c)
	err = rc.requestUsecase.UnassignRequest(c, authUserID, requestID, actorPerms)
	if err != nil {
		var (
			status  int
			message string
		)

		switch {
		case errors.Is(err, common.ErrRequestNotFound):
			status = http.StatusNotFound
			message = common.MessRequestNotFound

		case errors.Is(err, common.ErrRequestConflict):
			status = http.StatusConflict
			message = common.MessRequestConflict

		case errors.Is(err, common.ErrRequestIsLocked):
			status = http.StatusConflict
			message = common.MessRequestLocked

		case errors.Is(err, common.ErrRequestNotAssigned):
			status = http.StatusConflict
			message = common.MessRequestNotAssigned

		case errors.Is(err, common.ErrInvalidRequestTransition):
			status = http.StatusConflict
			message = "Invalid request status for this action"

		default:
			status = http.StatusInternalServerError
			message = common.MessInternalServerError
		}

		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{IsSuccessful: true, Message: "Request unassigned successfully"})
}

func (rc *requestController) GenerateRequestReport(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
		return
	}

	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequestData, Error: err.Error()})
		return
	}

	pdfBytes, filename, receiptNumber, err := rc.requestUsecase.GenerateRequestReport(c, authUserID, requestID)
	if err != nil {
		status := http.StatusInternalServerError
		message := common.MessInternalServerError
		switch {
		case errors.Is(err, common.ErrUnauthorized):
			status = http.StatusUnauthorized
			message = common.MessUnauthorized
		case errors.Is(err, common.ErrRequestNotFound):
			status = http.StatusNotFound
			message = common.MessRequestNotFound
		case errors.Is(err, common.ErrInvalidRequestTransition):
			status = http.StatusConflict
			message = "Report can only be generated for accepted requests"
		}
		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Header("X-Receipt-Number", receiptNumber)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func (rc *requestController) ValidateReceipt(c *gin.Context) {
	receiptNumber := c.Param("receiptNumber")
	if strings.TrimSpace(receiptNumber) == "" {
		c.JSON(http.StatusBadRequest, response.Status{Message: common.MessInvalidRequest, Error: "receipt number is required"})
		return
	}

	receipt, comparison, err := rc.requestUsecase.ValidateReceipt(c, receiptNumber)
	if err != nil {
		status := http.StatusInternalServerError
		message := common.MessInternalServerError
		if errors.Is(err, common.ErrReceiptNotFound) {
			status = http.StatusNotFound
			message = "Receipt not found"
		}
		c.JSON(status, response.Status{Message: message, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Status{
		IsSuccessful: true,
		Message:      "Receipt linked request found",
		Data: gin.H{
			"request":    receipt,
			"comparison": comparison,
		},
	})
}

