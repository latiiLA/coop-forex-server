package repository

import (
	"context"
	"fmt"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) model.UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
	}
}

func (ur *userRepository) Create(ctx context.Context, user *model.User) error {
	_, err := ur.collection.InsertOne(ctx, user)
	return err
}

func (ur *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "username", Value: username},
			{Key: "is_deleted", Value: false},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "roles"},
			{Key: "localField", Value: "role_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "role"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$role"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "profiles"},
			{Key: "localField", Value: "profile_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "profile"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$profile"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},

		// User - Creator
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "created_by"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "creator"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$creator"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Lookup the profile with top-level as "creator"
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "profiles"},
				{Key: "let", Value: bson.D{
					{Key: "profile_id", Value: "$creator.profile_id"},
				}},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$_id", "$$profile_id"}},
						}},
					}}},
				}},
				{Key: "as", Value: "profileObj"},
			}},
		},

		// Embed profile inside creator.profile
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "creator.profile", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$profileObj", 0}},
			}},
		}}},

		// Remove top-level profileObj
		bson.D{{Key: "$unset", Value: "profileObj"}},

		// User - Updater
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "created_by"},
				{Key: "foreignField", Value: "_id"},
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
				{Key: "as", Value: "profileObj"},
			}},
		},

		// Embed profile inside updater.profile
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "updater.profile", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$profileObj", 0}},
			}},
		}}},

		// Remove top-level profileObj
		bson.D{{Key: "$unset", Value: "profileObj"}},

		bson.D{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "username", Value: 1},
			{Key: "password", Value: 1},
			{Key: "permissions", Value: 1},
			{Key: "creator", Value: 1},
			{Key: "updater", Value: 1},
			{Key: "role", Value: 1},
			{Key: "profile", Value: 1},
			{Key: "status", Value: 1},
		}}},
	}

	cursor, err := ur.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregation error: %w", err)
	}
	defer cursor.Close(ctx)

	var user model.User
	if !cursor.Next(ctx) {
		return nil, mongo.ErrNoDocuments
	}

	if err := cursor.Decode(&user); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return &user, nil
}

func (ur *userRepository) FindByID(ctx context.Context, user_id primitive.ObjectID) (*model.User, error) {
	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "_id", Value: user_id},
				{Key: "is_deleted", Value: false},
			}},
		},
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "profiles"},
				{Key: "localField", Value: "profile_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "profile"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$profile"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},
		// Role
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "roles"},
				{Key: "localField", Value: "role_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "role"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$role"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},
		// Department
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "departments"},
				{Key: "localField", Value: "profile.department_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "department"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$department"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Lookup the subprocess with top-level as "subprocess"
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "subprocesses"},
				{Key: "let", Value: bson.D{
					{Key: "subprocessId", Value: "$department.subprocess_id"},
				}},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$_id", "$$subprocessId"}},
						}},
					}}},
				}},
				{Key: "as", Value: "subprocessObj"},
			}},
		},

		// Embed subprocess inside department.subprocess
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "department.subprocess", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$subprocessObj", 0}},
			}},
		}}},

		// Remove top-level subprocessObj
		bson.D{{Key: "$unset", Value: "subprocessObj"}},

		// Lookup the process with top-level as "process"
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "processes"},
				{Key: "let", Value: bson.D{
					{Key: "processId", Value: "$department.subprocess.process_id"},
				}},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$_id", "$$processId"}},
						}},
					}}},
				}},
				{Key: "as", Value: "processObj"},
			}},
		},

		// Embed process inside subprocess.process
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "department.subprocess.process", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$processObj", 0}},
			}},
		}}},

		// Remove top-level processObj
		bson.D{{Key: "$unset", Value: "processObj"}},

		// Branch
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "branches"},
				{Key: "localField", Value: "profile.branch_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "branch"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$branch"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// User - Creator
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "created_by"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "creator"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$creator"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Lookup the profile with top-level as "creator"
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "profiles"},
				{Key: "let", Value: bson.D{
					{Key: "profile_id", Value: "$creator.profile_id"},
				}},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$_id", "$$profile_id"}},
						}},
					}}},
				}},
				{Key: "as", Value: "profileObj"},
			}},
		},

		// Embed profile inside creator.profile
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "creator.profile", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$profileObj", 0}},
			}},
		}}},

		// Remove top-level profileObj
		bson.D{{Key: "$unset", Value: "profileObj"}},

		// User - Updater
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "created_by"},
				{Key: "foreignField", Value: "_id"},
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
				{Key: "as", Value: "profileObj"},
			}},
		},

		// Embed profile inside updater.profile
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "updater.profile", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$profileObj", 0}},
			}},
		}}},

		// Remove top-level profileObj
		bson.D{{Key: "$unset", Value: "profileObj"}},

		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 1},
				{Key: "role", Value: 1},
				{Key: "profile", Value: 1},
				{Key: "username", Value: 1},
				{Key: "status", Value: 1},
				{Key: "permissions", Value: 1},
				{Key: "department", Value: 1},
				{Key: "branch", Value: 1},
				{Key: "created_at", Value: 1},
				{Key: "updated_at", Value: 1},
				{Key: "created_by", Value: 1},
				{Key: "creator", Value: 1},
				{Key: "updater", Value: 1},
				{Key: "updated_by", Value: 1},
				{Key: "deleted_by", Value: 1},
				{Key: "deleted_at", Value: 1},
				{Key: "is_deleted", Value: 1},
				{Key: "debug_subprocess", Value: 1},
				{Key: "debug_process_id", Value: 1},
				{Key: "debug_processObj", Value: 1},
			}},
		},
	}

	cursor, err := ur.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	return &users[0], nil
}

