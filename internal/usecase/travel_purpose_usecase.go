package usecase

import (
	"context"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
)

type TravelPurposeUsecase interface {
	GetAllTravelPurposes(ctx context.Context) ([]model.TravelPurpose, error)
}

type travelPurposeUsecase struct {
	travelPurposeRepository model.TravelPurposeRepository
	contextTimeout          time.Duration
}

func NewTravelPurposeUsecase(travelPurposeRepository model.TravelPurposeRepository, timeout time.Duration) TravelPurposeUsecase {
	return &travelPurposeUsecase{
		travelPurposeRepository: travelPurposeRepository,
		contextTimeout:          timeout,
	}
}

func (tpu travelPurposeUsecase) GetAllTravelPurposes(ctx context.Context) ([]model.TravelPurpose, error) {
	ctx, cancel := context.WithTimeout(ctx, tpu.contextTimeout)
	defer cancel()
	return tpu.travelPurposeRepository.FindAll(ctx)
}
