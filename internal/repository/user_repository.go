package repository

import (
	"context"
	"fmt"
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
	filter := bson.M{"username": username, "is_deleted": false}
	var user model.User

	err := ur.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *userRepository) FindByID(ctx context.Context, user_id primitive.ObjectID) (*model.User, error) {
	filter := bson.M{"_id": user_id, "is_deleted": false}
	var user model.User

	err := ur.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *userRepository) FindAll(ctx context.Context) (*[]model.UserResponseDTO, error) {

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "is_deleted", Value: false},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "profiles"},
			{Key: "localField", Value: "profile_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "profile"},
		}}},
		{{Key: "$unwind", Value: "$profile"}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "role_id", Value: 1},
			{Key: "profile_id", Value: 1},
			{Key: "username", Value: 1},
			{Key: "email", Value: "$profile.email"},
			{Key: "status", Value: 1},
			{Key: "first_name", Value: "$profile.first_name"},
			{Key: "middle_name", Value: "$profile.middle_name"},
			{Key: "last_name", Value: "$profile.last_name"},
			{Key: "department_id", Value: "$profile.department_id"},
			{Key: "branch_id", Value: "$profile.branch_id"},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
			{Key: "created_by", Value: 1},
			{Key: "updated_by", Value: 1},
			{Key: "deleted_by", Value: 1},
			{Key: "deleted_at", Value: 1},
			{Key: "is_deleted", Value: 1},
		}}},
	}

	cursor, err := ur.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []model.UserResponseDTO
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return &users, nil
}

func (ur *userRepository) Update(ctx context.Context, user_id primitive.ObjectID, user *model.User) (*model.User, error) {
	filter := bson.M{"_id": user_id, "is_deleted": false}

	result, err := ur.collection.UpdateOne(ctx, filter, bson.M{"$set": user})

	if err != nil {
		return nil, err
	}

	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("user not found")
	}

	if result.ModifiedCount == 0 {
		return nil, fmt.Errorf("no changes were made")
	}

	var updatedUser model.User
	err = ur.collection.FindOne(ctx, filter).Decode(&updatedUser)
	if err != nil {
		return nil, err
	}

	return &updatedUser, nil
}

func (ur *userRepository) Delete(ctx context.Context, user_id primitive.ObjectID, user *model.User) error {
	filter := bson.M{"_id": user_id, "is_deleted": false}
	update := bson.M{"is_deleted": true, "status": "deleted"}

	result, err := ur.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 || result.ModifiedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
