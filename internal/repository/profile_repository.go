package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type profileRepository struct {
	collection *mongo.Collection
}

func NewProfileRepository(db *mongo.Database) model.ProfileRepository {
	return &profileRepository{
		collection: db.Collection("profiles"),
	}
}

func (pr *profileRepository) Create(ctx context.Context, profile *model.Profile) error {
	_, err := pr.collection.InsertOne(ctx, profile)
	return err
}

func (pr *profileRepository) FindByID(ctx context.Context, profile_id primitive.ObjectID) (*model.Profile, error) {
	filter := bson.M{"_id": profile_id, "is_deleted": false}
	var profile model.Profile

	err := pr.collection.FindOne(ctx, filter).Decode(&profile)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("profile not found")
		}
		return nil, err
	}

	return &profile, nil
}

func (pr *profileRepository) FindByEmail(ctx context.Context, email string) (*model.Profile, error) {
	filter := bson.M{"email": email, "is_deleted": false}
	var profile model.Profile

	err := pr.collection.FindOne(ctx, filter).Decode(&profile)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (pr *profileRepository) Update(ctx context.Context, profile_id primitive.ObjectID, profile *model.Profile) (*model.Profile, error) {
	filter := bson.M{"_id": profile_id, "is_deleted": false}

	result, err := pr.collection.UpdateOne(ctx, filter, bson.M{"$set": profile})

	if err != nil {
		return nil, err
	}

	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("profile not found")
	}

	if result.ModifiedCount == 0 {
		return nil, fmt.Errorf("no changes were made")
	}

	return pr.FindByID(ctx, profile_id)
}
