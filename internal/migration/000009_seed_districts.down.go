package migration

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func unseedDistricts(ctx context.Context, db *mongo.Database) error {
	userID, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345679000")
	if err != nil {
		return err
	}

	districts := []interface{}{
		bson.M{"_id": "676e44274d7cd2c7251d88e8", "name": "East Finfinne District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e48894d7cd2c7251d88f4", "name": "South Finfinne District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e48a34d7cd2c7251d8900", "name": "West Finfinne District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e48c24d7cd2c7251d890c", "name": "North Finfinne District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e48e34d7cd2c7251d891c", "name": "Central Finfinne District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e49084d7cd2c7251d8928", "name": "Adama District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e494c4d7cd2c7251d8934", "name": "Shashemene District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e49644d7cd2c7251d8940", "name": "Dire Dawa District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e49964d7cd2c7251d894a", "name": "Nekemte District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e49bb4d7cd2c7251d8954", "name": "Hawassa District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e49e34d7cd2c7251d8960", "name": "Jimma District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e4a1b4d7cd2c7251d896c", "name": "Bahirdar District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e4a3a4d7cd2c7251d8978", "name": "Mekelle District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e4a684d7cd2c7251d8984", "name": "Bale District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e4ac14d7cd2c7251d8990", "name": "Asella District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e4adc4d7cd2c7251d899c", "name": "Chiro District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "676e4b0a4d7cd2c7251d89a8", "name": "Hossana District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": "677430c85c2a0438962f6f32", "name": "Other", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
	}

	var deleteModels []mongo.WriteModel
	for _, d := range districts {
		district := d.(bson.M)

		model := mongo.NewDeleteOneModel().
			SetFilter(bson.M{"id": district["id"]})

		deleteModels = append(deleteModels, model)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err = db.Collection("districts").BulkWrite(ctx, deleteModels, opts)
	if err != nil {
		log.Fatal("Error during unseeding districts:", err)
	}

	return err
}
