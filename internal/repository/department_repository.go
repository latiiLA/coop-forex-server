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

		// Remove top-level subprocessObj
		bson.D{{Key: "$unset", Value: "subprocessObj"}},

		// Lookup the process with top-level as "process"
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "processes"},
				{Key: "let", Value: bson.D{
					{Key: "processId", Value: "$subprocess.process_id"},
				}},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$_id", "$$processId"}},
						}},
					}}},
				}},
				{Key: "as", Value: "processObj"},
			}},
		},

		// Embed process inside subprocess.process
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "subprocess.process", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$processObj", 0}},
			}},
		}}},

		// Remove top-level processObj
		bson.D{{Key: "$unset", Value: "processObj"}},

		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 1},
				{Key: "name", Value: 1},
				{Key: "subprocess", Value: 1},
				{Key: "department", Value: 1},
				{Key: "created_at", Value: 1},
				{Key: "updated_at", Value: 1},
				{Key: "created_by", Value: 1},
				{Key: "updated_by", Value: 1},
				{Key: "deleted_by", Value: 1},
				{Key: "deleted_at", Value: 1},
				{Key: "is_deleted", Value: 1},
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
