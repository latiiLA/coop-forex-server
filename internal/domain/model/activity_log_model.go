package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Activity action constants — keep in sync with frontend display labels.
const (
	ActionLogin              = "auth.login"
	ActionLogout             = "auth.logout"
	ActionRequestCreate      = "request.create"
	ActionRequestUpdate      = "request.update"
	ActionRequestAuthorize   = "request.authorize"
	ActionRequestValidate    = "request.validate"
	ActionRequestApprove     = "request.approve"
	ActionRequestAccept      = "request.accept"
	ActionRequestReject      = "request.reject"
	ActionRequestDecline     = "request.decline"
	ActionRequestSend        = "request.send"
	ActionRequestAssign      = "request.assign"
	ActionRequestUnassign    = "request.unassign"
	ActionUserCreate         = "user.create"
	ActionUserUpdate         = "user.update"
	ActionUserSuspend        = "user.suspend"
	ActionUserUnsuspend      = "user.unsuspend"
)

const (
	EntityAuth    = "auth"
	EntityRequest = "request"
	EntityUser    = "user"
)

type ActivityLog struct {
	ID           primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	ActorID      primitive.ObjectID  `json:"actor_id" bson:"actor_id"`
	Actor        *User               `json:"actor,omitempty" bson:"actor,omitempty"`
	Action       string              `json:"action" bson:"action"`
	EntityType   string              `json:"entity_type" bson:"entity_type"`
	EntityID     *primitive.ObjectID `json:"entity_id,omitempty" bson:"entity_id,omitempty"`
	Description  string              `json:"description" bson:"description"`
	FromState    string              `json:"from_state,omitempty" bson:"from_state,omitempty"`
	ToState      string              `json:"to_state,omitempty" bson:"to_state,omitempty"`
	Metadata     map[string]any      `json:"metadata,omitempty" bson:"metadata,omitempty"`
	DedupeKey    string              `json:"-" bson:"dedupe_key,omitempty"`
	CreatedAt    time.Time           `json:"created_at" bson:"created_at"`
}

type ActivityLogInput struct {
	ActorID     primitive.ObjectID
	Action      string
	EntityType  string
	EntityID    *primitive.ObjectID
	Description string
	FromState   string
	ToState     string
	Metadata    map[string]any
	DedupeKey   string
}

type UserStatusSummary struct {
	Total    int64            `json:"total"`
	ByStatus map[string]int64 `json:"by_status"`
}

type ActivityLogRepository interface {
	Create(ctx context.Context, log *ActivityLog) error
	FindRecent(ctx context.Context, limit int64) ([]ActivityLog, error)
}

type ActivityLogUsecase interface {
	Log(ctx context.Context, input ActivityLogInput) error
	GetRecent(ctx context.Context, limit int64) ([]ActivityLog, error)
}
