package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/common"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestUsecase interface {
	AddRequest(ctx context.Context, authUserID primitive.ObjectID, request *model.Request) error
	UpdateRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, request *model.Request) error
	GetAllRequests(ctx context.Context, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	ValidateRequest(ctx context.Context, authUserID primitive.ObjectID, request_id primitive.ObjectID, validated_currency_id primitive.ObjectID, request *model.RequestValidationDTO) error
	ApproveRequest(ctx context.Context, authUserID primitive.ObjectID, request_id primitive.ObjectID, request *model.RequestApprovalDTO) error
	GetAllOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	GetAuthorizedRequests(ctx context.Context, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	GetNewOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	GetApprovedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	GetRejectedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	GetRejectedRequests(ctx context.Context, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	GetAcceptedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	GetDeclinedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	GetAuthorizedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	GetDraftedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, viewerID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	GetRequestQueue(ctx context.Context, status, scope string, orgKey string, orgID, viewerID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error)

	AuthorizeOrgRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error
	RejectRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, rejection_reason string) error
	LockRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, assigneeID *primitive.ObjectID, actorPerms []string) (*model.AssignRequestResult, error)
	UnLockRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, actorPerms []string) error
	AssignRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, assigneeID *primitive.ObjectID, actorPerms []string) (*model.AssignRequestResult, error)
	UnassignRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, actorPerms []string) error

	DeleteRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error
	AcceptRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, request *model.RequestAcceptanceDTO, processingBranchID primitive.ObjectID) error
	SendRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error
	DeclineOrgRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error

	GetValidatedRequests(c context.Context, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	GetApprovedRequests(c context.Context, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	GetAcceptedRequests(c context.Context, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	GetDeclinedRequests(c context.Context, pageQuery model.PageQuery) (*model.PaginatedRequests, error)
	GetRequestStatusSummary(ctx context.Context, orgKey string, orgID primitive.ObjectID) (*model.RequestStatusSummary, error)

	GenerateRequestReport(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) (pdf []byte, filename string, receiptNumber string, err error)
	ValidateReceipt(ctx context.Context, receiptNumber string) (*model.Request, map[string]any, error)
}

type requestUsecase struct {
	requestRepository  model.RequestRepository
	userRepository     model.UserRepository
	receiptRepository  model.ReceiptRepository
	currencyRepository model.CurrencyRepository
	activityLog        model.ActivityLogUsecase
	contextTimeout     time.Duration
}

func NewRequestUsecase(
	requestRepository model.RequestRepository,
	userRepository model.UserRepository,
	receiptRepository model.ReceiptRepository,
	currencyRepository model.CurrencyRepository,
	activityLog model.ActivityLogUsecase,
	timeout time.Duration,
) RequestUsecase {
	return &requestUsecase{
		requestRepository:  requestRepository,
		userRepository:     userRepository,
		receiptRepository:  receiptRepository,
		currencyRepository: currencyRepository,
		activityLog:        activityLog,
		contextTimeout:     timeout,
	}
}

func hasPermission(perms []string, target string) bool {
	for _, p := range perms {
		if p == target {
			return true
		}
	}
	return false
}

func ensureRequestStatus(current model.RequestStatus, allowed ...model.RequestStatus) error {
	for _, status := range allowed {
		if current == status {
			return nil
		}
	}
	return common.ErrInvalidRequestTransition
}

func statusStrings(statuses ...model.RequestStatus) []string {
	out := make([]string, 0, len(statuses))
	for _, s := range statuses {
		out = append(out, string(s))
	}
	return out
}

// ensureSameOrgOrCreator restricts org-scoped mutations to creator or same branch/department.
func ensureSameOrgOrCreator(user *model.User, request *model.Request) error {
	if user == nil || request == nil {
		return common.ErrUnauthorized
	}
	if request.CreatedBy == user.ID {
		return nil
	}
	if user.Profile == nil {
		return common.ErrUnauthorized
	}
	sameBranch := user.Profile.BranchID != nil &&
		request.BranchID != nil &&
		*user.Profile.BranchID == *request.BranchID
	sameDepartment := user.Profile.DepartmentID != nil &&
		request.DepartmentID != nil &&
		*user.Profile.DepartmentID == *request.DepartmentID
	if sameBranch || sameDepartment {
		return nil
	}
	return common.ErrUnauthorized
}

func userProfileEmail(u *model.User) string {
	if u == nil || u.Profile == nil {
		return ""
	}
	return u.Profile.Email
}

func branchEmail(u *model.User) string {
	if u == nil || u.Profile == nil || u.Profile.Branch == nil {
		return ""
	}
	return u.Profile.Branch.Email
}

func requestBranchEmail(req *model.Request) string {
	if req == nil || req.Branch == nil {
		return ""
	}
	return req.Branch.Email
}

// notifyRequestActivity emails stakeholders for a lifecycle event (async, fail-safe).
func (ru *requestUsecase) notifyRequestActivity(
	requestID primitive.ObjectID,
	event utils.RequestMailEvent,
	configuredTo, configuredCc, configuredBcc []string,
	summary string,
	details []string,
	extraTo ...string,
) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req, err := ru.requestRepository.FindByID(ctx, requestID, true)
		if err != nil || req == nil {
			log.Warnf("email skipped for %s: request load failed: %v", requestID.Hex(), err)
			return
		}

		to := append([]string{}, configuredTo...)
		cc := append([]string{}, configuredCc...)
		bcc := append([]string{}, configuredBcc...)
		to = append(to, extraTo...)
		to = append(to, requestBranchEmail(req))

		if !req.CreatedBy.IsZero() {
			if creator, err := ru.userRepository.FindByID(ctx, req.CreatedBy); err == nil {
				to = append(to, userProfileEmail(creator), branchEmail(creator))
			}
		}
		if assignee := ru.activeAssignee(req); assignee != nil {
			if user, err := ru.userRepository.FindByID(ctx, *assignee); err == nil {
				to = append(to, userProfileEmail(user))
			}
		}

		if err := utils.NotifyRequestActivity(to, cc, bcc, event, *req, summary, details); err != nil {
			log.Warnf("Failed to send %s email for request %s: %v", event, requestID.Hex(), err)
		}
	}()
}

