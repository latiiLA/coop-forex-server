package repository

import (
	"context"
	"fmt"

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

	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "is_deleted", Value: false},
			}},
		},
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "let", Value: bson.D{{Key: "creater_id", Value: "$created_by"}}},
				{Key: "pipeline", Value: mongo.Pipeline{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$_id", "$$creater_id"}},
						}},
					}}},
					bson.D{{Key: "$project", Value: bson.D{
						{Key: "_id", Value: 1},
						{Key: "profile_id", Value: 1},
					}}},
				}},
				{Key: "as", Value: "creater"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$creater"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Lookup the profile with top-level as "creater"
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "profiles"},
				{Key: "let", Value: bson.D{
					{Key: "profile_id", Value: "$creater.profile_id"},
				}},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$_id", "$$profile_id"}},
						}},
					}}},
				}},
				{Key: "as", Value: "creater_profile"},
			}},
		},

		// Embed profile inside creatr.profile
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "creater.profile", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$creater_profile", 0}},
			}},
		}}},

		// Update user look up
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "let", Value: bson.D{{Key: "updater_id", Value: "$updated_by"}}},
				{Key: "pipeline", Value: mongo.Pipeline{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$_id", "$$updater_id"}},
						}},
					}}},
					bson.D{{Key: "$project", Value: bson.D{
						{Key: "_id", Value: 1},
						{Key: "profile_id", Value: 1},
					}}},
				}},
				{Key: "as", Value: "updater"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$updater"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Lookup the profile with top-level as "updater"
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "profiles"},
				{Key: "let", Value: bson.D{
					{Key: "profile_id", Value: "$updater.profile_id"},
				}},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$_id", "$$profile_id"}},
						}},
					}}},
				}},
				{Key: "as", Value: "updater_profile"},
			}},
		},

		// Embed profile inside updator.profile
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "updater.profile", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$updater_profile", 0}},
			}},
		}}},

		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 1},
				{Key: "name", Value: 1},
				{Key: "permissions", Value: 1},
				{Key: "creater", Value: 1},
				{Key: "created_at", Value: 1},
				{Key: "updated_at", Value: 1},
				{Key: "updater", Value: 1},
			}},
		},
	}

	cursor, err := rr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []model.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

func (rr *roleRepository) FindDeleted(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role

	cursor, err := rr.collection.Find(ctx, bson.M{"is_deleted": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

func (rr *roleRepository) FindRoleByName(ctx context.Context, name string) (*model.Role, error) {
	filter := bson.M{"name": name, "is_deleted": false}

	var result model.Role

	err := rr.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		// some other error occurred
		return nil, err
	}

	return &result, nil
}

func (rr *roleRepository) Update(ctx context.Context, role_id primitive.ObjectID, role *model.Role) error {
	filter := bson.M{"_id": role_id}

	update := bson.M{
		"$set": bson.M{
			"name":        role.Name,
			"permissions": role.Permissions,
			"updated_at":  role.UpdatedAt,
			"updated_by":  role.UpdatedBy,
		},
	}

	_, err := rr.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

func (rr *roleRepository) Delete(ctx context.Context, role_id primitive.ObjectID, role *model.Role) error {
	filter := bson.M{"_id": role_id}
	update := bson.M{
		"$set": bson.M{
			"is_deleted": role.IsDeleted,
			"deleted_at": role.DeletedAt,
			"deleted_by": role.DeletedBy,
		},
	}

	_, err := rr.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to soft-delete role: %w", err)
	}

	return nil
}
