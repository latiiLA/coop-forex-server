package utils

import "go.mongodb.org/mongo-driver/bson"

func LookupAndUnwind(from, localField, as string) []bson.D {
	return []bson.D{
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: from},
				{Key: "localField", Value: localField},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: as},
			}},
		},
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$" + as},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},
	}

}
