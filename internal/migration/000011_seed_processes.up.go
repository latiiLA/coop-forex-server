package migration

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func seedProcesses(ctx context.Context, db *mongo.Database) error {

	userID, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345679000")
	if err != nil {
		return err
	}

	id1, _ := primitive.ObjectIDFromHex("676e44274d7cd2c7251d99e8")
	id2, _ := primitive.ObjectIDFromHex("676e48894d7cd2c7251d99f4")
	id3, _ := primitive.ObjectIDFromHex("676e48a34d7cd2c7251d9900")
	id4, _ := primitive.ObjectIDFromHex("676e48c24d7cd2c7251d990c")
	id5, _ := primitive.ObjectIDFromHex("676e48e34d7cd2c7251d991c")

	processes := []interface{}{
		bson.M{"_id": id1, "name": "Growth and Operations", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id2, "name": "Internet Free Banking", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id3, "name": "Agriculture and Cooperative Relations", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id4, "name": "Credit Approval and Portfolio", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
		bson.M{"_id": id5, "name": "Technology and Strategy", "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
	}

	var models []mongo.WriteModel
	for _, p := range processes {
		process := p.(bson.M)
		id := process["_id"]
		delete(process, "_id") // remove _id from update to avoid immutable error

		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": id}).
			SetUpdate(bson.M{"$set": process}).
			SetUpsert(true)

		models = append(models, model)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err = db.Collection("processes").BulkWrite(ctx, models, opts)

	return err
}
