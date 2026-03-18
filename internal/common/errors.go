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

	ErrUnauthorized   = errors.New("unauthorized")
	ErrInternalServer = errors.New("internal server error")

	ErrRequestNotFound         = errors.New("request not found")
	ErrRequestIsLocked         = errors.New("request is already locked by another user")
	ErrRequestCannotBeDeleted  = errors.New("request cannot be deleted")
	ErrInvalidAcceptedAmount   = errors.New("accepted amount cannot be greater than approved amount")
	ErrInvalidAcceptedCurrency = errors.New("accepted currency must be the same approved currency amount")
)

var (
	MessInternalServerError = "Internal server error"
	MessUnauthorized        = "Unauthorized"
	MessInvalidRequest      = "Invalid request"
	MessInvalidRequestData  = "Invalid request data"
	MessInvalidRequestFile  = "Invalid request data"
	MessRequestLocked       = "Request is Locked by someone else"
	MessRequestNotFound     = "Request not found"
)
