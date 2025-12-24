package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Country struct {
	ID         primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	CountryID  int32               `json:"id,omitempty" bson:"id,omitempty"`
	Name       string              `json:"name" bson:"name"`
	ShortCode  string              `json:"short_code" bson:"short_code"`
	VisaStatus bool                `json:"visa_status" bson:"visa_status"`
	CreatedAt  time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at" bson:"updated_at"`
	CreatedBy  primitive.ObjectID  `json:"created_by" bson:"created_by"`
	UpdatedBy  *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	DeletedBy  *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt  *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted  bool                `json:"is_deleted" bson:"is_deleted"`

	Creator *User `json:"creator,omitempty" bson:"creator,omitempty"`
	Updater *User `json:"updater,omitempty" bson:"updater,omitempty"`
	Deleter *User `json:"deleter,omitempty" bson:"deleter,omitempty"`
}

type CountryResponseDTO struct {
}

type CountryRepository interface {
	Create(ctx context.Context, country *Country) error
	FindByID(ctx context.Context, country_id primitive.ObjectID) (*Country, error)
	FindAll(ctx context.Context) ([]Country, error)
	Update(ctx context.Context, country_id primitive.ObjectID, country *Country) (*Country, error)
	Delete(ctx context.Context, country_id primitive.ObjectID) error
}
