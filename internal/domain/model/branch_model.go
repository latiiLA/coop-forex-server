package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Branch struct {
	ID primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name string `json:"name" bson:"name"`
	DistrictID primitive.ObjectID `json:"district_id" bson:"district_id"`
	CreatedBy primitive.ObjectID `json:"created_by" bson:"created_by"`
	UpdatedBy *primitive.ObjectID `json:"updated_by" bson:"updated_by"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	DeletedBy *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	IsResultProcessor bool `json:"is_result_processor" bson:"is_result_processor"`
	IsDeleted bool `json:"-" bson:"is_deleted"`
}