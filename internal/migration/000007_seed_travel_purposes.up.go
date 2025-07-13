package migration

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func seedTravelPurpose(ctx context.Context, db *mongo.Database) error {
	travelPurposeID1, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678020")
	if err != nil {
		return err
	}
	travelPurposeID2, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678021")
	if err != nil {
		return err
	}

	travelPurposeID3, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678022")
	if err != nil {
		return err
	}

	travelPurposeID4, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678023")
	if err != nil {
		return err
	}

	CreatorUser, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345679000")
	if err != nil {
		return err
	}
	_, err = db.Collection("travel_purposes").UpdateOne(
		ctx,
		bson.M{"_id": travelPurposeID1},
		bson.M{
			"$setOnInsert": bson.M{
				"purpose":    "business",
				"created_at": time.Now(),
				"updated_at": time.Now(),
				"created_by": CreatorUser,
				"updated_by": nil,
				"deleted_by": nil,
				"deleted_at": nil,
				"is_deleted": false,
			},
		}, options.Update().SetUpsert(true))

	if err != nil {
		return err
	}

	_, err = db.Collection("travel_purpose").UpdateOne(
		ctx,
		bson.M{"_id": travelPurposeID2},
		bson.M{
			"$setOnInsert": bson.M{
				"purpose":    "visit",
				"created_at": time.Now(),
				"updated_at": time.Now(),
				"created_by": CreatorUser,
				"updated_by": nil,
				"deleted_by": nil,
				"deleted_at": nil,
				"is_deleted": false,
			},
		}, options.Update().SetUpsert(true))

	if err != nil {
		return err
	}

	_, err = db.Collection("travel_purpose").UpdateOne(
		ctx,
		bson.M{"_id": travelPurposeID3},
		bson.M{
			"$setOnInsert": bson.M{
				"purpose":    "education",
				"created_at": time.Now(),
				"updated_at": time.Now(),
				"created_by": CreatorUser,
				"updated_by": nil,
				"deleted_by": nil,
				"deleted_at": nil,
				"is_deleted": false,
			},
		}, options.Update().SetUpsert(true))

	if err != nil {
		return err
	}

	_, err = db.Collection("travel_purpose").UpdateOne(
		ctx,
		bson.M{"_id": travelPurposeID4},
		bson.M{
			"$setOnInsert": bson.M{
				"purpose":    "health",
				"created_at": time.Now(),
				"updated_at": time.Now(),
				"created_by": CreatorUser,
				"updated_by": nil,
				"deleted_by": nil,
				"deleted_at": nil,
				"is_deleted": false,
			},
		}, options.Update().SetUpsert(true))
	return err
}
