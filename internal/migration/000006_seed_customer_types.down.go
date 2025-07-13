package migration

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func unseedCustomerTypes(ctx context.Context, db *mongo.Database) error {
	_, err := db.Collection("customer_types").DeleteOne(ctx, bson.M{
		"_id": "66a7b8c9e4b0f12345678010",
	})

	if err != nil {
		return nil
	}

	_, err = db.Collection("customer_types").DeleteOne(ctx, bson.M{
		"_id": "66a7b8c9e4b0f12345678011",
	})
	return err
}
