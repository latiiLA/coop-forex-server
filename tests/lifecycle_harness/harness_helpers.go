// Isolated request-lifecycle harness helpers.
// Talks only to disposable local MongoDB and an in-process Gin router.
package lifecycle_harness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/router"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const clientIP = "127.0.0.1"

// Role permission sets copied from snapshot db/migrations/20251222112804_seed_roles.up.json.
var rolePerms = map[string][]string{
	"BRANCHUSER": {
		"request:add", "request:status", "request:view-orgapproved", "request:send",
		"request:view-orgdrafted", "request:update", "request:delete", "request:view-new",
		"request:view-orgauthorized", "request:view-orgrejected",
		"account:view", "manual:view", "help:view",
	},
	"BRANCHAUTHORIZER": {
		"request:status", "request:view-orgapproved", "request:view-new", "request:authorize",
		"request:reject", "request:view-orgauthorized", "request:view-orgrejected",
		"account:view", "manual:view", "help:view",
	},
	"PROCESSORUSER": {
		"request:add", "request:status", "request:view-approved", "request:send",
		"request:view-orgdrafted", "request:update", "request:delete", "request:view-new",
		"request:view-orgauthorized", "request:view-orgrejected", "request:decline",
		"request:view-declined", "request:process", "request:view-accepted",
		"request:generate-report", "account:view", "manual:view", "help:view",
	},
	"PROCESSORAUTHORIZER": {
		"request:status", "request:reject", "request:view-approved", "request:view-new",
		"request:authorize", "request:view-orgauthorized", "request:view-orgrejected",
		"request:decline", "request:view-declined", "request:process", "request:view-accepted",
		"request:generate-report", "account:view", "manual:view", "help:view",
	},
	"FOREXUSER": {
		"request:validate", "request:view-authorized", "request:view-validated",
		"request:view-declined", "request:view", "request:reject", "request:lock",
		"request:unlock", "request:view-assignee", "request:view-approved",
		"request:view-rejected", "request:view-accepted", "request:generate-report",
		"account:view", "manual:view", "help:view",
	},
	"FOREXAPPROVER": {
		"request:view-validated", "request:view-declined", "request:view", "request:approve",
		"request:reject", "request:lock", "request:unlock", "request:view-assignee",
		"request:view-approved", "request:view-rejected", "request:view-accepted",
		"request:generate-report", "account:view", "manual:view", "help:view",
	},
}

type actor struct {
	ID       primitive.ObjectID
	Username string
	Role     string
	Branch   primitive.ObjectID
	Perms    []string
}

type env struct {
	db       *mongo.Database
	engine   *gin.Engine
	etb      primitive.ObjectID
	usd      primitive.ObjectID
	branchA  primitive.ObjectID
	branchB  primitive.ObjectID
	branchC  primitive.ObjectID
	maker    actor
	checker  actor
	checker2 actor
	otherBA  actor
	proc     actor
	procAuth actor
	forex    actor
	forex2   actor
	approver actor
	outsider actor
	dual     actor
}

type createOpt func(*createSpec)

type createSpec struct {
	name string
}

func withName(name string) createOpt {
	return func(s *createSpec) { s.name = name }
}

