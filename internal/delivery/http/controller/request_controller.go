package controller

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RejectRequestPayload struct {
	RejectionReason string `json:"rejection_reason"`
}

var payload RejectRequestPayload

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

func (rc *requestController) AddRequest(c *gin.Context) {
	logEntry := utils.GetLogger(c)
	userID, err := utils.GetUserID(c)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid user ID")
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	var request model.RequestDTO

	err = c.ShouldBind(&request)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid request payload")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	logEntry.WithField("request", request).Debug("Request payload received")

	// Convert string fields to ObjectID
	travelPurposeObjID, err := primitive.ObjectIDFromHex(request.TravelPurposeID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid travel purpose ID")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid travel purpose ID"})
		return
	}

	travelCountryObjID, err := primitive.ObjectIDFromHex(request.TravelCountryID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid travel country ID")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid travel country ID"})
		return
	}

	accountCurrencyObjID, err := primitive.ObjectIDFromHex(request.AccountCurrencyID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid currency account ID")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid currency account ID"})
		return
	}

	fcyRequestedObjID, err := primitive.ObjectIDFromHex(request.FcyRequestedID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid currency ID")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid Currency ID"})
		return
	}

	deposit, err := strconv.ParseFloat(request.AverageDeposit, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid average deposit value")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid average deposit value"})
		return
	}

	totalFcy, err := strconv.ParseFloat(request.TotalFcyGenerated, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid total fcy performance value")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid total fcy performance value"})
		return
	}

	currentFcy, err := strconv.ParseFloat(request.CurrentFcyPerformance, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid current fcy performance value")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid current fcy performance value"})
		return
	}

	fcyAmount, err := strconv.ParseFloat(request.FcyRequestedAmount, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid fcy amount value")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid fcy amount value"})
		return
	}

	uploadPath := configs.FileUploadPath
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		logEntry.WithField("error", err.Error()).Warn("Unable to create a file upload folder")
		os.MkdirAll(uploadPath, os.ModePerm)
	}

	var passportID *primitive.ObjectID
	var ticketID *primitive.ObjectID
	var educationID *primitive.ObjectID
	var businessLicenseID *primitive.ObjectID
	var supportingLetterID *primitive.ObjectID

	passport, err := c.FormFile("passport_attachment")
	if err == nil {
		passportID, err = rc.fileUsecase.AddFile(c, passport, "passport")
		if err != nil {
			logEntry.WithField("error", err.Error()).Warn("Failed to upload education file")
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload education file"})
			return
		}
	}
	ticket, err := c.FormFile("ticket_attachment")
	if err == nil {
		ticketID, err = rc.fileUsecase.AddFile(c, ticket, "ticket")
		if err != nil {
			logEntry.WithField("error", err.Error()).Warn("Failed to upload ticket file")
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload ticket file"})
			return
		}
	}
	education, err := c.FormFile("education_loa_attachment")
	if err == nil {
		educationID, err = rc.fileUsecase.AddFile(c, education, "education")
		if err != nil {
			logEntry.WithField("error", err.Error()).Warn("Failed to upload education file")
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload education file"})
			return
		}
	}
	businessLicense, err := c.FormFile("business_license_attachment")
	if err == nil {
		businessLicenseID, err = rc.fileUsecase.AddFile(c, businessLicense, "business")
		if err != nil {
			logEntry.WithField("error", err.Error()).Warn("Failed to upload business license file")
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload business license file"})
			return
		}
	}
	supportLetter, err := c.FormFile("business_supporting_letter")
	if err == nil {
		supportingLetterID, err = rc.fileUsecase.AddFile(c, supportLetter, "support")
		if err != nil {
			logEntry.WithField("error", err.Error()).Warn("Failed to upload business letter file")
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload business letter file"})
			return
		}
	}

	if len(request.AccountsToDeduct) == 0 {
		logEntry.Warn("Account to deduct is required")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "accounts_to_deduct is required"})
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

		PassportAttachment:           passportID,
		TicketAttachment:             ticketID,
		BusinessLicenseAttachment:    businessLicenseID,
		EducationLoaAttachment:       educationID,
		BusinessSupportingAttachment: supportingLetterID,
	}

	err = rc.requestUsecase.AddRequest(c, userID, &fcyRequest)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Failed to add the request")
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	logEntry.Info("Request created successfully")
	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Request created successfully"})
}

