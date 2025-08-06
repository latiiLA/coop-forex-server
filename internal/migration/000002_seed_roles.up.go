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
	creatorUser, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345679000")
	if err != nil {
		return err
	}

	roleID, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678000")
	if err != nil {
		return err
	}

	roleID2, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678002")
	if err != nil {
		return err
	}

	roleID3, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678003")
	if err != nil {
		return err
	}

	roleID4, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678004")
	if err != nil {
		return err
	}

	roleID5, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678005")
	if err != nil {
		return err
	}

	roleID6, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678006")
	if err != nil {
		return err
	}

	roleID7, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345678007")
	if err != nil {
		return err
	}

	now := time.Now()

	roles := []bson.M{
		{
			"_id":         roleID,
			"name":        "superadmin",
			"created_at":  now,
			"updated_at":  now,
			"created_by":  creatorUser,
			"permissions": []string{"role:add", "role:view", "role:update", "role:view-deleted", "user:view", "user:add", "user:add-permission", "travel_purpose:add", "travel_purpose:view", "subprocess:add", "subprocess:view", "process:add", "process:view", "currency:add", "currency:view", "department:add", "department:view", "country:add", "country:view"},
			"updated_by":  nil,
			"deleted_by":  nil,
			"deleted_at":  nil,
			"is_deleted":  false,
		},
		{
			"_id":         roleID2,
			"name":        "forex_approver",
			"created_at":  now,
			"updated_at":  now,
			"created_by":  creatorUser,
			"permissions": []string{"request:view-validated", "request:view-declined", "request:view", "request:approve", "request:reject", "request:lock", "request:unlock", "request:view-approved", "request:view-rejected", "request:view-accepted"},
			"updated_by":  nil,
			"deleted_by":  nil,
			"deleted_at":  nil,
			"is_deleted":  false,
		},
		{
			"_id":         roleID3,
			"name":        "forex_user",
			"created_at":  now,
			"updated_at":  now,
			"created_by":  creatorUser,
			"permissions": []string{"request:validate", "request:view-authorized", "request:view-validated", "request:view-declined", "request:view", "request:reject", "request:lock", "request:unlock", "request:view-approved", "request:view-rejected", "request:view-accepted"},
			"updated_by":  nil,
			"deleted_by":  nil,
			"deleted_at":  nil,
			"is_deleted":  false,
		},
		{
			"_id":         roleID4,
			"name":        "authorizer",
			"created_at":  now,
			"updated_at":  now,
			"created_by":  creatorUser,
			"permissions": []string{"request:status", "request:view-orgreject", "request:view-orgapproved", "request:view-new", "request:authorize", "request:view-orgauthorized", "request:view-orgrejected"},
			"updated_by":  nil,
			"deleted_by":  nil,
			"deleted_at":  nil,
			"is_deleted":  false,
		},
		{
			"_id":         roleID5,
			"name":        "user",
			"created_at":  now,
			"updated_at":  now,
			"created_by":  creatorUser,
			"permissions": []string{"request:add", "request:status", "request:view-orgapproved", "request:send", "request:view-orgdrafted", "request:update", "request:delete", "request:view-orgauthorized", "request:view-orgrejected"},
			"updated_by":  nil,
			"deleted_by":  nil,
			"deleted_at":  nil,
			"is_deleted":  false,
		},
		{
			"_id":         roleID6,
			"name":        "processor_user",
			"created_at":  now,
			"updated_at":  now,
			"created_by":  creatorUser,
			"permissions": []string{"request:add", "request:status", "request:view-approved", "request:send", "request:view-orgdrafted", "request:update", "request:delete", "request:view-orgauthorized", "request:orgrejected", "request:decline", "request:view-declined", "request:process", "request:view-accepted"},
			"updated_by":  nil,
			"deleted_by":  nil,
			"deleted_at":  nil,
			"is_deleted":  false,
		},
		{
			"_id":         roleID7,
			"name":        "processor_authorizer",
			"created_at":  now,
			"updated_at":  now,
			"created_by":  creatorUser,
			"permissions": []string{"request:status", "request:orgreject", "request:approved", "request:view-new", "request:authorize", "request:view-orgauthorized", "request:view-orgrejected", "request:decline", "request:view-declined", "request:process", "request:view-accepted"},
			"updated_by":  nil,
			"deleted_by":  nil,
			"deleted_at":  nil,
			"is_deleted":  false,
		},
	}

	collection := db.Collection("roles")

	for _, role := range roles {
		filter := bson.M{"_id": role["_id"]}
		update := bson.M{"$setOnInsert": role}
		opts := options.Update().SetUpsert(true)

		_, err := collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return err
		}
	}

	return nil
}
