package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/copier"
	"github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestUsecase interface {
	AddRequest(ctx context.Context, authUserID primitive.ObjectID, request *model.Request) error
	UpdateRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, request *model.Request) error
	GetAllRequests(ctx context.Context) ([]model.Request, error)
	ValidateRequest(ctx context.Context, authUserID primitive.ObjectID, request_id primitive.ObjectID, validated_currency_id primitive.ObjectID, request *model.RequestValidationDTO) error
	ApproveRequest(ctx context.Context, authUserID primitive.ObjectID, request_id primitive.ObjectID, request *model.RequestApprovalDTO) error
	GetAllOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error)
	GetAuthorizedRequests(ctx context.Context) ([]model.Request, error)
	GetNewOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error)
	GetApprovedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error)
	GetRejectedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error)
	GetRejectedRequests(ctx context.Context) ([]model.Request, error)
	GetAcceptedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error)
	GetDeclinedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error)
	GetAuthorizedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error)
	GetDraftedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error)

	AuthorizeOrgRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error
	RejectRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, rejection_reason string) error
	LockRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error
	UnLockRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error

	DeleteRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error
	AcceptRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error
	SendRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error
	DeclineOrgRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error

	GetValidatedRequests(c context.Context) ([]model.Request, error)
	GetApprovedRequests(c context.Context) ([]model.Request, error)
	GetAcceptedRequests(c context.Context) ([]model.Request, error)
	GetDeclinedRequests(c context.Context) ([]model.Request, error)
}

type requestUsecase struct {
	requestRepository model.RequestRepository
	userRepository    model.UserRepository
	contextTimeout    time.Duration
}

func NewRequestUsecase(requestRepository model.RequestRepository, userRepository model.UserRepository, timeout time.Duration) RequestUsecase {
	return &requestUsecase{
		requestRepository: requestRepository,
		userRepository:    userRepository,
		contextTimeout:    timeout,
	}
}

func (ru *requestUsecase) AddRequest(ctx context.Context, authUserID primitive.ObjectID, request *model.Request) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingUser, err := ru.userRepository.FindByID(ctx, authUserID)
	if err != nil {
		fmt.Println("Invalid userID", err)
		return errors.New("unauthorized access")
	}

	// forexRequest := model.Request{}

	// copier.Copy(&forexRequest, request)
	if existingUser.Profile.BranchID != nil {
		request.BranchID = existingUser.Profile.BranchID
	}

	if existingUser.Profile.DepartmentID != nil {
		request.DepartmentID = existingUser.Profile.DepartmentID
	}

	request.RequestCode = utils.GenerateRequestCode()
	request.CreatedAt = time.Now()
	request.UpdatedAt = time.Now()
	request.CreatedBy = authUserID
	request.RequestStatus = "Drafted"
	request.IsDeleted = false

	err = ru.requestRepository.Create(ctx, request)

	if err != nil {
		fmt.Println("request creation error", err)
		return errors.New("internal server error")
	}

	return nil
}

