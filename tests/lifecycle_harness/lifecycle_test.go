// Isolated request-lifecycle harness against disposable local MongoDB + in-process Gin.
package lifecycle_harness

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/configs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	configs.JwtSecret = "synthetic-lifecycle-test-secret"
	configs.AccessTokenExpiry = time.Hour
	configs.FileUploadPath = os.TempDir()
	os.Exit(m.Run())
}

func TestRequestLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil {
		t.Fatalf("mongo connect: %v", err)
	}
	defer client.Disconnect(context.Background())
	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("mongo ping: %v", err)
	}

	dbName := fmt.Sprintf("forex_lifecycle_%d", time.Now().UnixNano())
	db := client.Database(dbName)
	t.Cleanup(func() {
		_ = db.Drop(context.Background())
	})

	e := seed(t, ctx, db)

	var happyID primitive.ObjectID
	t.Run("happy_path_six_roles", func(t *testing.T) {
		happyID = e.createDraft(t, e.maker, e.checker.ID, "TESTACCT00000011", e.usd, "500", "Cash", "")
		if st := e.statusOf(t, happyID); st != "Drafted" {
			t.Fatalf("status=%s", st)
		}
		if ids := e.queueIDs(t, e.maker, "Drafted"); !contains(ids, happyID) {
			t.Fatalf("maker drafted queue missing request: %v", ids)
		}
		e.expectStatus(t, e.maker, http.MethodPost, "/api/orgsendrequest/"+happyID.Hex(), nil, "", http.StatusOK)
		if st := e.statusOf(t, happyID); st != "New" {
			t.Fatalf("after send status=%s", st)
		}
		if ids := e.queueIDs(t, e.checker, "New"); !contains(ids, happyID) {
			t.Fatalf("assigned authorizer cannot see New queue: %v", ids)
		}
		if ids := e.queueIDs(t, e.checker2, "New"); contains(ids, happyID) {
			t.Fatalf("second authorizer sees request assigned to someone else: %v", ids)
		}
		if ids := e.queueIDs(t, e.maker, "New"); contains(ids, happyID) {
			t.Fatalf("maker New queue unexpectedly includes assigned request: %v", ids)
		}
		org := e.orgIDs(t, e.maker)
		if !contains(org, happyID) {
			t.Fatalf("maker org status list missing submitted request: %v", org)
		}
		e.expectStatus(t, e.checker, http.MethodPost, "/api/orgauthorizerequest/"+happyID.Hex(), nil, "", http.StatusOK)
		if st := e.statusOf(t, happyID); st != "Authorized" {
			t.Fatalf("after authorize status=%s locked=%v", st, e.lockedBy(t, happyID))
		}
		if e.lockedBy(t, happyID) != nil {
			t.Fatalf("authorize should clear assignment, locked_by=%v", e.lockedBy(t, happyID))
		}

		e.expectStatus(t, e.forex, http.MethodPost, "/api/assignrequest/"+happyID.Hex(), nil, "application/json", http.StatusOK)
		if ids := e.queueIDs(t, e.forex2, "Authorized"); contains(ids, happyID) {
			t.Fatalf("other validator sees assigned Authorized request")
		}
		if ids := e.queueIDs(t, e.forex, "Authorized"); !contains(ids, happyID) {
			t.Fatalf("assignee validator cannot see Authorized request")
		}
		valBody := []byte(fmt.Sprintf(`{"validated_average_deposit":10,"validated_current_balance":20,"validated_account_currency_id":"%s"}`, e.etb.Hex()))
		e.expectStatus(t, e.forex, http.MethodPost, "/api/validaterequest/"+happyID.Hex(), valBody, "application/json", http.StatusOK)
		if st := e.statusOf(t, happyID); st != "Validated" {
			t.Fatalf("after validate status=%s", st)
		}

		e.expectStatus(t, e.approver, http.MethodPost, "/api/assignrequest/"+happyID.Hex(), nil, "application/json", http.StatusOK)
		approve := []byte(fmt.Sprintf(`{"approved_currency_ids":["%s"],"approved_amounts":[400],"approved_amount_in_cash":[400],"approved_amount_in_card":[0]}`, e.usd.Hex()))
		e.expectStatus(t, e.approver, http.MethodPost, "/api/approverequest/"+happyID.Hex(), approve, "application/json", http.StatusOK)
		if st := e.statusOf(t, happyID); st != "Approved" {
			t.Fatalf("after approve status=%s", st)
		}

		accept := []byte(fmt.Sprintf(`{"accepted_currency_ids":["%s"],"accepted_amounts":[300],"accepted_amount_in_cash":[300],"accepted_amount_in_card":[0]}`, e.usd.Hex()))
		e.expectStatus(t, e.proc, http.MethodPost, "/api/acceptrequest/"+happyID.Hex(), accept, "application/json", http.StatusOK)
		if st := e.statusOf(t, happyID); st != "Accepted" {
			t.Fatalf("after accept status=%s", st)
		}
		if pb := e.processingBranch(t, happyID); pb == nil || *pb != e.branchB {
			t.Fatalf("processing branch=%v want %s", pb, e.branchB.Hex())
		}

		actions := e.actions(t, happyID)
		for _, want := range []string{
			"request.create", "request.send", "request.authorize",
			"request.assign", "request.validate", "request.approve", "request.accept",
		} {
			if !containsStr(actions, want) {
				t.Errorf("missing activity %s in %v", want, actions)
			}
		}
		assignN := 0
		for _, a := range actions {
			if a == "request.assign" {
				assignN++
			}
		}
		if assignN < 2 {
			t.Errorf("activity request.assign count=%d want 2 (validator assign and approver assign)", assignN)
		}
	})

	t.Run("reject_edit_resubmit", func(t *testing.T) {
		id := e.createDraft(t, e.maker, e.checker.ID, "TESTACCT00000022", e.usd, "250", "Cash", "")
		e.expectStatus(t, e.maker, http.MethodPost, "/api/orgsendrequest/"+id.Hex(), nil, "", http.StatusOK)
		rej := []byte(`{"rejection_reason":"synthetic incomplete documents"}`)
		e.expectStatus(t, e.checker, http.MethodPost, "/api/rejectrequest/"+id.Hex(), rej, "application/json", http.StatusOK)
		if st := e.statusOf(t, id); st != "Rejected" {
			t.Fatalf("after reject status=%s", st)
		}
		e.expectStatus(t, e.maker, http.MethodPost, "/api/orgsendrequest/"+id.Hex(), nil, "", http.StatusConflict)
		e.updateDraft(t, e.maker, id, e.checker.ID, "TESTACCT00000022", e.usd, "260", http.StatusOK)
		if st := e.statusOf(t, id); st != "Drafted" {
			t.Fatalf("after edit of rejected status=%s want Drafted", st)
		}
		logs := e.logs(t, id)
		var updateTo string
		for _, l := range logs {
			if l.Action == "request.update" {
				updateTo = l.To
			}
		}
		if updateTo != "Drafted" {
			t.Errorf("activity request.update to_state=%q; database status is Drafted (audit mismatch)", updateTo)
		}
		e.expectStatus(t, e.maker, http.MethodPost, "/api/orgsendrequest/"+id.Hex(), nil, "", http.StatusOK)
		if st := e.statusOf(t, id); st != "New" {
			t.Fatalf("after resubmit status=%s", st)
		}
	})

	t.Run("reject_from_authorized_validated_approved_and_invalid", func(t *testing.T) {
		id := e.advanceTo(t, "Authorized", "TESTACCT00000033")
		code, body := e.do(&e.forex, http.MethodPost, "/api/rejectrequest/"+id.Hex(), []byte(`{"rejection_reason":"bad"}`), "application/json")
		if code != http.StatusConflict {
			t.Errorf("unassigned reject status=%d body=%s (ErrRequestNotAssigned is not mapped; request left %s)", code, body, e.statusOf(t, id))
		}
		e.expectStatus(t, e.forex, http.MethodPost, "/api/assignrequest/"+id.Hex(), nil, "application/json", http.StatusOK)
		e.expectStatus(t, e.forex, http.MethodPost, "/api/rejectrequest/"+id.Hex(), []byte(`{"rejection_reason":"bad docs"}`), "application/json", http.StatusOK)
		if st := e.statusOf(t, id); st != "Rejected" {
			t.Fatalf("authorized reject status=%s", st)
		}

		id = e.advanceTo(t, "Validated", "TESTACCT00000034")
		e.expectStatus(t, e.approver, http.MethodPost, "/api/assignrequest/"+id.Hex(), nil, "application/json", http.StatusOK)
		e.expectStatus(t, e.approver, http.MethodPost, "/api/rejectrequest/"+id.Hex(), []byte(`{"rejection_reason":"limits"}`), "application/json", http.StatusOK)

		id = e.advanceTo(t, "Approved", "TESTACCT00000035")
		code, body = e.do(&e.approver, http.MethodPost, "/api/rejectrequest/"+id.Hex(), []byte(`{"rejection_reason":"late"}`), "application/json")
		if code != http.StatusConflict || e.statusOf(t, id) == "Rejected" {
			t.Errorf("unassigned approved-reject status=%d body=%s db=%s", code, body, e.statusOf(t, id))
		}

		draft := e.createDraft(t, e.maker, e.checker.ID, "TESTACCT00000036", e.usd, "100", "Cash", "")
		e.expectStatus(t, e.checker, http.MethodPost, "/api/rejectrequest/"+draft.Hex(), []byte(`{"rejection_reason":"nope"}`), "application/json", http.StatusConflict)
		e.expectStatus(t, e.maker, http.MethodPost, "/api/orgauthorizerequest/"+draft.Hex(), nil, "", http.StatusForbidden)
		e.expectStatus(t, e.checker, http.MethodPost, "/api/orgauthorizerequest/"+draft.Hex(), nil, "", http.StatusConflict)
	})

	t.Run("permissions_and_invalid_transitions", func(t *testing.T) {
		id := e.createDraft(t, e.maker, e.checker.ID, "TESTACCT00000041", e.usd, "100", "Cash", "")
		e.expectStatus(t, e.checker, http.MethodPost, "/api/request", nil, "", http.StatusForbidden)
		e.expectStatus(t, e.forex, http.MethodPost, "/api/orgauthorizerequest/"+id.Hex(), nil, "", http.StatusForbidden)
		e.expectStatus(t, e.approver, http.MethodPost, "/api/validaterequest/"+id.Hex(), []byte(`{"validated_average_deposit":1,"validated_current_balance":1,"validated_account_currency_id":"`+e.etb.Hex()+`"}`), "application/json", http.StatusForbidden)
		e.expectStatus(t, e.maker, http.MethodPost, "/api/approverequest/"+id.Hex(), []byte(`{}`), "application/json", http.StatusForbidden)
		e.expectStatus(t, e.maker, http.MethodPost, "/api/acceptrequest/"+id.Hex(), []byte(`{}`), "application/json", http.StatusForbidden)
		e.expectAnon(t, http.MethodPost, "/api/orgsendrequest/"+id.Hex(), nil, "", http.StatusUnauthorized)

		e.expectStatus(t, e.maker, http.MethodPost, "/api/orgsendrequest/"+id.Hex(), nil, "", http.StatusOK)
		e.expectStatus(t, e.maker, http.MethodPost, "/api/orgsendrequest/"+id.Hex(), nil, "", http.StatusConflict)
		e.expectStatus(t, e.forex, http.MethodPost, "/api/validaterequest/"+id.Hex(), []byte(`{"validated_average_deposit":1,"validated_current_balance":1,"validated_account_currency_id":"`+e.etb.Hex()+`"}`), "application/json", http.StatusConflict)
		code, body := e.do(&e.maker, http.MethodPatch, "/api/deleterequest/"+id.Hex(), nil, "")
		if e.statusOf(t, id) == "Deleted" {
			t.Fatalf("non-draft delete changed status body=%s", body)
		}
		if code != http.StatusConflict {
			t.Errorf("delete of New request status=%d body=%s (blocked, but response is not a transition conflict)", code, body)
		}
	})

	t.Run("maker_checker_separation", func(t *testing.T) {
		id := e.createDraft(t, e.maker, e.checker.ID, "TESTACCT00000051", e.usd, "100", "Cash", "")
		e.expectStatus(t, e.maker, http.MethodPost, "/api/orgsendrequest/"+id.Hex(), nil, "", http.StatusOK)
		code, _ := e.do(&e.maker, http.MethodPost, "/api/orgauthorizerequest/"+id.Hex(), nil, "")
		if code != http.StatusForbidden {
			t.Fatalf("maker authorize status=%d want 403", code)
		}
		self := e.createDraft(t, e.dual, e.dual.ID, "TESTACCT00000052", e.usd, "100", "Cash", "")
		e.expectStatus(t, e.dual, http.MethodPost, "/api/orgsendrequest/"+self.Hex(), nil, "", http.StatusOK)
		code, body := e.do(&e.dual, http.MethodPost, "/api/orgauthorizerequest/"+self.Hex(), nil, "")
		if code != http.StatusOK {
			t.Fatalf("dual-role self-authorize status=%d body=%s (expected current code to allow it)", code, body)
		}
		if e.statusOf(t, self) != "Authorized" {
			t.Fatalf("self-authorized status=%s", e.statusOf(t, self))
		}
		var doc bson.M
		if err := e.db.Collection("requests").FindOne(context.Background(), bson.M{"_id": self}).Decode(&doc); err != nil {
			t.Fatal(err)
		}
		if fmt.Sprint(doc["created_by"]) != fmt.Sprint(doc["authorized_by"]) {
			t.Fatalf("expected same actor, created=%v authorized=%v", doc["created_by"], doc["authorized_by"])
		}
	})

	t.Run("assignment_steal_by_branch_maker", func(t *testing.T) {
		id := e.advanceTo(t, "Authorized", "TESTACCT00000061")
		code, body := e.do(&e.maker, http.MethodPost, "/api/assignrequest/"+id.Hex(), nil, "application/json")
		if code == http.StatusOK {
			t.Fatalf("branch maker must not claim Authorized validation work body=%s", body)
		}
		if code != http.StatusForbidden {
			t.Fatalf("branch maker self-assign on Authorized status=%d want 403 body=%s", code, body)
		}
		if lb := e.lockedBy(t, id); lb != nil {
			t.Fatalf("Authorized request must stay unassigned after denied steal, locked_by=%v", lb)
		}
		e.expectStatus(t, e.forex, http.MethodPost, "/api/assignrequest/"+id.Hex(), nil, "application/json", http.StatusOK)
		if ids := e.queueIDs(t, e.forex, "Authorized"); !contains(ids, id) {
			t.Fatalf("validator cannot see request after denied steal")
		}
		valBody := []byte(fmt.Sprintf(`{"validated_average_deposit":1,"validated_current_balance":1,"validated_account_currency_id":"%s"}`, e.etb.Hex()))
		e.expectStatus(t, e.forex, http.MethodPost, "/api/validaterequest/"+id.Hex(), valBody, "application/json", http.StatusOK)
	})

	t.Run("input_validation_and_duplicates", func(t *testing.T) {
		e.createDraftExpect(t, e.maker, e.checker.ID, "SHORT", e.usd, "100", "Cash", "", http.StatusBadRequest)
		e.createDraftExpect(t, e.maker, e.checker.ID, "TESTACCT00000071", e.usd, "0", "Cash", "", http.StatusBadRequest)
		e.createDraftExpect(t, e.maker, e.checker.ID, "TESTACCT00000072", e.etb, "100", "Cash", "", http.StatusBadRequest)
		e.createDraftExpect(t, e.maker, e.checker.ID, "TESTACCT00000073", e.usd, "100", "Cash", "", http.StatusBadRequest, withName("A<B NAME"))
		e.createDraftExpect(t, e.maker, e.otherBA.ID, "TESTACCT00000074", e.usd, "100", "Cash", "", http.StatusBadRequest)
		cardID := e.createDraft(t, e.maker, e.checker.ID, "TESTACCT00000075", e.usd, "100", "Card", "")
		if e.statusOf(t, cardID) != "Drafted" {
			t.Fatalf("card without account was rejected")
		}
		a := e.createDraft(t, e.maker, e.checker.ID, "TESTACCT00000076", e.usd, "100", "Cash", "")
		b := e.createDraft(t, e.maker, e.checker.ID, "TESTACCT00000076", e.usd, "100", "Cash", "")
		if a == b {
			t.Fatal("duplicate creates returned the same id")
		}
		n, err := e.db.Collection("requests").CountDocuments(context.Background(), bson.M{"applicant_account_number": "TESTACCT00000076"})
		if err != nil {
			t.Fatal(err)
		}
		if n != 2 {
			t.Fatalf("duplicate account count=%d want 2", n)
		}

		id := e.advanceTo(t, "Approved", "TESTACCT00000077")
		over := []byte(fmt.Sprintf(`{"accepted_currency_ids":["%s"],"accepted_amounts":[9999],"accepted_amount_in_cash":[9999],"accepted_amount_in_card":[0]}`, e.usd.Hex()))
		e.expectStatus(t, e.proc, http.MethodPost, "/api/acceptrequest/"+id.Hex(), over, "application/json", http.StatusBadRequest)

		// Non-processor branch cannot decline; result-processor branch can.
		id = e.advanceTo(t, "Approved", "TESTACCT00000078")
		e.expectStatus(t, e.outsider, http.MethodPost, "/api/orgdeclinerequest/"+id.Hex(), nil, "", http.StatusForbidden)
		if st := e.statusOf(t, id); st != "Approved" {
			t.Fatalf("outsider decline changed status=%s", st)
		}
		e.expectStatus(t, e.proc, http.MethodPost, "/api/orgdeclinerequest/"+id.Hex(), nil, "", http.StatusOK)
		if st := e.statusOf(t, id); st != "Declined" {
			t.Fatalf("result-processor decline status=%s", st)
		}
	})

	t.Run("draft_delete_and_accepted_report", func(t *testing.T) {
		id := e.createDraft(t, e.maker, e.checker.ID, "TESTACCT00000091", e.usd, "100", "Cash", "")
		e.expectStatus(t, e.maker, http.MethodPatch, "/api/deleterequest/"+id.Hex(), nil, "", http.StatusOK)
		if st := e.statusOf(t, id); st != "Deleted" {
			t.Fatalf("after delete status=%s", st)
		}
		if containsStr(e.actions(t, id), "request.delete") {
			t.Errorf("unexpected delete activity")
		} else {
			t.Log("delete succeeded with no activity-log action")
		}
		code, body := e.do(&e.maker, http.MethodPost, "/api/orgsendrequest/"+id.Hex(), nil, "")
		if code == http.StatusOK {
			t.Fatalf("send after delete succeeded body=%s", body)
		}
		if e.proc.Perms == nil {
			t.Fatal("missing processor")
		}
		var accepted bson.M
		err := e.db.Collection("requests").FindOne(context.Background(), bson.M{"request_status": "Accepted"}).Decode(&accepted)
		if err != nil {
			t.Fatalf("no accepted request to report on: %v", err)
		}
		accID, _ := accepted["_id"].(primitive.ObjectID)
		draft := e.createDraft(t, e.maker, e.checker.ID, "TESTACCT00000092", e.usd, "100", "Cash", "")
		e.expectStatus(t, e.proc, http.MethodGet, "/api/request/"+draft.Hex()+"/report", nil, "", http.StatusConflict)
		code, body = e.do(&e.proc, http.MethodGet, "/api/request/"+accID.Hex()+"/report", nil, "")
		if code != http.StatusOK {
			t.Fatalf("accepted report status=%d body=%s", code, body)
		}
		if !bytes.Contains([]byte(body), []byte("%PDF")) {
			t.Fatalf("report is not a PDF, head=%q", body[:min(40, len(body))])
		}
	})

	t.Run("cross_org_send_error_mapping", func(t *testing.T) {
		id := e.createDraft(t, e.maker, e.checker.ID, "TESTACCT00000081", e.usd, "100", "Cash", "")
		code, body := e.do(&e.proc, http.MethodPost, "/api/orgsendrequest/"+id.Hex(), nil, "")
		if code != http.StatusForbidden && code != http.StatusUnauthorized {
			t.Errorf("cross-branch send status=%d body=%s (unauthorized mapped poorly if 500)", code, body)
		}
	})

	t.Run("processor_authorizer_decline", func(t *testing.T) {
		id := e.advanceTo(t, "Approved", "TESTACCT00000099")
		e.expectStatus(t, e.procAuth, http.MethodPost, "/api/orgdeclinerequest/"+id.Hex(), nil, "", http.StatusOK)
		if st := e.statusOf(t, id); st != "Declined" {
			t.Fatalf("processor authorizer decline status=%s", st)
		}
	})
}
