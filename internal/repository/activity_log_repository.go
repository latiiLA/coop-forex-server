package repository

import (
	"context"
	"errors"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type activityLogRepository struct {
	collection *mongo.Collection
}

func NewActivityLogRepository(db *mongo.Database) model.ActivityLogRepository {
	repo := &activityLogRepository{
		collection: db.Collection("activity_logs"),
	}
	// Best-effort unique index for dedupe_key to prevent duplicate logs.
	_, _ = repo.collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "dedupe_key", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true).SetName("uniq_activity_dedupe_key"),
	})
	_, _ = repo.collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "created_at", Value: -1}},
		Options: options.Index().SetName("idx_activity_created_at"),
	})
	return repo
}

func (r *activityLogRepository) Create(ctx context.Context, log *model.ActivityLog) error {
	if log.ID.IsZero() {
		log.ID = primitive.NewObjectID()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	_, err := r.collection.InsertOne(ctx, log)
	if err != nil {
		// Duplicate dedupe_key → treat as success (already logged).
		var writeErr mongo.WriteException
		if errors.As(err, &writeErr) {
			for _, we := range writeErr.WriteErrors {
				if we.Code == 11000 {
					return nil
				}
			}
		}
		return err
	}
	return nil
}

func (r *activityLogRepository) FindRecent(ctx context.Context, limit int64) ([]model.ActivityLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	pipeline := mongo.Pipeline{
		{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
	}
	pipeline = append(pipeline, utils.LookupUserWithProfile("actor_id", "actor")...)

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []model.ActivityLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	if logs == nil {
		logs = []model.ActivityLog{}
	}
	return logs, nil
}
