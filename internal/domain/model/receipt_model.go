package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Receipt links a generated PDF receipt number to an accepted request.
// Printable values are read from the request; this record is only the lookup key.
type Receipt struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	ReceiptNumber string             `json:"receipt_number" bson:"receipt_number"`
	RequestID     primitive.ObjectID `json:"request_id" bson:"request_id"`
	GeneratedBy   primitive.ObjectID `json:"generated_by" bson:"generated_by"`
	GeneratorName string             `json:"generator_name" bson:"generator_name"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
}

type ReceiptRepository interface {
	Create(ctx context.Context, receipt *Receipt) (*Receipt, error)
	FindByReceiptNumber(ctx context.Context, receiptNumber string) (*Receipt, error)
}
