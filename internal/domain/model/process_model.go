package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Process struct {
	ID        primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	Name      string              `json:"name" bson:"name"`
	CreatedAt time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time           `json:"updated_at" bson:"updated_at"`
	CreatedBy primitive.ObjectID  `json:"created_by" bson:"created_by"`
	UpdatedBy *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	DeletedBy *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted bool                `json:"is_deleted" bson:"is_deleted"`
}

type ProcessRepository interface {
	Create(ctx context.Context, process *Process) error
	FindByID(ctx context.Context, process_id primitive.ObjectID) (*Process, error)
	FindAll(ctx context.Context) ([]Process, error)
}
