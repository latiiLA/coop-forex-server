package utils

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func BuildCommonRequestPipelineStages() mongo.Pipeline {
	var pipeline mongo.Pipeline

	pipeline = append(pipeline, bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "currencies"},
			{Key: "let", Value: bson.D{
				// Fallback to empty array to avoid null issue
				{Key: "currency_ids", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$approved_currency_ids", bson.A{}}}}},
			}},
			{Key: "pipeline", Value: mongo.Pipeline{
				bson.D{
					{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$in", Value: bson.A{"$_id", "$$currency_ids"}},
						}},
					}},
				},
			}},
			{Key: "as", Value: "approved_currencies"},
		}},
	})

	pipeline = append(pipeline, LookupUserWithProfile("created_by", "creator")...)
	pipeline = append(pipeline, LookupUserWithProfile("requested_by", "requester")...)
	pipeline = append(pipeline, LookupUserWithProfile("authorized_by", "authorizer")...)
	pipeline = append(pipeline, LookupUserWithProfile("validated_by", "validater")...)
	pipeline = append(pipeline, LookupUserWithProfile("approved_by", "approver")...)
	pipeline = append(pipeline, LookupUserWithProfile("accepted_by", "accepter")...)
	pipeline = append(pipeline, LookupUserWithProfile("declined_by", "decliner")...)

	pipeline = append(pipeline, LookupAndUnwind("departments", "department_id", "department")...)
	pipeline = append(pipeline, LookupAndUnwind("branches", "branch_id", "branch")...)
	pipeline = append(pipeline, LookupAndUnwind("countries", "travel_country_id", "travel_country")...)
	pipeline = append(pipeline, LookupAndUnwind("travel_purposes", "travel_purpose_id", "travel_purpose")...)
	pipeline = append(pipeline, LookupAndUnwind("currencies", "account_currency_id", "account_currency")...)
	pipeline = append(pipeline, LookupAndUnwind("currencies", "fcy_requested_id", "fcy_requested")...)
	pipeline = append(pipeline, LookupAndUnwind("currencies", "validated_account_currency_id", "validated_account_currency")...)

	pipeline = append(pipeline, LookupAndUnwind("files", "passport_attachment", "passport")...)
	pipeline = append(pipeline, LookupAndUnwind("files", "ticket_attachment", "ticket")...)
	pipeline = append(pipeline, LookupAndUnwind("files", "visa_attachment", "visa")...)
	pipeline = append(pipeline, LookupAndUnwind("files", "education_loa_attachment", "education_loa")...)
	pipeline = append(pipeline, LookupAndUnwind("files", "business_letter_attachment", "business_letter")...)
	pipeline = append(pipeline, LookupAndUnwind("files", "business_supporting_attachment", "business_supporting")...)
	pipeline = append(pipeline, LookupAndUnwind("files", "health_letter_attachment", "health_letter")...)

	pipeline = append(pipeline, bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "request_code", Value: 1},
			{Key: "applicant_name", Value: 1},
			{Key: "applicant_account_number", Value: 1},
			{Key: "requesting_as", Value: 1},
			{Key: "account_currency_id", Value: 1},
			{Key: "account_currency", Value: 1},
			{Key: "fcy_requested_amount", Value: 1},
			{Key: "total_fcy_requested", Value: 1},
			{Key: "current_fcy_performance", Value: 1},
			{Key: "accounts_to_deduct", Value: 1},
			{Key: "fcy_acceptance_mode", Value: 1},
			{Key: "request_status", Value: 1},
			{Key: "department_id", Value: 1},
			{Key: "department", Value: 1},
			{Key: "branch_id", Value: 1},
			{Key: "branch", Value: 1},
			{Key: "travel_country_id", Value: 1},
			{Key: "travel_country", Value: 1},
			{Key: "travel_purpose_id", Value: 1},
			{Key: "travel_purpose", Value: 1},
			{Key: "fcy_requested_id", Value: 1},
			{Key: "fcy_requested", Value: 1},
			{Key: "branch_recommendation", Value: 1},
			{Key: "rejection_reason", Value: 1},

			{Key: "validated_average_deposit", Value: 1},
			{Key: "validated_current_balance", Value: 1},
			{Key: "validated_account_currency_id", Value: 1},
			{Key: "validated_account_currency", Value: 1},

			{Key: "passport_attachment", Value: 1},
			{Key: "passport", Value: 1},
			{Key: "ticket_attachment", Value: 1},
			{Key: "ticket", Value: 1},
			{Key: "education_loa_attachment", Value: 1},
			{Key: "education_loa", Value: 1},
			{Key: "business_letter_attachment", Value: 1},
			{Key: "business_letter", Value: 1},
			{Key: "business_supporting_attachment", Value: 1},
			{Key: "business_supporting", Value: 1},
			{Key: "health_letter_attachment", Value: 1},
			{Key: "health_letter", Value: 1},

			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
			{Key: "requested_at", Value: 1},
			{Key: "validated_at", Value: 1},

			{Key: "created_by", Value: 1},
			{Key: "updated_by", Value: 1},
			{Key: "requested_by", Value: 1},
			{Key: "authorized_by", Value: 1},
			{Key: "validated_by", Value: 1},
			{Key: "accepted_by", Value: 1},
			{Key: "declined_by", Value: 1},

			{Key: "creator", Value: bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$ifNull", Value: bson.A{"$creator._id", false}}},
				"$creator",
				"$$REMOVE",
			}}}},
			{Key: "updater", Value: bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$ifNull", Value: bson.A{"$updater._id", false}}},
				"$updater",
				"$$REMOVE",
			}}}},
			{Key: "requester", Value: bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$ifNull", Value: bson.A{"$requester._id", false}}},
				"$requester",
				"$$REMOVE",
			}}}},
			{Key: "authorizer", Value: bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$ifNull", Value: bson.A{"$authorizer._id", false}}},
				"$authorizer",
				"$$REMOVE",
			}}}},
			{Key: "validater", Value: bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$ifNull", Value: bson.A{"$validater._id", false}}},
				"$validater",
				"$$REMOVE",
			}}}},
			{Key: "approver", Value: bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$ifNull", Value: bson.A{"$approver._id", false}}},
				"$approver",
				"$$REMOVE",
			}}}},
			{Key: "rejecter", Value: bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$ifNull", Value: bson.A{"$rejecter._id", false}}},
				"$rejecter",
				"$$REMOVE",
			}}}},
			{Key: "accepter", Value: bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$ifNull", Value: bson.A{"$accepter._id", false}}},
				"$accepter",
				"$$REMOVE",
			}}}},
			{Key: "decliner", Value: bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$ifNull", Value: bson.A{"$decliner._id", false}}},
				"$decliner",
				"$$REMOVE",
			}}}},

			{Key: "approved_currency_ids", Value: 1},
			{Key: "approved_currencies", Value: 1},
			{Key: "approved_amounts", Value: 1},
			{Key: "approved_by", Value: 1},
			{Key: "approved_at", Value: 1},
			{Key: "authorized_at", Value: 1},
			{Key: "accepted_at", Value: 1},
			{Key: "declined_at", Value: 1},
			{Key: "rejected_at", Value: 1},
			{Key: "deleted_by", Value: 1},
			{Key: "deleted_at", Value: 1},
			{Key: "is_deleted", Value: 1},
		}},
	})

	return pipeline
}
