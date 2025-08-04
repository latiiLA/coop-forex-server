package repository

import (
	"context"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
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

	cursor, err := dr.collection.Find(ctx, bson.M{"is_deleted": false})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &districts); err != nil {
		return nil, err
	}

	return districts, nil
}