func (ur *userRepository) FindAll(ctx context.Context) (*[]model.UserResponseDTO, error) {

	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "is_deleted", Value: false},
			}},
		},
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "profiles"},
				{Key: "localField", Value: "profile_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "profile"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$profile"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},
		// Role
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "roles"},
				{Key: "localField", Value: "role_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "role"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$role"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},
		// Department
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "departments"},
				{Key: "localField", Value: "profile.department_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "department"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$department"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Lookup the subprocess with top-level as "subprocess"
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "subprocesses"},
				{Key: "let", Value: bson.D{
					{Key: "subprocessId", Value: "$department.subprocess_id"},
				}},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$_id", "$$subprocessId"}},
						}},
					}}},
				}},
				{Key: "as", Value: "subprocessObj"},
			}},
		},

		// Embed subprocess inside department.subprocess
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "department.subprocess", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$subprocessObj", 0}},
			}},
		}}},

		// Remove top-level subprocessObj
		bson.D{{Key: "$unset", Value: "subprocessObj"}},

		// Lookup the process with top-level as "process"
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "processes"},
				{Key: "let", Value: bson.D{
					{Key: "processId", Value: "$department.subprocess.process_id"},
				}},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$_id", "$$processId"}},
						}},
					}}},
				}},
				{Key: "as", Value: "processObj"},
			}},
		},

		// Embed process inside subprocess.process
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "department.subprocess.process", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$processObj", 0}},
			}},
		}}},

		// Remove top-level processObj
		bson.D{{Key: "$unset", Value: "processObj"}},

		// Branch
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "branches"},
				{Key: "localField", Value: "profile.branch_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "branch"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$branch"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Lookup the district with top-level as "branch"
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "districts"},
				{Key: "let", Value: bson.D{
					{Key: "districtId", Value: "$branch.district_id"},
				}},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$_id", "$$districtId"}},
						}},
					}}},
				}},
				{Key: "as", Value: "districtObj"},
			}},
		},

		// Embed district inside branch.district
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "branch.district", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$districtObj", 0}},
			}},
		}}},

		// Remove top-level districtObj
		bson.D{{Key: "$unset", Value: "districtObj"}},

		// User - Creator
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "created_by"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "creator"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$creator"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Lookup the profile with top-level as "creator"
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "profiles"},
				{Key: "let", Value: bson.D{
					{Key: "profile_id", Value: "$creator.profile_id"},
				}},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$_id", "$$profile_id"}},
						}},
					}}},
				}},
				{Key: "as", Value: "profileObj"},
			}},
		},

		// Embed profile inside creator.profile
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "creator.profile", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$profileObj", 0}},
			}},
		}}},

		// Remove top-level profileObj
		bson.D{{Key: "$unset", Value: "profileObj"}},

		// User - Updater
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "updated_by"},
				{Key: "foreignField", Value: "_id"},
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
				{Key: "as", Value: "profileObj"},
			}},
		},

		// Embed profile inside updater.profile
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "updater.profile", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$profileObj", 0}},
			}},
		}}},

		// Remove top-level profileObj
		bson.D{{Key: "$unset", Value: "profileObj"}},

		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 1},
				{Key: "role", Value: 1},
				{Key: "permissions", Value: 1},
				{Key: "profile", Value: 1},
				{Key: "username", Value: 1},
				{Key: "status", Value: 1},
				{Key: "department", Value: 1},
				{Key: "branch", Value: 1},
				{Key: "created_at", Value: 1},
				{Key: "creator", Value: 1},
				{Key: "updater", Value: 1},
				{Key: "updated_at", Value: 1},
				{Key: "created_by", Value: 1},
				{Key: "updated_by", Value: 1},
				{Key: "deleted_by", Value: 1},
				{Key: "deleted_at", Value: 1},
				{Key: "is_deleted", Value: 1},
				{Key: "debug_subprocess", Value: 1},
				{Key: "debug_process_id", Value: 1},
				{Key: "debug_processObj", Value: 1},
			}},
		},
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

	// prettyJSON, _ := json.MarshalIndent(users, "", "  ")
	// fmt.Println("users:", string(prettyJSON))

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
