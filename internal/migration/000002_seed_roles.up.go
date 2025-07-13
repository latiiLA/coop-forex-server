package migration

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func seedRoles(ctx context.Context, db *mongo.Database) error {
	roleID, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678000")
	if err != nil {
		return err
	}
	CreatorUser, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345679000")
	if err != nil {
		return err
	}
	_, err = db.Collection("roles").UpdateOne(
		ctx,
		bson.M{"_id": roleID},
		bson.M{
			"$setOnInsert": bson.M{
				"name":       "superadmin",
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
