package model

import (
	"context"
	"mime/multipart"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestStatus string

const (
	ReqStatusDrafted    RequestStatus = "Drafted"
	ReqStatusNew        RequestStatus = "New"
	ReqStatusAuthorized RequestStatus = "Authorized"
	ReqStatusValidated  RequestStatus = "Validated"
	ReqStatusRejected   RequestStatus = "Rejected"
	ReqStatusApproved   RequestStatus = "Approved"
	ReqStatusAccepted   RequestStatus = "Accepted"
	ReqStatusDeclined   RequestStatus = "Declined"
	ReqStatusDeleted    RequestStatus = "Deleted"
)

type Request struct {
	ID                     primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	BranchID               *primitive.ObjectID `json:"branch_id,omitempty" bson:"branch_id,omitempty"`
	Branch                 *Branch             `json:"branch,omitempty" bson:"branch,omitempty"`
	DepartmentID           *primitive.ObjectID `json:"department_id,omitempty" bson:"department_id,omitempty"`
	Department             *Department         `json:"department,omitempty" bson:"department,omitempty"`
	RequestCode            string              `json:"request_code" bson:"request_code"`
	ApplicantName          string              `json:"applicant_name" bson:"applicant_name"`
	ApplicantAccountNumber string              `json:"applicant_account_number" bson:"applicant_account_number"`
	AverageDeposit         float64             `json:"average_deposit" bson:"average_deposit"`
	TotalFcyGenerated      float64             `json:"total_fcy_generated" bson:"total_fcy_generated"`
	CurrentFcyPerformance  float64             `json:"current_fcy_performance" bson:"current_fcy_performance"`
	TravelPurposeID        primitive.ObjectID  `json:"travel_purpose_id" bson:"travel_purpose_id"`
	TravelPurpose          *TravelPurpose      `json:"travel_purpose,omitempty" bson:"travel_purpose,omitempty"`
	TravelCountryID        primitive.ObjectID  `json:"travel_country_id" bson:"travel_country_id"`
	TravelCountry          *Country            `json:"travel_country,omitempty" bson:"travel_country,omitempty"`
	RequestingAs           string              `json:"requesting_as" bson:"requesting_as"`
	AccountCurrencyID      primitive.ObjectID  `json:"account_currency_id" bson:"account_currency_id"`
	AccountCurrency        *Currency           `json:"account_currency,omitempty" bson:"account_currency,omitempty"`
	FcyRequestedID         primitive.ObjectID  `json:"fcy_requested_id" bson:"fcy_requested_id"`
	FcyRequested           *Currency           `json:"fcy_requested,omitempty" bson:"fcy_requested,omitempty"`
	FcyRequestedAmount     float64             `json:"fcy_requested_amount" bson:"fcy_requested_amount"`
	AccountsToDeduct       []string            `json:"accounts_to_deduct" bson:"accounts_to_deduct"`
	FcyAcceptanceMode      string              `json:"fcy_acceptance_mode" bson:"fcy_acceptance_mode"`
	CardAssociatedAccount  *string             `json:"card_associated_account,omitempty" bson:"card_associated_account,omitempty"`
	BranchRecommendation   *string             `json:"branch_recommendation,omitempty" bson:"branch_recommendation,omitempty"`

	// Attachments
	PassportAttachment           *primitive.ObjectID `json:"passport_attachment,omitempty" bson:"passport_attachment,omitempty"`
	TicketAttachment             *primitive.ObjectID `json:"ticket_attachment,omitempty" bson:"ticket_attachment,omitempty"`
	VisaAttachment               *primitive.ObjectID `json:"visa_attachment,omitempty" bson:"visa_attachment,omitempty"`
	BusinessLicenseAttachment    *primitive.ObjectID `json:"business_license_attachment,omitempty" bson:"business_license_attachment,omitempty"`
	EducationLoaAttachment       *primitive.ObjectID `json:"education_loa_attachment,omitempty" bson:"education_loa_attachment,omitempty"`
	HealthLetterAttachment       *primitive.ObjectID `json:"health_letter_attachment,omitempty" bson:"health_letter_attachment,omitempty"`
	BusinessSupportingAttachment *primitive.ObjectID `json:"business_supporting_attachment,omitempty" bson:"business_supporting_attachment,omitempty"`

	Passport           *File `json:"passport,omitempty" bson:"passport,omitempty"`
	Ticket             *File `json:"ticket,omitempty" bson:"ticket,omitempty"`
	Visa               *File `json:"visa,omitempty" bson:"visa,omitempty"`
	BusinessLicense    *File `json:"business_license,omitempty" bson:"business_license,omitempty"`
	EducationLoa       *File `json:"education_loa,omitempty" bson:"education_loa,omitempty"`
	HealthLetter       *File `json:"health_letter,omitempty" bson:"health_letter,omitempty"`
	BusinessSupporting *File `json:"business_supporting,omitempty" bson:"business_supporting,omitempty"`

	// Validation Fields
	ValidatedAverageDeposit    *float64            `json:"validated_average_deposit,omitempty" bson:"validated_average_deposit,omitempty"`
	ValidatedAccountCurrencyID *primitive.ObjectID `json:"validated_account_currency_id,omitempty" bson:"validated_account_currency_id,omitempty"`
	ValidatedAccountCurrency   *Currency           `json:"validated_account_currency,omitempty" bson:"validated_account_currency,omitempty"`
	ValidatedCurrentBalance    *float64            `json:"validated_current_balance,omitempty" bson:"validated_current_balance,omitempty"`

	// Approved Fields
	ApprovedCurrencyIDs  []primitive.ObjectID `json:"approved_currency_ids,omitempty" bson:"approved_currency_ids,omitempty"`
	ApprovedCurrencies   []Currency           `json:"approved_currencies,omitempty" bson:"approved_currencies,omitempty"`
	ApprovedAmounts      []float64            `json:"approved_amounts,omitempty" bson:"approved_amounts,omitempty"`
	ApprovedAmountInCash []float64            `json:"approved_amount_in_cash,omitempty" bson:"approved_amount_in_cash,omitempty"`
	ApprovedAmountInCard []float64            `json:"approved_amount_in_card,omitempty" bson:"approved_amount_in_card,omitempty"`

	// Accepted Fields
	AcceptedCurrencyIDs  []primitive.ObjectID `json:"accepted_currency_ids,omitempty" bson:"accepted_currency_ids,omitempty"`
	AcceptedCurrencies   []Currency           `json:"accepted_currencies,omitempty" bson:"accepted_currencies,omitempty"`
	AcceptedAmounts      []float64            `json:"accepted_amounts,omitempty" bson:"accepted_amounts,omitempty"`
	AcceptedAmountInCash []float64            `json:"accepted_amount_in_cash,omitempty" bson:"accepted_amount_in_cash,omitempty"`
	AcceptedAmountInCard []float64            `json:"accepted_amount_in_card,omitempty" bson:"accepted_amount_in_card,omitempty"`

	// Status & remarks
	RequestStatus   RequestStatus `json:"request_status" bson:"request_status"`
	Remark          string        `json:"remark,omitempty" bson:"remark,omitempty"`
	RejectionReason string        `json:"rejection_reason,omitempty" bson:"rejection_reason,omitempty"`
	ProcessedAmount *float64      `json:"processed_amount,omitempty" bson:"processed_amount,omitempty"`
	DueDate         *time.Time    `json:"due_date,omitempty" bson:"due_date,omitempty"`

	// Audit
	CreatedAt    time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" bson:"updated_at"`
	RequestedAt  *time.Time `json:"requested_at,omitempty" bson:"requested_at,omitempty"`
	AuthorizedAt *time.Time `json:"authorized_at,omitempty" bson:"authorized_at,omitempty"`
	RejectedAt   *time.Time `json:"rejected_at,omitempty" bson:"rejected_at,omitempty"`
	ValidatedAt  *time.Time `json:"validated_at,omitempty" bson:"validated_at,omitempty"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty" bson:"approved_at,omitempty"`
	AcceptedAt   *time.Time `json:"accepted_at,omitempty" bson:"accepted_at,omitempty"`
	DeclinedAt   *time.Time `json:"declined_at,omitempty" bson:"declined_at,omitempty"`

	CreatedBy    primitive.ObjectID  `json:"created_by" bson:"created_by"`
	Creator      *User               `json:"creator,omitempty" bson:"creator,omitempty"`
	RequestedBy  *primitive.ObjectID `json:"requested_by,omitempty" bson:"requested_by,omitempty"`
	Requester    *User               `json:"requester,omitempty" bson:"requester,omitempty"`
	UpdatedBy    *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	AuthorizedBy *primitive.ObjectID `json:"authorized_by,omitempty" bson:"authorized_by,omitempty"`
	Authorizer   *User               `json:"authorizer,omitempty" bson:"authorizer,omitempty"`
	ApprovedBy   *primitive.ObjectID `json:"approved_by,omitempty" bson:"approved_by,omitempty"`
	Approver     *User               `json:"approver,omitempty" bson:"approver,omitempty"`
	ValidatedBy  *primitive.ObjectID `json:"validated_by,omitempty" bson:"validated_by,omitempty"`
	Validater    *User               `json:"validater,omitempty" bson:"validater,omitempty"`
	RejectedBy   *primitive.ObjectID `json:"rejected_by,omitempty" bson:"rejected_by,omitempty"`
	Rejecter     *User               `json:"rejecter,omitempty" bson:"rejecter,omitempty"`
	AcceptedBy   *primitive.ObjectID `json:"accepted_by,omitempty" bson:"accepted_by,omitempty"`
	Accepter     *User               `json:"accepter,omitempty" bson:"accepter,omitempty"`
	DeclinedBy   *primitive.ObjectID `json:"declined_by,omitempty" bson:"declined_by,omitempty"`
	Decliner     *User               `json:"decliner,omitempty" bson:"decliner,omitempty"`
	DeletedBy    *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt    *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted    bool                `json:"is_deleted" bson:"is_deleted"`

	// AssignedTo is the validator/authorizer assignee (Mongo field locked_by for legacy).
	// Used for Authorized/Validated claim + queue visibility — not for Accept/Decline.
	AssignedTo    *primitive.ObjectID `json:"assigned_to,omitempty" bson:"locked_by,omitempty"`
	AssignedAt    *time.Time          `json:"assigned_at,omitempty" bson:"locked_at,omitempty"`
	Assignee      *User               `json:"assignee,omitempty" bson:"assignee,omitempty"`
	LockExpiresAt *time.Time          `json:"lock_expires_at,omitempty" bson:"lock_expires_at,omitempty"` // legacy; cleared on assign/unassign

	// ProcessingBranchID is the result-processor branch that accepted the request.
	ProcessingBranchID *primitive.ObjectID `json:"processing_branch_id,omitempty" bson:"processing_branch_id,omitempty"`
	ProcessingBranch   *Branch             `json:"processing_branch,omitempty" bson:"processing_branch,omitempty"`
}

type RequestUpdate struct {
	ID                     primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	BranchID               *primitive.ObjectID `json:"branch_id,omitempty" bson:"branch_id,omitempty"`
	DepartmentID           *primitive.ObjectID `json:"department_id,omitempty" bson:"department_id,omitempty"`
	RequestCode            string              `json:"request_code" bson:"request_code"`
	ApplicantName          string              `json:"applicant_name" bson:"applicant_name"`
	ApplicantAccountNumber string              `json:"applicant_account_number" bson:"applicant_account_number"`
	AverageDeposit         float64             `json:"average_deposit" bson:"average_deposit"`
	TotalFcyGenerated      float64             `json:"total_fcy_generated" bson:"total_fcy_generated"`
	CurrentFcyPerformance  float64             `json:"current_fcy_performance" bson:"current_fcy_performance"`
	TravelPurposeID        primitive.ObjectID  `json:"travel_purpose_id" bson:"travel_purpose_id"`
	TravelCountryID        primitive.ObjectID  `json:"travel_country_id" bson:"travel_country_id"`
	RequestingAs           string              `json:"requesting_as" bson:"requesting_as"`
	AccountCurrencyID      primitive.ObjectID  `json:"account_currency_id" bson:"account_currency_id"`
	FcyRequestedID         primitive.ObjectID  `json:"fcy_requested_id" bson:"fcy_requested_id"`
	FcyRequestedAmount     float64             `json:"fcy_requested_amount" bson:"fcy_requested_amount"`
	AccountsToDeduct       []string            `json:"accounts_to_deduct" bson:"accounts_to_deduct"`
	FcyAcceptanceMode      string              `json:"fcy_acceptance_mode" bson:"fcy_acceptance_mode"`
	CardAssociatedAccount  *string             `json:"card_associated_account,omitempty" bson:"card_associated_account,omitempty"`
	BranchRecommendation   *string             `json:"branch_recommendation,omitempty" bson:"branch_recommendation,omitempty"`

	// Attachments
	PassportAttachment           *primitive.ObjectID `json:"passport_attachment,omitempty" bson:"passport_attachment,omitempty"`
	TicketAttachment             *primitive.ObjectID `json:"ticket_attachment,omitempty" bson:"ticket_attachment,omitempty"`
	VisaAttachment               *primitive.ObjectID `json:"visa_attachment,omitempty" bson:"visa_attachment,omitempty"`
	BusinessLicenseAttachment    *primitive.ObjectID `json:"business_license_attachment,omitempty" bson:"business_license_attachment,omitempty"`
	EducationLoaAttachment       *primitive.ObjectID `json:"education_loa_attachment,omitempty" bson:"education_loa_attachment,omitempty"`
	HealthLetterAttachment       *primitive.ObjectID `json:"health_letter_attachment,omitempty" bson:"health_letter_attachment,omitempty"`
	BusinessSupportingAttachment *primitive.ObjectID `json:"business_supporting_attachment,omitempty" bson:"business_supporting_attachment,omitempty"`

	// Validation Fields
	ValidatedAverageDeposit    *float64            `json:"validated_average_deposit,omitempty" bson:"validated_average_deposit,omitempty"`
	ValidatedAccountCurrencyID *primitive.ObjectID `json:"validated_account_currency_id,omitempty" bson:"validated_account_currency_id,omitempty"`
	ValidatedCurrentBalance    *float64            `json:"validated_current_balance,omitempty" bson:"validated_current_balance,omitempty"`

	// Approved Fields
	ApprovedCurrencyIDs  []primitive.ObjectID `json:"approved_currency_ids,omitempty" bson:"approved_currency_ids,omitempty"`
	ApprovedAmounts      []float64            `json:"approved_amounts,omitempty" bson:"approved_amounts,omitempty"`
	ApprovedAmountInCash []float64            `json:"approved_amount_in_cash,omitempty" bson:"approved_amount_in_cash,omitempty"`
	ApprovedAmountInCard []float64            `json:"approved_amount_in_card,omitempty" bson:"approved_amount_in_card,omitempty"`

	AcceptedCurrencyIDs  []primitive.ObjectID `json:"accepted_currency_ids,omitempty" bson:"accepted_currency_ids,omitempty"`
	AcceptedAmounts      []float64            `json:"accepted_amounts,omitempty" bson:"accepted_amounts,omitempty"`
	AcceptedAmountInCash []float64            `json:"accepted_amount_in_cash,omitempty" bson:"accepted_amount_in_cash,omitempty"`
	AcceptedAmountInCard []float64            `json:"accepted_amount_in_card,omitempty" bson:"accepted_amount_in_card,omitempty"`

	// Status & remarks
	RequestStatus   string     `json:"request_status" bson:"request_status"`
	Remark          string     `json:"remark,omitempty" bson:"remark,omitempty"`
	RejectionReason string     `json:"rejection_reason,omitempty" bson:"rejection_reason,omitempty"`
	ProcessedAmount *float64   `json:"processed_amount,omitempty" bson:"processed_amount,omitempty"`
	DueDate         *time.Time `json:"due_date,omitempty" bson:"due_date,omitempty"`

	// Audit
	CreatedAt        time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" bson:"updated_at"`
	RequestedAt      *time.Time `json:"requested_at,omitempty" bson:"requested_at,omitempty"`
	AuthorizedAt     *time.Time `json:"authorized_at,omitempty" bson:"authorized_at,omitempty"`
	RejectedAt       *time.Time `json:"rejected_at,omitempty" bson:"rejected_at,omitempty"`
	ValidatedAt      *time.Time `json:"validated_at,omitempty" bson:"validated_at,omitempty"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty" bson:"approved_at,omitempty"`
	AcceptedAt       *time.Time `json:"accepted_at,omitempty" bson:"accepted_at,omitempty"`
	DeclinedAt       *time.Time `json:"declined_at,omitempty" bson:"declined_at,omitempty"`
	SystemDeclinedAt *time.Time `json:"system_declined_at,omitempty" bson:"system_declined_at,omitempty"`

	CreatedBy    primitive.ObjectID  `json:"created_by" bson:"created_by"`
	RequestedBy  *primitive.ObjectID `json:"requested_by,omitempty" bson:"requested_by,omitempty"`
	UpdatedBy    *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	AuthorizedBy *primitive.ObjectID `json:"authorized_by,omitempty" bson:"authorized_by,omitempty"`
	ApprovedBy   *primitive.ObjectID `json:"approved_by,omitempty" bson:"approved_by,omitempty"`
	ValidatedBy  *primitive.ObjectID `json:"validated_by,omitempty" bson:"validated_by,omitempty"`
	RejectedBy   *primitive.ObjectID `json:"rejected_by,omitempty" bson:"rejected_by,omitempty"`
	AcceptedBy   *primitive.ObjectID `json:"accepted_by,omitempty" bson:"accepted_by,omitempty"`
	DeclinedBy   *primitive.ObjectID `json:"declined_by,omitempty" bson:"declined_by,omitempty"`
	DeletedBy    *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt    *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted    bool                `json:"is_deleted" bson:"is_deleted"`

	AssignedTo    *primitive.ObjectID `json:"assigned_to,omitempty" bson:"locked_by,omitempty"`
	AssignedAt    *time.Time          `json:"assigned_at,omitempty" bson:"locked_at,omitempty"`
	LockExpiresAt *time.Time          `json:"lock_expires_at,omitempty" bson:"lock_expires_at,omitempty"`

	ProcessingBranchID *primitive.ObjectID `json:"processing_branch_id,omitempty" bson:"processing_branch_id,omitempty"`
}

type RequestDTO struct {
	ApplicantName          string   `form:"applicant_name" binding:"required,excludesall=<>"`
	ApplicantAccountNumber string   `form:"applicant_account_number" binding:"required,alphanum,len=16"`
	AverageDeposit         string   `form:"average_deposit" binding:"required,numeric"`
	TotalFcyGenerated      string   `form:"total_fcy_generated" binding:"required,numeric"`
	CurrentFcyPerformance  string   `form:"current_fcy_performance" binding:"required,numeric"`
	TravelPurposeID        string   `form:"travel_purpose_id" binding:"required,len=24,hexadecimal"`
	TravelCountryID        string   `form:"travel_country_id" binding:"required,len=24,hexadecimal"`
	RequestingAs           string   `form:"requesting_as" binding:"required,oneof=Individual Institution"`
	AccountCurrencyID      string   `form:"account_currency_id" binding:"required,len=24,hexadecimal"`
	FcyRequestedID         string   `form:"fcy_requested_id" binding:"required,len=24,hexadecimal"`
	FcyRequestedAmount     string   `form:"fcy_requested_amount" binding:"required,numeric"`
	AccountsToDeduct       []string `form:"accounts_to_deduct" binding:"required,min=1,dive,required,alphanum,len=16"`
	FcyAcceptanceMode      string   `form:"fcy_acceptance_mode" binding:"required,oneof=Cash Card Cash-and-Card"`
	CardAssociatedAccount  string   `form:"card_associated_account" binding:"omitempty,alphanum,len=16"`
	BranchRecommendation   string   `form:"branch_recommendation" binding:"required,min=3,max=200,excludesall=<>"`
	AuthorizerID           string   `form:"authorizer_id" binding:"required,len=24,hexadecimal"`

	// Attachments (must use `*multipart.FileHeader`)
	PassportAttachment           *multipart.FileHeader `form:"passport_attachment"`
	TicketAttachment             *multipart.FileHeader `form:"ticket_attachment"`
	VisaAttachment               *multipart.FileHeader `form:"visa_attachment"`
	BusinessLicenseAttachment    *multipart.FileHeader `form:"business_license_attachment"`
	EducationLoaAttachment       *multipart.FileHeader `form:"education_loa_attachment"`
	HealthLetterAttachment       *multipart.FileHeader `form:"health_letter_attachment"`
	BusinessSupportingAttachment *multipart.FileHeader `form:"business_supporting_attachment"`
}

type UpdateRequestDTO struct {
	ApplicantName          string   `form:"applicant_name" binding:"required,excludesall=<>"`
	ApplicantAccountNumber string   `form:"applicant_account_number" binding:"required,alphanum,len=16"`
	AverageDeposit         string   `form:"average_deposit" binding:"required,numeric"`
	TotalFcyGenerated      string   `form:"total_fcy_generated" binding:"required,numeric"`
	CurrentFcyPerformance  string   `form:"current_fcy_performance" binding:"required,numeric"`
	TravelPurposeID        string   `form:"travel_purpose_id" binding:"required,len=24,hexadecimal"`
	TravelCountryID        string   `form:"travel_country_id" binding:"required,len=24,hexadecimal"`
	RequestingAs           string   `form:"requesting_as" binding:"required,oneof=Individual Institution"`
	AccountCurrencyID      string   `form:"account_currency_id" binding:"required,len=24,hexadecimal"`
	FcyRequestedID         string   `form:"fcy_requested_id" binding:"required,len=24,hexadecimal"`
	FcyRequestedAmount     string   `form:"fcy_requested_amount" binding:"required,numeric"`
	AccountsToDeduct       []string `form:"accounts_to_deduct" binding:"required,min=1,dive,required,alphanum,len=16"`
	FcyAcceptanceMode      string   `form:"fcy_acceptance_mode" binding:"required,oneof=Cash Card Cash-and-Card"`
	CardAssociatedAccount  string   `form:"card_associated_account" binding:"omitempty,alphanum,len=16"`
	BranchRecommendation   string   `form:"branch_recommendation" binding:"required,min=3,max=200,excludesall=<>"`
	AuthorizerID           string   `form:"authorizer_id" binding:"required,len=24,hexadecimal"`

	// Attachments (must use `*multipart.FileHeader`)
	PassportAttachment           *multipart.FileHeader `form:"passport_attachment"`
	TicketAttachment             *multipart.FileHeader `form:"ticket_attachment"`
	VisaAttachment               *multipart.FileHeader `form:"visa_attachment"`
	BusinessLicenseAttachment    *multipart.FileHeader `form:"business_license_attachment"`
	EducationLoaAttachment       *multipart.FileHeader `form:"education_loa_attachment"`
	HealthLetterAttachment       *multipart.FileHeader `form:"health_letter_attachment"`
	BusinessSupportingAttachment *multipart.FileHeader `form:"business_supporting_attachment"`
}

type RequestValidationDTO struct {
	ValidatedAverageDeposit    float64 `json:"validated_average_deposit" binding:"gte=0"`
	ValidatedAccountCurrencyID string  `json:"validated_account_currency_id" binding:"required,len=24"`
	ValidatedCurrentBalance    float64 `json:"validated_current_balance" binding:"gte=0"`
}

type RequestApprovalDTO struct {
	ApprovedCurrencyIDs  []primitive.ObjectID `json:"approved_currency_ids" binding:"required"`
	ApprovedAmounts      []float64            `json:"approved_amounts" binding:"required,dive,gte=0"`
	ApprovedAmountInCash []float64            `json:"approved_amount_in_cash" binding:"required,dive,gte=0"`
	ApprovedAmountInCard []float64            `json:"approved_amount_in_card" binding:"required,dive,gte=0"`
}

type RequestAcceptanceDTO struct {
	AcceptedCurrencyIDs  []primitive.ObjectID `json:"accepted_currency_ids" binding:"required"`
	AcceptedAmounts      []float64            `json:"accepted_amounts" binding:"required,dive,gte=0"`
	AcceptedAmountInCash []float64            `json:"accepted_amount_in_cash" binding:"required,dive,gte=0"`
	AcceptedAmountInCard []float64            `json:"accepted_amount_in_card" binding:"required,dive,gte=0"`
}

// Internal model with ObjectIDs
type ParsedApproval struct {
	ApprovedCurrencyIDs []primitive.ObjectID
	ApprovedAmounts     []float64
}

type RequestRepository interface {
	Create(ctx context.Context, request *Request) error
	FindByID(ctx context.Context, request_id primitive.ObjectID, populate bool) (*Request, error)
	FindAll(ctx context.Context, populate bool, pageQuery PageQuery) (*PaginatedRequests, error)
	Validate(ctx context.Context, request_id primitive.ObjectID, request *Request) error
	FindAllByOrgID(ctx context.Context, orgKey string, orgID primitive.ObjectID, populate bool, pageQuery PageQuery) (*PaginatedRequests, error)
	// FindOrgByRequestStatus lists org-scoped rows. When assigneeViewer is set, only
	// unassigned/expired and assigned-to-viewer rows are returned.
	FindOrgByRequestStatus(ctx context.Context, orgID primitive.ObjectID, orgKey, request_status string, populate bool, pageQuery PageQuery, assigneeViewer *primitive.ObjectID) (*PaginatedRequests, error)
	Update(ctx context.Context, requestID primitive.ObjectID, request *RequestUpdate) error
	// UpdateIfStatus updates only when current request_status is in expectedStatuses (atomic transition).
	UpdateIfStatus(ctx context.Context, requestID primitive.ObjectID, expectedStatuses []string, request *RequestUpdate) error
	// ValidateIfStatus validates only when status matches expectedStatuses.
	ValidateIfStatus(ctx context.Context, requestID primitive.ObjectID, expectedStatuses []string, request *Request) error
	// FindByRequestStatus lists by status. When assigneeViewer is set, only unassigned
	// (or expired) and assigned-to-viewer rows are returned (authorizer/validator queues).
	FindByRequestStatus(ctx context.Context, request_status string, populate bool, pageQuery PageQuery, assigneeViewer *primitive.ObjectID) (*PaginatedRequests, error)
	FindByCreatorAndStatus(ctx context.Context, creatorID primitive.ObjectID, request_status string, populate bool, pageQuery PageQuery) (*PaginatedRequests, error)
	CountByStatus(ctx context.Context, orgKey string, orgID primitive.ObjectID) (*RequestStatusSummary, error)
	Assign(ctx context.Context, requestID, userID primitive.ObjectID, assignedAt time.Time) error
	// AssignIfAvailable atomically claims an unassigned (or self-owned / expired) request. Returns false on collision.
	AssignIfAvailable(ctx context.Context, requestID, userID primitive.ObjectID, assignedAt time.Time) (bool, error)
	ClearAssignment(ctx context.Context, requestID primitive.ObjectID) error
	// FindByAttachmentFileID finds a request that references the given uploaded file id.
	FindByAttachmentFileID(ctx context.Context, fileID primitive.ObjectID) (*Request, error)
}

// AssignRequestDTO allows approvers to assign to another eligible approver.
type AssignRequestDTO struct {
	AssigneeID *string `json:"assignee_id,omitempty"`
}

// AssignRequestResult is returned after a successful assignment.
type AssignRequestResult struct {
	AssignedTo *primitive.ObjectID `json:"assigned_to"`
	AssignedAt time.Time           `json:"assigned_at"`
	Assignee   *User               `json:"assignee,omitempty"`
}
