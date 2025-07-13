package usecase

import (
	"context"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubprocessUsecase interface {
	AddSubprocess(ctx context.Context, userID primitive.ObjectID, subprocess *model.Subprocess) error
	GetSubprocessByID(ctx context.Context, subprocess_id primitive.ObjectID) (*model.Subprocess, error)
	GetSubprocessByProcessID(ctx context.Context, process_id primitive.ObjectID) (*[]model.Subprocess, error)
	GetAllSubprocesses(ctx context.Context) ([]model.Subprocess, error)
}

type subprocessUsecase struct {
	subprocessRepository model.SubprocessRepository
	contextTimeout       time.Duration
}

func NewSubprocessUsecase(subprocessRepository model.SubprocessRepository, timeout time.Duration) SubprocessUsecase {
	return &subprocessUsecase{
		subprocessRepository: subprocessRepository,
		contextTimeout:       timeout,
	}
}

func (su *subprocessUsecase) AddSubprocess(ctx context.Context, userID primitive.ObjectID, subprocess *model.Subprocess) error {
	return nil
}

func (su *subprocessUsecase) GetSubprocessByID(ctx context.Context, subprocess_id primitive.ObjectID) (*model.Subprocess, error) {
	return nil, nil
}

func (su *subprocessUsecase) GetAllSubprocesses(ctx context.Context) ([]model.Subprocess, error) {
	ctx, cancel := context.WithTimeout(ctx, su.contextTimeout)
	defer cancel()
	return su.subprocessRepository.FindAll(ctx)
}

func (su *subprocessUsecase) GetSubprocessByProcessID(ctx context.Context, process_id primitive.ObjectID) (*[]model.Subprocess, error) {
	ctx, cancel := context.WithTimeout(ctx, su.contextTimeout)
	defer cancel()
	return su.subprocessRepository.FindByProcessID(ctx, process_id)
}