func (ru *requestUsecase) ensureForeignCurrencies(ctx context.Context, currencyIDs []primitive.ObjectID) error {
	for _, currencyID := range currencyIDs {
		if currencyID.IsZero() {
			return common.ErrInvalidCurrency
		}

		currency, err := ru.currencyRepository.FindByID(ctx, currencyID)
		if err != nil || currency == nil {
			return common.ErrInvalidCurrency
		}
		if strings.EqualFold(strings.TrimSpace(currency.ShortCode), "ETB") {
			return common.ErrLocalCurrencyNotAllowed
		}
	}
	return nil
}

func (ru *requestUsecase) ensureAccountCurrencyIsETB(ctx context.Context, currencyID primitive.ObjectID) error {
	if currencyID.IsZero() {
		return common.ErrInvalidCurrency
	}

	currency, err := ru.currencyRepository.FindByID(ctx, currencyID)
	if err != nil || currency == nil {
		return common.ErrInvalidCurrency
	}
	if !strings.EqualFold(strings.TrimSpace(currency.ShortCode), "ETB") {
		return common.ErrAccountCurrencyMustBeETB
	}
	return nil
}

func (ru *requestUsecase) ensureBranchAuthorizer(ctx context.Context, authorizerID primitive.ObjectID, branchID *primitive.ObjectID) error {
	if authorizerID.IsZero() {
		return common.ErrAssigneeNotEligible
	}
	if branchID == nil || branchID.IsZero() {
		return common.ErrAssigneeNotEligible
	}
	authorizer, err := ru.userRepository.FindByID(ctx, authorizerID)
	if err != nil || authorizer == nil {
		return common.ErrAssigneeNotEligible
	}
	if authorizer.Status != model.StatusActive && authorizer.Status != model.StatusNew {
		return common.ErrAssigneeNotEligible
	}
	if !userHasApproveOrAuthorize(authorizer, "request:authorize") || !sameBranchAsRequest(authorizer, &model.Request{BranchID: branchID}) {
		return common.ErrAssigneeNotEligible
	}
	return nil
}

func (ru *requestUsecase) AddRequest(ctx context.Context, authUserID primitive.ObjectID, request *model.Request) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingUser, err := ru.userRepository.FindByID(ctx, authUserID)
	if err != nil {
		return common.ErrUnauthorized
	}

	if existingUser.Profile.BranchID != nil {
		request.BranchID = existingUser.Profile.BranchID
	}

	if existingUser.Profile.DepartmentID != nil {
		request.DepartmentID = existingUser.Profile.DepartmentID
	}

	if err := ru.ensureAccountCurrencyIsETB(ctx, request.AccountCurrencyID); err != nil {
		return err
	}
	if err := ru.ensureForeignCurrencies(ctx, []primitive.ObjectID{request.FcyRequestedID}); err != nil {
		return err
	}

	if request.AssignedTo == nil || request.AssignedTo.IsZero() {
		return common.ErrAssigneeNotEligible
	}
	if err := ru.ensureBranchAuthorizer(ctx, *request.AssignedTo, request.BranchID); err != nil {
		return err
	}
	nowAssign := time.Now()
	request.AssignedAt = &nowAssign

	if request.ID.IsZero() {
		request.ID = primitive.NewObjectID()
	}
	request.RequestCode = utils.GenerateRequestCode()
	request.CreatedAt = time.Now()
	request.UpdatedAt = time.Now()
	request.CreatedBy = authUserID
	request.RequestStatus = model.ReqStatusDrafted
	request.IsDeleted = false

	err = ru.requestRepository.Create(ctx, request)
	if err != nil {
		return err
	}

	entityID := request.ID
	LogActivity(ru.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionRequestCreate,
		EntityType:  model.EntityRequest,
		EntityID:    &entityID,
		Description: "Request created",
		ToState:     string(model.ReqStatusDrafted),
	})

	return nil
}

func (ru *requestUsecase) UpdateRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, request *model.Request) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	// 1. Get user information
	existingUser, err := ru.userRepository.FindByID(ctx, authUserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", authUserID).Error("Failed to find user")
		return common.ErrUnauthorized
	}

	// 2. Get existing request
	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil || existingRequest == nil {
		logrus.WithError(err).WithField("requestID", requestID).Error("Failed to find request")
		return common.ErrRequestNotFound
	}

	if existingRequest.RequestStatus != model.ReqStatusDrafted &&
		existingRequest.RequestStatus != model.ReqStatusRejected {
		return common.ErrInvalidRequestTransition
	}

	if err := ensureSameOrgOrCreator(existingUser, existingRequest); err != nil {
		logrus.WithFields(logrus.Fields{
			"userID":    authUserID,
			"requestID": requestID,
		}).Warn("Unauthorized update attempt")
		return err
	}

	if err := ru.ensureAccountCurrencyIsETB(ctx, request.AccountCurrencyID); err != nil {
		return err
	}
	if err := ru.ensureForeignCurrencies(ctx, []primitive.ObjectID{request.FcyRequestedID}); err != nil {
		return err
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

	if request.PassportAttachment != nil && *request.PassportAttachment != primitive.NilObjectID {
		forexRequest.PassportAttachment = request.PassportAttachment
	}
	if request.TicketAttachment != nil && *request.TicketAttachment != primitive.NilObjectID {
		forexRequest.TicketAttachment = request.TicketAttachment
	}
	if request.VisaAttachment != nil && *request.VisaAttachment != primitive.NilObjectID {
		forexRequest.VisaAttachment = request.VisaAttachment
	}
	if request.EducationLoaAttachment != nil && *request.EducationLoaAttachment != primitive.NilObjectID {
		forexRequest.EducationLoaAttachment = request.EducationLoaAttachment
	}
	if request.BusinessLicenseAttachment != nil && *request.BusinessLicenseAttachment != primitive.NilObjectID {
		forexRequest.BusinessLicenseAttachment = request.BusinessLicenseAttachment
	}
	if request.BusinessSupportingAttachment != nil && *request.BusinessSupportingAttachment != primitive.NilObjectID {
		forexRequest.BusinessSupportingAttachment = request.BusinessSupportingAttachment
	}
	if request.HealthLetterAttachment != nil && *request.HealthLetterAttachment != primitive.NilObjectID {
		forexRequest.HealthLetterAttachment = request.HealthLetterAttachment
	}

	forexRequest.UpdatedAt = time.Now()
	forexRequest.UpdatedBy = &authUserID

	// Revising a rejected request returns it to Drafted so it can be resent.
	expectedStatuses := statusStrings(model.ReqStatusDrafted, model.ReqStatusRejected)
	if existingRequest.RequestStatus == model.ReqStatusRejected {
		forexRequest.RequestStatus = string(model.ReqStatusDrafted)
		forexRequest.RejectionReason = ""
		forexRequest.RejectedAt = nil
		forexRequest.RejectedBy = nil
	}

	// 6. Set user context if needed
	if existingUser.Profile != nil && existingUser.Profile.BranchID != nil {
		forexRequest.BranchID = existingUser.Profile.BranchID
	}
	if existingUser.Profile != nil && existingUser.Profile.DepartmentID != nil {
		forexRequest.DepartmentID = existingUser.Profile.DepartmentID
	}

	if request.AssignedTo != nil && !request.AssignedTo.IsZero() {
		if err := ru.ensureBranchAuthorizer(ctx, *request.AssignedTo, forexRequest.BranchID); err != nil {
			return err
		}
		nowAssign := time.Now()
		forexRequest.AssignedTo = request.AssignedTo
		forexRequest.AssignedAt = &nowAssign
	}

	fromStatus := string(existingRequest.RequestStatus)
	if err := ru.requestRepository.UpdateIfStatus(ctx, requestID, expectedStatuses, &forexRequest); err != nil {
		logrus.WithError(err).WithField("requestID", requestID).Error("Failed to update request")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"requestID": requestID,
		"userID":    authUserID,
	}).Info("Request updated successfully")

	entityID := requestID
	LogActivity(ru.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionRequestUpdate,
		EntityType:  model.EntityRequest,
		EntityID:    &entityID,
		Description: "Request updated",
		FromState:   fromStatus,
		ToState:     fromStatus,
	})

	return nil
}

