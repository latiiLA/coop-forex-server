package migration

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func seedProfiles(ctx context.Context, db *mongo.Database) error {
	profileID, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345689000")
	if err != nil {
		return err
	}

	_, err = db.Collection("profiles").UpdateOne(
		ctx,
		bson.M{"_id": profileID},
		bson.M{
			"$setOnInsert": bson.M{
				"email":         "superadmin@latii.com",
				"first_name":    "Super",
				"middle_name":   "Admin",
				"last_name":     "User",
				"department_id": nil,
				"branch_id":     nil,
				"is_deleted":    false,
			},
		},
		options.Update().SetUpsert(true),
	)

	return err
}
