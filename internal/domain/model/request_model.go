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
	AccountsToDeduct       []string            `json:"accounts_to_deduct[]" bson:"accounts_to_deduct"`
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
	BusinessLicense    *File `json:"business_license" bson:"business_license"`
	EducationLoa       *File `json:"education_loa" bson:"education_loa"`
	HealthLetter       *File `json:"health_letter" bson:"health_letter"`
	BusinessSupporting *File `json:"business_supporting" bson:"business_supporting"`

	// Validation Fields
	ValidatedAverageDeposit    *float64            `json:"validated_average_deposit,omitempty" bson:"validated_average_deposit,omitempty"`
	ValidatedAccountCurrencyID *primitive.ObjectID `json:"validated_account_currency_id,omitempty" bson:"validated_account_currency_id,omitempty"`
	ValidatedAccountCurrency   *Currency           `json:"validated_account_currency,omitempty" bson:"validated_account_currency,omitempty"`
	ValidatedCurrentBalance    *float64            `json:"validated_current_balance,omitempty" bson:"validated_current_balance,omitempty"`

	// Status & remarks
	RequestStatus      string    `json:"request_status" bson:"request_status"`
	Remark             string    `json:"remark,omitempty" bson:"remark,omitempty"`
	RejectionReason    string    `json:"rejection_reason,omitempty" bson:"rejection_reason,omitempty"`
	ApprovedCurrencies *[]string `json:"processed_currencies,omitempty" bson:"processed_currencies,omitempty"`
	ApprovedAmounts    *[]string `json:"approved_amounts,omitempty" bson:"approved_amounts,omitempty"`
	AcceptanceStatus   *string   `json:"acceptance_status,omitempty" bson:"acceptance_status,omitempty"`
	ProcessedAmount    *string   `json:"processed_amount,omitempty" bson:"processed_amount"`

	// Audit
	CreatedAt         time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at" bson:"updated_at"`
	ValidatedAt       *time.Time          `json:"validated_at,omitempty" bson:"validated_at,omitempty"`
	ApprovedAt        *time.Time          `json:"approved_at,omitempty" bson:"approved_at,omitempty"`
	ProcessedAt       *time.Time          `json:"processed_at,omitempty" bson:"processed_at,omitempty"`
	CreatedBy         primitive.ObjectID  `json:"created_by" bson:"created_by"`
	Creator           *User               `json:"creator,omitempty" bson:"creator,omitempty"`
	UpdatedBy         *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	AuthorizedBy      *primitive.ObjectID `json:"authorized_by,omitempty" bson:"authorized_by,omitempty"`
	Authorizer        *User               `json:"authorizer,omitempty" bson:"authorizer,omitempty"`
	ApprovedBy        *primitive.ObjectID `json:"approved_by,omitempty" bson:"approved_by,omitempty"`
	Approver          *User               `json:"approver,omitempty" bson:"approver,omitempty"`
	ValidatedBy       *primitive.ObjectID `json:"validated_by,omitempty" bson:"validated_by,omitempty"`
	Validator         *User               `json:"validator,omitempty" bson:"validator,omitempty"`
	RejectedBy        *primitive.ObjectID `json:"rejected_by,omitempty" bson:"rejected_by,omitempty"`
	Rejector          *User               `json:"rejector,omitempty" bson:"rejector,omitempty"`
	ResultProcessedBy *primitive.ObjectID `json:"result_processed_by,omitempty" bson:"result_processed_by,omitempty"`
	Processor         *User               `json:"processor,omitempty" bson:"processor,omitempty"`
	DeletedBy         *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt         *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted         bool                `json:"is_deleted" bson:"is_deleted"`
}

type RequestDTO struct {
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
	AccountsToDeduct       []string `form:"accounts_to_deduct[]"`
	FcyAcceptanceMode      string   `form:"fcy_acceptance_mode" binding:"required"`
	CardAssociatedAccount  string   `form:"card_associated_account"`
	BranchRecommendation   string   `form:"branch_recommendation" binding:"required"`

	// Attachments (must use `*multipart.FileHeader`)
	PassportAttachment           *multipart.FileHeader `form:"passport_attachment" binding:"required"`
	TicketAttachment             *multipart.FileHeader `form:"ticket_attachment" binding:"required"`
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
	ApprovedCurrencies []string  `json:"approved_currencies[]" binding:"required"`
	ApprovedAmounts    []float64 `json:"approved_amounts[]" binding:"gte=0"`
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
}
