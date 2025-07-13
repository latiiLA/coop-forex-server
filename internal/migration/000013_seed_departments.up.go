package migration

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func seedDepartments(ctx context.Context, db *mongo.Database) error {

	userID, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345679000")
	if err != nil {
		return err
	}

	subprocessid1, _ := primitive.ObjectIDFromHex("676e48894d7cd2c7251e69f4")

	departmentid1, _ := primitive.ObjectIDFromHex("676e44274d7cd2c7251d66d2")

	departments := []interface{}{
		bson.M{"_id": departmentid1, "name": "Forex Office", "subprocess_id": subprocessid1, "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
	}

	var models []mongo.WriteModel
	for _, d := range departments {
		department := d.(bson.M)
		id := department["_id"]
		delete(department, "_id") // remove _id from update to avoid immutable error

		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": id}).
			SetUpdate(bson.M{"$set": department}).
			SetUpsert(true)

		models = append(models, model)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err = db.Collection("departments").BulkWrite(ctx, models, opts)

	return err
}
