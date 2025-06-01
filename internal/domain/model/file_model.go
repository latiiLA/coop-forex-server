package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type File struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	URL       string             `json:"url" bson:"url"`
	Name      string             `json:"name" bson:"name"`
	Fid       string             `json:"fid" bson:"fid"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"update_at"`
}
