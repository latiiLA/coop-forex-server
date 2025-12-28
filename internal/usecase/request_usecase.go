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

	forexRequest := model.Request{}

	copier.Copy(&forexRequest, request)
	if existingUser.Profile.BranchID != nil {
		forexRequest.BranchID = existingUser.Profile.BranchID
	}

	if existingUser.Profile.DepartmentID != nil {
		forexRequest.DepartmentID = existingUser.Profile.DepartmentID
	}

	forexRequest.RequestCode = utils.GenerateRequestCode()
	forexRequest.CreatedAt = time.Now()
	forexRequest.UpdatedAt = time.Now()
	forexRequest.CreatedBy = authUserID
	forexRequest.RequestStatus = "Drafted"
	forexRequest.IsDeleted = false

	err = ru.requestRepository.Create(ctx, &forexRequest)

	if err != nil {
		fmt.Println("request creation error", err)
		return errors.New("internal server error")
	}

	return nil
}

func (ru *requestUsecase) UpdateRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, request *model.Request) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingUser, err := ru.userRepository.FindByID(ctx, authUserID)
	if err != nil {
		fmt.Println("Invalid userID", err)
		return errors.New("unauthorized access")
	}

	if (existingUser.Profile.BranchID != request.BranchID) || (existingUser.Profile.DepartmentID != request.DepartmentID) {
		return errors.New("unauthorized access")
	}

	forexRequest := model.Request{}

	copier.Copy(&forexRequest, request)

	forexRequest.UpdatedAt = time.Now()
	forexRequest.UpdatedBy = &authUserID

	err = ru.requestRepository.Update(ctx, requestID, &forexRequest)

	if err != nil {
		fmt.Println("request creation error", err)
		return errors.New("internal server error")
	}

	return nil
}

func (ru *requestUsecase) GetAllRequests(ctx context.Context) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()
	return ru.requestRepository.FindAll(ctx)
}

func (ru *requestUsecase) GetAuthorizedRequests(ctx context.Context) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()
	return ru.requestRepository.FindByRequestStatus(ctx, "Authorized")
}

func (ru *requestUsecase) GetValidatedRequests(ctx context.Context) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, "Validated")
}

func (ru *requestUsecase) GetApprovedRequests(ctx context.Context) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, "Approved")
}

func (ru *requestUsecase) GetAcceptedRequests(ctx context.Context) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, "Accepted")
}

func (ru *requestUsecase) GetDeclinedRequests(ctx context.Context) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, "Declined")
}

// Org Related Fetches

func (ru *requestUsecase) GetAllOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()
	fmt.Println("hello department", orgID)
	return ru.requestRepository.FindAllByOrgID(ctx, orgKey, orgID)
}

func (ru *requestUsecase) GetNewOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, "New")
}

func (ru *requestUsecase) GetApprovedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, "Approved")
}

func (ru *requestUsecase) GetAcceptedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, "Accepted")
}

func (ru *requestUsecase) GetDeclinedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, "Declined")
}

func (ru *requestUsecase) GetAuthorizedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, "Authorized")
}

func (ru *requestUsecase) GetDraftedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, "Drafted")
}

func (ru *requestUsecase) GetRejectedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, "Rejected")
}

func (ru *requestUsecase) GetRejectedRequests(ctx context.Context) ([]model.Request, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, "Rejected")
}

// Request operations

func (ru *requestUsecase) AuthorizeOrgRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	requestData, err := ru.requestRepository.FindByID(ctx, requestID)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	now := time.Now()
	requestData.AuthorizedAt = &now
	requestData.AuthorizedBy = &authUserID
	requestData.RequestStatus = "Authorized"

	// First update DB
	err = ru.requestRepository.Update(ctx, requestID, requestData)
	if err != nil {
		return err
	}

	// Then send email (fail-safe)
	subject := "New Foreign Currency Request"
	body := fmt.Sprintf("A new fcy request with request code %s has been created at %s.", requestData.RequestCode, now.Format("2006-01-02 15:04:05"))

	go func() {
		err := utils.SendEmail(configs.MailRequestAuthorizedTo, configs.MailRequestAuthorizedBcc, configs.MailRequestApprovedBcc, subject, body, *requestData)
		if err != nil {
			log.Warnf("Failed to send email for request %s: %v", requestID.Hex(), err)
		}
	}()

	return nil
}

func (ru *requestUsecase) ValidateRequest(ctx context.Context, authUserID primitive.ObjectID, request_id primitive.ObjectID, validated_account_currency_id primitive.ObjectID, request *model.RequestValidationDTO) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	requestData, err := ru.requestRepository.FindByID(ctx, request_id)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	if requestData.LockedBy != nil && *requestData.LockedBy != authUserID {
		return errors.New("request is already locked by another user")
	}

	requestData.ValidatedBy = &authUserID
	now := time.Now()
	requestData.ValidatedAt = &now
	requestData.ValidatedCurrentBalance = &request.ValidatedCurrentBalance
	requestData.ValidatedAverageDeposit = &request.ValidatedAverageDeposit
	requestData.ValidatedAccountCurrencyID = &validated_account_currency_id
	requestData.RequestStatus = "Validated"

	return ru.requestRepository.Validate(ctx, request_id, requestData)
}

