package repository

import (
	"context"
	"errors"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type travelPurposeRepository struct {
	collection *mongo.Collection
}

func NewTravelPurposeRepository(db *mongo.Database) model.TravelPurposeRepository {
	return &travelPurposeRepository{
		collection: db.Collection("travel_purposes"),
	}
}

func (tpr travelPurposeRepository) Create(ctx context.Context, travel_purpose *model.TravelPurpose) error {
	return errors.New("not implemented yet")
}

func (tpr travelPurposeRepository) FindByID(ctx context.Context, travel_purpose_id primitive.ObjectID) (*model.TravelPurpose, error) {
	return nil, errors.New("not implemented yet")
}

func (tpr travelPurposeRepository) FindAll(ctx context.Context) ([]model.TravelPurpose, error) {
	var travel_purpose []model.TravelPurpose

	cursor, err := tpr.collection.Find(ctx, bson.M{"is_deleted": false}, options.Find().SetSort(bson.D{{Key: "purpose", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &travel_purpose); err != nil {
		return nil, err
	}

	return travel_purpose, nil
}

func (tpr travelPurposeRepository) Update(ctx context.Context, travel_purpose_id primitive.ObjectID, travel_purpose *model.TravelPurpose) (*model.TravelPurpose, error) {
	return nil, errors.New("not implemented yet")

}

func (tpr travelPurposeRepository) Delete(ctx context.Context, travel_purpose_id primitive.ObjectID) error {
	return errors.New("not implemented yet")
}
