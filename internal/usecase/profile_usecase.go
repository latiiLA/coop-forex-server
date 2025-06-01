package usecase

import (
	"context"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProfileUsecase interface {
	AddProfile(ctx context.Context, userID primitive.ObjectID, profile *model.Profile) error
	GetProfileByID(ctx context.Context, profile_id primitive.ObjectID) (*model.Profile, error)
	UpdateProfile(ctx context.Context, profile_id primitive.ObjectID, profile *model.Profile) (*model.Profile, error)
}

type profileUsecase struct {
	profileRepository model.ProfileRepository
	contextTimeout    time.Duration
}

func NewProfileUsecase(profileRepository model.ProfileRepository, timeout time.Duration) ProfileUsecase {
	return &profileUsecase{
		profileRepository: profileRepository,
		contextTimeout:    timeout,
	}
}

func (pu *profileUsecase) AddProfile(ctx context.Context, userID primitive.ObjectID, profile *model.Profile) error {
	ctx, cancel := context.WithTimeout(ctx, pu.contextTimeout)
	defer cancel()

	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()

	return pu.profileRepository.Create(ctx, profile)
}

func (pu *profileUsecase) GetProfileByID(ctx context.Context, userID primitive.ObjectID) (*model.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, pu.contextTimeout)
	defer cancel()
	return pu.profileRepository.FindByID(ctx, userID)
}

func (pu *profileUsecase) UpdateProfile(ctx context.Context, userID primitive.ObjectID, profile *model.Profile) (*model.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, pu.contextTimeout)
	defer cancel()
	return pu.profileRepository.Update(ctx, userID, profile)
}
