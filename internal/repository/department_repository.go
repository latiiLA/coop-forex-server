package repository

import (
	"context"
	"errors"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type departmentRepository struct {
	collection *mongo.Collection
}

func NewDepartmentRepository(db *mongo.Database) model.DepartmentRepository {
	return &departmentRepository{
		collection: db.Collection("departments"),
	}
}

func (dr *departmentRepository) Create(ctx context.Context, subprocess *model.Department) error {
	_, err := dr.collection.InsertOne(ctx, subprocess)
	return err
}

func (dr *departmentRepository) FindByID(ctx context.Context, department_id primitive.ObjectID) (*model.Department, error) {
	var department model.Department
	filter := bson.M{"_id": department_id, "is_deleted": false}

	err := dr.collection.FindOne(ctx, filter).Decode(&department)
	if err != nil {
		return nil, err
	}

	return &department, nil
}

func (dr *departmentRepository) FindAll(ctx context.Context) ([]model.Department, error) {
	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "is_deleted", Value: false},
			}},
		},
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "subprocesses"},
				{Key: "localField", Value: "subprocess_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "subprocess"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$subprocess"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Lookup the process with top-level as "process"
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "processes"},
				{Key: "localField", Value: "subprocess.process_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "subprocess.process"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$subprocess.process"},
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
				{Key: "subprocess_id", Value: 1},
				{Key: "subprocess", Value: 1},
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

	cursor, err := dr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var departments []model.Department
	if err := cursor.All(ctx, &departments); err != nil {
		return nil, err
	}

	// prettyJSON, _ := json.MarshalIndent(users, "", "  ")
	// fmt.Println("users:", string(prettyJSON))

	return departments, nil
}

func (dr *departmentRepository) FindBySubprocessID(ctx context.Context, subprocess_id primitive.ObjectID) (*[]model.Department, error) {
	var departments []model.Department
	filter := bson.M{"subprocess_id": subprocess_id, "is_deleted": false}

	cursor, err := dr.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &departments); err != nil {
		return nil, err
	}

	return &departments, nil
}

func (dr *departmentRepository) Update(ctx context.Context, department_id primitive.ObjectID, department *model.Department) (*model.Department, error) {
	return nil, errors.New("not yet implemented")
}

func (dr *departmentRepository) Delete(ctx context.Context, department_id primitive.ObjectID) error {
	return errors.New("not yet implemented")
}
