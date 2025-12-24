package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Department struct {
	ID           primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	SubProcessID primitive.ObjectID  `json:"subprocess_id" bson:"subprocess_id"`
	SubProcess   *Subprocess         `json:"subprocess,omitempty" bson:"subprocess,omitempty"`
	Name         string              `json:"name" bson:"name"`
	CreatedAt    time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at" bson:"updated_at"`
	CreatedBy    primitive.ObjectID  `json:"created_by" bson:"created_by"`
	UpdatedBy    *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	DeletedBy    *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt    *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted    bool                `json:"is_deleted" bson:"is_deleted"`

	Creator *User `json:"creator,omitempty" bson:"creator,omitempty"`
	Updater *User `json:"updater,omitempty" bson:"updater,omitempty"`
	Deleter *User `json:"deleter,omitempty" bson:"deleter,omitempty"`
}

type DepartmentResponseDTO struct {
	ID           primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	SubProcessID primitive.ObjectID  `json:"subprocess_id" bson:"subprocess_id"`
	SubProcess   Subprocess          `json:"subprocess" bson:"subprocess"`
	Name         string              `json:"name" bson:"name"`
	CreatedAt    time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at" bson:"update_at"`
	CreatedBy    primitive.ObjectID  `json:"created_by" bson:"created_by"`
	UpdatedBy    *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	DeletedBy    *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt    *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted    bool                `json:"is_deleted" bson:"is_deleted"`
}

type DepartmentRepository interface {
	Create(c context.Context, department *Department) error
	FindAll(c context.Context) ([]Department, error)
	FindByID(c context.Context, department_id primitive.ObjectID) (*Department, error)
	FindBySubprocessID(ctx context.Context, subprocess_id primitive.ObjectID) (*[]Department, error)
	Update(c context.Context, department_id primitive.ObjectID, department *Department) (*Department, error)
	Delete(c context.Context, department_id primitive.ObjectID) error
}
