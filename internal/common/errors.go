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
	ErrSelfSuspension = errors.New("cannot suspend your own account")
	ErrInvalidUserStatus = errors.New("invalid user status for this action")
	ErrAssigneeNotEligible = errors.New("assignee is not an eligible approver")
	ErrSelfAssignOnly = errors.New("you can only assign this request to yourself")

	ErrRequestNotFound            = errors.New("request not found")
	ErrRequestIsLocked            = errors.New("request is already assigned to another user")
	ErrRequestNotAssigned         = errors.New("request must be assigned to you first")
	ErrRequestCannotBeDeleted     = errors.New("request cannot be deleted")
	ErrRequestConflict            = errors.New("request was modified by another action; refresh and retry")
	ErrInvalidRequest             = errors.New("invalid request")
	ErrInvalidRequestTransition   = errors.New("invalid request status transition")
	ErrInvalidAcceptedAmount      = errors.New("accepted amount cannot be greater than approved amount")
	ErrInvalidAcceptedCurrency    = errors.New("accepted currency must be the same approved currency amount")
	ErrLocalCurrencyNotAllowed    = errors.New("ETB cannot be used as an approved or accepted currency")
	ErrAccountCurrencyMustBeETB   = errors.New("account currency must be ETB")
	ErrInvalidCurrency            = errors.New("invalid currency")
	ErrReceiptNotFound            = errors.New("receipt not found")
	ErrInvalidUpload              = errors.New("invalid upload file")
	ErrFileNotFound               = errors.New("file not found")
	ErrFileAccessDenied           = errors.New("file access denied")
)

var (
	MessInternalServerError = "Internal server error"
	MessUnauthorized        = "Unauthorized"
	MessInvalidRequest      = "Invalid request"
	MessInvalidRequestData  = "Invalid request data"
	MessInvalidRequestFile  = "Invalid request data"
	MessRequestLocked       = "Request is assigned to someone else"
	MessRequestNotAssigned  = "Assign the request to yourself first"
	MessRequestNotFound     = "Request not found"
	MessRequestConflict     = "Request was modified by another action; refresh and retry"
)
