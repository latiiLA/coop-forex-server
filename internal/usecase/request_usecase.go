package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/copier"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestUsecase interface {
	AddRequest(ctx context.Context, authUserID primitive.ObjectID, request *model.Request) error
	GetAllRequests(ctx context.Context) ([]model.Request, error)
	ValidateRequest(ctx context.Context, authUserID primitive.ObjectID, request_id primitive.ObjectID, validated_currency_id primitive.ObjectID, request *model.RequestValidationDTO) error
	ApproveRequest(ctx context.Context, authUserID primitive.ObjectID, request_id primitive.ObjectID, request *model.ParsedApproval) error
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
	forexRequest.RequestStatus = "New"
	forexRequest.IsDeleted = false

	err = ru.requestRepository.Create(ctx, &forexRequest)

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

func (ru *requestUsecase) ValidateRequest(ctx context.Context, authUserID primitive.ObjectID, request_id primitive.ObjectID, validated_account_currency_id primitive.ObjectID, request *model.RequestValidationDTO) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	requestData, err := ru.requestRepository.FindByID(ctx, request_id)
	if err != nil {
		fmt.Println("the request_id doesn't exist")
		return errors.New("the request id doesn't exist")
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

func (ru *requestUsecase) ApproveRequest(ctx context.Context, authUserID primitive.ObjectID, request_id primitive.ObjectID, request *model.ParsedApproval) error {
	return nil
}
