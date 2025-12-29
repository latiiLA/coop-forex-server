package repository

import (
	"context"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type tokenBlacklistRepository struct {
	collection *mongo.Collection
}

// NewTokenRevocationRepository creates a new repository
func NewTokenBlacklistRepository(db *mongo.Database) model.TokenBlacklistRepository {
	return &tokenBlacklistRepository{
		collection: db.Collection("token_blacklists"),
	}
}

// BlacklistToken inserts the token into MongoDB if it doesn't already exist
func (r *tokenBlacklistRepository) BlacklistToken(ctx context.Context, token string, expiresAt time.Time, userID primitive.ObjectID, ip string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"token": token}, // filter by token
		bson.M{
			"$setOnInsert": bson.M{
				"user_id":   userID,
				"ip":        ip,
				"token":     token,
				"expiresAt": expiresAt,
				"createdAt": time.Now(),
			},
		},
		options.Update().SetUpsert(true), // insert if not exists
	)
	return err
}

// IsBlacklisted checks if the token exists and has not expired
func (r *tokenBlacklistRepository) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"token":     token,
		"expiresAt": bson.M{"$gt": time.Now()}, // only consider non-expired tokens
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
