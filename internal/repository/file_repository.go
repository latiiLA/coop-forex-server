package repository

import (
	"context"
	"fmt"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type fileRepository struct {
	collection *mongo.Collection
}

func NewFileRepository(db *mongo.Database) model.FileRepository {
	return &fileRepository{
		collection: db.Collection("files"),
	}
}

func (fr *fileRepository) Create(ctx context.Context, file *model.File) (*primitive.ObjectID, error) {
	result, err := fr.collection.InsertOne(ctx, file)
	if err != nil {
		return nil, err
	}
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("failed to convert inserted ID to ObjectID")
	}
	return &insertedID, nil
}

func (fr *fileRepository) FindByID(ctx context.Context, file_id primitive.ObjectID) (*model.File, error) {
	var file model.File
	filter := bson.M{"_id": file_id, "is_deleted": false}

	err := fr.collection.FindOne(ctx, filter).Decode(&file)
	if err != nil {
		return nil, err
	}

	return &file, nil
}
