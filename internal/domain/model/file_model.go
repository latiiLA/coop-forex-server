package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type File struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	URL       string             `json:"url" bson:"url"`
	Name      string             `json:"name" bson:"name"`
	Fid       string             `json:"fid" bson:"fid"`
	Size      int64              `json:"size" bson:"size"`
	MimeType  string             `json:"mime_type" bson:"mime_type"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

type FileRepository interface {
	Create(ctx context.Context, file *File) (*primitive.ObjectID, error)
	FindByID(ctx context.Context, file_id primitive.ObjectID) (*File, error)
}
