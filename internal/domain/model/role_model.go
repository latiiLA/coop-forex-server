package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role struct {
	ID          primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	Name        string              `json:"name" bson:"name"`
	Permissions []string            `json:"permissions" bson:"permissions"`
	CreatedAt   time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at" bson:"updated_at"`
	CreatedBy   primitive.ObjectID  `json:"created_by" bson:"created_by"`
	Creater     *User               `json:"creater,omitempty" bson:"creater,omitempty"`
	UpdatedBy   *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	Updater     *User               `json:"updater,omitempty" bson:"updater,omitempty"`
	DeletedBy   *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt   *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted   bool                `json:"is_deleted" bson:"is_deleted"`
}

type CreateRoleDTO struct {
	Name        string   `json:"name" binding:"required,min=3"`
	Permissions []string `json:"permissions" binding:"required,min=1"`
}

type UpdateRoleDTO struct {
	Name        string   `json:"name" binding:"required,min=3"`
	Permissions []string `json:"permissions" binding:"required,min=1"`
}

type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	FindByID(ctx context.Context, role_id primitive.ObjectID) (*Role, error)
	FindAll(ctx context.Context) ([]Role, error)
	FindRoleByName(ctx context.Context, name string) (*Role, error)
	Update(ctx context.Context, role_id primitive.ObjectID, role *Role) error
	Delete(ctx context.Context, role_id primitive.ObjectID, role *Role) error
	FindDeleted(ctx context.Context) ([]Role, error)
}
