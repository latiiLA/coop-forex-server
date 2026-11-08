package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type activityLogUsecase struct {
	repo           model.ActivityLogRepository
	contextTimeout time.Duration
}

func NewActivityLogUsecase(repo model.ActivityLogRepository, timeout time.Duration) model.ActivityLogUsecase {
	return &activityLogUsecase{
		repo:           repo,
		contextTimeout: timeout,
	}
}

func (uc *activityLogUsecase) Log(ctx context.Context, input model.ActivityLogInput) error {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	if input.ActorID.IsZero() || input.Action == "" {
		return nil
	}

	dedupe := input.DedupeKey
	if dedupe == "" {
		entityPart := "none"
		if input.EntityID != nil {
			entityPart = input.EntityID.Hex()
		}
		dedupe = fmt.Sprintf("%s:%s:%s:%s:%s",
			input.Action, input.EntityType, entityPart, input.FromState, input.ToState)
		if input.Action == model.ActionLogin || input.Action == model.ActionLogout {
			// Allow multiple auth events; bucket by minute to avoid double-submit duplicates.
			dedupe = fmt.Sprintf("%s:%s:%d", input.Action, input.ActorID.Hex(), time.Now().Unix()/60)
		}
	}

	log := &model.ActivityLog{
		ID:          primitive.NewObjectID(),
		ActorID:     input.ActorID,
		Action:      input.Action,
		EntityType:  input.EntityType,
		EntityID:    input.EntityID,
		Description: input.Description,
		FromState:   input.FromState,
		ToState:     input.ToState,
		Metadata:    input.Metadata,
		DedupeKey:   dedupe,
		CreatedAt:   time.Now(),
	}

	return uc.repo.Create(ctx, log)
}

func (uc *activityLogUsecase) GetRecent(ctx context.Context, limit int64) ([]model.ActivityLog, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.repo.FindRecent(ctx, limit)
}

// LogActivity is a nil-safe helper for controllers/usecases.
func LogActivity(uc model.ActivityLogUsecase, ctx context.Context, input model.ActivityLogInput) {
	if uc == nil {
		return
	}
	_ = uc.Log(ctx, input)
}
