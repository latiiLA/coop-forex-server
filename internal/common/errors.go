package common

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidCredentials         = errors.New("Invalid username or password")
	ErrUserAccessRevoked          = errors.New("User access has been revoked or user is deleted")
	ErrADUserNotFound             = errors.New("User not found in AD")
	ErrUserNotFound               = errors.New("User not found")
	ErrBranchOrDepartmentNotFound = errors.New("either BranchID or DepartmentID must be provided")
	ErrUsernameAlreadyExists      = errors.New("username already exists")

	ErrProfileNotFound = fmt.Errorf("profile not found")

	ErrRoleNotFound          = errors.New("role not found")
	ErrRoleNameAlreadyExists = errors.New("role with this name already exists")
	ErrRoleNameNotAllowed    = errors.New("role name not allowed")
)

var (
	MessInternalServerError = "Internal server error"
	MessUnathorized         = "Unathorized"
	MessInvalidRequest      = "Invalid request"
	MessInvalidRequestData  = "Invalid request data"
	MessInvalidRequestFile  = "Invalid request data"
)