func seed(t *testing.T, ctx context.Context, db *mongo.Database) *env {
	t.Helper()
	e := &env{db: db, etb: primitive.NewObjectID(), usd: primitive.NewObjectID(), branchA: primitive.NewObjectID(), branchB: primitive.NewObjectID(), branchC: primitive.NewObjectID()}
	now := time.Now()
	_, err := db.Collection("currencies").InsertMany(ctx, []any{
		bson.M{"_id": e.etb, "name": "Birr", "short_code": "ETB", "created_at": now, "updated_at": now, "is_deleted": false},
		bson.M{"_id": e.usd, "name": "US Dollar", "short_code": "USD", "created_at": now, "updated_at": now, "is_deleted": false},
	})
	must(t, err)
	_, err = db.Collection("branches").InsertMany(ctx, []any{
		bson.M{"_id": e.branchA, "name": "Synthetic Branch A", "branch_code": "TSTA", "email": "", "is_result_processor": false, "is_deleted": false, "created_at": now, "updated_at": now},
		bson.M{"_id": e.branchB, "name": "Synthetic Processor B", "branch_code": "TSTB", "email": "", "is_result_processor": true, "is_deleted": false, "created_at": now, "updated_at": now},
		bson.M{"_id": e.branchC, "name": "Synthetic Branch C", "branch_code": "TSTC", "email": "", "is_result_processor": false, "is_deleted": false, "created_at": now, "updated_at": now},
	})
	must(t, err)

	e.maker = e.addUser(t, ctx, "branchuser", "BRANCHUSER", &e.branchA)
	e.checker = e.addUser(t, ctx, "branchauthorizer", "BRANCHAUTHORIZER", &e.branchA)
	e.checker2 = e.addUser(t, ctx, "branchauthorizer2", "BRANCHAUTHORIZER", &e.branchA)
	e.otherBA = e.addUser(t, ctx, "otherbranchauthorizer", "BRANCHAUTHORIZER", &e.branchB)
	e.proc = e.addUser(t, ctx, "branchprocessoruser", "PROCESSORUSER", &e.branchB)
	e.procAuth = e.addUser(t, ctx, "branchprocessorauthorizer", "PROCESSORAUTHORIZER", &e.branchB)
	e.forex = e.addUser(t, ctx, "forexuser", "FOREXUSER", nil)
	e.forex2 = e.addUser(t, ctx, "forexuser2", "FOREXUSER", nil)
	e.approver = e.addUser(t, ctx, "forexapprover", "FOREXAPPROVER", nil)
	e.outsider = e.addUser(t, ctx, "outsiderprocessor", "PROCESSORUSER", &e.branchC)
	dualPerms := append([]string{}, rolePerms["BRANCHUSER"]...)
	dualPerms = append(dualPerms, "request:authorize", "request:reject")
	e.dual = e.addUserPerms(t, ctx, "dualmakerchecker", "BRANCHUSER", &e.branchA, dualPerms)

	r := gin.New()
	api := r.Group("/api")
	router.NewRequestRouter(db, 8*time.Second, api)
	e.engine = r
	return e
}

func (e *env) addUser(t *testing.T, ctx context.Context, username, role string, branch *primitive.ObjectID) actor {
	t.Helper()
	return e.addUserPerms(t, ctx, username, role, branch, rolePerms[role])
}

func (e *env) addUserPerms(t *testing.T, ctx context.Context, username, role string, branch *primitive.ObjectID, perms []string) actor {
	t.Helper()
	roleID := primitive.NewObjectID()
	profileID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	now := time.Now()
	_, err := e.db.Collection("roles").InsertOne(ctx, bson.M{
		"_id": roleID, "name": role, "permissions": perms, "is_deleted": false, "created_at": now, "updated_at": now,
	})
	must(t, err)
	profile := bson.M{
		"_id": profileID, "first_name": "Test", "middle_name": "Synthetic", "last_name": username,
		"display_name": username, "email": "", "created_at": now, "updated_at": now,
	}
	if branch != nil {
		profile["branch_id"] = *branch
	}
	_, err = e.db.Collection("profiles").InsertOne(ctx, profile)
	must(t, err)
	_, err = e.db.Collection("users").InsertOne(ctx, bson.M{
		"_id": userID, "username": username, "role_id": roleID, "profile_id": profileID,
		"permissions": []string{}, "status": "active", "is_deleted": false,
		"password": "not-a-real-credential", "created_at": now, "updated_at": now,
	})
	must(t, err)
	br := primitive.NilObjectID
	if branch != nil {
		br = *branch
	}
	return actor{ID: userID, Username: username, Role: role, Branch: br, Perms: perms}
}

func (e *env) token(a actor) string {
	tok, err := infrastructure.GenerateToken(a.ID, a.Role, a.Branch, primitive.NilObjectID, a.Perms, clientIP)
	if err != nil {
		panic(err)
	}
	return tok
}

func (e *env) do(a *actor, method, path string, body []byte, contentType string) (int, string) {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.RemoteAddr = clientIP + ":1234"
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if a != nil {
		req.Header.Set("Authorization", "Bearer "+e.token(*a))
	}
	rec := httptest.NewRecorder()
	e.engine.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func (e *env) expectStatus(t *testing.T, a actor, method, path string, body []byte, contentType string, want int) {
	t.Helper()
	e.expectActor(t, &a, method, path, body, contentType, want)
}

func (e *env) expectAnon(t *testing.T, method, path string, body []byte, contentType string, want int) {
	t.Helper()
	e.expectActor(t, nil, method, path, body, contentType, want)
}

func (e *env) expectActor(t *testing.T, a *actor, method, path string, body []byte, contentType string, want int) {
	t.Helper()
	code, resp := e.do(a, method, path, body, contentType)
	if code != want {
		t.Fatalf("%s %s status=%d want %d body=%s", method, path, code, want, resp)
	}
}

func (e *env) createDraft(t *testing.T, a actor, authorizer primitive.ObjectID, account string, fcy primitive.ObjectID, amount, mode, card string) primitive.ObjectID {
	t.Helper()
	code, body := e.postCreate(a, authorizer, account, fcy, amount, mode, card, "TEST APPLICANT")
	if code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", code, body)
	}
	return parseID(t, body)
}

