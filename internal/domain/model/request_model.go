package model

import (
	"context"
	"mime/multipart"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
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
	PassportAttachment           primitive.ObjectID  `json:"passport_attachment" bson:"passport_attachment"`
	TicketAttachment             primitive.ObjectID  `json:"ticket_attachment" bson:"ticket_attachment"`
	VisaAttachment               *primitive.ObjectID `json:"visa_attachment,omitempty" bson:"visa_attachment,omitempty"`
	BusinessLicenseAttachment    *primitive.ObjectID `json:"business_license_attachment,omitempty" bson:"business_license_attachment,omitempty"`
	EducationLoaAttachment       *primitive.ObjectID `json:"education_loa_attachment,omitempty" bson:"education_loa_attachment,omitempty"`
	HealthLetterAttachment       *primitive.ObjectID `json:"health_letter_attachment,omitempty" bson:"health_letter_attachment,omitempty"`
	BusinessSupportingAttachment *primitive.ObjectID `json:"business_supporting_attachment,omitempty" bson:"business_supporting_attachment,omitempty"`

	Passport           File  `json:"passport" bson:"passport"`
	Ticket             File  `json:"ticket" bson:"ticket"`
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
	ApprovedCurrencyIDs *[]primitive.ObjectID `json:"approved_currency_ids,omitempty" bson:"approved_currency_ids,omitempty"`
	ApprovedCurrencies  *[]Currency           `json:"approved_currencies,omitempty" bson:"approved_currencies,omitempty"`
	ApprovedAmounts     *[]float64            `json:"approved_amounts,omitempty" bson:"approved_amounts,omitempty"`
	AcceptanceStatus    *string               `json:"acceptance_status,omitempty" bson:"acceptance_status,omitempty"`

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

	CreatedBy primitive.ObjectID `json:"created_by" bson:"created_by"`

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

	LockedBy      *primitive.ObjectID `json:"locked_by,omitempty" bson:"locked_by"`
	LockedAt      *time.Time          `json:"locked_at,omitempty" bson:"locked_at"`
	LockExpiresAt *time.Time          `json:"lock_expires_at,omitempty" bson:"lock_expires_at"`
}

type RequestDTO struct {
	ApplicantName          string   `form:"applicant_name" binding:"required"`
	ApplicantAccountNumber string   `form:"applicant_account_number" binding:"required"`
	AverageDeposit         string   `form:"average_deposit" binding:"required"`
	TotalFcyGenerated      string   `form:"total_fcy_generated" binding:"required"`
	CurrentFcyPerformance  string   `form:"current_fcy_performance" binding:"required"`
	TravelPurposeID        string   `form:"travel_purpose_id" binding:"required"`
	TravelCountryID        string   `form:"travel_country_id" binding:"required"`
	RequestingAs           string   `form:"requesting_as" binding:"required"`
	AccountCurrencyID      string   `form:"account_currency_id" binding:"required"`
	FcyRequestedID         string   `form:"fcy_requested_id" binding:"required"`
	FcyRequestedAmount     string   `form:"fcy_requested_amount" binding:"required"`
	AccountsToDeduct       []string `form:"accounts_to_deduct"`
	FcyAcceptanceMode      string   `form:"fcy_acceptance_mode" binding:"required"`
	CardAssociatedAccount  string   `form:"card_associated_account"`
	BranchRecommendation   string   `form:"branch_recommendation" binding:"required"`

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
	ApplicantName          string   `form:"applicant_name" binding:"required"`
	ApplicantAccountNumber string   `form:"applicant_account_number" binding:"required"`
	AverageDeposit         string   `form:"average_deposit" binding:"required,gte=0"`
	TotalFcyGenerated      string   `form:"total_fcy_generated" binding:"required,gte=0"`
	CurrentFcyPerformance  string   `form:"current_fcy_performance" binding:"required,gte=0"`
	TravelPurposeID        string   `form:"travel_purpose_id" binding:"required,len=24"`
	TravelCountryID        string   `form:"travel_country_id" binding:"required,len=24"`
	RequestingAs           string   `form:"requesting_as" binding:"required"`
	AccountCurrencyID      string   `form:"account_currency_id" binding:"required,len=24"`
	FcyRequestedID         string   `form:"fcy_requested_id" binding:"required,len=24"`
	FcyRequestedAmount     string   `form:"fcy_requested_amount" binding:"required,gt=0"`
	AccountsToDeduct       []string `form:"accounts_to_deduct"`
	FcyAcceptanceMode      string   `form:"fcy_acceptance_mode" binding:"required"`
	CardAssociatedAccount  string   `form:"card_associated_account"`
	BranchRecommendation   string   `form:"branch_recommendation" binding:"required"`

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
	ApprovedCurrencyIDs []primitive.ObjectID `json:"approved_currency_ids" binding:"required"`
	ApprovedAmounts     []float64            `json:"approved_amounts" binding:"gte=0"`
}

// Internal model with ObjectIDs
type ParsedApproval struct {
	ApprovedCurrencyIDs []primitive.ObjectID
	ApprovedAmounts     []float64
}

type RequestRepository interface {
	Create(ctx context.Context, request *Request) error
	FindByID(ctx context.Context, request_id primitive.ObjectID) (*Request, error)
	FindAll(ctx context.Context) ([]Request, error)
	Validate(ctx context.Context, request_id primitive.ObjectID, request *Request) error
	FindAllByOrgID(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]Request, error)
	FindOrgByRequestStatus(ctx context.Context, orgID primitive.ObjectID, orgKey, request_status string) ([]Request, error)
	Update(ctx context.Context, requestID primitive.ObjectID, request *Request) error
	FindByRequestStatus(ctx context.Context, request_status string) ([]Request, error)
}
