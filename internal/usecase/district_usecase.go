package usecase

import (
	"context"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
)

type DistrictUsecase interface {
	GetAllDistricts(ctx context.Context) ([]model.District, error)
}

type districtUsecase struct {
	districtRepository model.DistrictRepository
	contextTimeout     time.Duration
}

func NewDistrictUsecase(districtRepository model.DistrictRepository, timeout time.Duration) DistrictUsecase {
	return &districtUsecase{
		districtRepository: districtRepository,
		contextTimeout:     timeout,
	}
}

func (du *districtUsecase) GetAllDistricts(ctx context.Context) ([]model.District, error) {
	ctx, cancel := context.WithTimeout(ctx, du.contextTimeout)
	defer cancel()
	return du.districtRepository.FindAll(ctx)
}
