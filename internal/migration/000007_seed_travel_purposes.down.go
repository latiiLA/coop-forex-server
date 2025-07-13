package migration

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func unseedTravelPurpose(ctx context.Context, db *mongo.Database) error {
	_, err := db.Collection("travel_purpose").DeleteOne(ctx, bson.M{
		"_id": "66a7b8c9e4b0f12345678020",
	})

	if err != nil {
		return nil
	}

	_, err = db.Collection("travel_purpose").DeleteOne(ctx, bson.M{
		"_id": "66a7b8c9e4b0f12345678021",
	})

	if err != nil {
		return nil
	}

	_, err = db.Collection("travel_purpose").DeleteOne(ctx, bson.M{
		"_id": "66a7b8c9e4b0f12345678022",
	})

	if err != nil {
		return nil
	}

	_, err = db.Collection("travel_purpose").DeleteOne(ctx, bson.M{
		"_id": "66a7b8c9e4b0f12345678023",
	})
	return err
}
