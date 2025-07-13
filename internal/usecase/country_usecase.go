package usecase

import (
	"context"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
)

type CountryUsecase interface {
	GetAllCountry(ctx context.Context) ([]model.Country, error)
}

type countryUsecase struct {
	countryRepository model.CountryRepository
	contextTimeout    time.Duration
}

func NewCountryUsecase(countryRepository model.CountryRepository, timeout time.Duration) CountryUsecase {
	return &countryUsecase{
		countryRepository: countryRepository,
		contextTimeout:    timeout,
	}
}

func (cu *countryUsecase) GetAllCountry(ctx context.Context) ([]model.Country, error) {
	ctx, cancel := context.WithTimeout(ctx, cu.contextTimeout)
	defer cancel()
	return cu.countryRepository.FindAll(ctx)
}
