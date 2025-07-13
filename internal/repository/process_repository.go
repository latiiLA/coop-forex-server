package repository

import (
	"context"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type processRepository struct {
	collection *mongo.Collection
}

func NewProcessRepository(db *mongo.Database) model.ProcessRepository {
	return &processRepository{
		collection: db.Collection("processes"),
	}
}

func (pr *processRepository) Create(ctx context.Context, process *model.Process) error {
	_, err := pr.collection.InsertOne(ctx, process)
	return err
}

func (pr *processRepository) FindByID(ctx context.Context, process_id primitive.ObjectID) (*model.Process, error) {
	var process model.Process
	filter := bson.M{"_id": process_id, "is_deleted": false}

	err := pr.collection.FindOne(ctx, filter).Decode(&process)
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (pr *processRepository) FindAll(ctx context.Context) ([]model.Process, error) {
	var processes []model.Process

	cursor, err := pr.collection.Find(ctx, bson.M{"is_deleted": false})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &processes); err != nil {
		return nil, err
	}

	return processes, nil
}
