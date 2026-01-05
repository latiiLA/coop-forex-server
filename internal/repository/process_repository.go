package repository

import (
	"context"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type processRepository struct {
	collection *mongo.Collection
}

func NewProcessRepository(db *mongo.Database) model.ProcessRepository {
	return &processRepository{
		collection: db.Collection("processes"),
	}
}

func (pr *processRepository) Create(ctx context.Context, process *model.Process) error {
	_, err := pr.collection.InsertOne(ctx, process)
	return err
}

func (pr *processRepository) FindByID(ctx context.Context, process_id primitive.ObjectID) (*model.Process, error) {
	var process model.Process
	filter := bson.M{"_id": process_id, "is_deleted": false}

	err := pr.collection.FindOne(ctx, filter).Decode(&process)
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (pr *processRepository) FindAll(ctx context.Context) ([]model.Process, error) {
	var processes []model.Process

	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "is_deleted", Value: false},
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

	cursor, err := pr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &processes); err != nil {
		return nil, err
	}

	return processes, nil
}
