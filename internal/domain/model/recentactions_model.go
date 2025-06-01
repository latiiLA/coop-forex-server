package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RecentAction struct {
	ID        primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID  `json:"user_id" bson:"user_id"`
	Type      string			  `json:"type" bson:"type"`
	Description string            `json:"description" bson:"description"`
	CreatedAt time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time           `json:"updated_at" bson:"update_at"`
}