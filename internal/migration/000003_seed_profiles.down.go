package migration

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func unseedProfiles(ctx context.Context, db *mongo.Database) error {
	_, err := db.Collection("profiles").DeleteOne(ctx, bson.M{
		"_id": "66a7b8c9e4b0f12345689000",
	})
	return err
}