func (e *env) createDraftExpect(t *testing.T, a actor, authorizer primitive.ObjectID, account string, fcy primitive.ObjectID, amount, mode, card string, want int, opts ...createOpt) {
	t.Helper()
	spec := createSpec{name: "TEST APPLICANT"}
	for _, o := range opts {
		o(&spec)
	}
	code, body := e.postCreate(a, authorizer, account, fcy, amount, mode, card, spec.name)
	if code != want {
		t.Fatalf("create status=%d want %d body=%s", code, want, body)
	}
}

func (e *env) postCreate(a actor, authorizer primitive.ObjectID, account string, fcy primitive.ObjectID, amount, mode, card, name string) (int, string) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fields := map[string]string{
		"applicant_name":           name,
		"applicant_account_number": account,
		"average_deposit":          "1000",
		"total_fcy_generated":      "0",
		"current_fcy_performance":  "0",
		"travel_purpose_id":        primitive.NewObjectID().Hex(),
		"travel_country_id":        primitive.NewObjectID().Hex(),
		"requesting_as":            "Individual",
		"account_currency_id":      e.etb.Hex(),
		"fcy_requested_id":         fcy.Hex(),
		"fcy_requested_amount":     amount,
		"fcy_acceptance_mode":      mode,
		"branch_recommendation":    "Synthetic recommendation",
		"authorizer_id":            authorizer.Hex(),
	}
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	if card != "" || mode == "Card" || mode == "Cash-and-Card" {
		_ = w.WriteField("card_associated_account", card)
	}
	_ = w.WriteField("accounts_to_deduct", "DEDUCT0000000001")
	_ = w.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/request", &buf)
	req.RemoteAddr = clientIP + ":1234"
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+e.token(a))
	rec := httptest.NewRecorder()
	e.engine.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func (e *env) updateDraft(t *testing.T, a actor, id, authorizer primitive.ObjectID, account string, fcy primitive.ObjectID, amount string, want int) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fields := map[string]string{
		"applicant_name":           "TEST APPLICANT",
		"applicant_account_number": account,
		"average_deposit":          "1000",
		"total_fcy_generated":      "0",
		"current_fcy_performance":  "0",
		"travel_purpose_id":        primitive.NewObjectID().Hex(),
		"travel_country_id":        primitive.NewObjectID().Hex(),
		"requesting_as":            "Individual",
		"account_currency_id":      e.etb.Hex(),
		"fcy_requested_id":         fcy.Hex(),
		"fcy_requested_amount":     amount,
		"fcy_acceptance_mode":      "Cash",
		"branch_recommendation":    "Revised synthetic recommendation",
		"authorizer_id":            authorizer.Hex(),
	}
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	_ = w.WriteField("accounts_to_deduct", "DEDUCT0000000001")
	_ = w.Close()
	req := httptest.NewRequest(http.MethodPut, "/api/updaterequest/"+id.Hex(), &buf)
	req.RemoteAddr = clientIP + ":1234"
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+e.token(a))
	rec := httptest.NewRecorder()
	e.engine.ServeHTTP(rec, req)
	if rec.Code != want {
		t.Fatalf("update status=%d want %d body=%s", rec.Code, want, rec.Body.String())
	}
}

func (e *env) advanceTo(t *testing.T, target, account string) primitive.ObjectID {
	t.Helper()
	id := e.createDraft(t, e.maker, e.checker.ID, account, e.usd, "500", "Cash", "")
	if target == "Drafted" {
		return id
	}
	e.expectStatus(t, e.maker, http.MethodPost, "/api/orgsendrequest/"+id.Hex(), nil, "", http.StatusOK)
	if target == "New" {
		return id
	}
	e.expectStatus(t, e.checker, http.MethodPost, "/api/orgauthorizerequest/"+id.Hex(), nil, "", http.StatusOK)
	if target == "Authorized" {
		return id
	}
	e.expectStatus(t, e.forex, http.MethodPost, "/api/assignrequest/"+id.Hex(), nil, "application/json", http.StatusOK)
	valBody := []byte(fmt.Sprintf(`{"validated_average_deposit":10,"validated_current_balance":20,"validated_account_currency_id":"%s"}`, e.etb.Hex()))
	e.expectStatus(t, e.forex, http.MethodPost, "/api/validaterequest/"+id.Hex(), valBody, "application/json", http.StatusOK)
	if target == "Validated" {
		return id
	}
	e.expectStatus(t, e.approver, http.MethodPost, "/api/assignrequest/"+id.Hex(), nil, "application/json", http.StatusOK)
	approve := []byte(fmt.Sprintf(`{"approved_currency_ids":["%s"],"approved_amounts":[400],"approved_amount_in_cash":[400],"approved_amount_in_card":[0]}`, e.usd.Hex()))
	e.expectStatus(t, e.approver, http.MethodPost, "/api/approverequest/"+id.Hex(), approve, "application/json", http.StatusOK)
	if target != "Approved" {
		t.Fatalf("unsupported target %s", target)
	}
	return id
}

