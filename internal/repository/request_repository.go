package repository

import (
	"context"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/common"
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

func (rr *requestRepository) aggregatePaginated(ctx context.Context, match bson.D, pageQuery model.PageQuery, populate bool) (*model.PaginatedRequests, error) {
	// Apply date range filter on created_at
	if pageQuery.DateFrom != nil || pageQuery.DateTo != nil {
		dateFilter := bson.D{}
		if pageQuery.DateFrom != nil {
			dateFilter = append(dateFilter, bson.E{Key: "$gte", Value: *pageQuery.DateFrom})
		}
		if pageQuery.DateTo != nil {
			dateFilter = append(dateFilter, bson.E{Key: "$lte", Value: *pageQuery.DateTo})
		}
		match = append(match, bson.E{Key: "created_at", Value: dateFilter})
	}

	dataStages := mongo.Pipeline{}
	if pageQuery.Limit > 0 {
		skip := int64((pageQuery.Page - 1) * pageQuery.Limit)
		dataStages = append(dataStages,
			bson.D{{Key: "$skip", Value: skip}},
			bson.D{{Key: "$limit", Value: int64(pageQuery.Limit)}},
		)
	}
	dataStages = append(dataStages, utils.BuildCommonRequestPipelineStages(populate)...)

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}}}},
		bson.D{{Key: "$facet", Value: bson.D{
			{Key: "metadata", Value: bson.A{
				bson.D{{Key: "$count", Value: "total"}},
			}},
			{Key: "data", Value: dataStages},
		}}},
	}

	cursor, err := rr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var facetResults []struct {
		Metadata []struct {
			Total int64 `bson:"total"`
		} `bson:"metadata"`
		Data []model.Request `bson:"data"`
	}
	if err := cursor.All(ctx, &facetResults); err != nil {
		return nil, err
	}

	result := &model.PaginatedRequests{
		Items: []model.Request{},
		Page:  pageQuery.Page,
		Limit: pageQuery.Limit,
	}
	if len(facetResults) == 0 {
		return result, nil
	}

	if len(facetResults[0].Metadata) > 0 {
		result.Total = facetResults[0].Metadata[0].Total
	}
	if facetResults[0].Data != nil {
		result.Items = facetResults[0].Data
	}
	if pageQuery.Limit == 0 {
		result.Limit = int(result.Total)
		if result.Page < 1 {
			result.Page = 1
		}
	}

	return result, nil
}

func (rr *requestRepository) FindAll(ctx context.Context, populate bool, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	match := bson.D{
		{Key: "is_deleted", Value: false},
		{Key: "request_status", Value: bson.D{
			{Key: "$nin", Value: bson.A{"New", "Rejected", "Drafted"}},
		}},
	}
	return rr.aggregatePaginated(ctx, match, pageQuery, populate)
}

func (rr *requestRepository) Validate(ctx context.Context, request_id primitive.ObjectID, request *model.Request) error {
	update := bson.M{
		"$set": request,
	}

	filter := bson.M{"_id": request_id}
	result, err := rr.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return common.ErrRequestNotFound
	}

	return nil
}

func (rr *requestRepository) FindByID(ctx context.Context, requestID primitive.ObjectID, populate bool) (*model.Request, error) {
	var results []model.Request

	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "_id", Value: requestID},
				{Key: "is_deleted", Value: false},
			}},
		},
	}

	pipeline = append(pipeline, utils.BuildCommonRequestPipelineStages(populate)...)

	cursor, err := rr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

// orgInvolvementMatch returns requests this org unit is tied to: requesting
// branch/department, or (for branches) processing branch after accept.
func orgInvolvementMatch(orgKey string, orgID primitive.ObjectID) bson.D {
	if orgKey == "branch_id" {
		return bson.D{{Key: "$or", Value: bson.A{
			bson.D{{Key: "branch_id", Value: orgID}},
			bson.D{{Key: "processing_branch_id", Value: orgID}},
		}}}
	}
	return bson.D{{Key: orgKey, Value: orgID}}
}

