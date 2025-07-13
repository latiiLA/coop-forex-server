package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Account refers to Institution or self account
type CustomerType struct {
	ID        primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	Type      string              `json:"type" bson:"type"`
	CreatedBy primitive.ObjectID  `json:"created_by" bson:"created_by"`
	UpdatedBy *primitive.ObjectID `json:"updated_by" bson:"updated_by"`
	CreatedAt time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time           `json:"updated_at" bson:"updated_at"`
	DeletedAt *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	DeletedBy *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	IsDeleted bool                `json:"is_deleted" bson:"is_deleted"`
}

type CustomerTypesRepository interface {
	Create(ctx context.Context, customer_type *CustomerType) error
	FindByID(ctx context.Context, customer_type_id primitive.ObjectID) (*CustomerType, error)
	FindAll(ctx context.Context) ([]CustomerType, error)
	Update(ctx context.Context, customer_type_id primitive.ObjectID, customer_type *CustomerType) (*CustomerType, error)
	Delete(ctx context.Context, customer_type_id primitive.ObjectID) error
}
