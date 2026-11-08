package repository

import (
	"context"
	"fmt"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type receiptRepository struct {
	collection *mongo.Collection
}

func NewReceiptRepository(db *mongo.Database) model.ReceiptRepository {
	collection := db.Collection("receipts")
	_, _ = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "receipt_number", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	_, _ = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{Key: "request_id", Value: 1}, {Key: "created_at", Value: -1}},
	})
	return &receiptRepository{collection: collection}
}

func (rr *receiptRepository) Create(ctx context.Context, receipt *model.Receipt) (*model.Receipt, error) {
	result, err := rr.collection.InsertOne(ctx, receipt)
	if err != nil {
		return nil, err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("failed to convert inserted ID to ObjectID")
	}
	receipt.ID = id
	return receipt, nil
}

func (rr *receiptRepository) FindByReceiptNumber(ctx context.Context, receiptNumber string) (*model.Receipt, error) {
	var receipt model.Receipt
	err := rr.collection.FindOne(ctx, bson.M{"receipt_number": receiptNumber}).Decode(&receipt)
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}