func (ru *requestUsecase) GetAllRequests(ctx context.Context, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindAll(ctx, true, pageQuery)
}

func (ru *requestUsecase) GetAuthorizedRequests(ctx context.Context, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, string(model.ReqStatusAuthorized), true, pageQuery, nil)
}

func (ru *requestUsecase) GetValidatedRequests(ctx context.Context, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, string(model.ReqStatusValidated), true, pageQuery, nil)
}

func (ru *requestUsecase) GetApprovedRequests(ctx context.Context, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, string(model.ReqStatusApproved), true, pageQuery, nil)
}

func (ru *requestUsecase) GetAcceptedRequests(ctx context.Context, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, string(model.ReqStatusAccepted), true, pageQuery, nil)
}

func (ru *requestUsecase) GetDeclinedRequests(ctx context.Context, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, string(model.ReqStatusDeclined), true, pageQuery, nil)
}

// Org Related Fetches
func (ru *requestUsecase) GetAllOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindAllByOrgID(ctx, orgKey, orgID, true, pageQuery)
}

func (ru *requestUsecase) GetNewOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, string(model.ReqStatusNew), true, pageQuery, nil)
}

func (ru *requestUsecase) GetApprovedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, string(model.ReqStatusApproved), true, pageQuery, nil)
}

func (ru *requestUsecase) GetAcceptedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, string(model.ReqStatusAccepted), true, pageQuery, nil)
}

func (ru *requestUsecase) GetDeclinedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, string(model.ReqStatusDeclined), true, pageQuery, nil)
}

func (ru *requestUsecase) GetAuthorizedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, string(model.ReqStatusAuthorized), true, pageQuery, nil)
}

func (ru *requestUsecase) GetDraftedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, viewerID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	return ru.GetRequestQueue(ctx, string(model.ReqStatusDrafted), "org", orgKey, orgID, viewerID, pageQuery)
}

func emptyQueue(pageQuery model.PageQuery) *model.PaginatedRequests {
	return &model.PaginatedRequests{
		Items: []model.Request{},
		Total: 0,
		Page:  pageQuery.Page,
		Limit: pageQuery.Limit,
	}
}

// GetRequestQueue returns a status queue for bank or org scope.
// Org without branch/department: Drafted → own drafts; other statuses → empty (no 400).
func (ru *requestUsecase) GetRequestQueue(ctx context.Context, status, scope string, orgKey string, orgID, viewerID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	switch scope {
	case "bank":
		// Authorized/Validated: only unassigned (pool) or assigned to this viewer.
		var assigneeViewer *primitive.ObjectID
		if status == string(model.ReqStatusAuthorized) || status == string(model.ReqStatusValidated) {
			assigneeViewer = &viewerID
		}
		return ru.requestRepository.FindByRequestStatus(ctx, status, true, pageQuery, assigneeViewer)
	case "org":
		if orgKey == "" || orgID.IsZero() {
			if status == string(model.ReqStatusDrafted) {
				return ru.requestRepository.FindByCreatorAndStatus(ctx, viewerID, status, true, pageQuery)
			}
			return emptyQueue(pageQuery), nil
		}
		// New: branch authorizer assignment — unassigned pool or assigned-to-me.
		var orgAssigneeViewer *primitive.ObjectID
		if status == string(model.ReqStatusNew) {
			orgAssigneeViewer = &viewerID
		}
		return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, status, true, pageQuery, orgAssigneeViewer)
	default:
		return nil, common.ErrInvalidRequest
	}
}

func (ru *requestUsecase) GetRejectedOrgRequests(ctx context.Context, orgKey string, orgID primitive.ObjectID, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindOrgByRequestStatus(ctx, orgID, orgKey, string(model.ReqStatusRejected), true, pageQuery, nil)
}

func (ru *requestUsecase) GetRejectedRequests(ctx context.Context, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.FindByRequestStatus(ctx, string(model.ReqStatusRejected), true, pageQuery, nil)
}

func (ru *requestUsecase) GetRequestStatusSummary(ctx context.Context, orgKey string, orgID primitive.ObjectID) (*model.RequestStatusSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	return ru.requestRepository.CountByStatus(ctx, orgKey, orgID)
}

