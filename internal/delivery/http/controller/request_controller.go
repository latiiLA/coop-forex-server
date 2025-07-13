package controller

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"github.com/latiiLA/coop-forex-server/internal/usecase"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestController interface {
	AddRequest(c *gin.Context)
	GetAllRequests(c *gin.Context)
	ValidateRequest(c *gin.Context)
	ApproveRequest(c *gin.Context)
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
	userID, err := utils.GetUserID(c)
	if err != nil {
		fmt.Println(userID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	fmt.Println("request controller'")
	var request model.RequestDTO

	err = c.ShouldBind(&request)
	if err != nil {
		fmt.Println("err binding", err)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

	fmt.Println("after unmarshaling", request)
	fmt.Printf("%+v\n", request.AccountsToDeduct)

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid multipart form data"})
		return
	}

	request.AccountsToDeduct = form.Value["accounts_to_deduct"]
	fmt.Println(request.AccountsToDeduct, "accounts")

	// Convert string fields to ObjectID
	travelPurposeObjID, err := primitive.ObjectIDFromHex(request.TravelPurposeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid Travel Purpose ID"})
		return
	}

	travelCountryObjID, err := primitive.ObjectIDFromHex(request.TravelCountryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid Travel Country ID"})
		return
	}

	// requestingAsObjID, err := primitive.ObjectIDFromHex(request.RequestingAs)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid Requesting As ID"})
	// 	return
	// }

	accountCurrencyObjID, err := primitive.ObjectIDFromHex(request.AccountCurrencyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid Currency Account ID"})
		return
	}

	fcyRequestedObjID, err := primitive.ObjectIDFromHex(request.FcyRequestedID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid Currency ID"})
		return
	}

	deposit, err := strconv.ParseFloat(request.AverageDeposit, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid Average Deposit value"})
		return
	}

	totalFcy, err := strconv.ParseFloat(request.TotalFcyGenerated, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid Total Fccy Performance value"})
		return
	}

	currentFcy, err := strconv.ParseFloat(request.CurrentFcyPerformance, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid Current Fcy Performance value"})
		return
	}

	fcyAmount, err := strconv.ParseFloat(request.FcyRequestedAmount, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "Invalid Fcy Amount value"})
		return
	}

	uploadPath := configs.FileUploadPath
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
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
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload education file"})
			return
		}
	}
	ticket, err := c.FormFile("ticket_attachment")
	if err == nil {
		ticketID, err = rc.fileUsecase.AddFile(c, ticket, "ticket")
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload ticket file"})
			return
		}
	}
	education, err := c.FormFile("education_loa_attachment")
	if err == nil {
		educationID, err = rc.fileUsecase.AddFile(c, education, "education")
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload education file"})
			return
		}
	}
	businessLicense, err := c.FormFile("business_license_attachment")
	if err == nil {
		businessLicenseID, err = rc.fileUsecase.AddFile(c, businessLicense, "business")
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload business license file"})
			return
		}
	}
	supportLetter, err := c.FormFile("business_supporting_letter")
	if err == nil {
		supportingLetterID, err = rc.fileUsecase.AddFile(c, supportLetter, "support")
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to upload education file"})
			return
		}
	}

	if len(request.AccountsToDeduct) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "accounts_to_deduct is required"})
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

		PassportAttachment:           *passportID,
		TicketAttachment:             *ticketID,
		BusinessLicenseAttachment:    businessLicenseID,
		EducationLoaAttachment:       educationID,
		BusinessSupportingAttachment: supportingLetterID,
	}

	err = rc.requestUsecase.AddRequest(c, userID, &fcyRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Request created successfully"})
}

func (rc *requestController) GetAllRequests(c *gin.Context) {
	roles, err := rc.requestUsecase.GetAllRequests(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, roles)
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

func (rc *requestController) ApproveRequest(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		fmt.Println(userID, c)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: err.Error()})
		return
	}

	var request model.RequestApprovalDTO

	err = c.ShouldBind(&request)
	if err != nil {
		fmt.Println("err binding", err)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: err.Error()})
		return
	}

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

	var objectIDs []primitive.ObjectID
	for _, idStr := range request.ApprovedCurrencies {
		objID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "Invalid currency ID format",
				"invalid_id": idStr,
			})
			return
		}
		objectIDs = append(objectIDs, objID)
	}

	parsed := model.ParsedApproval{
		ApprovedCurrencyIDs: objectIDs,
		ApprovedAmounts:     request.ApprovedAmounts,
	}

	err = rc.requestUsecase.ApproveRequest(c, userID, requestObjID, parsed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "Request approved successfully"})

}
