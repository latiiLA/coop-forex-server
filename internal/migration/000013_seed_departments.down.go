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

func unseedDepartments(ctx context.Context, db *mongo.Database) error {
	userID, err := primitive.ObjectIDFromHex("66a7b8c9e4b0f12345679000")
	if err != nil {
		return err
	}

	subprocessid1, _ := primitive.ObjectIDFromHex("676e48894d7cd2c7251e69f4")

	departmentid1, _ := primitive.ObjectIDFromHex("676e44274d7cd2c7251d66d2")

	departments := []interface{}{
		bson.M{"_id": departmentid1, "name": "Trade and Service Banking", "subprocess_id": subprocessid1, "created_at": time.Now(), "updated_at": time.Now(), "created_by": userID, "deleted_by": nil, "deleted_at": nil, "is_deleted": false},
	}

	var deleteModels []mongo.WriteModel
	for _, d := range departments {
		department := d.(bson.M)

		model := mongo.NewDeleteOneModel().
			SetFilter(bson.M{"id": department["id"]})

		deleteModels = append(deleteModels, model)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err = db.Collection("departments").BulkWrite(ctx, deleteModels, opts)
	if err != nil {
		log.Fatal("Error during unseeding departments:", err)
	}

	return err
}
