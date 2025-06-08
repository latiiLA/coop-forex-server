package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProfileUsecase interface {
	AddProfile(ctx context.Context, user_id primitive.ObjectID, profile *model.Profile) error
	GetProfileByID(ctx context.Context, user_id primitive.ObjectID, profile_id primitive.ObjectID) (*model.Profile, error)
	UpdateProfileByUserID(ctx context.Context, user_id primitive.ObjectID, profile_id primitive.ObjectID, profile *model.Profile) (*model.Profile, error)
}

type profileUsecase struct {
	profileRepository model.ProfileRepository
	userRepository    model.UserRepository
	contextTimeout    time.Duration
}

func NewProfileUsecase(profileRepository model.ProfileRepository, userRepository model.UserRepository, timeout time.Duration) ProfileUsecase {
	return &profileUsecase{
		profileRepository: profileRepository,
		userRepository:    userRepository,
		contextTimeout:    timeout,
	}
}

func (pu *profileUsecase) AddProfile(ctx context.Context, user_id primitive.ObjectID, profile *model.Profile) error {
	ctx, cancel := context.WithTimeout(ctx, pu.contextTimeout)
	defer cancel()

	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()

	return pu.profileRepository.Create(ctx, profile)
}

func (pu *profileUsecase) GetProfileByID(ctx context.Context, user_id primitive.ObjectID, profile_id primitive.ObjectID) (*model.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, pu.contextTimeout)
	defer cancel()

	fmt.Print(user_id, profile_id, "inside get profile")

	existingUser, err := pu.userRepository.FindByID(ctx, user_id)
	if err != nil {
		return nil, err
	}

	if existingUser.ProfileID != profile_id {
		return nil, errors.New("unauthorized user")
	}

	return pu.profileRepository.FindByID(ctx, profile_id)
}

func (pu *profileUsecase) UpdateProfileByUserID(ctx context.Context, user_id primitive.ObjectID, profile_id primitive.ObjectID, profile *model.Profile) (*model.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, pu.contextTimeout)
	defer cancel()

	existingUser, err := pu.userRepository.FindByID(ctx, user_id)
	if err != nil {
		return nil, err
	}

	if existingUser.ProfileID != profile_id {
		return nil, errors.New("unauthorized user")
	}

	existingProfile, err := pu.profileRepository.FindByID(ctx, profile_id)
	if err != nil {
		return nil, errors.New("profile not found")
	}

	if profile.Birthday != nil && !profile.Birthday.IsZero() {
		existingProfile.Birthday = profile.Birthday
	}
	if profile.ShortBio != nil {
		existingProfile.ShortBio = profile.ShortBio
	}
	if profile.Gender != nil {
		existingProfile.Gender = profile.Gender
	}
	if profile.Phone != nil {
		existingProfile.Phone = profile.Phone
	}
	existingProfile.UpdatedAt = time.Now()

	return pu.profileRepository.Update(ctx, profile_id, existingProfile)
}