func (ru *requestUsecase) UpdateRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, request *model.Request) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	// 1. Get user information
	existingUser, err := ru.userRepository.FindByID(ctx, authUserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", authUserID).Error("Failed to find user")
		return errors.New("unauthorized access")
	}

	// 2. Get existing request
	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil {
		logrus.WithError(err).WithField("requestID", requestID).Error("Failed to find request")
		return errors.New("request doesn't exist")
	}

	// 3. Authorization check - fixed logical condition
	if existingUser.Profile.BranchID != nil && existingRequest.BranchID != nil &&
		*existingUser.Profile.BranchID != *existingRequest.BranchID {
		logrus.WithFields(logrus.Fields{
			"userBranchID":    existingUser.Profile.BranchID,
			"requestBranchID": existingRequest.BranchID,
			"userID":          authUserID,
			"requestID":       requestID,
		}).Warn("Unauthorized branch access attempt")
		return errors.New("unauthorized access")
	}

	if existingUser.Profile.DepartmentID != nil && existingRequest.DepartmentID != nil &&
		*existingUser.Profile.DepartmentID != *existingRequest.DepartmentID {
		logrus.WithFields(logrus.Fields{
			"userDeptID":    existingUser.Profile.DepartmentID,
			"requestDeptID": existingRequest.DepartmentID,
			"userID":        authUserID,
			"requestID":     requestID,
		}).Warn("Unauthorized department access attempt")
		return errors.New("unauthorized access")
	}

	// 4. Create update object
	forexRequest := model.RequestUpdate{}

	// Copy existing request data
	copier.Copy(&forexRequest, existingRequest)

	// 5. Apply updates
	forexRequest.ApplicantName = request.ApplicantName
	forexRequest.ApplicantAccountNumber = request.ApplicantAccountNumber
	forexRequest.AverageDeposit = request.AverageDeposit
	forexRequest.TotalFcyGenerated = request.TotalFcyGenerated
	forexRequest.CurrentFcyPerformance = request.CurrentFcyPerformance
	forexRequest.FcyRequestedAmount = request.FcyRequestedAmount
	forexRequest.AccountsToDeduct = request.AccountsToDeduct
	forexRequest.FcyAcceptanceMode = request.FcyAcceptanceMode
	forexRequest.CardAssociatedAccount = request.CardAssociatedAccount
	forexRequest.BranchRecommendation = request.BranchRecommendation
	forexRequest.TravelPurposeID = request.TravelPurposeID
	forexRequest.TravelCountryID = request.TravelCountryID
	forexRequest.RequestingAs = request.RequestingAs
	forexRequest.AccountCurrencyID = request.AccountCurrencyID
	forexRequest.FcyRequestedID = request.FcyRequestedID

	if request.VisaAttachment != nil && request.PassportAttachment != &primitive.NilObjectID {
		forexRequest.FcyRequestedID = request.FcyRequestedID
	}
	if request.VisaAttachment != nil && request.TicketAttachment != &primitive.NilObjectID {
		forexRequest.TicketAttachment = request.TicketAttachment
	}
	if request.VisaAttachment != nil && *request.EducationLoaAttachment != primitive.NilObjectID {
		forexRequest.VisaAttachment = request.VisaAttachment

	}
	if request.EducationLoaAttachment != nil && *request.EducationLoaAttachment != primitive.NilObjectID {
		forexRequest.EducationLoaAttachment = request.EducationLoaAttachment
	}
	if request.BusinessLicenseAttachment != nil && *request.BusinessLicenseAttachment != primitive.NilObjectID {
		forexRequest.BusinessLicenseAttachment = request.BusinessLicenseAttachment
	}
	if request.BusinessSupportingAttachment != nil && *request.BusinessLicenseAttachment != primitive.NilObjectID {
		forexRequest.BusinessSupportingAttachment = request.BusinessSupportingAttachment
	}

	forexRequest.UpdatedAt = time.Now()
	forexRequest.UpdatedBy = &authUserID

	// 6. Set user context if needed
	if existingUser.Profile.BranchID != nil {
		forexRequest.BranchID = existingUser.Profile.BranchID
	}
	if existingUser.Profile.DepartmentID != nil {
		forexRequest.DepartmentID = existingUser.Profile.DepartmentID
	}

	if err := ru.requestRepository.Update(ctx, requestID, &forexRequest); err != nil {
		logrus.WithError(err).WithField("requestID", requestID).Error("Failed to update request")
		return fmt.Errorf("failed to update request: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"requestID": requestID,
		"userID":    authUserID,
	}).Info("Request updated successfully")

	return nil
}

func (ru *requestUsecase) GetAllRequests(ctx context.Context) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()
	return ru.requestRepository.FindAll(ctx, true)
}

func (ru *requestUsecase) GetAuthorizedRequests(ctx context.Context) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()
	return ru.requestRepository.FindByRequestStatus(ctx, "Authorized", true)
}

func (ru *requestUsecase) GetValidatedRequests(ctx context.Context) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, "Validated", true)
}

func (ru *requestUsecase) GetApprovedRequests(ctx context.Context) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, "Approved", true)
}

func (ru *requestUsecase) GetAcceptedRequests(ctx context.Context) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, "Accepted", true)
}

func (ru *requestUsecase) GetDeclinedRequests(ctx context.Context) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, "Declined", true)
}

// Org Related Fetches

func (ru *requestUsecase) GetAllOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()
	fmt.Println("hello department", orgID)
	return ru.requestRepository.FindAllByOrgID(ctx, orgKey, orgID, true)
}

func (ru *requestUsecase) GetNewOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, "New", true)
}

func (ru *requestUsecase) GetApprovedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, "Approved", true)
}

func (ru *requestUsecase) GetAcceptedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, "Accepted", true)
}

func (ru *requestUsecase) GetDeclinedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, "Declined", true)
}

func (ru *requestUsecase) GetAuthorizedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, "Authorized", true)
}

func (ru *requestUsecase) GetDraftedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, "Drafted", true)
}