func (e *env) statusOf(t *testing.T, id primitive.ObjectID) string {
	t.Helper()
	var doc bson.M
	if err := e.db.Collection("requests").FindOne(context.Background(), bson.M{"_id": id}).Decode(&doc); err != nil {
		t.Fatalf("load request: %v", err)
	}
	return fmt.Sprint(doc["request_status"])
}

func (e *env) lockedBy(t *testing.T, id primitive.ObjectID) *primitive.ObjectID {
	t.Helper()
	var doc struct {
		LockedBy *primitive.ObjectID `bson:"locked_by"`
	}
	if err := e.db.Collection("requests").FindOne(context.Background(), bson.M{"_id": id}).Decode(&doc); err != nil {
		t.Fatalf("load lock: %v", err)
	}
	return doc.LockedBy
}

func (e *env) processingBranch(t *testing.T, id primitive.ObjectID) *primitive.ObjectID {
	t.Helper()
	var doc struct {
		ProcessingBranchID *primitive.ObjectID `bson:"processing_branch_id"`
	}
	if err := e.db.Collection("requests").FindOne(context.Background(), bson.M{"_id": id}).Decode(&doc); err != nil {
		t.Fatalf("load processing branch: %v", err)
	}
	return doc.ProcessingBranchID
}

type actLog struct {
	Action string `bson:"action"`
	From   string `bson:"from_state"`
	To     string `bson:"to_state"`
}

func (e *env) logs(t *testing.T, id primitive.ObjectID) []actLog {
	t.Helper()
	cur, err := e.db.Collection("activity_logs").Find(context.Background(), bson.M{"entity_id": id})
	if err != nil {
		t.Fatal(err)
	}
	var out []actLog
	if err := cur.All(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func (e *env) actions(t *testing.T, id primitive.ObjectID) []string {
	t.Helper()
	logs := e.logs(t, id)
	out := make([]string, 0, len(logs))
	for _, l := range logs {
		out = append(out, l.Action)
	}
	return out
}

func (e *env) queueIDs(t *testing.T, a actor, status string) []primitive.ObjectID {
	t.Helper()
	code, body := e.do(&a, http.MethodGet, "/api/requests/queue?status="+status+"&limit=100", nil, "")
	if code != http.StatusOK {
		t.Fatalf("queue %s as %s status=%d body=%s", status, a.Username, code, body)
	}
	return parseDataIDs(t, body)
}

func (e *env) orgIDs(t *testing.T, a actor) []primitive.ObjectID {
	t.Helper()
	code, body := e.do(&a, http.MethodGet, "/api/orgrequests?limit=100", nil, "")
	if code != http.StatusOK {
		t.Fatalf("orgrequests status=%d body=%s", code, body)
	}
	return parseDataIDs(t, body)
}

func parseID(t *testing.T, body string) primitive.ObjectID {
	t.Helper()
	var wrap struct {
		Data struct {
			ID string `json:"_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &wrap); err != nil {
		t.Fatalf("parse create: %v body=%s", err, body)
	}
	id, err := primitive.ObjectIDFromHex(wrap.Data.ID)
	if err != nil {
		t.Fatalf("id: %v body=%s", err, body)
	}
	return id
}

func parseDataIDs(t *testing.T, body string) []primitive.ObjectID {
	t.Helper()
	var wrap struct {
		Data []struct {
			ID string `json:"_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &wrap); err != nil {
		t.Fatalf("parse list: %v body=%s", err, body)
	}
	out := make([]primitive.ObjectID, 0, len(wrap.Data))
	for _, row := range wrap.Data {
		id, err := primitive.ObjectIDFromHex(row.ID)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, id)
	}
	return out
}

func contains(ids []primitive.ObjectID, id primitive.ObjectID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func containsStr(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