func (ru *requestUsecase) ApproveRequest(ctx context.Context, authUserID primitive.ObjectID, request_id primitive.ObjectID, request *model.RequestApprovalDTO) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	requestData, err := ru.requestRepository.FindByID(ctx, request_id)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	if requestData.LockedBy != nil && *requestData.LockedBy != authUserID {
		return errors.New("request is already locked by another user")
	}

	requestData.ApprovedBy = &authUserID
	now := time.Now()
	requestData.ApprovedAt = &now
	requestData.ApprovedCurrencyIDs = &request.ApprovedCurrencyIDs
	requestData.ApprovedAmounts = &request.ApprovedAmounts
	requestData.RequestStatus = "Approved"

	return ru.requestRepository.Update(ctx, request_id, requestData)
}

func (ru *requestUsecase) DeleteRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	requestData, err := ru.requestRepository.FindByID(ctx, requestID)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	now := time.Now()
	requestData.DeletedAt = &now
	requestData.DeletedBy = &authUserID
	requestData.IsDeleted = true
	requestData.RequestStatus = "Deleted"

	return ru.requestRepository.Update(ctx, requestID, requestData)
}

func (ru *requestUsecase) SendRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	requestData, err := ru.requestRepository.FindByID(ctx, requestID)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	now := time.Now()
	requestData.RequestedAt = &now
	requestData.RequestedBy = &authUserID
	requestData.RequestStatus = "New"

	// First update DB
	err = ru.requestRepository.Update(ctx, requestID, requestData)
	if err != nil {
		return err
	}

	// Then send email (fail-safe)
	subject := "New Foreign Currency Request"
	body := fmt.Sprintf("A new fcy request with request code %s has been created at %s.", requestData.RequestCode, now.Format("2006-01-02 15:04:05"))

	go func() {
		err := utils.SendAcknowledgementEmail(configs.MailRequestSentTo, configs.MailRequestSentCc, configs.MailRequestSentBcc, subject, body, *requestData)
		if err != nil {
			log.Warnf("Failed to send email for request %s: %v", requestID.Hex(), err)
		}
	}()

	fmt.Print(configs.MailRequestSentTo, configs.MailRequestSentCc, configs.MailRequestSentBcc, configs.MailUsername)

	return nil
}

func (ru *requestUsecase) DeclineOrgRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	requestData, err := ru.requestRepository.FindByID(ctx, requestID)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	now := time.Now()
	requestData.DeclinedAt = &now
	requestData.DeclinedBy = &authUserID
	requestData.RequestStatus = "Declined"

	// // Then send email (fail-safe)
	// subject := "Foreign Currency Request Rejected"
	// body := fmt.Sprintf("A request with code %s has been rejected at %s.", requestID.Hex(), now.Format("2006-01-02 15:04:05"))

	// go func() {
	// 	err := utils.SendEmail(configs.MailReciever, subject, body, *requestData)
	// 	if err != nil {
	// 		log.Warnf("Failed to send rejection email for request %s: %v", requestID.Hex(), err)
	// 	}
	// }()

	return ru.requestRepository.Update(ctx, requestID, requestData)
}

func (ru *requestUsecase) RejectRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, rejection_reason string) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	requestData, err := ru.requestRepository.FindByID(ctx, requestID)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	if requestData.LockedBy != nil && *requestData.LockedBy != authUserID {
		return errors.New("request is already locked by another user")
	}

	now := time.Now()
	requestData.RejectedAt = &now
	requestData.RejectedBy = &authUserID
	requestData.RejectionReason = rejection_reason
	requestData.RequestStatus = "Rejected"

	return ru.requestRepository.Update(ctx, requestID, requestData)
}

func (ru *requestUsecase) LockRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	requestData, err := ru.requestRepository.FindByID(ctx, requestID)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	// If already locked by someone else
	if requestData.LockedBy != nil && *requestData.LockedBy != authUserID {
		return errors.New("request is already locked by another user")
	}

	now := time.Now()
	expiry := now.Add(15 * time.Minute)
	requestData.LockedAt = &now
	requestData.LockedBy = &authUserID
	requestData.LockExpiresAt = &expiry

	return ru.requestRepository.Update(ctx, requestID, requestData)
}

func (ru *requestUsecase) UnLockRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	requestData, err := ru.requestRepository.FindByID(ctx, requestID)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	if requestData.LockedBy != nil && *requestData.LockedBy != authUserID {
		return errors.New("request is already locked by another user")
	}

	requestData.LockedAt = nil
	requestData.LockedBy = nil
	requestData.LockExpiresAt = nil

	fmt.Println("unlocked", requestData.LockedBy)

	return ru.requestRepository.Update(ctx, requestID, requestData)
}

func (ru *requestUsecase) AcceptRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	requestData, err := ru.requestRepository.FindByID(ctx, requestID)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
	}

	now := time.Now()
	requestData.AcceptedAt = &now
	requestData.AcceptedBy = &authUserID
	requestData.RequestStatus = "Accepted"

	return ru.requestRepository.Update(ctx, requestID, requestData)
}
