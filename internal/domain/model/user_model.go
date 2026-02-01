package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID          primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	RoleID      primitive.ObjectID  `json:"role_id" bson:"role_id"`
	Role        *Role               `json:"role,omitempty" bson:"role,omitempty"`
	Permissions *[]string           `json:"permissions,omitempty" bson:"permissions,omitempty"`
	ProfileID   primitive.ObjectID  `json:"profile_id" bson:"profile_id"`
	Profile     *Profile            `json:"profile,omitempty" bson:"profile,omitempty"`
	Username    string              `json:"username" bson:"username"`
	Password    string              `json:"-" bson:"password,omitempty"`
	Status      string              `json:"status" bson:"status"`
	LastLogin   *time.Time          `json:"last_login,omitempty" bson:"last_login,omitempty"`
	CreatedAt   time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at" bson:"updated_at"`
	CreatedBy   primitive.ObjectID  `json:"created_by" bson:"created_by"`
	Creator     *User               `json:"creator,omitempty" bson:"omitempty"`
	Updater     *User               `json:"updater,omitempty" bson:"updater,omitempty"`
	UpdatedBy   *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	DeletedBy   *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt   *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted   bool                `json:"is_deleted" bson:"is_deleted"`
}

type RegisterRequestDTO struct {
	Username     string              `json:"username" binding:"required,min=3"`
	Password     string              `json:"password" binding:"omitempty,alphanum,min=6"`
	FirstName    string              `json:"first_name" binding:"required,min=3"`
	MiddleName   string              `json:"middle_name" binding:"required,min=3"`
	LastName     string              `json:"last_name" binding:"required,min=3"`
	Role         primitive.ObjectID  `json:"role" binding:"required"`
	DepartmentID *primitive.ObjectID `json:"department"`
	BranchID     *primitive.ObjectID `json:"branch"`
}

type RegisterUsecaseRequestDTO struct {
	Username     string
	Password     string
	FirstName    string
	MiddleName   string
	DisplayName  string
	LastName     string
	Role         primitive.ObjectID
	Email        string
	DepartmentID *primitive.ObjectID
	BranchID     *primitive.ObjectID
}

type UpdateUserRequestDTO struct {
	Username     string              `json:"username" binding:"omitempty,min=3"`
	Password     string              `json:"password" binding:"omitempty,alphanum,min=6"`
	FirstName    string              `json:"first_name" binding:"omitempty,min=3"`
	MiddleName   string              `json:"middle_name" binding:"omitempty,min=3"`
	LastName     string              `json:"last_name" binding:"omitempty,min=3"`
	Role         primitive.ObjectID  `json:"role" binding:"omitempty"`
	DepartmentID *primitive.ObjectID `json:"department_id" binding:"omitempty"`
	BranchID     *primitive.ObjectID `json:"branch_id" binding:"omitempty"`
}

type LoginRequestDTO struct {
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponseDTO struct {
	ID           primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	FirstName    string              `json:"first_name" bson:"first_name"`
	MiddleName   string              `json:"middle_name" bson:"middle_name"`
	Role         string              `json:"role" bson:"role"`
	Permissions  []string            `json:"permissions" bson:"permissions"`
	Username     string              `json:"username" bson:"username"`
	Email        string              `json:"email" bson:"email"`
	DepartmentID *primitive.ObjectID `json:"department_id,omitempty" bson:"department_id,omitempty"`
	BranchID     *primitive.ObjectID `json:"branch_id,omitempty" bson:"branch_id,omitempty"`
	Token        string              `json:"token" bson:"-"`
	RefreshToken string              `json:"refresh_token" bson:"-"`
}

type UserResponseDTO struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Role        Role               `json:"role" bson:"role"`
	Permissions *[]string          `json:"permissions,omitempty" bson:"permissions,omitempty"`
	Profile     Profile            `json:"profile" bson:"profile"`
	Username    string             `json:"username" bson:"username"`
	Status      string             `json:"status" bson:"status"`
	Department  *Department        `json:"department,omitempty" bson:"department,omitempty"`
	Branch      *Branch            `json:"branch,omitempty" bson:"branch,omitempty"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
	CreatedBy   string             `json:"created_by" bson:"created_by"`
	Creator     *User              `json:"creator,omitempty" bson:"creator,omitempty"`
	Updater     *User              `json:"updater,omitempty" bson:"updater,omitempty"`
	UpdatedBy   *string            `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	DeletedBy   *string            `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt   *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted   bool               `json:"is_deleted" bson:"is_deleted"`
}

type UserRepository interface {
	Create(c context.Context, user *User) error
	FindByUsername(c context.Context, username string) (*User, error)
	FindAll(c context.Context) (*[]UserResponseDTO, error)
	FindByID(c context.Context, user_id primitive.ObjectID) (*User, error)
	Update(c context.Context, user_id primitive.ObjectID, user *User) (*User, error)
	Delete(c context.Context, user_id primitive.ObjectID, user *User) error
}
