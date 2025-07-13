package migration

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func seedUsers(ctx context.Context, db *mongo.Database) error {

	userID, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345679000")
	if err != nil {
		return err
	}

	roleID, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678000")
	if err != nil {
		return err
	}
	profileID, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345689000")
	if err != nil {
		return err
	}

	_, err = db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{
			"$setOnInsert": bson.M{
				"username":   "superadmin",
				"role_id":    roleID,
				"profile_id": profileID,
				"password":   "$2a$14$Losw5N2WTYnqYa99B69.2eRxWL2nPF3f2gepXnCNkKGAzYGv/aG5G",
				"status":     "active",
				"created_at": time.Now(),
				"updated_at": time.Now(),
				"created_by": userID,
				"updated_by": nil,
				"deleted_by": nil,
				"deleted_at": nil,
				"is_deleted": false,
			},
		},
		options.Update().SetUpsert(true),
	)
	return err
}
