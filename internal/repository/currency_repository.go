package repository

import (
	"context"
	"errors"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type currencyRepository struct {
	collection *mongo.Collection
}

func NewCurrencyRepository(db *mongo.Database) model.CurrencyRepository {
	return &currencyRepository{
		collection: db.Collection("currencies"),
	}
}

func (cr *currencyRepository) Create(ctx context.Context, currency *model.Currency) error {
	return errors.New("not implemented yet")
}

func (cr *currencyRepository) FindByID(ctx context.Context, currency_id primitive.ObjectID) (*model.Currency, error) {
	return nil, errors.New("not implemented yet")
}

func (cr *currencyRepository) FindAll(ctx context.Context) ([]model.Currency, error) {
	var currencies []model.Currency

	cursor, err := cr.collection.Find(ctx, bson.M{"is_deleted": false})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &currencies); err != nil {
		return nil, err
	}

	return currencies, nil
}

func (cr *currencyRepository) Update(ctx context.Context, currency_id primitive.ObjectID, country *model.Currency) (*model.Currency, error) {
	return nil, errors.New("not implemented yet")
}

func (cr *currencyRepository) Delete(ctx context.Context, currency_id primitive.ObjectID) error {
	return errors.New("not implemented")
}
