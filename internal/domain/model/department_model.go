package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Department struct {
	ID        primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	SubProcessID primitive.ObjectID `json:"sub_process_id" bson:"sub_process_id"`
	Name      string              `json:"name" bson:"name"`
	CreatedAt time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time           `json:"updated_at" bson:"update_at"`
	CreatedBy primitive.ObjectID  `json:"created_by" bson:"created_by"`
	UpdatedBy *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	DeletedBy *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted bool                `json:"-" bson:"is_deleted"`
}