// Request operations
func (ru *requestUsecase) AuthorizeOrgRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil || existingRequest == nil {
		return common.ErrRequestNotFound
	}

	if err := ensureRequestStatus(existingRequest.RequestStatus, model.ReqStatusNew); err != nil {
		return err
	}

	sender, err := ru.userRepository.FindByID(ctx, authUserID)
	if err != nil {
		return err
	}
	if err := ensureSameOrgOrCreator(sender, existingRequest); err != nil {
		return err
	}
	// Branch authorizer must be the assignee (assigned at initiate / self-claim).
	if err := ensureAssignedTo(ru.activeAssignee(existingRequest), authUserID); err != nil {
		return err
	}

	// 4. Create update object
	forexRequest := model.RequestUpdate{}

	// Copy existing request data
	copier.Copy(&forexRequest, existingRequest)

	now := time.Now()
	forexRequest.AuthorizedAt = &now
	forexRequest.AuthorizedBy = &authUserID
	forexRequest.RequestStatus = string(model.ReqStatusAuthorized)

	fromStatus := string(existingRequest.RequestStatus)
	err = ru.requestRepository.UpdateIfStatus(
		ctx,
		requestID,
		statusStrings(model.ReqStatusNew),
		&forexRequest,
	)
	if err != nil {
		return err
	}
	_ = ru.requestRepository.ClearAssignment(ctx, requestID)

	entityID := requestID
	LogActivity(ru.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionRequestAuthorize,
		EntityType:  model.EntityRequest,
		EntityID:    &entityID,
		Description: "Request authorized",
		FromState:   fromStatus,
		ToState:     string(model.ReqStatusAuthorized),
	})

	ru.notifyRequestActivity(
		requestID,
		utils.MailEventAuthorized,
		configs.MailRequestAuthorizedTo,
		configs.MailRequestAuthorizedCc,
		configs.MailRequestAuthorizedBcc,
		fmt.Sprintf("Request %s was authorized on %s.", forexRequest.RequestCode, now.Format("2006-01-02 15:04")),
		nil,
	)

	return nil
}

func ensureAssignedTo(assignedTo *primitive.ObjectID, authUserID primitive.ObjectID) error {
	if assignedTo == nil {
		return common.ErrRequestNotAssigned
	}
	if *assignedTo != authUserID {
		return common.ErrRequestIsLocked
	}
	return nil
}

func (ru *requestUsecase) activeAssignee(req *model.Request) *primitive.ObjectID {
	if req == nil || req.AssignedTo == nil {
		return nil
	}
	// Legacy timed locks: expired claims are free to take.
	if req.LockExpiresAt != nil && req.LockExpiresAt.Before(time.Now()) {
		return nil
	}
	return req.AssignedTo
}

func (ru *requestUsecase) ValidateRequest(ctx context.Context, authUserID primitive.ObjectID, request_id primitive.ObjectID, validated_account_currency_id primitive.ObjectID, request *model.RequestValidationDTO) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	forexRequest, err := ru.requestRepository.FindByID(ctx, request_id, false)
	if err != nil || forexRequest == nil {
		return common.ErrRequestNotFound
	}

	if err := ensureRequestStatus(forexRequest.RequestStatus, model.ReqStatusAuthorized, model.ReqStatusValidated); err != nil {
		return err
	}

	if err := ensureAssignedTo(ru.activeAssignee(forexRequest), authUserID); err != nil {
		return err
	}

	if err := ru.ensureAccountCurrencyIsETB(ctx, validated_account_currency_id); err != nil {
		return err
	}

	fromStatus := string(forexRequest.RequestStatus)
	expected := statusStrings(model.ReqStatusAuthorized, model.ReqStatusValidated)
	forexRequest.ValidatedBy = &authUserID
	now := time.Now()
	forexRequest.ValidatedAt = &now
	forexRequest.ValidatedCurrentBalance = &request.ValidatedCurrentBalance
	forexRequest.ValidatedAverageDeposit = &request.ValidatedAverageDeposit
	forexRequest.ValidatedAccountCurrencyID = &validated_account_currency_id
	forexRequest.RequestStatus = model.ReqStatusValidated

	if err := ru.requestRepository.ValidateIfStatus(ctx, request_id, expected, forexRequest); err != nil {
		return err
	}
	_ = ru.requestRepository.ClearAssignment(ctx, request_id)

	entityID := request_id
	LogActivity(ru.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionRequestValidate,
		EntityType:  model.EntityRequest,
		EntityID:    &entityID,
		Description: "Request validated",
		FromState:   fromStatus,
		ToState:     string(model.ReqStatusValidated),
	})

	ru.notifyRequestActivity(
		request_id,
		utils.MailEventValidated,
		configs.MailRequestValidatedTo,
		configs.MailRequestValidatedCc,
		configs.MailRequestValidatedBcc,
		fmt.Sprintf("Request %s was validated on %s.", forexRequest.RequestCode, now.Format("2006-01-02 15:04")),
		nil,
	)
	return nil
}

func (ru *requestUsecase) ApproveRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, request *model.RequestApprovalDTO) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil || existingRequest == nil {
		return common.ErrRequestNotFound
	}

	// Approve once: only from Validated (prevents rewriting approved amounts).
	if err := ensureRequestStatus(existingRequest.RequestStatus, model.ReqStatusValidated); err != nil {
		return err
	}

	if err := ru.ensureForeignCurrencies(ctx, request.ApprovedCurrencyIDs); err != nil {
		return err
	}

	// Copy the data
	forexRequest := model.RequestUpdate{}
	copier.Copy(&forexRequest, existingRequest)

	if err := ensureAssignedTo(ru.activeAssignee(existingRequest), authUserID); err != nil {
		return err
	}

	forexRequest.ApprovedBy = &authUserID
	now := time.Now()
	forexRequest.ApprovedAt = &now
	forexRequest.ApprovedCurrencyIDs = request.ApprovedCurrencyIDs
	forexRequest.ApprovedAmounts = request.ApprovedAmounts
	forexRequest.ApprovedAmountInCash = request.ApprovedAmountInCash
	forexRequest.ApprovedAmountInCard = request.ApprovedAmountInCard
	forexRequest.RequestStatus = string(model.ReqStatusApproved)

	fromStatus := string(existingRequest.RequestStatus)
	err = ru.requestRepository.UpdateIfStatus(
		ctx,
		requestID,
		statusStrings(model.ReqStatusValidated),
		&forexRequest,
	)
	if err != nil {
		return err
	}
	_ = ru.requestRepository.ClearAssignment(ctx, requestID)

	entityID := requestID
	LogActivity(ru.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionRequestApprove,
		EntityType:  model.EntityRequest,
		EntityID:    &entityID,
		Description: "Request approved",
		FromState:   fromStatus,
		ToState:     string(model.ReqStatusApproved),
	})

	ru.notifyRequestActivity(
		requestID,
		utils.MailEventApproved,
		configs.MailRequestApprovedTo,
		configs.MailRequestApprovedCc,
		configs.MailRequestApprovedBcc,
		fmt.Sprintf("Request %s was approved on %s.", forexRequest.RequestCode, now.Format("2006-01-02 15:04")),
		nil,
	)

	return nil
}

