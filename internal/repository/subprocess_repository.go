package repository

import (
	"context"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type subprocessRepository struct {
	collection *mongo.Collection
}

func NewSubprocessRepository(db *mongo.Database) model.SubprocessRepository {
	return &subprocessRepository{
		collection: db.Collection("subprocesses"),
	}
}

func (sr *subprocessRepository) Create(ctx context.Context, subprocess *model.Subprocess) error {
	_, err := sr.collection.InsertOne(ctx, subprocess)
	return err
}

func (sr *subprocessRepository) FindByID(ctx context.Context, subprocess_id primitive.ObjectID) (*model.Subprocess, error) {
	var subprocess model.Subprocess
	filter := bson.M{"_id": subprocess_id, "is_deleted": false}

	err := sr.collection.FindOne(ctx, filter).Decode(&subprocess)
	if err != nil {
		return nil, err
	}

	return &subprocess, nil
}

func (sr *subprocessRepository) FindAll(ctx context.Context) ([]model.Subprocess, error) {
	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "is_deleted", Value: false},
			}},
		},
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "processes"},
				{Key: "localField", Value: "process_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "process"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$process"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "created_by"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "creator"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$creator"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "profiles"},
				{Key: "localField", Value: "creator.profile_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "creator.profile"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$creator.profile"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "updated_by"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "updater"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$updater"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "profiles"},
				{Key: "localField", Value: "updater.profile_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "updater.profile"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$updater.profile"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "name", Value: 1},
		}}},

		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 1},
				{Key: "name", Value: 1},
				{Key: "process", Value: 1},
				{Key: "created_at", Value: 1},
				{Key: "updated_at", Value: 1},
				{Key: "created_by", Value: 1},
				{Key: "updated_by", Value: 1},
				{Key: "is_deleted", Value: 1},

				{Key: "creator", Value: bson.D{{Key: "$cond", Value: bson.A{
					bson.D{{Key: "$ifNull", Value: bson.A{"$creator._id", false}}},
					"$creator",
					"$$REMOVE",
				}}}},
				{Key: "updater", Value: bson.D{{Key: "$cond", Value: bson.A{
					bson.D{{Key: "$ifNull", Value: bson.A{"$updater._id", false}}},
					"$updater",
					"$$REMOVE",
				}}}},
			}},
		},
	}

	cursor, err := sr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subprocesses []model.Subprocess
	if err := cursor.All(ctx, &subprocesses); err != nil {
		return nil, err
	}

	// prettyJSON, _ := json.MarshalIndent(users, "", "  ")
	// fmt.Println("users:", string(prettyJSON))

	return subprocesses, nil
}

func (sr *subprocessRepository) FindByProcessID(ctx context.Context, process_id primitive.ObjectID) (*[]model.Subprocess, error) {
	var subprocesses []model.Subprocess
	filter := bson.M{"process_id": process_id, "is_deleted": false}

	cursor, err := sr.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &subprocesses); err != nil {
		return nil, err
	}

	return &subprocesses, nil
}
