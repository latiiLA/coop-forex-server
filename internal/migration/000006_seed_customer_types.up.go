package migration

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func seedCustomerTypes(ctx context.Context, db *mongo.Database) error {
	customerTypeID1, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678010")
	if err != nil {
		return err
	}
	customerTypeID2, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678011")
	if err != nil {
		return err
	}

	CreatorUser, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345679000")
	if err != nil {
		return err
	}
	_, err = db.Collection("customer_types").UpdateOne(
		ctx,
		bson.M{"_id": customerTypeID1},
		bson.M{
			"$setOnInsert": bson.M{
				"type":       "personal",
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

	_, err = db.Collection("customer_types").UpdateOne(
		ctx,
		bson.M{"_id": customerTypeID2},
		bson.M{
			"$setOnInsert": bson.M{
				"type":       "institution",
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
