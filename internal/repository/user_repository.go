package repository

import (
	"context"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type userRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
}

func NewUserRepository(db *mongo.Database, timeout time.Duration) model.UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
		timeout:    timeout,
	}
}

func (ur *userRepository) Create(ctx context.Context, user *model.User) error {
	_, err := ur.collection.InsertOne(ctx, user)
	return err
}

func (ur *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	filter := bson.M{"username": username}
	var user model.User

	err := ur.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *userRepository) FindByID(ctx context.Context, user_id primitive.ObjectID) (*model.UserResponseDTO, error) {
	filter := bson.M{"_id": user_id}
	var user model.UserResponseDTO

	err := ur.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *userRepository) FindAll(ctx context.Context) (*[]model.UserResponseDTO, error) {
	var users []model.UserResponseDTO

	cursor, err := ur.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return &users, err
}
