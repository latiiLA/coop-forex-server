package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Branch struct {
	ID                primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	Name              string              `json:"name" bson:"name"`
	BranchCode        string              `json:"branch_code" bson:"branch_code"`
	Email             string              `json:"email" bson:"email"`
	Address           string              `json:"address" bson:"address"`
	DistrictID        primitive.ObjectID  `json:"district_id" bson:"district_id"`
	District          *District           `json:"district,omitempty" bson:"district,omitempty"`
	CreatedBy         primitive.ObjectID  `json:"created_by" bson:"created_by"`
	UpdatedBy         *primitive.ObjectID `json:"updated_by" bson:"updated_by"`
	CreatedAt         time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at" bson:"updated_at"`
	DeletedAt         *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	DeletedBy         *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	IsResultProcessor bool                `json:"is_result_processor" bson:"is_result_processor"`
	IsDeleted         bool                `json:"is_deleted" bson:"is_deleted"`

	Creator *User `json:"creator,omitempty" bson:"creator,omitempty"`
	Updater *User `json:"updater,omitempty" bson:"updater,omitempty"`
	Deleter *User `json:"deleter,omitempty" bson:"deleter,omitempty"`
}

type BranchRepository interface {
	Create(ctx context.Context, branch *Branch) error
	FindByID(ctx context.Context, branchID primitive.ObjectID) (*Branch, error)
	FindByDistrictID(ctx context.Context, districtID primitive.ObjectID) (*[]Branch, error)
	FindAll(ctx context.Context) ([]Branch, error)
}