func (rc *requestController) GetAllRequests(c *gin.Context) {
	requests, err := rc.requestUsecase.GetAllRequests(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, requests)
}

func (rc *requestController) ValidateRequest(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		fmt.Println(userID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	fmt.Println("request controller'")
	var request model.RequestValidationDTO

	err = c.ShouldBind(&request)
	if err != nil {
		fmt.Println("err binding", err)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	fmt.Println("after unmarshaling", request)
	fmt.Printf("%+v\n", request)

	// get user id from param
	requestID := c.Param("id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "user ID is required"})
		return
	}

	requestObjID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	validatedCurrencyObjID, err := primitive.ObjectIDFromHex(request.ValidatedAccountCurrencyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = rc.requestUsecase.ValidateRequest(c, userID, requestObjID, validatedCurrencyObjID, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Request validated successfully"})
}

func (rc *requestController) UpdateRequest(c *gin.Context) {
	logEntry := utils.GetLogger(c)
	userID, err := utils.GetUserID(c)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid user ID")
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	// get user id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	var request model.RequestDTO

	err = c.ShouldBind(&request)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid request payload")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	logEntry.WithField("request", request).Debug("Request payload received")

	form, err := c.MultipartForm()
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid multipart form data")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid multipart form data"})
		return
	}

	request.AccountsToDeduct = form.Value["accounts_to_deduct"]

	// Convert string fields to ObjectID
	travelPurposeObjID, err := primitive.ObjectIDFromHex(request.TravelPurposeID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid travel purpose ID")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid travel purpose ID"})
		return
	}

	travelCountryObjID, err := primitive.ObjectIDFromHex(request.TravelCountryID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid travel country ID")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid travel country ID"})
		return
	}

	accountCurrencyObjID, err := primitive.ObjectIDFromHex(request.AccountCurrencyID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid currency account ID")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid currency account ID"})
		return
	}

	fcyRequestedObjID, err := primitive.ObjectIDFromHex(request.FcyRequestedID)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid currency ID")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid Currency ID"})
		return
	}

	deposit, err := strconv.ParseFloat(request.AverageDeposit, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid average deposit value")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid average deposit value"})
		return
	}

	totalFcy, err := strconv.ParseFloat(request.TotalFcyGenerated, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid total fcy performance value")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid total fcy performance value"})
		return
	}

	currentFcy, err := strconv.ParseFloat(request.CurrentFcyPerformance, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid current fcy performance value")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid current fcy performance value"})
		return
	}

	fcyAmount, err := strconv.ParseFloat(request.FcyRequestedAmount, 64)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Invalid fcy amount value")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid fcy amount value"})
		return
	}

	uploadPath := configs.FileUploadPath
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		logEntry.WithField("error", err.Error()).Warn("Unable to create a file upload folder")
		os.MkdirAll(uploadPath, os.ModePerm)
	}

	var passportID *primitive.ObjectID
	var ticketID *primitive.ObjectID
	var educationID *primitive.ObjectID
	var businessLicenseID *primitive.ObjectID
	var supportingLetterID *primitive.ObjectID

	passport, err := c.FormFile("passport_attachment")
	if err == nil {
		passportID, err = rc.fileUsecase.AddFile(c, passport, "passport")
		if err != nil {
			logEntry.WithField("error", err.Error()).Warn("Failed to upload passport file")
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload passport file"})
			return
		}
	}
	ticket, err := c.FormFile("ticket_attachment")
	if err == nil {
		ticketID, err = rc.fileUsecase.AddFile(c, ticket, "ticket")
		if err != nil {
			logEntry.WithField("error", err.Error()).Warn("Failed to upload ticket file")
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload ticket file"})
			return
		}
	}
	education, err := c.FormFile("education_loa_attachment")
	if err == nil {
		educationID, err = rc.fileUsecase.AddFile(c, education, "education")
		if err != nil {
			logEntry.WithField("error", err.Error()).Warn("Failed to upload education file")
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload education file"})
			return
		}
	}
	businessLicense, err := c.FormFile("business_license_attachment")
	if err == nil {
		businessLicenseID, err = rc.fileUsecase.AddFile(c, businessLicense, "business")
		if err != nil {
			logEntry.WithField("error", err.Error()).Warn("Failed to upload business license file")
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload business license file"})
			return
		}
	}
	supportLetter, err := c.FormFile("business_supporting_letter")
	if err == nil {
		supportingLetterID, err = rc.fileUsecase.AddFile(c, supportLetter, "support")
		if err != nil {
			logEntry.WithField("error", err.Error()).Warn("Failed to upload business letter file")
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload business letter file"})
			return
		}
	}

	if len(request.AccountsToDeduct) == 0 {
		logEntry.Warn("Account to deduct is required")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "accounts_to_deduct is required"})
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
	}

	if passportID != &primitive.NilObjectID {
		fcyRequest.PassportAttachment = passportID
	}
	if ticketID != &primitive.NilObjectID {
		fcyRequest.TicketAttachment = ticketID
	}
	if educationID != nil {
		fcyRequest.EducationLoaAttachment = educationID
	}
	if businessLicenseID != nil {
		fcyRequest.BusinessLicenseAttachment = businessLicenseID
	}
	if supportingLetterID != nil {
		fcyRequest.BusinessSupportingAttachment = supportingLetterID
	}

	err = rc.requestUsecase.UpdateRequest(c, userID, requestID, &fcyRequest)
	if err != nil {
		logEntry.WithField("error", err.Error()).Warn("Failed to update the request")
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	logEntry.Info("Request updated successfully")
	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Request updated successfully"})
}

