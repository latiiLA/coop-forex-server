package repository

import (
	"context"
	"errors"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type countryRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
}

func NewCountryRepository(db *mongo.Database, timeout time.Duration) model.CountryRepository {
	return &countryRepository{
		collection: db.Collection("countries"),
		timeout:    timeout,
	}
}

func (cr *countryRepository) Create(ctx context.Context, country *model.Country) error {
	_, err := cr.collection.InsertOne(ctx, country)
	return err
}

func (cr *countryRepository) FindByID(ctx context.Context, country_id primitive.ObjectID) (*model.Country, error) {
	filter := bson.M{"_id": country_id, "is_deleted": false}
	var country model.Country

	err := cr.collection.FindOne(ctx, filter).Decode(&country)
	if err != nil {
		return nil, err
	}
	return nil, errors.New("not implemented yet")
}

func (cr *countryRepository) FindAll(ctx context.Context) ([]model.Country, error) {
	var countries []model.Country

	cursor, err := cr.collection.Find(ctx, bson.M{"is_deleted": false})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &countries); err != nil {
		return nil, err
	}

	return countries, nil
}

func (cr *countryRepository) Update(ctx context.Context, country_id primitive.ObjectID, country *model.Country) (*model.Country, error) {
	return nil, errors.New("not implemented yet")
}

func (cr *countryRepository) Delete(ctx context.Context, country_id primitive.ObjectID) error {
	return nil
}
