package repository

import (
	"context"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type branchRepository struct {
	collection *mongo.Collection
}

func NewBranchRepository(db *mongo.Database) model.BranchRepository {
	return &branchRepository{
		collection: db.Collection("branches"),
	}
}

func (br *branchRepository) Create(ctx context.Context, branch *model.Branch) error {
	_, err := br.collection.InsertOne(ctx, branch)
	return err
}

func (br *branchRepository) FindByID(ctx context.Context, branchID primitive.ObjectID) (*model.Branch, error) {
	var branches model.Branch
	filter := bson.M{"_id": branchID, "is_deleted": false}

	err := br.collection.FindOne(ctx, filter).Decode(&branches)
	if err != nil {
		return nil, err
	}

	return &branches, nil
}

func (br *branchRepository) FindAll(ctx context.Context) ([]model.Branch, error) {
	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "is_deleted", Value: false},
			}},
		},
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "districts"},
				{Key: "localField", Value: "district_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "district"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$district"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 1},
				{Key: "name", Value: 1},
				{Key: "address", Value: 1},
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

	cursor, err := br.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var branches []model.Branch
	if err := cursor.All(ctx, &branches); err != nil {
		return nil, err
	}

	return branches, nil
}

func (br *branchRepository) FindByDistrictID(ctx context.Context, districtID primitive.ObjectID) (*[]model.Branch, error) {
	var districts []model.Branch
	filter := bson.M{"district_id": districtID, "is_deleted": false}

	cursor, err := br.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &districts); err != nil {
		return nil, err
	}

	return &districts, nil
}
