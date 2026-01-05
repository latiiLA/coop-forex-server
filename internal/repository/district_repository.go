package repository

import (
	"context"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type districtRepository struct {
	collection *mongo.Collection
}

func NewDistrictRepository(db *mongo.Database) model.DistrictRepository {
	return &districtRepository{
		collection: db.Collection("districts"),
	}
}

func (dr *districtRepository) FindAll(ctx context.Context) ([]model.District, error) {
	var districts []model.District

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "is_deleted", Value: false},
		}}},

		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "name", Value: 1},
		}}},
	}

	pipeline = append(pipeline, utils.LookupUserWithProfile("created_by", "creator")...)

	cursor, err := dr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &districts); err != nil {
		return nil, err
	}

	return districts, nil
}