func (rr *requestRepository) FindAllByOrgID(ctx context.Context, orgKey string, orgID primitive.ObjectID, populate bool, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	match := bson.D{{Key: "is_deleted", Value: false}}
	match = append(match, orgInvolvementMatch(orgKey, orgID)...)
	return rr.aggregatePaginated(ctx, match, pageQuery, populate)
}

func (rr *requestRepository) FindOrgByRequestStatus(ctx context.Context, orgID primitive.ObjectID, orgKey, request_status string, populate bool, pageQuery model.PageQuery, assigneeViewer *primitive.ObjectID) (*model.PaginatedRequests, error) {
	match := bson.D{
		{Key: "is_deleted", Value: false},
		{Key: "request_status", Value: request_status},
	}
	match = append(match, orgInvolvementMatch(orgKey, orgID)...)
	if assigneeViewer != nil && !assigneeViewer.IsZero() {
		match = append(match, assigneeVisibleClause(*assigneeViewer))
	}
	return rr.aggregatePaginated(ctx, match, pageQuery, populate)
}

func (rr *requestRepository) Update(ctx context.Context, requestID primitive.ObjectID, request *model.RequestUpdate) error {
	update := bson.M{
		"$set": request,
	}
	filter := bson.M{"_id": requestID}
	result, err := rr.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return common.ErrRequestNotFound
	}

	return nil
}

func (rr *requestRepository) updateConflictOrNotFound(ctx context.Context, requestID primitive.ObjectID) error {
	count, err := rr.collection.CountDocuments(ctx, bson.M{"_id": requestID})
	if err != nil {
		return err
	}
	if count == 0 {
		return common.ErrRequestNotFound
	}
	return common.ErrRequestConflict
}

func (rr *requestRepository) UpdateIfStatus(ctx context.Context, requestID primitive.ObjectID, expectedStatuses []string, request *model.RequestUpdate) error {
	filter := bson.M{
		"_id":            requestID,
		"request_status": bson.M{"$in": expectedStatuses},
		"is_deleted":     false,
	}
	result, err := rr.collection.UpdateOne(ctx, filter, bson.M{"$set": request})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return rr.updateConflictOrNotFound(ctx, requestID)
	}
	return nil
}

func (rr *requestRepository) ValidateIfStatus(ctx context.Context, requestID primitive.ObjectID, expectedStatuses []string, request *model.Request) error {
	filter := bson.M{
		"_id":            requestID,
		"request_status": bson.M{"$in": expectedStatuses},
		"is_deleted":     false,
	}
	result, err := rr.collection.UpdateOne(ctx, filter, bson.M{"$set": request})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return rr.updateConflictOrNotFound(ctx, requestID)
	}
	return nil
}

func (rr *requestRepository) FindByAttachmentFileID(ctx context.Context, fileID primitive.ObjectID) (*model.Request, error) {
	filter := bson.M{
		"is_deleted": false,
		"$or": []bson.M{
			{"passport_attachment": fileID},
			{"ticket_attachment": fileID},
			{"visa_attachment": fileID},
			{"business_license_attachment": fileID},
			{"education_loa_attachment": fileID},
			{"health_letter_attachment": fileID},
			{"business_supporting_attachment": fileID},
		},
	}
	var request model.Request
	err := rr.collection.FindOne(ctx, filter).Decode(&request)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, common.ErrRequestNotFound
		}
		return nil, err
	}
	return &request, nil
}

func assigneeVisibleClause(viewerID primitive.ObjectID) bson.E {
	now := time.Now()
	return bson.E{Key: "$or", Value: bson.A{
		bson.M{"locked_by": bson.M{"$exists": false}},
		bson.M{"locked_by": nil},
		bson.M{"locked_by": viewerID},
		bson.M{"lock_expires_at": bson.M{"$lte": now}},
	}}
}