func (ru *requestUsecase) GetRejectedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, "Rejected", true)
}

func (ru *requestUsecase) GetRejectedRequests(ctx context.Context) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, "Rejected", true)
}

// Request operations

func (ru *requestUsecase) AuthorizeOrgRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	// 4. Create update object
	forexRequest := model.RequestUpdate{}

	// Copy existing request data
	copier.Copy(&forexRequest, existingRequest)

	now := time.Now()
	forexRequest.AuthorizedAt = &now
	forexRequest.AuthorizedBy = &authUserID
	forexRequest.RequestStatus = "Authorized"

	// Fetch sender
	sender, err := ru.userRepository.FindByID(ctx, authUserID)
	if err != nil {
		return err
	}

	// First update DB
	err = ru.requestRepository.Update(ctx, requestID, &forexRequest)
	if err != nil {
		return err
	}

	to := append([]string{}, configs.MailRequestAuthorizedTo...)
	cc := append([]string{}, configs.MailRequestAuthorizedCc...)
	bcc := append([]string{}, configs.MailRequestAuthorizedBcc...)

	if sender.Profile.Branch != nil && sender.Profile.Branch.Email != "" {
		to = append(to, sender.Profile.Branch.Email)
	}

	if sender.Profile != nil && sender.Profile.Email != "" {
		to = append(to, sender.Profile.Email)
	}

	// Then send email (fail-safe)
	subject := "New Foreign Currency Request"
	body := fmt.Sprintf("A new fcy request with request code %s has been created at %s.", forexRequest.RequestCode, now.Format("2006-01-02 15:04:05"))

	go func() {
		err := utils.SendEmail(to, cc, bcc, subject, body, *existingRequest)
		if err != nil {
			log.Warnf("Failed to send email for request %s: %v", requestID.Hex(), err)
		}
	}()

	return nil
}

func (ru *requestUsecase) ValidateRequest(ctx context.Context, authUserID primitive.ObjectID, request_id primitive.ObjectID, validated_account_currency_id primitive.ObjectID, request *model.RequestValidationDTO) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	forexRequest, err := ru.requestRepository.FindByID(ctx, request_id, false)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	if forexRequest.LockedBy != nil && *forexRequest.LockedBy != authUserID {
		return errors.New("request is already locked by another user")
	}

	forexRequest.ValidatedBy = &authUserID
	now := time.Now()
	forexRequest.ValidatedAt = &now
	forexRequest.ValidatedCurrentBalance = &request.ValidatedCurrentBalance
	forexRequest.ValidatedAverageDeposit = &request.ValidatedAverageDeposit
	forexRequest.ValidatedAccountCurrencyID = &validated_account_currency_id
	forexRequest.RequestStatus = "Validated"

	return ru.requestRepository.Validate(ctx, request_id, forexRequest)
}

func (ru *requestUsecase) ApproveRequest(ctx context.Context, authUserID primitive.ObjectID, request_id primitive.ObjectID, request *model.RequestApprovalDTO) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, request_id, false)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	// Copy the data
	forexRequest := model.RequestUpdate{}
	copier.Copy(&forexRequest, existingRequest)

	if forexRequest.LockedBy != nil && *forexRequest.LockedBy != authUserID {
		return errors.New("request is already locked by another user")
	}

	forexRequest.ApprovedBy = &authUserID
	now := time.Now()
	forexRequest.ApprovedAt = &now
	forexRequest.ApprovedCurrencyIDs = &request.ApprovedCurrencyIDs
	forexRequest.ApprovedAmounts = &request.ApprovedAmounts
	forexRequest.RequestStatus = "Approved"

	return ru.requestRepository.Update(ctx, request_id, &forexRequest)
}

func (ru *requestUsecase) DeleteRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	// Copy the data
	forexRequest := model.RequestUpdate{}
	copier.Copy(&forexRequest, existingRequest)

	now := time.Now()
	forexRequest.DeletedAt = &now
	forexRequest.DeletedBy = &authUserID
	forexRequest.IsDeleted = true
	forexRequest.RequestStatus = "Deleted"

	return ru.requestRepository.Update(ctx, requestID, &forexRequest)
}

