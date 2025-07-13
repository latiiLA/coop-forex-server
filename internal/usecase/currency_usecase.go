package usecase

import (
	"context"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
)

type CurrencyUsecase interface {
	GetAllCurrency(ctx context.Context) ([]model.Currency, error)
}

type currencyUsecase struct {
	currencyRepository model.CurrencyRepository
	contextTimeout     time.Duration
}

func NewCurrencyUsecase(currencyRepository model.CurrencyRepository, timeout time.Duration) CurrencyUsecase {
	return &currencyUsecase{
		currencyRepository: currencyRepository,
		contextTimeout:     timeout,
	}
}

func (cu *currencyUsecase) GetAllCurrency(ctx context.Context) ([]model.Currency, error) {
	ctx, cancel := context.WithTimeout(ctx, cu.contextTimeout)
	defer cancel()
	return cu.currencyRepository.FindAll(ctx)
}
