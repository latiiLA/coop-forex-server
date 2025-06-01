package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Currency struct {
	ID        primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	Name      string              `json:"name" bson:"name"`
	ShortCode string              `json:"short_code" bson:"short_code"`
	CreatedAt time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time           `json:"updated_at" bson:"update_at"`
	CreatedBy primitive.ObjectID  `json:"created_by" bson:"created_by"`
	UpdatedBy *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	DeletedBy *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted bool                `json:"-" bson:"is_deleted"`
}
