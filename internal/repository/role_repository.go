package repository

import (
	"context"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type roleRepository struct {
	collection *mongo.Collection
}

func NewRoleRepository(db *mongo.Database) model.RoleRepository {
	return &roleRepository{
		collection: db.Collection("roles"),
	}
}

func (rr *roleRepository) Create(ctx context.Context, role *model.Role) error {
	_, err := rr.collection.InsertOne(ctx, role)
	return err
}

func (rr *roleRepository) FindByID(ctx context.Context, role_id primitive.ObjectID) (*model.Role, error) {
	var role model.Role
	filter := bson.M{"_id": role_id, "is_deleted": false}

	err := rr.collection.FindOne(ctx, filter).Decode(&role)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (rr *roleRepository) FindAll(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role

	cursor, err := rr.collection.Find(ctx, bson.M{"is_deleted": false})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

func (rr *roleRepository) ExistsRoleByName(ctx context.Context, name string) (bool, error) {
	filter := bson.M{"name": name, "is_deleted": false}

	var result model.Role

	err := rr.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		// some other error occurred
		return false, err
	}

	return true, nil
}