func (rr *requestRepository) FindByRequestStatus(ctx context.Context, request_status string, populate bool, pageQuery model.PageQuery, assigneeViewer *primitive.ObjectID) (*model.PaginatedRequests, error) {
	match := bson.D{
		{Key: "is_deleted", Value: false},
		{Key: "request_status", Value: request_status},
	}
	if assigneeViewer != nil && !assigneeViewer.IsZero() {
		match = append(match, assigneeVisibleClause(*assigneeViewer))
	}
	return rr.aggregatePaginated(ctx, match, pageQuery, populate)
}

func (rr *requestRepository) FindByCreatorAndStatus(ctx context.Context, creatorID primitive.ObjectID, request_status string, populate bool, pageQuery model.PageQuery) (*model.PaginatedRequests, error) {
	match := bson.D{
		{Key: "is_deleted", Value: false},
		{Key: "request_status", Value: request_status},
		{Key: "created_by", Value: creatorID},
	}
	return rr.aggregatePaginated(ctx, match, pageQuery, populate)
}

func (rr *requestRepository) CountByStatus(ctx context.Context, orgKey string, orgID primitive.ObjectID) (*model.RequestStatusSummary, error) {
	match := bson.D{{Key: "is_deleted", Value: false}}
	if orgKey != "" && !orgID.IsZero() {
		match = append(match, orgInvolvementMatch(orgKey, orgID)...)
	} else {
		// Bank-wide home summary matches GetAllRequests scope.
		match = append(match, bson.E{Key: "request_status", Value: bson.D{
			{Key: "$nin", Value: bson.A{"New", "Rejected", "Drafted"}},
		}})
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$request_status"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := rr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID    string `bson:"_id"`
		Count int64  `bson:"count"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	summary := &model.RequestStatusSummary{
		ByStatus: map[string]int64{},
	}
	for _, row := range rows {
		summary.ByStatus[row.ID] = row.Count
		summary.Total += row.Count
	}
	return summary, nil
}

func (rr *requestRepository) Assign(ctx context.Context, requestID, userID primitive.ObjectID, assignedAt time.Time) error {
	claimed, err := rr.AssignIfAvailable(ctx, requestID, userID, assignedAt)
	if err != nil {
		return err
	}
	if !claimed {
		return common.ErrRequestIsLocked
	}
	return nil
}

func (rr *requestRepository) AssignIfAvailable(ctx context.Context, requestID, userID primitive.ObjectID, assignedAt time.Time) (bool, error) {
	now := assignedAt
	if now.IsZero() {
		now = time.Now()
	}
	// First-wins: claim only when unassigned, expired, or already owned by the same user.
	// Status-gated so assign cannot land after Approve/Accept/etc.
	filter := bson.M{
		"_id": requestID,
		"is_deleted": false,
		"request_status": bson.M{"$in": []string{
			string(model.ReqStatusNew),
			string(model.ReqStatusAuthorized),
			string(model.ReqStatusValidated),
		}},
		"$or": []bson.M{
			{"locked_by": bson.M{"$exists": false}},
			{"locked_by": nil},
			{"locked_by": userID},
			{"lock_expires_at": bson.M{"$lte": now}},
		},
	}
	result, err := rr.collection.UpdateOne(ctx, filter, bson.M{
		"$set": bson.M{
			"locked_by":  userID,
			"locked_at":  now,
			"updated_at": now,
		},
		"$unset": bson.M{
			"lock_expires_at": "",
		},
	})
	if err != nil {
		return false, err
	}
	if result.MatchedCount == 0 {
		// Distinguish not-found vs collision.
		count, cErr := rr.collection.CountDocuments(ctx, bson.M{"_id": requestID})
		if cErr != nil {
			return false, cErr
		}
		if count == 0 {
			return false, common.ErrRequestNotFound
		}
		return false, nil
	}
	return true, nil
}

func (rr *requestRepository) ClearAssignment(ctx context.Context, requestID primitive.ObjectID) error {
	result, err := rr.collection.UpdateOne(ctx, bson.M{"_id": requestID}, bson.M{
		"$unset": bson.M{
			"locked_by":       "",
			"locked_at":       "",
			"lock_expires_at": "",
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return common.ErrRequestNotFound
	}
	return nil
}
