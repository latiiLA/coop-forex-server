package usecase

import (
	"context"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProcessUsecase interface {
	AddProcess(ctx context.Context, userID primitive.ObjectID, process *model.Process) error
	GetProcessByID(ctx context.Context, process_id primitive.ObjectID) (*model.Process, error)
	GetAllProcesses(ctx context.Context) ([]model.Process, error)
}

type processUsecase struct {
	processRepository model.ProcessRepository
	contextTimeout    time.Duration
}

func NewProcessUsecase(processRepository model.ProcessRepository, timeout time.Duration) ProcessUsecase {
	return &processUsecase{
		processRepository: processRepository,
		contextTimeout:    timeout,
	}
}

func (pu *processUsecase) AddProcess(ctx context.Context, userID primitive.ObjectID, process *model.Process) error {
	return nil
}

func (pu *processUsecase) GetProcessByID(ctx context.Context, process_id primitive.ObjectID) (*model.Process, error) {
	return nil, nil
}

func (pu *processUsecase) GetAllProcesses(ctx context.Context) ([]model.Process, error) {
	ctx, cancel := context.WithTimeout(ctx, pu.contextTimeout)
	defer cancel()
	return pu.processRepository.FindAll(ctx)
}