func (rc *requestController) ApproveRequest(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		fmt.Println(userID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
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
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Approved amount must be greater than 0",
				})
				return
			}

			// fallback for other validation errors
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Validation failed",
			})
			return
		}

		// fallback for non-validator errors
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return

	}

	for i := range request.ApprovedAmounts {
		cash := request.ApprovedAmountInCash[i]
		card := request.ApprovedAmountInCard[i]
		total := cash + card

		if total != request.ApprovedAmounts[i] {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: fmt.Sprintf(
				"Approved amount at index %d must equal to cash + card (%.2f + %.2f = %.2f)",
				i, cash, card, total,
			)})
			return
		}
	}

	// get user id from param
	requestID := c.Param("id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "request ID is required"})
		return
	}

	requestObjID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = rc.requestUsecase.ApproveRequest(c, userID, requestObjID, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Request approved successfully"})

}

func (rc *requestController) GetAllOrgRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		fmt.Println(departmentID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		fmt.Println(branchID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	var orgRequests interface{}

	if !branchID.IsZero() {
		orgRequests, err = rc.requestUsecase.GetAllOrgRequests(c, "branch_id", branchID)
	} else if !departmentID.IsZero() {
		fmt.Println("department id request", departmentID)
		orgRequests, err = rc.requestUsecase.GetAllOrgRequests(c, "department_id", departmentID)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid orgKey. Must be 'branchID' or 'departmentID'"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, orgRequests)

}

func (rc *requestController) GetAuthorizedRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		fmt.Println(departmentID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		fmt.Println(branchID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	requests, err := rc.requestUsecase.GetAuthorizedRequests(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, requests)
}

func (rc *requestController) GetNewOrgRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		fmt.Println(departmentID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		fmt.Println(branchID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	var (
		newOrgRequests interface{}
		orgKey         string
		orgID          primitive.ObjectID
	)

	if !branchID.IsZero() {
		orgKey = "branch_id"
		orgID = branchID
	} else if !departmentID.IsZero() {
		orgKey = "department_id"
		orgID = departmentID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization context"})
		return
	}

	newOrgRequests, err = rc.requestUsecase.GetNewOrgRequests(c, orgKey, orgID)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("inside controller here is the org Key and ID", orgKey, orgID)

	c.JSON(http.StatusOK, newOrgRequests)
}

func (rc *requestController) GetRejectedOrgRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		fmt.Println(departmentID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		fmt.Println(branchID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	var (
		newOrgRequests interface{}
		orgKey         string
		orgID          primitive.ObjectID
	)

	if !branchID.IsZero() {
		orgKey = "branch_id"
		orgID = branchID
	} else if !departmentID.IsZero() {
		orgKey = "department_id"
		orgID = departmentID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization context"})
		return
	}

	newOrgRequests, err = rc.requestUsecase.GetRejectedOrgRequests(c, orgKey, orgID)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("inside controller here is the org Key and ID", orgKey, orgID)

	c.JSON(http.StatusOK, newOrgRequests)
}

func (rc *requestController) GetRejectedRequests(c *gin.Context) {
	rejectedRequests, err := rc.requestUsecase.GetRejectedRequests(c)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rejectedRequests)
}

func (rc *requestController) GetApprovedOrgRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		fmt.Println(departmentID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		fmt.Println(branchID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	var (
		approvedOrgRequests interface{}
		orgKey              string
		orgID               primitive.ObjectID
	)

	if !branchID.IsZero() {
		orgKey = "branch_id"
		orgID = branchID
	} else if !departmentID.IsZero() {
		orgKey = "department_id"
		orgID = departmentID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization context"})
		return
	}

	approvedOrgRequests, err = rc.requestUsecase.GetApprovedOrgRequests(c, orgKey, orgID)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("inside controller here is the org Key and ID", orgKey, orgID)

	c.JSON(http.StatusOK, approvedOrgRequests)
}

func (rc *requestController) GetAcceptedOrgRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		fmt.Println(departmentID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		fmt.Println(branchID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	var (
		orgRequests interface{}
		orgKey      string
		orgID       primitive.ObjectID
	)

	if !branchID.IsZero() {
		orgKey = "branch_id"
		orgID = branchID
	} else if !departmentID.IsZero() {
		orgKey = "department_id"
		orgID = departmentID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization context"})
		return
	}

	orgRequests, err = rc.requestUsecase.GetAcceptedOrgRequests(c, orgKey, orgID)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("inside controller here is the org Key and ID", orgKey, orgID)

	c.JSON(http.StatusOK, orgRequests)
}

func (rc *requestController) GetDeclinedOrgRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		fmt.Println(departmentID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		fmt.Println(branchID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	var (
		orgRequests interface{}
		orgKey      string
		orgID       primitive.ObjectID
	)

	if !branchID.IsZero() {
		orgKey = "branch_id"
		orgID = branchID
	} else if !departmentID.IsZero() {
		orgKey = "department_id"
		orgID = departmentID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization context"})
		return
	}

	orgRequests, err = rc.requestUsecase.GetDeclinedOrgRequests(c, orgKey, orgID)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("inside controller here is the org Key and ID", orgKey, orgID)

	c.JSON(http.StatusOK, orgRequests)
}

func (rc *requestController) GetAuthorizedOrgRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		fmt.Println(departmentID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		fmt.Println(branchID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	var (
		orgRequests interface{}
		orgKey      string
		orgID       primitive.ObjectID
	)

	if !branchID.IsZero() {
		orgKey = "branch_id"
		orgID = branchID
	} else if !departmentID.IsZero() {
		orgKey = "department_id"
		orgID = departmentID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization context"})
		return
	}

	orgRequests, err = rc.requestUsecase.GetAuthorizedOrgRequests(c, orgKey, orgID)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("inside controller here is the org Key and ID", orgKey, orgID)

	c.JSON(http.StatusOK, orgRequests)
}

func (rc *requestController) GetDraftedOrgRequests(c *gin.Context) {
	departmentID, err := utils.GetDepartmentID(c)
	if err != nil {
		fmt.Println(departmentID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	branchID, err := utils.GetBranchID(c)
	if err != nil {
		fmt.Println(branchID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	var (
		orgRequests interface{}
		orgKey      string
		orgID       primitive.ObjectID
	)

	if !branchID.IsZero() {
		orgKey = "branch_id"
		orgID = branchID
	} else if !departmentID.IsZero() {
		orgKey = "department_id"
		orgID = departmentID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization context"})
		return
	}

	orgRequests, err = rc.requestUsecase.GetDraftedOrgRequests(c, orgKey, orgID)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orgRequests)
}

func (rc *requestController) AuthorizeOrgRequest(c *gin.Context) {
	fmt.Println("inside authorize controller")
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		fmt.Println(authUserID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	// get user id from param
	requestID := c.Param("id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "request ID is required"})
		return
	}

	requestObjID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = rc.requestUsecase.AuthorizeOrgRequest(c, authUserID, requestObjID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Request authorized successfully"})
}

func (rc *requestController) RejectRequest(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		fmt.Println(authUserID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	// get user id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = c.ShouldBind(&payload)
	if err != nil {
		fmt.Println("err binding", err)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = rc.requestUsecase.RejectRequest(c, authUserID, requestID, payload.RejectionReason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Request rejected successfully"})
}

func (rc *requestController) DeleteRequest(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		fmt.Println(authUserID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	// get user id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = rc.requestUsecase.DeleteRequest(c, authUserID, requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Request deleted successfully"})
}

func (rc *requestController) AcceptRequest(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		fmt.Println(authUserID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	// get user id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = rc.requestUsecase.AcceptRequest(c, authUserID, requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Request accepted successfully"})
}

func (rc *requestController) SendRequest(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		fmt.Println(authUserID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	// get user id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = rc.requestUsecase.SendRequest(c, authUserID, requestID)
	if err != nil {

		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Request sent successfully"})
}

func (rc *requestController) DeclineOrgRequest(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)
	if err != nil {
		fmt.Println(authUserID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	// get user id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = rc.requestUsecase.DeclineOrgRequest(c, authUserID, requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Request declined successfully"})
}

func (rc *requestController) GetValidatedRequests(c *gin.Context) {
	requests, err := rc.requestUsecase.GetValidatedRequests(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, requests)
}

func (rc *requestController) GetApprovedRequests(c *gin.Context) {
	fmt.Println("inside get approved requests")
	requests, err := rc.requestUsecase.GetApprovedRequests(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, requests)
}

func (rc *requestController) GetAcceptedRequests(c *gin.Context) {
	requests, err := rc.requestUsecase.GetAcceptedRequests(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, requests)
}

func (rc *requestController) GetDeclinedRequests(c *gin.Context) {
	requests, err := rc.requestUsecase.GetDeclinedRequests(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, requests)
}

func (rc *requestController) LockRequest(c *gin.Context) {

	authUserID, err := utils.GetUserID(c)
	fmt.Println("----------------------------------------------------------------")
	fmt.Println(authUserID)
	fmt.Println("----------------------------------------------------------------")

	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	// get request id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = rc.requestUsecase.LockRequest(c, authUserID, requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "request lock successful"})
}

func (rc *requestController) UnLockRequest(c *gin.Context) {
	authUserID, err := utils.GetUserID(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	// get request id from param
	requestIDStr := c.Param("id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "request ID is required"})
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	err = rc.requestUsecase.UnLockRequest(c, authUserID, requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "request unlock successful"})
}
