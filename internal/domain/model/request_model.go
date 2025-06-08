package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Request struct {
	ID                        primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	BranchID                  primitive.ObjectID  `json:"branch_id" bson:"branch_id"`
	RequestCode               string              `json:"request_code" bson:"request_code"`
	ApplicantName             string              `json:"applicant_name" bson:"applicant_name"`
	ApplicantAccountNumber    []string            `json:"applicant_account_number" bson:"applicant_account_number"`
	AverageDeposit            float64             `json:"average_deposit" bson:"average_deposit"`
	TotalFcyGenerated         time.Time           `json:"total_fcy_generated" bson:"total_fcy_generated"`
	CurrentFcyPerformance     string              `json:"current_fcy_performance" bson:"current_fcy_performance"`
	TravelPurpose             primitive.ObjectID  `json:"travel_purpose" bson:"travel_purpose"`
	TravelCountry             primitive.ObjectID  `json:"travel_country" bson:"travel_country"`
	AccountType               primitive.ObjectID  `json:"account_type" bson:"account_type"`
	AccountCurrency           primitive.ObjectID  `json:"account_currency" bson:"account_currency"`
	DeductionAccount          []string            `json:"deduction_account" bson:"deduction_account"`
	FcyCurrencyRequested      primitive.ObjectID  `json:"fcy_currency_request" bson:"fcy_currecy_request"`
	FcyAmountRequested        float64             `json:"fcy_amount_requested" bson:"fcy_amount_requested"`
	CashAcceptanceMode        string              `json:"cash_acceptance_mode" bson:"cash_acceptance_mode"`
	CardAssociatedAccount     *string             `json:"cash_associated_account,omitempty" bson:"cash_associated_account,omitempty"`
	BranchRecommendation      *string             `json:"branch_recommendation,omitempty" bson:"branch_recommendation,omitempty"`
	BranchComment             *string             `json:"branch_comment,omitempty" bson:"branch_comment,omitempty"`
	PassportAttachment        primitive.ObjectID  `json:"passport_attachment" bson:"passport_attachment"`
	VisaAttachment            primitive.ObjectID  `json:"visa_attachment" bson:"visa_attachment"`
	TicketAttachment          primitive.ObjectID  `json:"ticket_attachment" bson:"ticket_attachment"`
	BusinessLicenseAttachment primitive.ObjectID  `json:"business_license_attachment" bson:"business_license_attachment"`
	EducationLoaAttachment    primitive.ObjectID  `json:"education_loa_attachment" bson:"education_loa_attachment"`
	RequestStatus             string              `json:"request_status" bson:"request_status"`
	Remark                    string              `json:"remark,omitempty" bson:"remark,omitempty"`
	RejectionReason           string              `json:"rejection_reason,omitempty" bson:"rejection_reason,omitempty"`
	ApprovedCurrencies        *[]string           `json:"processed_currencies,omitempty" bson:"processed_currencies,omitempty"`
	AcceptanceStatus          *string             `json:"acceptance_status,omitempty" bson:"acceptance_status,omitempty"`
	ProcessedAmount           *string             `json:"processed_amount,omitempty" bson:"processed_amount"`
	CreatedAt                 time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt                 time.Time           `json:"updated_at" bson:"update_at"`
	CreatedBy                 primitive.ObjectID  `json:"created_by" bson:"created_by"`
	UpdatedBy                 *primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	AuthorizedBy              *primitive.ObjectID `json:"authorized_by,omitempty" bson:"authorized_by,omitempty"`
	ApprovedBy                *primitive.ObjectID `json:"approved_by,omitempty" bson:"approved_by,omitempty"`
	RequestProcessedBy        *primitive.ObjectID `json:"request_processed_by,omitempty" bson:"request_processed_by,empty"`
	RejectedBy                *primitive.ObjectID `json:"rejected_by,omitempty" bson:"processed_by,omitempty"`
	ResultProcessedBy         *primitive.ObjectID `json:"result_processed_by,omitempty" bson:"result_processed_by,omitempty"`
	DeletedBy                 *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	DeletedAt                 *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	IsDeleted                 bool                `json:"id_deleted,omitempty" bson:"is_deleted"`
}