func (ru *requestUsecase) DeleteRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil || existingRequest == nil {
		return common.ErrRequestNotFound
	}

	if existingRequest.RequestStatus != model.ReqStatusDrafted {
		return common.ErrRequestCannotBeDeleted
	}

	existingUser, err := ru.userRepository.FindByID(ctx, authUserID)
	if err != nil {
		return common.ErrUnauthorized
	}
	if err := ensureSameOrgOrCreator(existingUser, existingRequest); err != nil {
		return err
	}

	// Copy the data
	forexRequest := model.RequestUpdate{}
	copier.Copy(&forexRequest, existingRequest)

	now := time.Now()
	forexRequest.DeletedAt = &now
	forexRequest.DeletedBy = &authUserID
	forexRequest.IsDeleted = true
	forexRequest.RequestStatus = string(model.ReqStatusDeleted)

	return ru.requestRepository.UpdateIfStatus(
		ctx,
		requestID,
		statusStrings(model.ReqStatusDrafted),
		&forexRequest,
	)
}

func (ru *requestUsecase) SendRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil || existingRequest == nil {
		return common.ErrRequestNotFound
	}

	if err := ensureRequestStatus(existingRequest.RequestStatus, model.ReqStatusDrafted); err != nil {
		return err
	}

	// Fetch sender
	sender, err := ru.userRepository.FindByID(ctx, authUserID)
	if err != nil {
		return err
	}
	if err := ensureSameOrgOrCreator(sender, existingRequest); err != nil {
		return err
	}

	// Copy the data
	forexRequest := model.RequestUpdate{}
	copier.Copy(&forexRequest, existingRequest)

	now := time.Now()
	forexRequest.RequestedAt = &now
	forexRequest.RequestedBy = &authUserID
	forexRequest.RequestStatus = string(model.ReqStatusNew)

	fromStatus := string(existingRequest.RequestStatus)
	err = ru.requestRepository.UpdateIfStatus(
		ctx,
		requestID,
		statusStrings(model.ReqStatusDrafted),
		&forexRequest,
	)
	if err != nil {
		return err
	}

	entityID := requestID
	LogActivity(ru.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionRequestSend,
		EntityType:  model.EntityRequest,
		EntityID:    &entityID,
		Description: "Request sent",
		FromState:   fromStatus,
		ToState:     string(model.ReqStatusNew),
	})

	ru.notifyRequestActivity(
		requestID,
		utils.MailEventSent,
		configs.MailRequestSentTo,
		configs.MailRequestSentCc,
		configs.MailRequestSentBcc,
		fmt.Sprintf("Request %s was submitted on %s.", forexRequest.RequestCode, now.Format("2006-01-02 15:04")),
		nil,
	)

	return nil
}

func (ru *requestUsecase) DeclineOrgRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil || existingRequest == nil {
		return common.ErrRequestNotFound
	}

	if err := ensureRequestStatus(existingRequest.RequestStatus, model.ReqStatusApproved); err != nil {
		return err
	}

	// Decline is a branch-level post-approval action (same tier as Accept) — not org-scoped.
	if _, err := ru.userRepository.FindByID(ctx, authUserID); err != nil {
		return common.ErrUnauthorized
	}

	// Copy the data
	forexRequest := model.RequestUpdate{}
	copier.Copy(&forexRequest, existingRequest)

	fromStatus := string(existingRequest.RequestStatus)
	now := time.Now()
	forexRequest.DeclinedAt = &now
	forexRequest.DeclinedBy = &authUserID
	forexRequest.RequestStatus = string(model.ReqStatusDeclined)

	if err := ru.requestRepository.UpdateIfStatus(
		ctx,
		requestID,
		statusStrings(model.ReqStatusApproved),
		&forexRequest,
	); err != nil {
		return err
	}
	_ = ru.requestRepository.ClearAssignment(ctx, requestID)

	entityID := requestID
	LogActivity(ru.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionRequestDecline,
		EntityType:  model.EntityRequest,
		EntityID:    &entityID,
		Description: "Request declined",
		FromState:   fromStatus,
		ToState:     string(model.ReqStatusDeclined),
	})

	ru.notifyRequestActivity(
		requestID,
		utils.MailEventDeclined,
		configs.MailRequestDeclinedTo,
		configs.MailRequestDeclinedCc,
		configs.MailRequestDeclinedBcc,
		fmt.Sprintf("Request %s was declined on %s.", forexRequest.RequestCode, now.Format("2006-01-02 15:04")),
		nil,
	)
	return nil
}