func (ru *requestUsecase) SendRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	// Fetch sender
	sender, err := ru.userRepository.FindByID(ctx, authUserID)
	if err != nil {
		return err
	}

	// Copy the data
	forexRequest := model.RequestUpdate{}
	copier.Copy(&forexRequest, existingRequest)

	now := time.Now()
	forexRequest.RequestedAt = &now
	forexRequest.RequestedBy = &authUserID
	forexRequest.RequestStatus = "New"

	// First update DB
	err = ru.requestRepository.Update(ctx, requestID, &forexRequest)
	if err != nil {
		return err
	}

	to := append([]string{}, configs.MailRequestSentTo...)
	cc := append([]string{}, configs.MailRequestSentCc...)
	bcc := append([]string{}, configs.MailRequestSentBcc...)

	if sender.Profile.Branch != nil && sender.Profile.Branch.Email != "" {
		to = append(to, sender.Profile.Branch.Email)
	}

	if sender.Profile != nil && sender.Profile.Email != "" {
		to = append(to, sender.Profile.Email)
	}

	// Then send email (fail-safe)
	subject := "New Foreign Currency Request"
	body := fmt.Sprintf("A new fcy request with request code %s has been created at %s.", forexRequest.RequestCode, now.Format("2006-01-02 15:04:05"))

	// Send email async (fail-safe)
	go func() {
		err := utils.SendAcknowledgementEmail(
			to,
			cc,
			bcc,
			subject,
			body,
			*existingRequest,
		)
		if err != nil {
			log.Warnf(
				"Failed to send email for request %s: %v",
				requestID.Hex(),
				err,
			)
		}
	}()

	return nil
}

func (ru *requestUsecase) DeclineOrgRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	// Copy the data
	forexRequest := model.RequestUpdate{}
	copier.Copy(&forexRequest, existingRequest)

	now := time.Now()
	forexRequest.DeclinedAt = &now
	forexRequest.DeclinedBy = &authUserID
	forexRequest.RequestStatus = "Declined"

	// // Then send email (fail-safe)
	// subject := "Foreign Currency Request Rejected"
	// body := fmt.Sprintf("A request with code %s has been rejected at %s.", requestID.Hex(), now.Format("2006-01-02 15:04:05"))

	// go func() {
	// 	err := utils.SendEmail(configs.MailReciever, subject, body, *forexRequest)
	// 	if err != nil {
	// 		log.Warnf("Failed to send rejection email for request %s: %v", requestID.Hex(), err)
	// 	}
	// }()

	return ru.requestRepository.Update(ctx, requestID, &forexRequest)
}

func (ru *requestUsecase) RejectRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, rejection_reason string) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	// Copy the data
	forexRequest := model.RequestUpdate{}
	copier.Copy(&forexRequest, existingRequest)

	if forexRequest.LockedBy != nil && *forexRequest.LockedBy != authUserID {
		return errors.New("request is already locked by another user")
	}

	now := time.Now()
	forexRequest.RejectedAt = &now
	forexRequest.RejectedBy = &authUserID
	forexRequest.RejectionReason = rejection_reason
	forexRequest.RequestStatus = "Rejected"

	return ru.requestRepository.Update(ctx, requestID, &forexRequest)
}

func (ru *requestUsecase) LockRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	// Copy the data
	forexRequest := model.RequestUpdate{}
	copier.Copy(&forexRequest, existingRequest)

	// If already locked by someone else
	if forexRequest.LockedBy != nil && *forexRequest.LockedBy != authUserID {
		return errors.New("request is already locked by another user")
	}

	now := time.Now()
	expiry := now.Add(15 * time.Minute)
	forexRequest.LockedAt = &now
	forexRequest.LockedBy = &authUserID
	forexRequest.LockExpiresAt = &expiry

	return ru.requestRepository.Update(ctx, requestID, &forexRequest)
}

func (ru *requestUsecase) UnLockRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	// Copy the data
	forexRequest := model.RequestUpdate{}
	copier.Copy(&forexRequest, existingRequest)

	if forexRequest.LockedBy != nil && *forexRequest.LockedBy != authUserID {
		return errors.New("request is already locked by another user")
	}

	forexRequest.LockedAt = nil
	forexRequest.LockedBy = nil
	forexRequest.LockExpiresAt = nil

	fmt.Println("unlocked", forexRequest.LockedBy)

	return ru.requestRepository.Update(ctx, requestID, &forexRequest)
}

func (ru *requestUsecase) AcceptRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	// Copy the data
	forexRequest := model.RequestUpdate{}
	copier.Copy(&forexRequest, existingRequest)

	now := time.Now()
	forexRequest.AcceptedAt = &now
	forexRequest.AcceptedBy = &authUserID
	forexRequest.RequestStatus = "Accepted"

	return ru.requestRepository.Update(ctx, requestID, &forexRequest)
}
