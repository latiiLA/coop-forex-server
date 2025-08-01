package migration

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func seedDistricts(ctx context.Context, db *mongo.Database) error {

	userID, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345679000")
	if err != nil {
		return err
	}

	id1, _ := primitive.ObjectIDFromHex("676e44274d7cd2c7251d88e8")
	id2, _ := primitive.ObjectIDFromHex("676e48894d7cd2c7251d88f4")
	id3, _ := primitive.ObjectIDFromHex("676e48a34d7cd2c7251d8900")
	id4, _ := primitive.ObjectIDFromHex("676e48c24d7cd2c7251d890c")
	id5, _ := primitive.ObjectIDFromHex("676e48e34d7cd2c7251d891c")

	id6, _ := primitive.ObjectIDFromHex("676e49084d7cd2c7251d8928")
	id7, _ := primitive.ObjectIDFromHex("676e494c4d7cd2c7251d8934")
	id8, _ := primitive.ObjectIDFromHex("676e49644d7cd2c7251d8940")
	id9, _ := primitive.ObjectIDFromHex("676e49964d7cd2c7251d894a")
	id10, _ := primitive.ObjectIDFromHex("676e49bb4d7cd2c7251d8954")

	id11, _ := primitive.ObjectIDFromHex("676e49e34d7cd2c7251d8960")
	id12, _ := primitive.ObjectIDFromHex("676e4a1b4d7cd2c7251d896c")
	id13, _ := primitive.ObjectIDFromHex("676e4a3a4d7cd2c7251d8978")
	id14, _ := primitive.ObjectIDFromHex("676e4a684d7cd2c7251d8984")
	id15, _ := primitive.ObjectIDFromHex("676e4ac14d7cd2c7251d8990")

	id16, _ := primitive.ObjectIDFromHex("676e4adc4d7cd2c7251d899c")
	id17, _ := primitive.ObjectIDFromHex("676e4b0a4d7cd2c7251d89a8")
	id18, _ := primitive.ObjectIDFromHex("677430c85c2a0438962f6f32")

	districts := []interface{}{
		bson.M{"_id": id1, "name": "East Finfinne District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id2, "name": "South Finfinne District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id3, "name": "West Finfinne District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id4, "name": "North Finfinne District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id5, "name": "Central Finfinne District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id6, "name": "Adama District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id7, "name": "Shashemene District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id8, "name": "Dire Dawa District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id9, "name": "Nekemte District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id10, "name": "Hawassa District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id11, "name": "Jimma District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id12, "name": "Bahirdar District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id13, "name": "Mekelle District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id14, "name": "Bale District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id15, "name": "Asella District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id16, "name": "Chiro District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id17, "name": "Hossana District", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id18, "name": "Head Office", "address": "Finfinne", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
	}

	var models []mongo.WriteModel
	for _, d := range districts {
		district := d.(bson.M)
		id := district["_id"]
		delete(district, "_id") // remove _id from update to avoid immutable error

		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": id}).
			SetUpdate(bson.M{"$set": district}).
			SetUpsert(true)

		models = append(models, model)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err = db.Collection("districts").BulkWrite(ctx, models, opts)

	return err
}