func (ru *requestUsecase) RejectRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, rejection_reason string) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil || existingRequest == nil {
		return common.ErrRequestNotFound
	}

	if err := ensureRequestStatus(
		existingRequest.RequestStatus,
		model.ReqStatusNew,
		model.ReqStatusAuthorized,
		model.ReqStatusValidated,
		model.ReqStatusApproved,
	); err != nil {
		return err
	}

	existingUser, err := ru.userRepository.FindByID(ctx, authUserID)
	if err != nil {
		return common.ErrUnauthorized
	}

	if existingUser.Role != nil && existingUser.Role.Name == "BRANCHAUTHORIZER" {
		sameBranch := existingUser.Profile != nil &&
			existingUser.Profile.BranchID != nil &&
			existingRequest.BranchID != nil &&
			*existingUser.Profile.BranchID == *existingRequest.BranchID
		sameDepartment := existingUser.Profile != nil &&
			existingUser.Profile.DepartmentID != nil &&
			existingRequest.DepartmentID != nil &&
			*existingUser.Profile.DepartmentID == *existingRequest.DepartmentID
		if !sameBranch && !sameDepartment {
			return common.ErrUnauthorized
		}
	}

	// Copy the data
	forexRequest := model.RequestUpdate{}
	copier.Copy(&forexRequest, existingRequest)

	// Assignment required for bank workflow statuses (same as validate/approve).
	// "New" org-authorizer rejects do not use assignment.
	if existingRequest.RequestStatus == model.ReqStatusAuthorized ||
		existingRequest.RequestStatus == model.ReqStatusValidated ||
		existingRequest.RequestStatus == model.ReqStatusApproved {
		if err := ensureAssignedTo(ru.activeAssignee(existingRequest), authUserID); err != nil {
			return err
		}
	} else if existingRequest.AssignedTo != nil {
		if err := ensureAssignedTo(ru.activeAssignee(existingRequest), authUserID); err != nil {
			return err
		}
	}

	fromStatus := string(existingRequest.RequestStatus)
	expectedReject := statusStrings(
		model.ReqStatusNew,
		model.ReqStatusAuthorized,
		model.ReqStatusValidated,
		model.ReqStatusApproved,
	)
	now := time.Now()
	forexRequest.RejectedAt = &now
	forexRequest.RejectedBy = &authUserID
	forexRequest.RejectionReason = rejection_reason
	forexRequest.RequestStatus = string(model.ReqStatusRejected)

	err = ru.requestRepository.UpdateIfStatus(ctx, requestID, expectedReject, &forexRequest)
	if err != nil {
		return err
	}
	_ = ru.requestRepository.ClearAssignment(ctx, requestID)

	entityID := requestID
	LogActivity(ru.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionRequestReject,
		EntityType:  model.EntityRequest,
		EntityID:    &entityID,
		Description: "Request rejected",
		FromState:   fromStatus,
		ToState:     string(model.ReqStatusRejected),
	})

	rejectDetails := []string{}
	if strings.TrimSpace(rejection_reason) != "" {
		rejectDetails = append(rejectDetails, "Reason: "+strings.TrimSpace(rejection_reason))
	}
	ru.notifyRequestActivity(
		requestID,
		utils.MailEventRejected,
		configs.MailRequestRejectedTo,
		configs.MailRequestRejectedCc,
		configs.MailRequestRejectedBcc,
		fmt.Sprintf("Request %s was rejected on %s.", forexRequest.RequestCode, now.Format("2006-01-02 15:04")),
		rejectDetails,
	)
	return nil
}

func (ru *requestUsecase) LockRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, assigneeID *primitive.ObjectID, actorPerms []string) (*model.AssignRequestResult, error) {
	return ru.AssignRequest(ctx, authUserID, requestID, assigneeID, actorPerms)
}

func (ru *requestUsecase) UnLockRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, actorPerms []string) error {
	return ru.UnassignRequest(ctx, authUserID, requestID, actorPerms)
}

func userHasApproveOrAuthorize(u *model.User, need string) bool {
	if u == nil {
		return false
	}
	rolePerms := []string{}
	if u.Role != nil {
		rolePerms = u.Role.Permissions
	}
	return hasPermission(utils.MergePermissions(rolePerms, u.Permissions), need)
}

func sameBranchAsRequest(u *model.User, req *model.Request) bool {
	if u == nil || u.Profile == nil || u.Profile.BranchID == nil || req == nil || req.BranchID == nil {
		return false
	}
	return *u.Profile.BranchID == *req.BranchID
}

func (ru *requestUsecase) AssignRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, assigneeID *primitive.ObjectID, actorPerms []string) (*model.AssignRequestResult, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil || existingRequest == nil {
		return nil, common.ErrRequestNotFound
	}

	// New → branch authorizer; Authorized → validator; Validated → forex authorizer.
	if err := ensureRequestStatus(
		existingRequest.RequestStatus,
		model.ReqStatusNew,
		model.ReqStatusAuthorized,
		model.ReqStatusValidated,
	); err != nil {
		return nil, err
	}

	actor, err := ru.userRepository.FindByID(ctx, authUserID)
	if err != nil || actor == nil {
		return nil, common.ErrUnauthorized
	}

	targetID := authUserID
	if assigneeID != nil && !assigneeID.IsZero() {
		targetID = *assigneeID
	}

	switch existingRequest.RequestStatus {
	case model.ReqStatusNew:
		canPick := hasPermission(actorPerms, "request:send") || hasPermission(actorPerms, "request:authorize")
		if !canPick {
			return nil, common.ErrUnauthorized
		}
		if err := ensureSameOrgOrCreator(actor, existingRequest); err != nil {
			return nil, err
		}
		if targetID == authUserID {
			if !hasPermission(actorPerms, "request:authorize") {
				return nil, common.ErrSelfAssignOnly
			}
		} else {
			assigneeUser, err := ru.userRepository.FindByID(ctx, targetID)
			if err != nil || assigneeUser == nil {
				return nil, common.ErrAssigneeNotEligible
			}
			if !userHasApproveOrAuthorize(assigneeUser, "request:authorize") ||
				!sameBranchAsRequest(assigneeUser, existingRequest) ||
				(assigneeUser.Status != model.StatusActive && assigneeUser.Status != model.StatusNew) {
				return nil, common.ErrAssigneeNotEligible
			}
		}
	case model.ReqStatusAuthorized:
		if targetID != authUserID {
			return nil, common.ErrSelfAssignOnly
		}
	case model.ReqStatusValidated:
		canAssignOthers := hasPermission(actorPerms, "request:approve")
		if !canAssignOthers {
			if targetID != authUserID {
				return nil, common.ErrSelfAssignOnly
			}
		} else if targetID != authUserID {
			assigneeUser, err := ru.userRepository.FindByID(ctx, targetID)
			if err != nil || assigneeUser == nil {
				return nil, common.ErrAssigneeNotEligible
			}
			if !userHasApproveOrAuthorize(assigneeUser, "request:approve") ||
				(assigneeUser.Status != model.StatusActive && assigneeUser.Status != model.StatusNew) {
				return nil, common.ErrAssigneeNotEligible
			}
		}
	}

	assignedAt := time.Now()
	claimed, err := ru.requestRepository.AssignIfAvailable(ctx, requestID, targetID, assignedAt)
	if err != nil {
		return nil, err
	}
	if !claimed {
		return nil, common.ErrRequestIsLocked
	}

	assigneeUser, err := ru.userRepository.FindByID(ctx, targetID)
	if err != nil {
		assigneeUser = nil
	}

	result := &model.AssignRequestResult{
		AssignedTo: &targetID,
		AssignedAt: assignedAt,
		Assignee:   assigneeUser,
	}

	entityID := requestID
	LogActivity(ru.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionRequestAssign,
		EntityType:  model.EntityRequest,
		EntityID:    &entityID,
		Description: "Request assigned",
		Metadata: map[string]any{
			"assignee_id": targetID.Hex(),
		},
	})

	if targetID != authUserID {
		assigneeEmail := userProfileEmail(assigneeUser)
		summary := fmt.Sprintf(
			"Request %s was assigned to you on %s.",
			existingRequest.RequestCode,
			assignedAt.Format("2006-01-02 15:04"),
		)
		go func(req model.Request, toEmail, note string) {
			if err := utils.NotifyRequestActivity(
				[]string{toEmail}, nil, nil,
				utils.MailEventAssigned,
				req,
				note,
				nil,
			); err != nil {
				log.Warnf("Failed to send assign email for request %s: %v", req.ID.Hex(), err)
			}
		}(*existingRequest, assigneeEmail, summary)
	}

	return result, nil
}

