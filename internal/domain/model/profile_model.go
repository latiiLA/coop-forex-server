package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Profile struct {
	ID           primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	FirstName    string              `json:"first_name" bson:"first_name"`
	MiddleName   string              `json:"middle_name" bson:"middle_name"`
	LastName     string              `json:"last_name" bson:"last_name"`
	DisplayName  string              `json:"display_name" bson:"display_name"`
	Birthday     *time.Time          `json:"birthday,omitempty" bson:"birthday,omitempty"`
	ShortBio     *string             `json:"short_bio,omitempty" bson:"short_bio,omitempty"`
	Gender       *string             `json:"gender,omitempty" bson:"gender,omitempty"`
	Phone        *string             `json:"phone,omitempty" bson:"phone,omitempty"`
	Email        string              `json:"email" bson:"email"`
	DepartmentID *primitive.ObjectID `json:"department_id,omitempty" bson:"department_id,omitempty"`
	Department   *Department         `json:"department,omitempty" bson:"department,omitempty"`
	BranchID     *primitive.ObjectID `json:"branch_id,omitempty" bson:"branch_id,omitempty"`
	Branch       *Branch             `json:"branch,omitempty" bson:"branch,omitempty"`
	CreatedAt    time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at" bson:"updated_at"`
}

type ProfileRepository interface {
	Create(ctx context.Context, profile *Profile) error
	FindByID(ctx context.Context, profile_id primitive.ObjectID) (*Profile, error)
	FindByEmail(ctx context.Context, email string) (*Profile, error)
	Update(ctx context.Context, profile_id primitive.ObjectID, profile *Profile) (*Profile, error)
}
