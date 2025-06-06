package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	RoleID    primitive.ObjectID  `json:"role_id" bson:"role_id"`
	ProfileID primitive.ObjectID  `json:"profile_id" bson:"profile_id"`
	Username  string              `json:"username" bson:"username"`
	Password  string              `json:"-" bson:"password"`
	Status    string              `json:"status" bson:"status"`
	CreatedAt time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time           `json:"updated_at" bson:"updated_at"`
	CreatedBy primitive.ObjectID  `json:"created_by" bson:"created_by"`
	UpdatedBy *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	DeletedBy *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted bool                `json:"is_deleted,omitempty" bson:"is_deleted,omitempty"`
}

type RegisterRequestDTO struct {
	Username     string              `json:"username" binding:"required,min=3"`
	Password     string              `json:"password" binding:"required,alphanum,min=6"`
	FirstName    string              `json:"first_name" binding:"required,min=3"`
	MiddleName   string              `json:"middle_name" binding:"required,min=3"`
	LastName     string              `json:"last_name" binding:"required,min=3"`
	Email        string              `json:"email" binding:"email"`
	Role         primitive.ObjectID  `json:"role" binding:"required"`
	DepartmentID *primitive.ObjectID `json:"department_id"`
	BranchID     *primitive.ObjectID `json:"branch_id"`
}

type UpdateRequestDTO struct {
	Username     string              `json:"username" binding:"required,min=3"`
	FirstName    string              `json:"first_name" binding:"required,min=3"`
	MiddleName   string              `json:"middle_name" binding:"required,min=3"`
	LastName     string              `json:"last_name" binding:"required,min=3"`
	Email        string              `json:"email" binding:"required"`
	Role         primitive.ObjectID  `json:"role" binding:"required"`
	DepartmentID *primitive.ObjectID `json:"department_id"`
	BranchID     *primitive.ObjectID `json:"branch_id"`
}

type LoginRequestDTO struct {
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponseDTO struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	FirstName  string             `json:"FirstName" bson:"firstName"`
	MiddleName string             `json:"MiddleName" bson:"middleName"`
	Role       string             `json:"role" bson:"role"`
	Username   string             `json:"username" bson:"username"`
	Token      string             `json:"token" bson:"-"`
}

type UserResponseDTO struct {
	ID           primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	RoleID       primitive.ObjectID  `json:"role_id" bson:"role_id"`
	ProfileID    primitive.ObjectID  `json:"profile_id" bson:"profile_id"`
	Username     string              `json:"username" bson:"username"`
	Email        string              `json:"email" bson:"email"`
	Status       string              `json:"status" bson:"status"`
	FirstName    *string             `json:"first_name" bson:"first_name"`
	MiddleName   *string             `json:"middle_name" bson:"middle_name"`
	LastName     *string             `json:"last_name" bson:"last_name"`
	DepartmentID *primitive.ObjectID `json:"department_id" bson:"department_id"`
	BranchID     *primitive.ObjectID `json:"branch_id" bson:"branch_id"`
	CreatedAt    time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at" bson:"updated_at"`
	CreatedBy    *primitive.ObjectID `json:"created_by" bson:"created_by"`
	UpdatedBy    *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	DeletedBy    *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt    *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted    bool                `json:"is_deleted,omitempty" bson:"is_deleted,omitempty"`
}

type UserRepository interface {
	Create(c context.Context, user *User) error
	FindByUsername(c context.Context, username string) (*User, error)
	FindAll(c context.Context) (*[]UserResponseDTO, error)
	FindByID(c context.Context, user_id primitive.ObjectID) (*User, error)
	Update(c context.Context, user_id primitive.ObjectID, user *User) (*User, error)
	Delete(c context.Context, user_id primitive.ObjectID, user *User) error
}