func (ru *requestUsecase) UnassignRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, actorPerms []string) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil || existingRequest == nil {
		return common.ErrRequestNotFound
	}

	assignee := ru.activeAssignee(existingRequest)
	isAssignee := assignee != nil && *assignee == authUserID

	canManage := false
	switch existingRequest.RequestStatus {
	case model.ReqStatusNew:
		canManage = hasPermission(actorPerms, "request:send") || hasPermission(actorPerms, "request:authorize")
	case model.ReqStatusAuthorized:
		canManage = hasPermission(actorPerms, "request:validate")
	case model.ReqStatusValidated:
		canManage = hasPermission(actorPerms, "request:approve")
	}

	if !isAssignee && !canManage {
		return common.ErrRequestIsLocked
	}

	if err := ru.requestRepository.ClearAssignment(ctx, requestID); err != nil {
		return err
	}

	entityID := requestID
	LogActivity(ru.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionRequestUnassign,
		EntityType:  model.EntityRequest,
		EntityID:    &entityID,
		Description: "Request unassigned",
	})
	return nil
}

func (ru *requestUsecase) AcceptRequest(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID, request *model.RequestAcceptanceDTO, processingBranchID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, false)
	if err != nil || existingRequest == nil {
		return common.ErrRequestNotFound
	}

	if err := ensureRequestStatus(existingRequest.RequestStatus, model.ReqStatusApproved); err != nil {
		return err
	}

	// Accept is branch-level post-approval — no validator/authorizer claim required.
	if err := ru.ensureForeignCurrencies(ctx, request.AcceptedCurrencyIDs); err != nil {
		return err
	}

	approvedLen := len(existingRequest.ApprovedAmounts)
	if approvedLen == 0 ||
		len(existingRequest.ApprovedCurrencyIDs) != approvedLen ||
		len(existingRequest.ApprovedAmountInCash) != approvedLen ||
		len(existingRequest.ApprovedAmountInCard) != approvedLen {
		return common.ErrInvalidRequest
	}
	if len(request.AcceptedCurrencyIDs) != approvedLen ||
		len(request.AcceptedAmounts) != approvedLen ||
		len(request.AcceptedAmountInCash) != approvedLen ||
		len(request.AcceptedAmountInCard) != approvedLen {
		return common.ErrInvalidAcceptedCurrency
	}

	for i := 0; i < approvedLen; i++ {
		if existingRequest.ApprovedCurrencyIDs[i] != request.AcceptedCurrencyIDs[i] {
			return common.ErrInvalidAcceptedCurrency
		}

		if existingRequest.ApprovedAmounts[i] < request.AcceptedAmounts[i] ||
			existingRequest.ApprovedAmountInCard[i] < request.AcceptedAmountInCard[i] ||
			existingRequest.ApprovedAmountInCash[i] < request.AcceptedAmountInCash[i] {
			return common.ErrInvalidAcceptedAmount
		}
	}

	// Copy the data
	forexRequest := model.RequestUpdate{}
	copier.Copy(&forexRequest, existingRequest)

	fromStatus := string(existingRequest.RequestStatus)
	now := time.Now()
	forexRequest.AcceptedAt = &now
	forexRequest.AcceptedBy = &authUserID
	forexRequest.AcceptedCurrencyIDs = request.AcceptedCurrencyIDs
	forexRequest.AcceptedAmounts = request.AcceptedAmounts
	forexRequest.AcceptedAmountInCash = request.AcceptedAmountInCash
	forexRequest.AcceptedAmountInCard = request.AcceptedAmountInCard
	forexRequest.RequestStatus = string(model.ReqStatusAccepted)

	// Processing branch is the accepter's branch (result-processor), independent of requesting branch.
	if !processingBranchID.IsZero() {
		forexRequest.ProcessingBranchID = &processingBranchID
	} else if accepter, err := ru.userRepository.FindByID(ctx, authUserID); err == nil &&
		accepter != nil && accepter.Profile != nil && accepter.Profile.BranchID != nil {
		forexRequest.ProcessingBranchID = accepter.Profile.BranchID
	}

	if err := ru.requestRepository.UpdateIfStatus(
		ctx,
		requestID,
		statusStrings(model.ReqStatusApproved),
		&forexRequest,
	); err != nil {
		return err
	}
	_ = ru.requestRepository.ClearAssignment(ctx, requestID)

	entityID := requestID
	LogActivity(ru.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionRequestAccept,
		EntityType:  model.EntityRequest,
		EntityID:    &entityID,
		Description: "Request accepted",
		FromState:   fromStatus,
		ToState:     string(model.ReqStatusAccepted),
	})

	ru.notifyRequestActivity(
		requestID,
		utils.MailEventAccepted,
		configs.MailRequestAcceptedTo,
		configs.MailRequestAcceptedCc,
		configs.MailRequestAcceptedBcc,
		fmt.Sprintf("Request %s was accepted on %s.", forexRequest.RequestCode, now.Format("2006-01-02 15:04")),
		nil,
	)
	return nil
}

