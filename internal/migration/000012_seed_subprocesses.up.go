package migration

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func seedSubprocesses(ctx context.Context, db *mongo.Database) error {

	userID, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345679000")
	if err != nil {
		return err
	}

	processid1, _ := primitive.ObjectIDFromHex("676e44274d7cd2c7251d99e8")

	subprocessid1, _ := primitive.ObjectIDFromHex("676e48894d7cd2c7251e69f4")

	subprocesses := []interface{}{
		bson.M{"_id": subprocessid1, "name": "Trade and Service Banking", "process_id": processid1, "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
	}

	var models []mongo.WriteModel
	for _, p := range subprocesses {
		subprocess := p.(bson.M)
		id := subprocess["_id"]
		delete(subprocess, "_id") // remove _id from update to avoid immutable error

		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": id}).
			SetUpdate(bson.M{"$set": subprocess}).
			SetUpsert(true)

		models = append(models, model)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err = db.Collection("subprocesses").BulkWrite(ctx, models, opts)

	return err
}
