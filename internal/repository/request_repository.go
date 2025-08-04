package repository

import (
	"context"
	"fmt"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type requestRepository struct {
	collection *mongo.Collection
}

func NewRequestRepository(db *mongo.Database) model.RequestRepository {
	return &requestRepository{
		collection: db.Collection("requests"),
	}
}

func (rr *requestRepository) Create(ctx context.Context, request *model.Request) error {
	_, err := rr.collection.InsertOne(ctx, request)
	return err
}

func (rr *requestRepository) FindAll(ctx context.Context) ([]model.Request, error) {

	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "is_deleted", Value: false},
				{Key: "request_status", Value: bson.D{
					{Key: "$nin", Value: bson.A{"New", "Rejected", "Drafted"}},
				}},
			}},
		},
	}

	pipeline = append(pipeline, utils.BuildCommonRequestPipelineStages()...)

	cursor, err := rr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []model.Request
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		return []model.Request{}, nil
	}

	return requests, nil
}

func (rr *requestRepository) Validate(ctx context.Context, request_id primitive.ObjectID, request *model.Request) error {
	// Prepare the update document using $set to only update the fields you pass
	update := bson.M{
		"$set": request,
	}

	// Perform the update
	filter := bson.M{"_id": request_id}
	result, err := rr.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// Check if any document was actually modified
	if result.MatchedCount == 0 {
		return fmt.Errorf("no request found with id %s", request_id.Hex())
	}

	return nil
}

func (rr *requestRepository) FindByID(ctx context.Context, request_id primitive.ObjectID) (*model.Request, error) {
	var request model.Request
	filter := bson.M{"_id": request_id, "is_deleted": false}

	err := rr.collection.FindOne(ctx, filter).Decode(&request)
	if err != nil {
		return nil, err
	}

	return &request, nil
}

func (rr *requestRepository) FindAllByOrgID(ctx context.Context, orgKey string, orgID primitive.ObjectID) ([]model.Request, error) {
	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "is_deleted", Value: false},
				{Key: orgKey, Value: orgID},
			}},
		},
	}

	pipeline = append(pipeline, utils.BuildCommonRequestPipelineStages()...)

	cursor, err := rr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []model.Request
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		return []model.Request{}, nil
	}

	return requests, nil
}

func (rr *requestRepository) FindOrgByRequestStatus(ctx context.Context, orgID primitive.ObjectID, orgKey, request_status string) ([]model.Request, error) {
	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "is_deleted", Value: false},
				{Key: "request_status", Value: request_status},
				{Key: orgKey, Value: orgID},
			}},
		},
	}

	pipeline = append(pipeline, utils.BuildCommonRequestPipelineStages()...)

	cursor, err := rr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []model.Request
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		return []model.Request{}, nil
	}

	return requests, nil
}

func (rr *requestRepository) Update(ctx context.Context, requestID primitive.ObjectID, request *model.Request) error {
	update := bson.M{
		"$set": request,
	}
	// Perform the update
	filter := bson.M{"_id": requestID}
	result, err := rr.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// Check if any document was actually modified
	if result.MatchedCount == 0 {
		return fmt.Errorf("no request found with id %s", requestID)
	}

	return nil
}

func (rr *requestRepository) FindByRequestStatus(ctx context.Context, request_status string) ([]model.Request, error) {
	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "is_deleted", Value: false},
				{Key: "request_status", Value: request_status},
			}},
		},
	}

	pipeline = append(pipeline, utils.BuildCommonRequestPipelineStages()...)

	cursor, err := rr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []model.Request
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, err
	}

	fmt.Println("requests", len(requests))

	if len(requests) == 0 {
		return []model.Request{}, nil
	}

	return requests, nil
}
