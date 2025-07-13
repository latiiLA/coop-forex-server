package repository

import (
	"context"
	"fmt"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
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

		// ___________________________________________________
		// User - Authorizer
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "authorized_by"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "authorizer"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$authorizer"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Lookup the profile with top-level as "authorizer"
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "profiles"},
				{Key: "let", Value: bson.D{
					{Key: "profile_id", Value: "$authorizer.profile_id"},
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

		// Embed profile inside authorizer.profile
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "authorizer.profile", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$profileObj", 0}},
			}},
		}}},

		// Remove top-level profileObj
		bson.D{{Key: "$unset", Value: "profileObj"}},

		// User - Validator
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "validated_by"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "validator"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$validator"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Lookup the profile with top-level as "validator"
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "profiles"},
				{Key: "let", Value: bson.D{
					{Key: "profile_id", Value: "$validator.profile_id"},
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

		// Embed profile inside validator.profile
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "validator.profile", Value: bson.D{
				{Key: "$arrayElemAt", Value: bson.A{"$profileObj", 0}},
			}},
		}}},

		// Remove top-level profileObj
		bson.D{{Key: "$unset", Value: "profileObj"}},

		// Department
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "departments"},
				{Key: "localField", Value: "department_id"},
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

		// Branch
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "branches"},
				{Key: "localField", Value: "branch_id"},
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

		// Travel Country
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "countries"},
				{Key: "localField", Value: "travel_country_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "travel_country"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$travel_country"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Travel Purpose
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "travel_purposes"},
				{Key: "localField", Value: "travel_purpose_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "travel_purpose"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$travel_purpose"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Account Currency
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "currencies"},
				{Key: "localField", Value: "account_currency_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "account_currency"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$account_currency"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Validated Account Currency
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "currencies"},
				{Key: "localField", Value: "validated_account_currency_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "validated_account_currency"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$validated_account_currency"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Travel Purpose
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "currencies"},
				{Key: "localField", Value: "fcy_requested_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "fcy_requested"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$fcy_requested"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Passport Attachment
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "files"},
				{Key: "localField", Value: "passport_attachment"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "passport"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$passport"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Visa Attachment
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "files"},
				{Key: "localField", Value: "visa_attachment"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "visa"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$visa"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// ticket Attachment
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "files"},
				{Key: "localField", Value: "ticket_attachment"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "ticket"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$ticket"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Business License Attachment
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "files"},
				{Key: "localField", Value: "business_license_attachment"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "business_license"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$business_license"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Business Supporting Attachment
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "files"},
				{Key: "localField", Value: "business_supporting_attachment"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "business_supporting"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$business_supporting"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		// Business Supporting Attachment
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "files"},
				{Key: "localField", Value: "education_loa_attachment"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "education_loa"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$education_loa"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},

		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 1},
				{Key: "request_code", Value: 1},
				{Key: "applicant_name", Value: 1},
				{Key: "applicant_account_number", Value: 1},
				{Key: "requesting_as", Value: 1},
				{Key: "account_currency", Value: 1},
				{Key: "fcy_requested_amount", Value: 1},
				{Key: "total_fcy_requested", Value: 1},
				{Key: "current_fcy_performance", Value: 1},
				{Key: "accounts_to_deduct", Value: 1},
				{Key: "fcy_acceptance_mode", Value: 1},
				{Key: "request_status", Value: 1},

				{Key: "department", Value: 1},
				{Key: "branch", Value: 1},
				{Key: "travel_country", Value: 1},
				{Key: "travel_purpose", Value: 1},
				{Key: "fcy_requested", Value: 1},

				{Key: "validated_average_deposit", Value: 1},
				{Key: "validated_current_balance", Value: 1},
				{Key: "validated_account_currency", Value: 1},

				{Key: "passport", Value: 1},
				{Key: "ticket", Value: 1},
				{Key: "education_loa", Value: 1},
				{Key: "business_letter", Value: 1},
				{Key: "business_supporting", Value: 1},
				{Key: "health_letter", Value: 1},

				{Key: "created_at", Value: 1},
				{Key: "updated_at", Value: 1},
				{Key: "created_by", Value: 1},
				{Key: "creator", Value: 1},
				{Key: "updated_by", Value: 1},
				{Key: "authorized_by", Value: 1},
				{Key: "authorizer", Value: 1},
				{Key: "validator", Value: 1},
				{Key: "validated_at", Value: 1},
				{Key: "deleted_by", Value: 1},
				{Key: "deleted_at", Value: 1},
				{Key: "is_deleted", Value: 1},
				{Key: "debug_subprocess", Value: 1},
				{Key: "debug_process_id", Value: 1},
				{Key: "debug_processObj", Value: 1},
			}},
		},
	}

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
		return nil, mongo.ErrNoDocuments
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
