package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subprocess struct {
	ID        primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	ProcessID primitive.ObjectID  `json:"process_id" bson:"process_id"`
	Process   *Process            `json:"process,omitempty" bson:"process,omitempty"`
	Name      string              `json:"name" bson:"name"`
	CreatedAt time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time           `json:"updated_at" bson:"updated_at"`
	CreatedBy primitive.ObjectID  `json:"created_by" bson:"created_by"`
	UpdatedBy *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	DeletedBy *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted bool                `json:"is_deleted" bson:"is_deleted"`

	Creator *User `json:"creator,omitempty" bson:"creator,omitempty"`
	Updater *User `json:"updater,omitempty" bson:"updater,omitempty"`
	Deleter *User `json:"deleter,omitempty" bson:"deleter,omitempty"`
}

type SubprocessRepository interface {
	Create(ctx context.Context, subprocess *Subprocess) error
	FindByID(ctx context.Context, subprocess_id primitive.ObjectID) (*Subprocess, error)
	FindByProcessID(ctx context.Context, process_id primitive.ObjectID) (*[]Subprocess, error)
	FindAll(ctx context.Context) ([]Subprocess, error)
}