func (ru *requestUsecase) GenerateRequestReport(ctx context.Context, authUserID primitive.ObjectID, requestID primitive.ObjectID) ([]byte, string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	existingRequest, err := ru.requestRepository.FindByID(ctx, requestID, true)
	if err != nil || existingRequest == nil {
		return nil, "", "", common.ErrRequestNotFound
	}

	if existingRequest.RequestStatus != model.ReqStatusAccepted {
		return nil, "", "", common.ErrInvalidRequestTransition
	}

	generator, err := ru.userRepository.FindByID(ctx, authUserID)
	if err != nil {
		return nil, "", "", common.ErrUnauthorized
	}

	generatedBy := "N/A"
	if generator.Profile != nil {
		generatedBy = fmt.Sprintf("%s %s", generator.Profile.FirstName, generator.Profile.MiddleName)
	}

	// Each download issues a new receipt number linked to the same request.
	receiptNumber, err := utils.GenerateReceiptNumber()
	if err != nil {
		return nil, "", "", err
	}

	now := time.Now()
	if _, err := ru.receiptRepository.Create(ctx, &model.Receipt{
		ReceiptNumber: receiptNumber,
		RequestID:     existingRequest.ID,
		GeneratedBy:   authUserID,
		GeneratorName: generatedBy,
		CreatedAt:     now,
	}); err != nil {
		return nil, "", "", err
	}

	pdfBytes, err := utils.GenerateRequestInvoicePDF(utils.ReceiptPDFInput{
		Request:       existingRequest,
		ReceiptNumber: receiptNumber,
		GeneratedBy:   generatedBy,
	})
	if err != nil {
		return nil, "", "", err
	}

	filename := fmt.Sprintf("%s.pdf", receiptNumber)
	return pdfBytes, filename, receiptNumber, nil
}

func (ru *requestUsecase) ValidateReceipt(ctx context.Context, receiptNumber string) (*model.Request, map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	receipt, err := ru.receiptRepository.FindByReceiptNumber(ctx, strings.TrimSpace(receiptNumber))
	if err != nil {
		return nil, nil, common.ErrReceiptNotFound
	}

	req, err := ru.requestRepository.FindByID(ctx, receipt.RequestID, true)
	if err != nil || req == nil {
		return nil, nil, common.ErrRequestNotFound
	}

	currency := "N/A"
	if req.FcyRequested != nil {
		currency = utils.CurrencyLabel(req.FcyRequested)
	}

	amountLines := buildReceiptAmountLines(req)

	requestingBranch := "N/A"
	if req.Branch != nil {
		requestingBranch = req.Branch.Name
	}
	processingBranch := resolveProcessingBranchName(req)

	comparison := map[string]any{
		"receipt_number":           receipt.ReceiptNumber,
		"generated_by":             receipt.GeneratorName,
		"generated_at":             receipt.CreatedAt,
		"request_id":               req.ID.Hex(),
		"request_code":             req.RequestCode,
		"request_status":           req.RequestStatus,
		"applicant_name":           req.ApplicantName,
		"applicant_account_number": req.ApplicantAccountNumber,
		"requested_currency":       currency,
		"requested_amount":         req.FcyRequestedAmount,
		"requesting_branch":        requestingBranch,
		"processing_branch":        processingBranch,
		"amount_lines":             amountLines,
		"approved_amounts":         req.ApprovedAmounts,
		"approved_amount_in_cash":  req.ApprovedAmountInCash,
		"approved_amount_in_card":  req.ApprovedAmountInCard,
		"accepted_amounts":         req.AcceptedAmounts,
		"accepted_amount_in_cash":  req.AcceptedAmountInCash,
		"accepted_amount_in_card":  req.AcceptedAmountInCard,
		"requester_name":           utils.PersonName(req.Requester),
		"authorizer_name":          utils.PersonName(req.Authorizer),
		"validator_name":           utils.PersonName(req.Validater),
		"approver_name":            utils.PersonName(req.Approver),
		"accepter_name":            utils.PersonName(req.Accepter),
		"valid":                    req.RequestStatus == model.ReqStatusAccepted,
		"note":                     "Compare these request values with the printed PDF. Accepted request data is the source of truth.",
	}

	return req, comparison, nil
}

func resolveProcessingBranchName(req *model.Request) string {
	if req.ProcessingBranch != nil && strings.TrimSpace(req.ProcessingBranch.Name) != "" {
		return req.ProcessingBranch.Name
	}
	return "N/A"
}

func buildReceiptAmountLines(req *model.Request) []map[string]any {
	maxRows := len(req.ApprovedCurrencyIDs)
	if len(req.AcceptedCurrencyIDs) > maxRows {
		maxRows = len(req.AcceptedCurrencyIDs)
	}
	for _, n := range []int{
		len(req.ApprovedAmounts),
		len(req.ApprovedAmountInCash),
		len(req.ApprovedAmountInCard),
		len(req.AcceptedAmounts),
		len(req.AcceptedAmountInCash),
		len(req.AcceptedAmountInCard),
	} {
		if n > maxRows {
			maxRows = n
		}
	}

	lines := make([]map[string]any, 0, maxRows)
	for i := 0; i < maxRows; i++ {
		approvedCurrency := "N/A"
		if i < len(req.ApprovedCurrencies) {
			approvedCurrency = utils.CurrencyLabel(&req.ApprovedCurrencies[i])
		}
		acceptedCurrency := "N/A"
		if i < len(req.AcceptedCurrencies) {
			acceptedCurrency = utils.CurrencyLabel(&req.AcceptedCurrencies[i])
		}

		approvedCash := amountAt(req.ApprovedAmountInCash, i)
		approvedCard := amountAt(req.ApprovedAmountInCard, i)
		approvedTotal := amountAt(req.ApprovedAmounts, i)
		if approvedTotal == nil && (approvedCash != nil || approvedCard != nil) {
			sum := float64(0)
			if approvedCash != nil {
				sum += *approvedCash
			}
			if approvedCard != nil {
				sum += *approvedCard
			}
			approvedTotal = &sum
		}

		acceptedCash := amountAt(req.AcceptedAmountInCash, i)
		acceptedCard := amountAt(req.AcceptedAmountInCard, i)
		acceptedTotal := amountAt(req.AcceptedAmounts, i)
		if acceptedTotal == nil && (acceptedCash != nil || acceptedCard != nil) {
			sum := float64(0)
			if acceptedCash != nil {
				sum += *acceptedCash
			}
			if acceptedCard != nil {
				sum += *acceptedCard
			}
			acceptedTotal = &sum
		}

		lines = append(lines, map[string]any{
			"currency":          acceptedCurrency,
			"approved_currency": approvedCurrency,
			"accepted_currency": acceptedCurrency,
			"approved_cash":     approvedCash,
			"approved_card":     approvedCard,
			"approved_total":    approvedTotal,
			"accepted_cash":     acceptedCash,
			"accepted_card":     acceptedCard,
			"accepted_total":    acceptedTotal,
		})
	}
	return lines
}

func amountAt(values []float64, i int) *float64 {
	if i < 0 || i >= len(values) {
		return nil
	}
	v := values[i]
	return &v
}
