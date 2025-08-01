package usecase

import (
	"context"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BranchUsecase interface {
	AddBranch(ctx context.Context, userID primitive.ObjectID, branch *model.Branch) error
	GetBranchByID(ctx context.Context, branchID primitive.ObjectID) (*model.Branch, error)
	GetBranchesByDistrictID(ctx context.Context, districtID primitive.ObjectID) (*[]model.Branch, error)
	GetAllBranches(ctx context.Context) ([]model.Branch, error)
}

type branchUsecase struct {
	branchRepository model.BranchRepository
	contextTimeout   time.Duration
}

func NewBranchUsecase(branchRepository model.BranchRepository, timeout time.Duration) BranchUsecase {
	return &branchUsecase{
		branchRepository: branchRepository,
		contextTimeout:   timeout,
	}
}

func (bu *branchUsecase) AddBranch(ctx context.Context, userID primitive.ObjectID, branch *model.Branch) error {
	return nil
}

func (bu *branchUsecase) GetBranchByID(ctx context.Context, branchID primitive.ObjectID) (*model.Branch, error) {
	return nil, nil
}

func (bu *branchUsecase) GetAllBranches(ctx context.Context) ([]model.Branch, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	return bu.branchRepository.FindAll(ctx)
}

func (bu *branchUsecase) GetBranchesByDistrictID(ctx context.Context, districtID primitive.ObjectID) (*[]model.Branch, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	return bu.branchRepository.FindByDistrictID(ctx, districtID)
}
