package utils

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Optimized lookup to fetch user with profile
func LookupUserWithProfile(localField, alias string) mongo.Pipeline {
	userAlias := alias
	profileAlias := fmt.Sprintf("%s_profile", alias)

	return mongo.Pipeline{
		// Lookup user
		{{
			Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: localField},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: userAlias},
			},
		}},
		// Unwind user array
		{{
			Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$" + userAlias},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			},
		}},
		// Lookup profile from users.profile_id
		{{
			Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "profiles"},
				{Key: "let", Value: bson.D{
					{Key: "profile_id", Value: fmt.Sprintf("$%s.profile_id", userAlias)},
				}},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$_id", "$$profile_id"}},
						}},
					}}},
					// Project only required fields
					bson.D{{Key: "$project", Value: bson.D{
						{Key: "first_name", Value: 1},
						{Key: "middle_name", Value: 1},
						{Key: "last_name", Value: 1}},
					}}, // optimization
				}},
				{Key: "as", Value: profileAlias},
			},
		}},
		// Merge profile into user object
		{{
			Key: "$set", Value: bson.D{
				{Key: fmt.Sprintf("%s.profile", userAlias), Value: bson.D{
					{Key: "$arrayElemAt", Value: bson.A{"$" + profileAlias, 0}},
				}},
			},
		}},
		// Optionally remove intermediate array
		{{
			Key: "$unset", Value: profileAlias,
		}},
	}
}
