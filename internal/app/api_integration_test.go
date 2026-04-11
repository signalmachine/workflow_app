package app_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"workflow_app/internal/accounting"
	"workflow_app/internal/ai"
	"workflow_app/internal/app"
	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/inventoryops"
	"workflow_app/internal/parties"
	"workflow_app/internal/reporting"
	"workflow_app/internal/testsupport/dbtest"
	"workflow_app/internal/workflow"
	"workflow_app/internal/workforce"
	"workflow_app/internal/workorders"
)

type inboundRequestMutationTestResponse struct {
	RequestID           string     `json:"request_id"`
	RequestReference    string     `json:"request_reference"`
	Status              string     `json:"status"`
	MessageID           string     `json:"message_id,omitempty"`
	AttachmentIDs       []string   `json:"attachment_ids,omitempty"`
	CancellationReason  string     `json:"cancellation_reason,omitempty"`
	FailureReason       string     `json:"failure_reason,omitempty"`
	ReceivedAt          time.Time  `json:"received_at"`
	QueuedAt            *time.Time `json:"queued_at,omitempty"`
	ProcessingStartedAt *time.Time `json:"processing_started_at,omitempty"`
	ProcessedAt         *time.Time `json:"processed_at,omitempty"`
	ActedOnAt           *time.Time `json:"acted_on_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	FailedAt            *time.Time `json:"failed_at,omitempty"`
	CancelledAt         *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func TestAgentAPISessionLoginCurrentSessionAndLogoutIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	orgSlug, userEmail := loadOrgSlugAndUserEmail(t, ctx, db, orgID, operatorUserID)

	handler := app.NewServedAgentAPIHandler(db)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/session/login", bytes.NewBufferString(`{
		"org_slug":"`+orgSlug+`",
		"email":"`+userEmail+`",
		"password":"`+testLoginPassword+`",
		"device_label":"browser-integration"
	}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, loginReq)

	if loginRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected login status: got %d body=%s", loginRecorder.Code, loginRecorder.Body.String())
	}

	var loginResponse struct {
		SessionID       string `json:"session_id"`
		OrgID           string `json:"org_id"`
		OrgSlug         string `json:"org_slug"`
		UserID          string `json:"user_id"`
		UserEmail       string `json:"user_email"`
		UserDisplayName string `json:"user_display_name"`
		RoleCode        string `json:"role_code"`
		DeviceLabel     string `json:"device_label"`
	}
	if err := json.Unmarshal(loginRecorder.Body.Bytes(), &loginResponse); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginResponse.OrgID != orgID || loginResponse.UserID != operatorUserID {
		t.Fatalf("unexpected login identity: %+v", loginResponse)
	}
	if loginResponse.OrgSlug != orgSlug || loginResponse.UserEmail != userEmail {
		t.Fatalf("unexpected login profile: %+v", loginResponse)
	}
	if loginResponse.RoleCode != identityaccess.RoleOperator || loginResponse.DeviceLabel != "browser-integration" {
		t.Fatalf("unexpected login session metadata: %+v", loginResponse)
	}

	currentReq := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	applyResponseCookies(currentReq, loginRecorder.Result().Cookies())
	currentRecorder := httptest.NewRecorder()
	handler.ServeHTTP(currentRecorder, currentReq)

	if currentRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected current-session status: got %d body=%s", currentRecorder.Code, currentRecorder.Body.String())
	}

	var currentResponse struct {
		SessionID string `json:"session_id"`
		OrgID     string `json:"org_id"`
		UserID    string `json:"user_id"`
	}
	if err := json.Unmarshal(currentRecorder.Body.Bytes(), &currentResponse); err != nil {
		t.Fatalf("decode current-session response: %v", err)
	}
	if currentResponse.SessionID != loginResponse.SessionID || currentResponse.OrgID != orgID || currentResponse.UserID != operatorUserID {
		t.Fatalf("unexpected current session payload: %+v", currentResponse)
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/session/logout", nil)
	applyResponseCookies(logoutReq, loginRecorder.Result().Cookies())
	logoutRecorder := httptest.NewRecorder()
	handler.ServeHTTP(logoutRecorder, logoutReq)

	if logoutRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected logout status: got %d body=%s", logoutRecorder.Code, logoutRecorder.Body.String())
	}

	postLogoutReq := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	applyResponseCookies(postLogoutReq, loginRecorder.Result().Cookies())
	postLogoutRecorder := httptest.NewRecorder()
	handler.ServeHTTP(postLogoutRecorder, postLogoutReq)
	if postLogoutRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected post-logout current-session status: got %d body=%s", postLogoutRecorder.Code, postLogoutRecorder.Body.String())
	}
}

func TestAgentAPIDefaultSvelteFrontendServesPromotedRoutesIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()

	handler := app.NewServedAgentAPIHandler(db)
	routes := []string{
		"/app",
		"/app/login",
		"/app/routes",
		"/app/settings",
		"/app/admin",
		"/app/admin/master-data",
		"/app/admin/lists",
		"/app/admin/accounting",
		"/app/admin/parties",
		"/app/admin/parties/party-123",
		"/app/admin/access",
		"/app/admin/inventory",
		"/app/operations",
		"/app/review",
		"/app/inventory",
		"/app/submit-inbound-request",
		"/app/operations-feed",
		"/app/agent-chat",
		"/app/inbound-requests/REQ-000123",
		"/app/review/inbound-requests",
		"/app/review/approvals",
		"/app/review/approvals/approval-123",
		"/app/review/proposals",
		"/app/review/proposals/recommendation-123",
		"/app/review/documents",
		"/app/review/documents/document-123",
		"/app/review/accounting",
		"/app/review/accounting/journal-entries",
		"/app/review/accounting/control-balances",
		"/app/review/accounting/tax-summaries",
		"/app/review/accounting/trial-balance",
		"/app/review/accounting/balance-sheet",
		"/app/review/accounting/income-statement",
		"/app/review/accounting/entry-123",
		"/app/review/inventory",
		"/app/review/inventory/movement-123",
		"/app/review/work-orders",
		"/app/review/work-orders/work-order-123",
		"/app/review/audit",
		"/app/review/audit/event-123",
	}

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, route, nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("unexpected status for %s: got %d body=%s", route, recorder.Code, recorder.Body.String())
			}

			body := recorder.Body.String()
			requireContains(t, body, `data-sveltekit-preload-data="hover"`)
			requireContains(t, body, `/app/_app/immutable/entry/start.`)
		})
	}
}

func TestAgentAPIDefaultSvelteFrontendServesStaticAssetsAndDoesNotFallbackMissingAssetsIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()

	handler := app.NewServedAgentAPIHandler(db)

	versionReq := httptest.NewRequest(http.MethodGet, "/app/_app/version.json", nil)
	versionRecorder := httptest.NewRecorder()
	handler.ServeHTTP(versionRecorder, versionReq)

	if versionRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected version asset status: got %d body=%s", versionRecorder.Code, versionRecorder.Body.String())
	}
	if got := versionRecorder.Header().Get("Content-Type"); !strings.Contains(got, "application/json") && !strings.Contains(got, "text/plain") {
		t.Fatalf("expected json-ish version asset content type, got %q body=%s", got, versionRecorder.Body.String())
	}
	requireContains(t, versionRecorder.Body.String(), `"version":`)

	missingReq := httptest.NewRequest(http.MethodGet, "/app/_app/immutable/entry/missing.js", nil)
	missingRecorder := httptest.NewRecorder()
	handler.ServeHTTP(missingRecorder, missingReq)

	if missingRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected missing asset to return 404, got %d body=%s", missingRecorder.Code, missingRecorder.Body.String())
	}
	if strings.Contains(strings.ToLower(missingRecorder.Body.String()), "<!doctype html>") {
		t.Fatalf("expected missing asset response to avoid SPA shell fallback, got %s", missingRecorder.Body.String())
	}
}

func TestAgentAPIAdminAccountingMaintenanceIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, adminUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleAdmin)
	_, operatorUserID := seedOrgAndUserInOrg(t, ctx, db, identityaccess.RoleOperator, orgID)

	handler := app.NewServedAgentAPIHandler(db)
	adminCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, adminUserID)
	operatorCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, operatorUserID)

	createLedgerReq := httptest.NewRequest(http.MethodPost, "/api/admin/accounting/ledger-accounts", bytes.NewBufferString(`{
		"code":"AR1000",
		"name":"Accounts Receivable",
		"account_class":"asset",
		"control_type":"receivable",
		"allows_direct_posting":false
	}`))
	createLedgerReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(createLedgerReq, adminCookies)
	createLedgerRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createLedgerRecorder, createLedgerReq)
	if createLedgerRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected create ledger account status: got %d body=%s", createLedgerRecorder.Code, createLedgerRecorder.Body.String())
	}

	var ledgerResponse struct {
		ID   string `json:"id"`
		Code string `json:"code"`
	}
	if err := json.Unmarshal(createLedgerRecorder.Body.Bytes(), &ledgerResponse); err != nil {
		t.Fatalf("decode ledger account response: %v", err)
	}
	if ledgerResponse.Code != "AR1000" || strings.TrimSpace(ledgerResponse.ID) == "" {
		t.Fatalf("unexpected ledger account response: %+v", ledgerResponse)
	}

	createTaxControlReq := httptest.NewRequest(http.MethodPost, "/api/admin/accounting/ledger-accounts", bytes.NewBufferString(`{
		"code":"GST2200",
		"name":"GST Output",
		"account_class":"liability",
		"control_type":"gst_output",
		"allows_direct_posting":false
	}`))
	createTaxControlReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(createTaxControlReq, adminCookies)
	createTaxControlRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createTaxControlRecorder, createTaxControlReq)
	if createTaxControlRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected create tax-control account status: got %d body=%s", createTaxControlRecorder.Code, createTaxControlRecorder.Body.String())
	}

	var taxControlResponse struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createTaxControlRecorder.Body.Bytes(), &taxControlResponse); err != nil {
		t.Fatalf("decode tax-control account response: %v", err)
	}

	listLedgerReq := httptest.NewRequest(http.MethodGet, "/api/admin/accounting/ledger-accounts", nil)
	applyResponseCookies(listLedgerReq, adminCookies)
	listLedgerRecorder := httptest.NewRecorder()
	handler.ServeHTTP(listLedgerRecorder, listLedgerReq)
	if listLedgerRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected list ledger accounts status: got %d body=%s", listLedgerRecorder.Code, listLedgerRecorder.Body.String())
	}
	requireContains(t, listLedgerRecorder.Body.String(), `"code":"AR1000"`)

	operatorLedgerReq := httptest.NewRequest(http.MethodGet, "/api/admin/accounting/ledger-accounts", nil)
	applyResponseCookies(operatorLedgerReq, operatorCookies)
	operatorLedgerRecorder := httptest.NewRecorder()
	handler.ServeHTTP(operatorLedgerRecorder, operatorLedgerReq)
	if operatorLedgerRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected operator ledger-account denial, got %d body=%s", operatorLedgerRecorder.Code, operatorLedgerRecorder.Body.String())
	}

	updateLedgerReq := httptest.NewRequest(http.MethodPost, "/api/admin/accounting/ledger-accounts/"+ledgerResponse.ID+"/status", bytes.NewBufferString(`{
		"status":"inactive"
	}`))
	updateLedgerReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(updateLedgerReq, adminCookies)
	updateLedgerRecorder := httptest.NewRecorder()
	handler.ServeHTTP(updateLedgerRecorder, updateLedgerReq)
	if updateLedgerRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected update ledger account status: got %d body=%s", updateLedgerRecorder.Code, updateLedgerRecorder.Body.String())
	}
	requireContains(t, updateLedgerRecorder.Body.String(), `"status":"inactive"`)

	createTaxReq := httptest.NewRequest(http.MethodPost, "/api/admin/accounting/tax-codes", bytes.NewBufferString(`{
		"code":"GST18",
		"name":"GST 18%",
		"tax_type":"gst",
		"rate_basis_points":1800,
		"payable_account_id":"`+taxControlResponse.ID+`"
	}`))
	createTaxReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(createTaxReq, adminCookies)
	createTaxRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createTaxRecorder, createTaxReq)
	if createTaxRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected create tax code status: got %d body=%s", createTaxRecorder.Code, createTaxRecorder.Body.String())
	}
	requireContains(t, createTaxRecorder.Body.String(), `"code":"GST18"`)
	var taxCodeResponse struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createTaxRecorder.Body.Bytes(), &taxCodeResponse); err != nil {
		t.Fatalf("decode tax code response: %v", err)
	}

	listTaxReq := httptest.NewRequest(http.MethodGet, "/api/admin/accounting/tax-codes", nil)
	applyResponseCookies(listTaxReq, adminCookies)
	listTaxRecorder := httptest.NewRecorder()
	handler.ServeHTTP(listTaxRecorder, listTaxReq)
	if listTaxRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected list tax codes status: got %d body=%s", listTaxRecorder.Code, listTaxRecorder.Body.String())
	}
	requireContains(t, listTaxRecorder.Body.String(), `"code":"GST18"`)

	updateTaxReq := httptest.NewRequest(http.MethodPost, "/api/admin/accounting/tax-codes/"+taxCodeResponse.ID+"/status", bytes.NewBufferString(`{
		"status":"inactive"
	}`))
	updateTaxReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(updateTaxReq, adminCookies)
	updateTaxRecorder := httptest.NewRecorder()
	handler.ServeHTTP(updateTaxRecorder, updateTaxReq)
	if updateTaxRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected update tax code status: got %d body=%s", updateTaxRecorder.Code, updateTaxRecorder.Body.String())
	}
	requireContains(t, updateTaxRecorder.Body.String(), `"status":"inactive"`)

	createPeriodReq := httptest.NewRequest(http.MethodPost, "/api/admin/accounting/periods", bytes.NewBufferString(`{
		"period_code":"FY2026-04",
		"start_on":"2026-04-01",
		"end_on":"2026-04-30"
	}`))
	createPeriodReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(createPeriodReq, adminCookies)
	createPeriodRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createPeriodRecorder, createPeriodReq)
	if createPeriodRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected create accounting period status: got %d body=%s", createPeriodRecorder.Code, createPeriodRecorder.Body.String())
	}

	var periodResponse struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(createPeriodRecorder.Body.Bytes(), &periodResponse); err != nil {
		t.Fatalf("decode accounting period response: %v", err)
	}
	if strings.TrimSpace(periodResponse.ID) == "" || periodResponse.Status != "open" {
		t.Fatalf("unexpected accounting period response: %+v", periodResponse)
	}

	closePeriodReq := httptest.NewRequest(http.MethodPost, "/api/admin/accounting/periods/"+periodResponse.ID+"/close", nil)
	applyResponseCookies(closePeriodReq, adminCookies)
	closePeriodRecorder := httptest.NewRecorder()
	handler.ServeHTTP(closePeriodRecorder, closePeriodReq)
	if closePeriodRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected close accounting period status: got %d body=%s", closePeriodRecorder.Code, closePeriodRecorder.Body.String())
	}
	requireContains(t, closePeriodRecorder.Body.String(), `"status":"closed"`)
}

func TestAgentAPIAdminPartyMaintenanceIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, adminUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleAdmin)
	_, operatorUserID := seedOrgAndUserInOrg(t, ctx, db, identityaccess.RoleOperator, orgID)

	handler := app.NewServedAgentAPIHandler(db)
	adminCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, adminUserID)
	operatorCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, operatorUserID)

	createPartyReq := httptest.NewRequest(http.MethodPost, "/api/admin/parties", bytes.NewBufferString(`{
		"party_code":"CUST-100",
		"display_name":"Northwind Service",
		"legal_name":"Northwind Service Pvt Ltd",
		"party_kind":"customer"
	}`))
	createPartyReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(createPartyReq, adminCookies)
	createPartyRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createPartyRecorder, createPartyReq)
	if createPartyRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected create party status: got %d body=%s", createPartyRecorder.Code, createPartyRecorder.Body.String())
	}

	var partyResponse struct {
		ID        string `json:"id"`
		PartyCode string `json:"party_code"`
		PartyKind string `json:"party_kind"`
	}
	if err := json.Unmarshal(createPartyRecorder.Body.Bytes(), &partyResponse); err != nil {
		t.Fatalf("decode party response: %v", err)
	}
	if strings.TrimSpace(partyResponse.ID) == "" || partyResponse.PartyCode != "CUST-100" || partyResponse.PartyKind != parties.PartyKindCustomer {
		t.Fatalf("unexpected party response: %+v", partyResponse)
	}

	createContactReq := httptest.NewRequest(http.MethodPost, "/api/admin/parties/"+partyResponse.ID+"/contacts", bytes.NewBufferString(`{
		"full_name":"Asha Nair",
		"role_title":"Accounts",
		"email":"asha@example.com",
		"is_primary":true
	}`))
	createContactReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(createContactReq, adminCookies)
	createContactRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createContactRecorder, createContactReq)
	if createContactRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected create contact status: got %d body=%s", createContactRecorder.Code, createContactRecorder.Body.String())
	}
	requireContains(t, createContactRecorder.Body.String(), `"full_name":"Asha Nair"`)

	listReq := httptest.NewRequest(http.MethodGet, "/api/admin/parties?party_kind=customer", nil)
	applyResponseCookies(listReq, adminCookies)
	listRecorder := httptest.NewRecorder()
	handler.ServeHTTP(listRecorder, listReq)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected list parties status: got %d body=%s", listRecorder.Code, listRecorder.Body.String())
	}
	requireContains(t, listRecorder.Body.String(), `"party_code":"CUST-100"`)

	detailReq := httptest.NewRequest(http.MethodGet, "/api/admin/parties/"+partyResponse.ID, nil)
	applyResponseCookies(detailReq, adminCookies)
	detailRecorder := httptest.NewRecorder()
	handler.ServeHTTP(detailRecorder, detailReq)
	if detailRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected party detail status: got %d body=%s", detailRecorder.Code, detailRecorder.Body.String())
	}
	requireContains(t, detailRecorder.Body.String(), `"full_name":"Asha Nair"`)

	operatorReq := httptest.NewRequest(http.MethodGet, "/api/admin/parties", nil)
	applyResponseCookies(operatorReq, operatorCookies)
	operatorRecorder := httptest.NewRecorder()
	handler.ServeHTTP(operatorRecorder, operatorReq)
	if operatorRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected operator party-admin denial, got %d body=%s", operatorRecorder.Code, operatorRecorder.Body.String())
	}

	updatePartyReq := httptest.NewRequest(http.MethodPost, "/api/admin/parties/"+partyResponse.ID+"/status", bytes.NewBufferString(`{
		"status":"inactive"
	}`))
	updatePartyReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(updatePartyReq, adminCookies)
	updatePartyRecorder := httptest.NewRecorder()
	handler.ServeHTTP(updatePartyRecorder, updatePartyReq)
	if updatePartyRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected update party status: got %d body=%s", updatePartyRecorder.Code, updatePartyRecorder.Body.String())
	}
	requireContains(t, updatePartyRecorder.Body.String(), `"status":"inactive"`)
}

func TestAgentAPIAdminAccessMaintenanceIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, adminUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleAdmin)
	_, operatorUserID := seedOrgAndUserInOrg(t, ctx, db, identityaccess.RoleOperator, orgID)

	handler := app.NewServedAgentAPIHandler(db)
	adminCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, adminUserID)
	operatorCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, operatorUserID)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/access/users", bytes.NewBufferString(`{
		"email":"approver@example.com",
		"display_name":"Approver One",
		"role_code":"approver",
		"password":"approver-password-123"
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(createReq, adminCookies)
	createRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createRecorder, createReq)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected create access user status: got %d body=%s", createRecorder.Code, createRecorder.Body.String())
	}

	var created struct {
		MembershipID string `json:"membership_id"`
		UserID       string `json:"user_id"`
		UserEmail    string `json:"user_email"`
		RoleCode     string `json:"role_code"`
	}
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode access create response: %v", err)
	}
	if strings.TrimSpace(created.MembershipID) == "" || strings.TrimSpace(created.UserID) == "" {
		t.Fatalf("expected membership and user identifiers, body=%s", createRecorder.Body.String())
	}
	if created.UserEmail != "approver@example.com" || created.RoleCode != identityaccess.RoleApprover {
		t.Fatalf("unexpected created access record: %+v", created)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/admin/access/users", nil)
	applyResponseCookies(listReq, adminCookies)
	listRecorder := httptest.NewRecorder()
	handler.ServeHTTP(listRecorder, listReq)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected list access users status: got %d body=%s", listRecorder.Code, listRecorder.Body.String())
	}
	requireContains(t, listRecorder.Body.String(), `"user_email":"approver@example.com"`)

	updateReq := httptest.NewRequest(http.MethodPost, "/api/admin/access/users/"+created.MembershipID+"/role", bytes.NewBufferString(`{
		"role_code":"operator"
	}`))
	updateReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(updateReq, adminCookies)
	updateRecorder := httptest.NewRecorder()
	handler.ServeHTTP(updateRecorder, updateReq)
	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected update access role status: got %d body=%s", updateRecorder.Code, updateRecorder.Body.String())
	}
	requireContains(t, updateRecorder.Body.String(), `"role_code":"operator"`)

	operatorReq := httptest.NewRequest(http.MethodGet, "/api/admin/access/users", nil)
	applyResponseCookies(operatorReq, operatorCookies)
	operatorRecorder := httptest.NewRecorder()
	handler.ServeHTTP(operatorRecorder, operatorReq)
	if operatorRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected operator access-admin denial, got %d body=%s", operatorRecorder.Code, operatorRecorder.Body.String())
	}

	selfDemotionReq := httptest.NewRequest(http.MethodPost, "/api/admin/access/users/"+loadMembershipIDForUser(t, ctx, db, orgID, adminUserID)+"/role", bytes.NewBufferString(`{
		"role_code":"operator"
	}`))
	selfDemotionReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(selfDemotionReq, adminCookies)
	selfDemotionRecorder := httptest.NewRecorder()
	handler.ServeHTTP(selfDemotionRecorder, selfDemotionReq)
	if selfDemotionRecorder.Code != http.StatusConflict {
		t.Fatalf("expected protected-admin conflict, got %d body=%s", selfDemotionRecorder.Code, selfDemotionRecorder.Body.String())
	}
}

func TestAgentAPIAdminInventoryMaintenanceIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, adminUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleAdmin)
	_, operatorUserID := seedOrgAndUserInOrg(t, ctx, db, identityaccess.RoleOperator, orgID)

	handler := app.NewServedAgentAPIHandler(db)
	adminCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, adminUserID)
	operatorCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, operatorUserID)

	createItemReq := httptest.NewRequest(http.MethodPost, "/api/admin/inventory/items", bytes.NewBufferString(`{
		"sku":"PUMP-100",
		"name":"Warehouse Pump",
		"item_role":"traceable_equipment",
		"tracking_mode":"serial"
	}`))
	createItemReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(createItemReq, adminCookies)
	createItemRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createItemRecorder, createItemReq)
	if createItemRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected create inventory item status: got %d body=%s", createItemRecorder.Code, createItemRecorder.Body.String())
	}

	var itemResponse struct {
		ID           string `json:"id"`
		SKU          string `json:"sku"`
		ItemRole     string `json:"item_role"`
		TrackingMode string `json:"tracking_mode"`
	}
	if err := json.Unmarshal(createItemRecorder.Body.Bytes(), &itemResponse); err != nil {
		t.Fatalf("decode inventory item response: %v", err)
	}
	if strings.TrimSpace(itemResponse.ID) == "" || itemResponse.SKU != "PUMP-100" || itemResponse.ItemRole != inventoryops.ItemRoleTraceableEquipment {
		t.Fatalf("unexpected inventory item response: %+v", itemResponse)
	}

	createLocationReq := httptest.NewRequest(http.MethodPost, "/api/admin/inventory/locations", bytes.NewBufferString(`{
		"code":"WH-A",
		"name":"Main Warehouse",
		"location_role":"warehouse"
	}`))
	createLocationReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(createLocationReq, adminCookies)
	createLocationRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createLocationRecorder, createLocationReq)
	if createLocationRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected create inventory location status: got %d body=%s", createLocationRecorder.Code, createLocationRecorder.Body.String())
	}
	requireContains(t, createLocationRecorder.Body.String(), `"code":"WH-A"`)

	listItemsReq := httptest.NewRequest(http.MethodGet, "/api/admin/inventory/items?item_role=traceable_equipment", nil)
	applyResponseCookies(listItemsReq, adminCookies)
	listItemsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(listItemsRecorder, listItemsReq)
	if listItemsRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected list inventory items status: got %d body=%s", listItemsRecorder.Code, listItemsRecorder.Body.String())
	}
	requireContains(t, listItemsRecorder.Body.String(), `"sku":"PUMP-100"`)

	listLocationsReq := httptest.NewRequest(http.MethodGet, "/api/admin/inventory/locations?location_role=warehouse", nil)
	applyResponseCookies(listLocationsReq, adminCookies)
	listLocationsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(listLocationsRecorder, listLocationsReq)
	if listLocationsRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected list inventory locations status: got %d body=%s", listLocationsRecorder.Code, listLocationsRecorder.Body.String())
	}
	requireContains(t, listLocationsRecorder.Body.String(), `"code":"WH-A"`)

	operatorReq := httptest.NewRequest(http.MethodGet, "/api/admin/inventory/items", nil)
	applyResponseCookies(operatorReq, operatorCookies)
	operatorRecorder := httptest.NewRecorder()
	handler.ServeHTTP(operatorRecorder, operatorReq)
	if operatorRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected operator inventory-admin denial, got %d body=%s", operatorRecorder.Code, operatorRecorder.Body.String())
	}

	updateItemReq := httptest.NewRequest(http.MethodPost, "/api/admin/inventory/items/"+itemResponse.ID+"/status", bytes.NewBufferString(`{
		"status":"inactive"
	}`))
	updateItemReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(updateItemReq, adminCookies)
	updateItemRecorder := httptest.NewRecorder()
	handler.ServeHTTP(updateItemRecorder, updateItemReq)
	if updateItemRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected update inventory item status: got %d body=%s", updateItemRecorder.Code, updateItemRecorder.Body.String())
	}
	requireContains(t, updateItemRecorder.Body.String(), `"status":"inactive"`)

	var locationResponse struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createLocationRecorder.Body.Bytes(), &locationResponse); err != nil {
		t.Fatalf("decode inventory location response: %v", err)
	}

	updateLocationReq := httptest.NewRequest(http.MethodPost, "/api/admin/inventory/locations/"+locationResponse.ID+"/status", bytes.NewBufferString(`{
		"status":"inactive"
	}`))
	updateLocationReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(updateLocationReq, adminCookies)
	updateLocationRecorder := httptest.NewRecorder()
	handler.ServeHTTP(updateLocationRecorder, updateLocationReq)
	if updateLocationRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected update inventory location status: got %d body=%s", updateLocationRecorder.Code, updateLocationRecorder.Body.String())
	}
	requireContains(t, updateLocationRecorder.Body.String(), `"status":"inactive"`)
}

func TestAgentAPITokenSessionIssueRefreshAndRevokeIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	orgSlug, userEmail := loadOrgSlugAndUserEmail(t, ctx, db, orgID, operatorUserID)

	handler := app.NewServedAgentAPIHandler(db)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/session/token", bytes.NewBufferString(`{
		"org_slug":"`+orgSlug+`",
		"email":"`+userEmail+`",
		"password":"`+testLoginPassword+`",
		"device_label":"mobile-integration"
	}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, loginReq)

	if loginRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected token login status: got %d body=%s", loginRecorder.Code, loginRecorder.Body.String())
	}
	if len(loginRecorder.Result().Cookies()) != 0 {
		t.Fatalf("expected token login to avoid cookies, got %d", len(loginRecorder.Result().Cookies()))
	}

	var loginResponse struct {
		SessionID             string    `json:"session_id"`
		OrgID                 string    `json:"org_id"`
		UserID                string    `json:"user_id"`
		DeviceLabel           string    `json:"device_label"`
		AccessToken           string    `json:"access_token"`
		AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
		RefreshToken          string    `json:"refresh_token"`
		RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	}
	if err := json.Unmarshal(loginRecorder.Body.Bytes(), &loginResponse); err != nil {
		t.Fatalf("decode token login response: %v", err)
	}
	if loginResponse.OrgID != orgID || loginResponse.UserID != operatorUserID {
		t.Fatalf("unexpected token login identity: %+v", loginResponse)
	}
	if loginResponse.DeviceLabel != "mobile-integration" || loginResponse.AccessToken == "" || loginResponse.RefreshToken == "" {
		t.Fatalf("unexpected token login credentials: %+v", loginResponse)
	}
	if !loginResponse.AccessTokenExpiresAt.Before(loginResponse.RefreshTokenExpiresAt) {
		t.Fatalf("expected shorter access token expiry: %+v", loginResponse)
	}

	currentReq := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	applyBearer(currentReq, loginResponse.AccessToken)
	currentRecorder := httptest.NewRecorder()
	handler.ServeHTTP(currentRecorder, currentReq)
	if currentRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected bearer current-session status: got %d body=%s", currentRecorder.Code, currentRecorder.Body.String())
	}

	body := bytes.NewBufferString(`{
		"origin_type":"human",
		"channel":"mobile",
		"metadata":{"submitter_label":"field app"},
		"message":{"message_role":"request","text_content":"Submit this request with bearer auth."}
	}`)
	submitReq := httptest.NewRequest(http.MethodPost, "/api/inbound-requests", body)
	submitReq.Header.Set("Content-Type", "application/json")
	applyBearer(submitReq, loginResponse.AccessToken)
	submitRecorder := httptest.NewRecorder()
	handler.ServeHTTP(submitRecorder, submitReq)
	if submitRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected bearer submit status: got %d body=%s", submitRecorder.Code, submitRecorder.Body.String())
	}

	var submitResponse struct {
		RequestID string `json:"request_id"`
		Status    string `json:"status"`
	}
	if err := json.Unmarshal(submitRecorder.Body.Bytes(), &submitResponse); err != nil {
		t.Fatalf("decode bearer submit response: %v", err)
	}
	if submitResponse.Status != "queued" || submitResponse.RequestID == "" {
		t.Fatalf("unexpected bearer submit response: %+v", submitResponse)
	}

	var storedSessionID sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT session_id FROM ai.inbound_requests WHERE id = $1`, submitResponse.RequestID).Scan(&storedSessionID); err != nil {
		t.Fatalf("load submitted request session: %v", err)
	}
	if !storedSessionID.Valid || storedSessionID.String != loginResponse.SessionID {
		t.Fatalf("unexpected submitted session linkage: %+v want %s", storedSessionID, loginResponse.SessionID)
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/api/session/refresh", bytes.NewBufferString(`{
		"session_id":"`+loginResponse.SessionID+`",
		"refresh_token":"`+loginResponse.RefreshToken+`"
	}`))
	refreshReq.Header.Set("Content-Type", "application/json")
	refreshRecorder := httptest.NewRecorder()
	handler.ServeHTTP(refreshRecorder, refreshReq)
	if refreshRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected refresh status: got %d body=%s", refreshRecorder.Code, refreshRecorder.Body.String())
	}

	var refreshResponse struct {
		SessionID            string    `json:"session_id"`
		AccessToken          string    `json:"access_token"`
		AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
		RefreshToken         string    `json:"refresh_token"`
	}
	if err := json.Unmarshal(refreshRecorder.Body.Bytes(), &refreshResponse); err != nil {
		t.Fatalf("decode refresh response: %v", err)
	}
	if refreshResponse.SessionID != loginResponse.SessionID {
		t.Fatalf("unexpected refreshed session id: %+v", refreshResponse)
	}
	if refreshResponse.AccessToken == "" || refreshResponse.RefreshToken == "" {
		t.Fatalf("missing refreshed credentials: %+v", refreshResponse)
	}
	if refreshResponse.AccessToken == loginResponse.AccessToken || refreshResponse.RefreshToken == loginResponse.RefreshToken {
		t.Fatalf("expected rotated credentials, got %+v", refreshResponse)
	}

	oldCurrentReq := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	applyBearer(oldCurrentReq, loginResponse.AccessToken)
	oldCurrentRecorder := httptest.NewRecorder()
	handler.ServeHTTP(oldCurrentRecorder, oldCurrentReq)
	if oldCurrentRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected rotated access token to fail, got %d body=%s", oldCurrentRecorder.Code, oldCurrentRecorder.Body.String())
	}

	refreshedCurrentReq := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	applyBearer(refreshedCurrentReq, refreshResponse.AccessToken)
	refreshedCurrentRecorder := httptest.NewRecorder()
	handler.ServeHTTP(refreshedCurrentRecorder, refreshedCurrentReq)
	if refreshedCurrentRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected refreshed bearer current-session status: got %d body=%s", refreshedCurrentRecorder.Code, refreshedCurrentRecorder.Body.String())
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/session/logout", nil)
	applyBearer(logoutReq, refreshResponse.AccessToken)
	logoutRecorder := httptest.NewRecorder()
	handler.ServeHTTP(logoutRecorder, logoutReq)
	if logoutRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected bearer logout status: got %d body=%s", logoutRecorder.Code, logoutRecorder.Body.String())
	}

	postLogoutReq := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	applyBearer(postLogoutReq, refreshResponse.AccessToken)
	postLogoutRecorder := httptest.NewRecorder()
	handler.ServeHTTP(postLogoutRecorder, postLogoutReq)
	if postLogoutRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected post-logout bearer status: got %d body=%s", postLogoutRecorder.Code, postLogoutRecorder.Body.String())
	}
}

func TestAgentAPISubmitInboundRequestWithBrowserSessionCookies(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	orgSlug, userEmail := loadOrgSlugAndUserEmail(t, ctx, db, orgID, operatorUserID)

	handler := app.NewServedAgentAPIHandler(db)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/session/login", bytes.NewBufferString(`{
		"org_slug":"`+orgSlug+`",
		"email":"`+userEmail+`",
		"password":"`+testLoginPassword+`"
	}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, loginReq)
	if loginRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected login status: got %d body=%s", loginRecorder.Code, loginRecorder.Body.String())
	}

	body := bytes.NewBufferString(`{
		"origin_type":"human",
		"channel":"browser",
		"metadata":{"submitter_label":"front desk"},
		"message":{"message_role":"request","text_content":"Submit this request with cookie auth."}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/inbound-requests", body)
	req.Header.Set("Content-Type", "application/json")
	applyResponseCookies(req, loginRecorder.Result().Cookies())

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		RequestID string `json:"request_id"`
		Status    string `json:"status"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode submit response: %v", err)
	}
	if response.RequestID == "" || response.Status != "queued" {
		t.Fatalf("unexpected submit response: %+v", response)
	}

	var storedSessionID sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT session_id FROM ai.inbound_requests WHERE id = $1`, response.RequestID).Scan(&storedSessionID); err != nil {
		t.Fatalf("load submitted request session: %v", err)
	}
	if !storedSessionID.Valid {
		t.Fatal("expected persisted browser session linkage")
	}
}

func TestAgentAPIInboundRequestCancelRejectsInvalidJSONIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)

	handler := app.NewAgentAPIHandlerWithServices(nil, app.NewSubmissionService(db))

	service := app.NewSubmissionService(db)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}
	draft, err := service.SaveInboundDraft(ctx, app.SaveInboundDraftInput{
		OriginType:  intake.OriginHuman,
		Channel:     "browser",
		MessageRole: intake.MessageRoleRequest,
		MessageText: "Draft request that will be queued.",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("save draft: %v", err)
	}
	if _, err := service.QueueInboundRequest(ctx, app.QueueInboundRequestInput{
		RequestID: draft.Request.ID,
		Actor:     operator,
	}); err != nil {
		t.Fatalf("queue draft: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/inbound-requests/"+draft.Request.ID+"/cancel", bytes.NewBufferString(`{"reason":"bad"`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", session.ID)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected invalid-cancel status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM ai.inbound_requests WHERE id = $1`, draft.Request.ID).Scan(&status); err != nil {
		t.Fatalf("load request status: %v", err)
	}
	if status != intake.StatusQueued {
		t.Fatalf("expected invalid cancel payload to preserve queued status, got %s", status)
	}
}

func TestAgentAPIInboundRequestLifecycleActionsIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	otherOrgID, otherOperatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	handler := app.NewServedAgentAPIHandler(db)
	operatorCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, operatorUserID)
	otherOperatorCookies := issueBrowserSessionCookies(t, ctx, db, handler, otherOrgID, otherOperatorUserID)

	draftReq := httptest.NewRequest(http.MethodPost, "/api/inbound-requests", bytes.NewBufferString(`{
		"origin_type":"human",
		"channel":"browser",
		"metadata":{"submitter_label":"front desk"},
		"message":{"message_role":"request","text_content":"Initial draft details"},
		"queue_for_review":false
	}`))
	draftReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(draftReq, operatorCookies)

	draftRecorder := httptest.NewRecorder()
	handler.ServeHTTP(draftRecorder, draftReq)
	if draftRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected draft-create status: got %d body=%s", draftRecorder.Code, draftRecorder.Body.String())
	}

	var draftResponse inboundRequestMutationTestResponse
	if err := json.Unmarshal(draftRecorder.Body.Bytes(), &draftResponse); err != nil {
		t.Fatalf("decode draft-create response: %v", err)
	}
	if draftResponse.RequestID == "" || draftResponse.MessageID == "" || draftResponse.Status != intake.StatusDraft {
		t.Fatalf("unexpected draft-create response: %+v", draftResponse)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/api/inbound-requests/"+draftResponse.RequestID+"/draft", bytes.NewBufferString(`{
		"origin_type":"human",
		"channel":"browser",
		"metadata":{"submitter_label":"front desk","operator_note":"ready for queue"},
		"message_id":"`+draftResponse.MessageID+`",
		"message":{"message_role":"request","text_content":"Updated draft details for queue"}
	}`))
	updateReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(updateReq, operatorCookies)

	updateRecorder := httptest.NewRecorder()
	handler.ServeHTTP(updateRecorder, updateReq)
	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected draft-update status: got %d body=%s", updateRecorder.Code, updateRecorder.Body.String())
	}

	var updateResponse inboundRequestMutationTestResponse
	if err := json.Unmarshal(updateRecorder.Body.Bytes(), &updateResponse); err != nil {
		t.Fatalf("decode draft-update response: %v", err)
	}
	if updateResponse.RequestID != draftResponse.RequestID || updateResponse.MessageID != draftResponse.MessageID || updateResponse.Status != intake.StatusDraft {
		t.Fatalf("unexpected draft-update response: %+v", updateResponse)
	}

	queueReq := httptest.NewRequest(http.MethodPost, "/api/inbound-requests/"+draftResponse.RequestID+"/queue", nil)
	applyResponseCookies(queueReq, operatorCookies)
	queueRecorder := httptest.NewRecorder()
	handler.ServeHTTP(queueRecorder, queueReq)
	if queueRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected queue status: got %d body=%s", queueRecorder.Code, queueRecorder.Body.String())
	}

	var queueResponse inboundRequestMutationTestResponse
	if err := json.Unmarshal(queueRecorder.Body.Bytes(), &queueResponse); err != nil {
		t.Fatalf("decode queue response: %v", err)
	}
	if queueResponse.Status != intake.StatusQueued || queueResponse.QueuedAt == nil {
		t.Fatalf("unexpected queue response: %+v", queueResponse)
	}

	crossOrgCancelReq := httptest.NewRequest(http.MethodPost, "/api/inbound-requests/"+draftResponse.RequestID+"/cancel", bytes.NewBufferString(`{"reason":"wrong org should not mutate"}`))
	crossOrgCancelReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(crossOrgCancelReq, otherOperatorCookies)
	crossOrgCancelRecorder := httptest.NewRecorder()
	handler.ServeHTTP(crossOrgCancelRecorder, crossOrgCancelReq)
	if crossOrgCancelRecorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected cross-org cancel status: got %d body=%s", crossOrgCancelRecorder.Code, crossOrgCancelRecorder.Body.String())
	}

	var statusAfterCrossOrgAttempt string
	if err := db.QueryRowContext(ctx, `SELECT status FROM ai.inbound_requests WHERE id = $1`, draftResponse.RequestID).Scan(&statusAfterCrossOrgAttempt); err != nil {
		t.Fatalf("load request after cross-org cancel: %v", err)
	}
	if statusAfterCrossOrgAttempt != intake.StatusQueued {
		t.Fatalf("expected cross-org cancel to preserve queued status, got %s", statusAfterCrossOrgAttempt)
	}

	cancelReq := httptest.NewRequest(http.MethodPost, "/api/inbound-requests/"+draftResponse.RequestID+"/cancel", bytes.NewBufferString(`{"reason":"operator paused request"}`))
	cancelReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(cancelReq, operatorCookies)
	cancelRecorder := httptest.NewRecorder()
	handler.ServeHTTP(cancelRecorder, cancelReq)
	if cancelRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected cancel status: got %d body=%s", cancelRecorder.Code, cancelRecorder.Body.String())
	}

	var cancelResponse inboundRequestMutationTestResponse
	if err := json.Unmarshal(cancelRecorder.Body.Bytes(), &cancelResponse); err != nil {
		t.Fatalf("decode cancel response: %v", err)
	}
	if cancelResponse.Status != intake.StatusCancelled || cancelResponse.CancellationReason != "operator paused request" || cancelResponse.CancelledAt == nil {
		t.Fatalf("unexpected cancel response: %+v", cancelResponse)
	}

	amendReq := httptest.NewRequest(http.MethodPost, "/api/inbound-requests/"+draftResponse.RequestID+"/amend", nil)
	applyResponseCookies(amendReq, operatorCookies)
	amendRecorder := httptest.NewRecorder()
	handler.ServeHTTP(amendRecorder, amendReq)
	if amendRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected amend status: got %d body=%s", amendRecorder.Code, amendRecorder.Body.String())
	}

	var amendResponse inboundRequestMutationTestResponse
	if err := json.Unmarshal(amendRecorder.Body.Bytes(), &amendResponse); err != nil {
		t.Fatalf("decode amend response: %v", err)
	}
	if amendResponse.Status != intake.StatusDraft || amendResponse.QueuedAt != nil || amendResponse.CancelledAt != nil || amendResponse.CancellationReason != "" {
		t.Fatalf("unexpected amend response: %+v", amendResponse)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/inbound-requests/"+draftResponse.RequestID+"/delete", nil)
	applyResponseCookies(deleteReq, operatorCookies)
	deleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deleteRecorder, deleteReq)
	if deleteRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected delete status: got %d body=%s", deleteRecorder.Code, deleteRecorder.Body.String())
	}

	var deleteResponse struct {
		Deleted bool `json:"deleted"`
	}
	if err := json.Unmarshal(deleteRecorder.Body.Bytes(), &deleteResponse); err != nil {
		t.Fatalf("decode delete response: %v", err)
	}
	if !deleteResponse.Deleted {
		t.Fatalf("unexpected delete response: %+v", deleteResponse)
	}

	var remaining int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai.inbound_requests WHERE id = $1`, draftResponse.RequestID).Scan(&remaining); err != nil {
		t.Fatalf("count deleted request: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected deleted draft request to be removed, found %d", remaining)
	}
}

func TestAgentAPISubmitInboundRequestRejectsUnknownJSONFieldsIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)

	handler := app.NewAgentAPIHandlerWithServices(nil, app.NewSubmissionService(db))

	req := httptest.NewRequest(http.MethodPost, "/api/inbound-requests", bytes.NewBufferString(`{
		"channel":"browser",
		"message":{"text_content":"Unknown field should fail."},
		"unexpected":"value"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", session.ID)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected unknown-field status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var requestCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai.inbound_requests WHERE org_id = $1`, orgID).Scan(&requestCount); err != nil {
		t.Fatalf("count inbound requests: %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("expected unknown-field request rejection before persistence, found %d requests", requestCount)
	}
}

func TestAgentAPIProcessNextFailureExposesReviewContinuityIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}
	request := createQueuedRequest(t, ctx, db, operator, "Provider failure should remain review-visible.")

	handler := app.NewServedAgentAPIHandlerWithDependencies(func() (app.ProcessNextQueuedInboundRequester, error) {
		return app.NewAgentProcessor(db, fakeCoordinatorProvider{err: errors.New("upstream provider timeout")})
	}, app.NewSubmissionService(db), reporting.NewService(db), nil, identityaccess.NewService(db))
	cookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, operatorUserID)

	processReq := httptest.NewRequest(http.MethodPost, "/api/agent/process-next-queued-inbound-request", bytes.NewBufferString(`{"channel":"browser"}`))
	processReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(processReq, cookies)
	processRecorder := httptest.NewRecorder()
	handler.ServeHTTP(processRecorder, processReq)
	if processRecorder.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected process failure status: got %d body=%s", processRecorder.Code, processRecorder.Body.String())
	}

	var processResponse struct {
		Error            string `json:"error"`
		RequestReference string `json:"request_reference"`
		RunID            string `json:"run_id"`
	}
	if err := json.Unmarshal(processRecorder.Body.Bytes(), &processResponse); err != nil {
		t.Fatalf("decode process failure response: %v", err)
	}
	if processResponse.Error != "failed to process queued inbound request" || processResponse.RequestReference != request.RequestReference || processResponse.RunID == "" {
		t.Fatalf("unexpected process failure response: %+v", processResponse)
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/review/inbound-requests/"+request.RequestReference, nil)
	applyResponseCookies(detailReq, cookies)
	detailRecorder := httptest.NewRecorder()
	handler.ServeHTTP(detailRecorder, detailReq)
	if detailRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected failed request detail status: got %d body=%s", detailRecorder.Code, detailRecorder.Body.String())
	}

	var detailResponse struct {
		Request struct {
			RequestReference string  `json:"request_reference"`
			Status           string  `json:"status"`
			FailureReason    string  `json:"failure_reason"`
			FailedAt         *string `json:"failed_at"`
			LastRunID        *string `json:"last_run_id"`
			LastRunStatus    *string `json:"last_run_status"`
		} `json:"request"`
		Runs []struct {
			RunID   string `json:"run_id"`
			Status  string `json:"status"`
			Summary string `json:"summary"`
		} `json:"runs"`
		Steps []struct {
			RunID         string          `json:"run_id"`
			StepTitle     string          `json:"step_title"`
			Status        string          `json:"status"`
			OutputPayload json.RawMessage `json:"output_payload"`
		} `json:"steps"`
	}
	if err := json.Unmarshal(detailRecorder.Body.Bytes(), &detailResponse); err != nil {
		t.Fatalf("decode failed request detail response: %v", err)
	}
	if detailResponse.Request.RequestReference != request.RequestReference || detailResponse.Request.Status != intake.StatusFailed || detailResponse.Request.FailureReason != "upstream provider timeout" || detailResponse.Request.FailedAt == nil {
		t.Fatalf("unexpected failed request review state: %+v", detailResponse.Request)
	}
	if detailResponse.Request.LastRunID == nil || *detailResponse.Request.LastRunID != processResponse.RunID || detailResponse.Request.LastRunStatus == nil || *detailResponse.Request.LastRunStatus != ai.RunStatusFailed {
		t.Fatalf("unexpected failed request latest run linkage: %+v", detailResponse.Request)
	}
	if len(detailResponse.Runs) != 1 || detailResponse.Runs[0].RunID != processResponse.RunID || detailResponse.Runs[0].Status != ai.RunStatusFailed || detailResponse.Runs[0].Summary != "upstream provider timeout" {
		t.Fatalf("unexpected failed run review state: %+v", detailResponse.Runs)
	}
	if len(detailResponse.Steps) != 1 || detailResponse.Steps[0].RunID != processResponse.RunID || detailResponse.Steps[0].Status != ai.StepStatusFailed || detailResponse.Steps[0].StepTitle != "Provider execution failed" {
		t.Fatalf("unexpected failed step review state: %+v", detailResponse.Steps)
	}
	requireContains(t, string(detailResponse.Steps[0].OutputPayload), `"error":"upstream provider timeout"`)
}

func TestAgentAPIReviewSurfacesIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	request := createQueuedRequest(t, ctx, db, operator, "Urgent pump issue reported from the warehouse.")

	processor, err := app.NewAgentProcessor(db, fakeCoordinatorProvider{
		output: ai.CoordinatorProviderOutput{
			ProviderName:       "openai",
			ProviderResponseID: "resp_review_api_test_123",
			Model:              "gpt-5.2",
			Summary:            "Operator review is required for the urgent pump issue.",
			Priority:           "urgent",
			ArtifactTitle:      "Inbound request review brief",
			ArtifactBody:       "The request describes an urgent equipment problem that should be reviewed immediately.",
			Rationale: []string{
				"The request describes a time-sensitive equipment failure.",
			},
			NextActions: []string{
				"Confirm the site details and route controlled follow-up.",
			},
		},
	})
	if err != nil {
		t.Fatalf("new agent processor: %v", err)
	}

	processResult, err := processor.ProcessNextQueuedInboundRequest(ctx, app.ProcessNextQueuedInboundRequestInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("process next queued inbound request: %v", err)
	}

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)
	aiService := ai.NewService(db)

	approval, _ := createPendingApproval(t, ctx, documentService, workflowService, operator)
	if _, err := aiService.LinkRecommendationApproval(ctx, ai.LinkRecommendationApprovalInput{
		RecommendationID: processResult.Recommendation.ID,
		ApprovalID:       approval.ID,
		Actor:            operator,
	}); err != nil {
		t.Fatalf("link recommendation approval: %v", err)
	}

	handler := app.NewServedAgentAPIHandler(db)
	cookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, operatorUserID)
	req := httptest.NewRequest(http.MethodGet, "/api/review/inbound-requests?status=processed", nil)
	applyResponseCookies(req, cookies)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected inbound request list status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var listResponse struct {
		Items []struct {
			RequestReference         string `json:"request_reference"`
			Status                   string `json:"status"`
			LastRecommendationStatus string `json:"last_recommendation_status"`
		} `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &listResponse); err != nil {
		t.Fatalf("decode inbound request list: %v", err)
	}
	if len(listResponse.Items) != 1 {
		t.Fatalf("unexpected inbound request item count: %d", len(listResponse.Items))
	}
	if listResponse.Items[0].RequestReference != request.RequestReference || listResponse.Items[0].Status != intake.StatusProcessed {
		t.Fatalf("unexpected inbound request list item: %+v", listResponse.Items[0])
	}
	if listResponse.Items[0].LastRecommendationStatus != ai.RecommendationStatusApprovalRequested {
		t.Fatalf("unexpected recommendation status: %+v", listResponse.Items[0])
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/inbound-request-status-summary", nil)
	applyResponseCookies(req, cookies)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected inbound request summary status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var summaryResponse struct {
		Items []struct {
			Status       string `json:"status"`
			RequestCount int    `json:"request_count"`
		} `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &summaryResponse); err != nil {
		t.Fatalf("decode inbound request summary: %v", err)
	}
	if len(summaryResponse.Items) == 0 || summaryResponse.Items[0].Status != intake.StatusProcessed {
		t.Fatalf("unexpected inbound request summary items: %+v", summaryResponse.Items)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/inbound-requests/"+request.RequestReference, nil)
	applyResponseCookies(req, cookies)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected inbound request detail status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var detailResponse struct {
		Request struct {
			RequestReference string `json:"request_reference"`
		} `json:"request"`
		Runs            []struct{} `json:"runs"`
		Recommendations []struct {
			ApprovalID *string `json:"approval_id"`
		} `json:"recommendations"`
		Proposals []struct {
			ApprovalID *string `json:"approval_id"`
		} `json:"proposals"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &detailResponse); err != nil {
		t.Fatalf("decode inbound request detail: %v", err)
	}
	if detailResponse.Request.RequestReference != request.RequestReference {
		t.Fatalf("unexpected inbound request detail reference: %+v", detailResponse.Request)
	}
	if len(detailResponse.Runs) == 0 || len(detailResponse.Recommendations) == 0 || len(detailResponse.Proposals) == 0 {
		t.Fatalf("expected review detail slices, got %+v", detailResponse)
	}
	if detailResponse.Recommendations[0].ApprovalID == nil || *detailResponse.Recommendations[0].ApprovalID != approval.ID {
		t.Fatalf("unexpected recommendation approval linkage: %+v", detailResponse.Recommendations[0])
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/processed-proposals?request_reference="+request.RequestReference, nil)
	applyResponseCookies(req, cookies)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected processed proposal list status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var proposalListResponse struct {
		Items []struct {
			RecommendationID     string  `json:"recommendation_id"`
			RecommendationStatus string  `json:"recommendation_status"`
			ApprovalID           *string `json:"approval_id"`
			ApprovalStatus       *string `json:"approval_status"`
		} `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &proposalListResponse); err != nil {
		t.Fatalf("decode processed proposal list: %v", err)
	}
	if len(proposalListResponse.Items) != 1 {
		t.Fatalf("unexpected processed proposal count: %d", len(proposalListResponse.Items))
	}
	if proposalListResponse.Items[0].ApprovalID == nil || *proposalListResponse.Items[0].ApprovalID != approval.ID {
		t.Fatalf("unexpected processed proposal approval linkage: %+v", proposalListResponse.Items[0])
	}
	if proposalListResponse.Items[0].RecommendationStatus != ai.RecommendationStatusApprovalRequested {
		t.Fatalf("unexpected processed proposal status: %+v", proposalListResponse.Items[0])
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/processed-proposals?recommendation_id="+proposalListResponse.Items[0].RecommendationID, nil)
	applyResponseCookies(req, cookies)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact processed proposal list status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &proposalListResponse); err != nil {
		t.Fatalf("decode exact processed proposal list: %v", err)
	}
	if len(proposalListResponse.Items) != 1 || proposalListResponse.Items[0].RecommendationID == "" {
		t.Fatalf("unexpected exact processed proposal items: %+v", proposalListResponse.Items)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/processed-proposal-status-summary", nil)
	applyResponseCookies(req, cookies)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected processed proposal summary status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var proposalSummaryResponse struct {
		Items []struct {
			RecommendationStatus string `json:"recommendation_status"`
			ProposalCount        int    `json:"proposal_count"`
		} `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &proposalSummaryResponse); err != nil {
		t.Fatalf("decode processed proposal summary: %v", err)
	}
	if len(proposalSummaryResponse.Items) == 0 || proposalSummaryResponse.Items[0].RecommendationStatus != ai.RecommendationStatusApprovalRequested {
		t.Fatalf("unexpected processed proposal summary items: %+v", proposalSummaryResponse.Items)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/approval-queue?status=pending", nil)
	applyResponseCookies(req, cookies)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected approval queue status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var queueResponse struct {
		Items []struct {
			ApprovalID     string `json:"approval_id"`
			ApprovalStatus string `json:"approval_status"`
			QueueCode      string `json:"queue_code"`
		} `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &queueResponse); err != nil {
		t.Fatalf("decode approval queue: %v", err)
	}
	if len(queueResponse.Items) != 1 || queueResponse.Items[0].ApprovalID != approval.ID || queueResponse.Items[0].ApprovalStatus != "pending" {
		t.Fatalf("unexpected approval queue items: %+v", queueResponse.Items)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/approval-queue?approval_id="+approval.ID, nil)
	applyResponseCookies(req, cookies)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact approval queue status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &queueResponse); err != nil {
		t.Fatalf("decode exact approval queue: %v", err)
	}
	if len(queueResponse.Items) != 1 || queueResponse.Items[0].ApprovalID != approval.ID {
		t.Fatalf("unexpected exact approval queue items: %+v", queueResponse.Items)
	}

}

func TestAgentAPIRequestProcessedProposalApprovalIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	intakeService := intake.NewService(db)
	request := createQueuedRequest(t, ctx, db, operator, "Request approval for the submitted invoice proposal.")
	if _, err := intakeService.ClaimNextQueued(ctx, intake.ClaimNextQueuedInput{
		Channel: "browser",
		Actor:   operator,
	}); err != nil {
		t.Fatalf("claim queued request: %v", err)
	}
	if _, err := intakeService.AdvanceRequest(ctx, intake.AdvanceRequestInput{
		RequestID: request.ID,
		Status:    intake.StatusProcessed,
		Actor:     operator,
	}); err != nil {
		t.Fatalf("mark request processed: %v", err)
	}

	documentService := documents.NewService(db)
	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: "invoice",
		Title:    "Proposal-backed invoice",
		Actor:    operator,
	})
	if err != nil {
		t.Fatalf("create document draft: %v", err)
	}
	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("submit document: %v", err)
	}

	aiService := ai.NewService(db)
	run, err := aiService.StartRun(ctx, ai.StartRunInput{
		AgentRole:        ai.RunRoleSpecialist,
		CapabilityCode:   "workflow.approvals",
		InboundRequestID: request.ID,
		RequestText:      "request approval for submitted invoice proposal",
		Metadata: map[string]any{
			"request_reference": request.RequestReference,
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("start run: %v", err)
	}
	recommendation, err := aiService.CreateRecommendation(ctx, ai.CreateRecommendationInput{
		RunID:              run.ID,
		RecommendationType: "request_approval",
		Summary:            "Request finance approval for the submitted invoice.",
		Payload: map[string]any{
			"document_id": doc.ID,
			"queue_code":  "finance-review",
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("create recommendation: %v", err)
	}

	handler := app.NewServedAgentAPIHandler(db)
	cookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, operatorUserID)

	preReq := httptest.NewRequest(http.MethodGet, "/api/review/processed-proposals?recommendation_id="+recommendation.ID, nil)
	applyResponseCookies(preReq, cookies)
	preRecorder := httptest.NewRecorder()
	handler.ServeHTTP(preRecorder, preReq)
	if preRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected pre-action proposal status: got %d body=%s", preRecorder.Code, preRecorder.Body.String())
	}

	var preResponse struct {
		Items []struct {
			RecommendationID   string  `json:"recommendation_id"`
			SuggestedQueueCode *string `json:"suggested_queue_code"`
			DocumentID         *string `json:"document_id"`
			ApprovalID         *string `json:"approval_id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(preRecorder.Body.Bytes(), &preResponse); err != nil {
		t.Fatalf("decode pre-action proposal review: %v", err)
	}
	if len(preResponse.Items) != 1 {
		t.Fatalf("unexpected pre-action proposal count: %d", len(preResponse.Items))
	}
	if preResponse.Items[0].DocumentID == nil || *preResponse.Items[0].DocumentID != doc.ID {
		t.Fatalf("unexpected pre-action proposal document: %+v", preResponse.Items[0])
	}
	if preResponse.Items[0].SuggestedQueueCode == nil || *preResponse.Items[0].SuggestedQueueCode != "finance-review" {
		t.Fatalf("unexpected suggested queue: %+v", preResponse.Items[0])
	}
	if preResponse.Items[0].ApprovalID != nil {
		t.Fatalf("expected no approval before action: %+v", preResponse.Items[0])
	}

	req := httptest.NewRequest(http.MethodPost, "/api/review/processed-proposals/"+recommendation.ID+"/request-approval", bytes.NewBufferString(`{"reason":"finance review required before posting"}`))
	req.Header.Set("Content-Type", "application/json")
	applyResponseCookies(req, cookies)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("unexpected request-approval status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		RecommendationID     string  `json:"recommendation_id"`
		RecommendationStatus string  `json:"recommendation_status"`
		ApprovalID           string  `json:"approval_id"`
		ApprovalStatus       string  `json:"approval_status"`
		ApprovalQueueCode    string  `json:"approval_queue_code"`
		DocumentID           string  `json:"document_id"`
		DocumentStatus       *string `json:"document_status"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode request-approval response: %v", err)
	}
	if response.RecommendationID != recommendation.ID || response.ApprovalID == "" {
		t.Fatalf("unexpected request-approval ids: %+v", response)
	}
	if response.RecommendationStatus != ai.RecommendationStatusApprovalRequested || response.ApprovalStatus != "pending" {
		t.Fatalf("unexpected request-approval states: %+v", response)
	}
	if response.ApprovalQueueCode != "finance-review" || response.DocumentID != doc.ID {
		t.Fatalf("unexpected request-approval linkage: %+v", response)
	}
	if response.DocumentStatus == nil || *response.DocumentStatus != string(documents.StatusSubmitted) {
		t.Fatalf("unexpected request-approval document status: %+v", response)
	}

	postReq := httptest.NewRequest(http.MethodGet, "/api/review/processed-proposals?recommendation_id="+recommendation.ID, nil)
	applyResponseCookies(postReq, cookies)
	postRecorder := httptest.NewRecorder()
	handler.ServeHTTP(postRecorder, postReq)
	if postRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected post-action proposal status: got %d body=%s", postRecorder.Code, postRecorder.Body.String())
	}

	var postResponse struct {
		Items []struct {
			RecommendationStatus string  `json:"recommendation_status"`
			ApprovalID           *string `json:"approval_id"`
			ApprovalStatus       *string `json:"approval_status"`
			ApprovalQueueCode    *string `json:"approval_queue_code"`
			DocumentID           *string `json:"document_id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(postRecorder.Body.Bytes(), &postResponse); err != nil {
		t.Fatalf("decode post-action proposal review: %v", err)
	}
	if len(postResponse.Items) != 1 {
		t.Fatalf("unexpected post-action proposal count: %d", len(postResponse.Items))
	}
	if postResponse.Items[0].ApprovalID == nil || *postResponse.Items[0].ApprovalID != response.ApprovalID {
		t.Fatalf("unexpected post-action approval linkage: %+v", postResponse.Items[0])
	}
	if postResponse.Items[0].ApprovalStatus == nil || *postResponse.Items[0].ApprovalStatus != "pending" {
		t.Fatalf("unexpected post-action approval status: %+v", postResponse.Items[0])
	}
	if postResponse.Items[0].ApprovalQueueCode == nil || *postResponse.Items[0].ApprovalQueueCode != "finance-review" {
		t.Fatalf("unexpected post-action queue code: %+v", postResponse.Items[0])
	}
	if postResponse.Items[0].DocumentID == nil || *postResponse.Items[0].DocumentID != doc.ID {
		t.Fatalf("unexpected post-action document: %+v", postResponse.Items[0])
	}
}

func TestAgentAPIRequestProcessedProposalApprovalRejectsCrossOrgActionWithoutMutationIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}
	otherOrgID, otherOperatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)

	documentService := documents.NewService(db)
	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: "invoice",
		Title:    "Foreign proposal approval boundary invoice",
		Actor:    operator,
	})
	if err != nil {
		t.Fatalf("create foreign-boundary document: %v", err)
	}
	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("submit foreign-boundary document: %v", err)
	}

	request := createQueuedRequest(t, ctx, db, operator, "Create a proposal that another org must not approve.")
	intakeService := intake.NewService(db)
	request, err = intakeService.ClaimNextQueued(ctx, intake.ClaimNextQueuedInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("claim foreign-boundary request: %v", err)
	}
	request, err = intakeService.AdvanceRequest(ctx, intake.AdvanceRequestInput{
		RequestID: request.ID,
		Status:    intake.StatusProcessed,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("advance foreign-boundary request: %v", err)
	}

	aiService := ai.NewService(db)
	run, err := aiService.StartRun(ctx, ai.StartRunInput{
		InboundRequestID: request.ID,
		AgentRole:        ai.RunRoleCoordinator,
		CapabilityCode:   ai.DefaultCoordinatorCapabilityCode,
		Actor:            operator,
	})
	if err != nil {
		t.Fatalf("start foreign-boundary run: %v", err)
	}
	recommendation, err := aiService.CreateRecommendation(ctx, ai.CreateRecommendationInput{
		RunID:              run.ID,
		RecommendationType: "request_approval",
		Summary:            "Request finance approval for a foreign-boundary invoice.",
		Payload:            map[string]any{"document_id": doc.ID, "queue_code": "finance-review"},
		Actor:              operator,
	})
	if err != nil {
		t.Fatalf("create foreign-boundary recommendation: %v", err)
	}

	handler := app.NewServedAgentAPIHandler(db)
	otherOrgCookies := issueBrowserSessionCookies(t, ctx, db, handler, otherOrgID, otherOperatorUserID)

	req := httptest.NewRequest(http.MethodPost, "/api/review/processed-proposals/"+recommendation.ID+"/request-approval", bytes.NewBufferString(`{"reason":"wrong org should not mutate"}`))
	req.Header.Set("Content-Type", "application/json")
	applyResponseCookies(req, otherOrgCookies)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("unexpected cross-org request-approval status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	requireContains(t, recorder.Body.String(), `"error":"record not found"`)

	var approvalID sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT approval_id FROM ai.agent_recommendations WHERE id = $1`, recommendation.ID).Scan(&approvalID); err != nil {
		t.Fatalf("load recommendation approval after cross-org request-approval attempt: %v", err)
	}
	if approvalID.Valid {
		t.Fatalf("expected recommendation approval link to remain empty after cross-org request-approval attempt, got %s", approvalID.String)
	}

	var approvalCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM workflow.approvals WHERE document_id = $1`, doc.ID).Scan(&approvalCount); err != nil {
		t.Fatalf("count approvals after cross-org request-approval attempt: %v", err)
	}
	if approvalCount != 0 {
		t.Fatalf("expected no approval rows after cross-org request-approval attempt, got %d", approvalCount)
	}

	var documentStatus string
	if err := db.QueryRowContext(ctx, `SELECT status FROM documents.documents WHERE id = $1`, doc.ID).Scan(&documentStatus); err != nil {
		t.Fatalf("load document after cross-org request-approval attempt: %v", err)
	}
	if documentStatus != string(documents.StatusSubmitted) {
		t.Fatalf("expected cross-org request-approval attempt to preserve submitted document, got %s", documentStatus)
	}
}

func TestAgentAPIProcessedProposalApprovalDecisionContinuityIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	_, approverUserID := seedOrgAndUserInOrg(t, ctx, db, identityaccess.RoleApprover, orgID)
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	documentService := documents.NewService(db)
	aiService := ai.NewService(db)
	intakeService := intake.NewService(db)

	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: "invoice",
		Title:    "Submitted continuity invoice",
		Actor:    operator,
	})
	if err != nil {
		t.Fatalf("create continuity document: %v", err)
	}
	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("submit continuity document: %v", err)
	}

	request := createQueuedRequest(t, ctx, db, operator, "Create an approval-producing proposal for the submitted continuity invoice.")
	request, err = intakeService.ClaimNextQueued(ctx, intake.ClaimNextQueuedInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("claim continuity request: %v", err)
	}
	request, err = intakeService.AdvanceRequest(ctx, intake.AdvanceRequestInput{
		RequestID: request.ID,
		Status:    intake.StatusProcessed,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("advance continuity request: %v", err)
	}

	run, err := aiService.StartRun(ctx, ai.StartRunInput{
		InboundRequestID: request.ID,
		AgentRole:        ai.RunRoleCoordinator,
		CapabilityCode:   ai.DefaultCoordinatorCapabilityCode,
		Actor:            operator,
	})
	if err != nil {
		t.Fatalf("start continuity run: %v", err)
	}

	recommendation, err := aiService.CreateRecommendation(ctx, ai.CreateRecommendationInput{
		RunID:              run.ID,
		RecommendationType: "request_approval",
		Summary:            "Request finance approval for the submitted continuity invoice.",
		Payload:            map[string]any{"document_id": doc.ID, "queue_code": "finance-review"},
		Actor:              operator,
	})
	if err != nil {
		t.Fatalf("create continuity recommendation: %v", err)
	}

	handler := app.NewServedAgentAPIHandler(db)
	operatorCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, operatorUserID)
	approverCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, approverUserID)

	requestApprovalReq := httptest.NewRequest(http.MethodPost, "/api/review/processed-proposals/"+recommendation.ID+"/request-approval", bytes.NewBufferString(`{"reason":"finance review required before approval decision"}`))
	requestApprovalReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(requestApprovalReq, operatorCookies)
	requestApprovalRecorder := httptest.NewRecorder()
	handler.ServeHTTP(requestApprovalRecorder, requestApprovalReq)
	if requestApprovalRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected request-approval status: got %d body=%s", requestApprovalRecorder.Code, requestApprovalRecorder.Body.String())
	}

	var requestApprovalResponse struct {
		RecommendationID     string `json:"recommendation_id"`
		RecommendationStatus string `json:"recommendation_status"`
		ApprovalID           string `json:"approval_id"`
		ApprovalStatus       string `json:"approval_status"`
		ApprovalQueueCode    string `json:"approval_queue_code"`
		DocumentID           string `json:"document_id"`
	}
	if err := json.Unmarshal(requestApprovalRecorder.Body.Bytes(), &requestApprovalResponse); err != nil {
		t.Fatalf("decode request-approval response: %v", err)
	}
	if requestApprovalResponse.RecommendationID != recommendation.ID || requestApprovalResponse.RecommendationStatus != ai.RecommendationStatusApprovalRequested || requestApprovalResponse.ApprovalID == "" || requestApprovalResponse.ApprovalStatus != "pending" || requestApprovalResponse.ApprovalQueueCode != "finance-review" || requestApprovalResponse.DocumentID != doc.ID {
		t.Fatalf("unexpected request-approval continuity response: %+v", requestApprovalResponse)
	}

	decisionReq := httptest.NewRequest(http.MethodPost, "/api/approvals/"+requestApprovalResponse.ApprovalID+"/decision", bytes.NewBufferString(`{"decision":"approved","decision_note":"Approved from continuity flow."}`))
	decisionReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(decisionReq, approverCookies)
	decisionRecorder := httptest.NewRecorder()
	handler.ServeHTTP(decisionRecorder, decisionReq)
	if decisionRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected approval decision status: got %d body=%s", decisionRecorder.Code, decisionRecorder.Body.String())
	}

	var decisionResponse struct {
		ApprovalID      string  `json:"approval_id"`
		Status          string  `json:"status"`
		QueueCode       string  `json:"queue_code"`
		DocumentID      string  `json:"document_id"`
		DocumentStatus  string  `json:"document_status"`
		DecisionNote    *string `json:"decision_note"`
		DecidedByUserID *string `json:"decided_by_user_id"`
	}
	if err := json.Unmarshal(decisionRecorder.Body.Bytes(), &decisionResponse); err != nil {
		t.Fatalf("decode approval decision response: %v", err)
	}
	if decisionResponse.ApprovalID != requestApprovalResponse.ApprovalID || decisionResponse.Status != "approved" || decisionResponse.QueueCode != "finance-review" || decisionResponse.DocumentID != doc.ID || decisionResponse.DocumentStatus != string(documents.StatusApproved) {
		t.Fatalf("unexpected approval decision continuity: %+v", decisionResponse)
	}
	if decisionResponse.DecisionNote == nil || *decisionResponse.DecisionNote != "Approved from continuity flow." || decisionResponse.DecidedByUserID == nil || *decisionResponse.DecidedByUserID != approverUserID {
		t.Fatalf("unexpected approval decision metadata: %+v", decisionResponse)
	}

	proposalsReq := httptest.NewRequest(http.MethodGet, "/api/review/processed-proposals?recommendation_id="+recommendation.ID, nil)
	applyResponseCookies(proposalsReq, operatorCookies)
	proposalsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(proposalsRecorder, proposalsReq)
	if proposalsRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected processed proposals status: got %d body=%s", proposalsRecorder.Code, proposalsRecorder.Body.String())
	}
	requireContains(t, proposalsRecorder.Body.String(), `"recommendation_id":"`+recommendation.ID+`"`)
	requireContains(t, proposalsRecorder.Body.String(), `"approval_id":"`+requestApprovalResponse.ApprovalID+`"`)
	requireContains(t, proposalsRecorder.Body.String(), `"approval_status":"approved"`)
	requireContains(t, proposalsRecorder.Body.String(), `"document_id":"`+doc.ID+`"`)

	approvalQueueReq := httptest.NewRequest(http.MethodGet, "/api/review/approval-queue?approval_id="+requestApprovalResponse.ApprovalID, nil)
	applyResponseCookies(approvalQueueReq, approverCookies)
	approvalQueueRecorder := httptest.NewRecorder()
	handler.ServeHTTP(approvalQueueRecorder, approvalQueueReq)
	if approvalQueueRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact approval queue status: got %d body=%s", approvalQueueRecorder.Code, approvalQueueRecorder.Body.String())
	}
	requireContains(t, approvalQueueRecorder.Body.String(), `"approval_id":"`+requestApprovalResponse.ApprovalID+`"`)
	requireContains(t, approvalQueueRecorder.Body.String(), `"approval_status":"approved"`)
	requireContains(t, approvalQueueRecorder.Body.String(), `"document_id":"`+doc.ID+`"`)

	documentsReq := httptest.NewRequest(http.MethodGet, "/api/review/documents?document_id="+doc.ID, nil)
	applyResponseCookies(documentsReq, approverCookies)
	documentsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(documentsRecorder, documentsReq)
	if documentsRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact documents status: got %d body=%s", documentsRecorder.Code, documentsRecorder.Body.String())
	}
	requireContains(t, documentsRecorder.Body.String(), `"document_id":"`+doc.ID+`"`)
	requireContains(t, documentsRecorder.Body.String(), `"approval_id":"`+requestApprovalResponse.ApprovalID+`"`)
	requireContains(t, documentsRecorder.Body.String(), `"approval_status":"approved"`)

}

func TestAgentAPIReviewSurfacesRejectInvalidExactIDFiltersIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)

	handler := app.NewServedAgentAPIHandler(db)
	cookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, operatorUserID)

	testCases := []struct {
		name string
		path string
	}{
		{name: "approval queue", path: "/api/review/approval-queue?approval_id=not-a-uuid"},
		{name: "documents", path: "/api/review/documents?document_id=not-a-uuid"},
		{name: "journal entries", path: "/api/review/accounting/journal-entries?entry_id=not-a-uuid"},
		{name: "control balances", path: "/api/review/accounting/control-account-balances?account_id=not-a-uuid"},
		{name: "inventory stock", path: "/api/review/inventory/stock?item_id=not-a-uuid"},
		{name: "inventory movements", path: "/api/review/inventory/movements?movement_id=not-a-uuid"},
		{name: "inventory reconciliation", path: "/api/review/inventory/reconciliation?document_id=not-a-uuid"},
		{name: "work order list", path: "/api/review/work-orders?work_order_id=not-a-uuid"},
		{name: "work order detail", path: "/api/review/work-orders/not-a-uuid"},
		{name: "processed proposals", path: "/api/review/processed-proposals?recommendation_id=not-a-uuid"},
		{name: "processed proposals request", path: "/api/review/processed-proposals?request_id=not-a-uuid"},
		{name: "inbound request detail", path: "/api/review/inbound-requests/not-a-uuid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			applyResponseCookies(req, cookies)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("unexpected status for %s: got %d body=%s", tc.path, recorder.Code, recorder.Body.String())
			}
			requireContains(t, recorder.Body.String(), `"error":"invalid review filter"`)
		})
	}
}

func TestAgentAPIDecideApprovalIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	_, approverUserID := seedOrgAndUserInOrg(t, ctx, db, identityaccess.RoleApprover, orgID)

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)
	approval, doc := createPendingApproval(t, ctx, documentService, workflowService, operator)

	handler := app.NewServedAgentAPIHandler(db)
	approverCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, approverUserID)

	req := httptest.NewRequest(http.MethodPost, "/api/approvals/"+approval.ID+"/decision", bytes.NewBufferString(`{"decision":"approved","decision_note":"Looks correct."}`))
	req.Header.Set("Content-Type", "application/json")
	applyResponseCookies(req, approverCookies)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected approval decision status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		ApprovalID     string  `json:"approval_id"`
		Status         string  `json:"status"`
		DocumentID     string  `json:"document_id"`
		DocumentStatus string  `json:"document_status"`
		DecisionNote   *string `json:"decision_note"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode approval decision response: %v", err)
	}
	if response.ApprovalID != approval.ID || response.DocumentID != doc.ID {
		t.Fatalf("unexpected approval decision response: %+v", response)
	}
	if response.Status != "approved" || response.DocumentStatus != string(documents.StatusApproved) {
		t.Fatalf("unexpected approval decision states: %+v", response)
	}
	if response.DecisionNote == nil || *response.DecisionNote != "Looks correct." {
		t.Fatalf("unexpected decision note: %+v", response)
	}

	var documentStatus string
	if err := db.QueryRowContext(ctx, `SELECT status FROM documents.documents WHERE id = $1`, doc.ID).Scan(&documentStatus); err != nil {
		t.Fatalf("load document status: %v", err)
	}
	if documentStatus != string(documents.StatusApproved) {
		t.Fatalf("unexpected persisted document status: %s", documentStatus)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/approval-queue?status=closed", nil)
	applyResponseCookies(req, approverCookies)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected closed approval queue status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAgentAPIDecideApprovalRejectsCrossOrgDecisionWithoutMutationIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}
	otherOrgID, otherApproverUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleApprover)

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)
	approval, doc := createPendingApproval(t, ctx, documentService, workflowService, operator)

	handler := app.NewServedAgentAPIHandler(db)
	otherApproverCookies := issueBrowserSessionCookies(t, ctx, db, handler, otherOrgID, otherApproverUserID)

	req := httptest.NewRequest(http.MethodPost, "/api/approvals/"+approval.ID+"/decision", bytes.NewBufferString(`{"decision":"approved","decision_note":"wrong org should not mutate"}`))
	req.Header.Set("Content-Type", "application/json")
	applyResponseCookies(req, otherApproverCookies)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("unexpected cross-org approval decision status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	requireContains(t, recorder.Body.String(), `"error":"approval not found"`)

	var approvalStatus string
	var decidedByUserID sql.NullString
	var decisionNote sql.NullString
	var decidedAt sql.NullTime
	if err := db.QueryRowContext(
		ctx,
		`SELECT status, decided_by_user_id, decision_note, decided_at FROM workflow.approvals WHERE id = $1`,
		approval.ID,
	).Scan(&approvalStatus, &decidedByUserID, &decisionNote, &decidedAt); err != nil {
		t.Fatalf("load approval after cross-org decision attempt: %v", err)
	}
	if approvalStatus != "pending" || decidedByUserID.Valid || decisionNote.Valid || decidedAt.Valid {
		t.Fatalf("expected cross-org decision attempt to preserve pending approval, got status=%s decided_by=%v note=%v decided_at=%v", approvalStatus, decidedByUserID, decisionNote, decidedAt)
	}

	var documentStatus string
	if err := db.QueryRowContext(ctx, `SELECT status FROM documents.documents WHERE id = $1`, doc.ID).Scan(&documentStatus); err != nil {
		t.Fatalf("load document after cross-org decision attempt: %v", err)
	}
	if documentStatus != string(documents.StatusSubmitted) {
		t.Fatalf("expected cross-org decision attempt to preserve submitted document, got %s", documentStatus)
	}
}

func TestAgentAPIDecideApprovalRejectsInvalidApprovalIDIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, approverUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleApprover)

	handler := app.NewServedAgentAPIHandler(db)
	approverCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, approverUserID)

	req := httptest.NewRequest(http.MethodPost, "/api/approvals/not-a-uuid/decision", bytes.NewBufferString(`{"decision":"approved"}`))
	req.Header.Set("Content-Type", "application/json")
	applyResponseCookies(req, approverCookies)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected invalid approval status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	requireContains(t, recorder.Body.String(), `"error":"invalid approval"`)
}

func TestAgentAPIDecideApprovalRejectsInvalidRequestBodyIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, approverUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleApprover)

	handler := app.NewServedAgentAPIHandler(db)
	approverCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, approverUserID)

	testCases := []struct {
		name          string
		body          string
		expectedError string
	}{
		{name: "empty body", body: "", expectedError: `"error":"request body is required"`},
		{name: "unknown field", body: `{"decision":"approved","unexpected":true}`, expectedError: `"error":"invalid JSON request body"`},
		{name: "whitespace note", body: `{"decision":"approved","decision_note":"   "}`, expectedError: `"error":"invalid approval decision"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/approvals/11111111-1111-4111-8111-111111111111/decision", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			applyResponseCookies(req, approverCookies)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
			}
			requireContains(t, recorder.Body.String(), tc.expectedError)
		})
	}
}

func TestAgentAPIDecideApprovalConflictReturnsCurrentStateIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	_, approverUserID := seedOrgAndUserInOrg(t, ctx, db, identityaccess.RoleApprover, orgID)

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)
	approval, _ := createPendingApproval(t, ctx, documentService, workflowService, operator)

	handler := app.NewServedAgentAPIHandler(db)
	approverCookies := issueBrowserSessionCookies(t, ctx, db, handler, orgID, approverUserID)

	approveReq := httptest.NewRequest(http.MethodPost, "/api/approvals/"+approval.ID+"/decision", bytes.NewBufferString(`{"decision":"approved","decision_note":"Looks correct."}`))
	approveReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(approveReq, approverCookies)
	approveRecorder := httptest.NewRecorder()
	handler.ServeHTTP(approveRecorder, approveReq)
	if approveRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected first approval status: got %d body=%s", approveRecorder.Code, approveRecorder.Body.String())
	}

	conflictReq := httptest.NewRequest(http.MethodPost, "/api/approvals/"+approval.ID+"/decision", bytes.NewBufferString(`{"decision":"rejected"}`))
	conflictReq.Header.Set("Content-Type", "application/json")
	applyResponseCookies(conflictReq, approverCookies)
	conflictRecorder := httptest.NewRecorder()
	handler.ServeHTTP(conflictRecorder, conflictReq)

	if conflictRecorder.Code != http.StatusConflict {
		t.Fatalf("unexpected conflict status: got %d body=%s", conflictRecorder.Code, conflictRecorder.Body.String())
	}

	var response struct {
		Error          string  `json:"error"`
		ApprovalID     string  `json:"approval_id"`
		Status         string  `json:"status"`
		DocumentID     string  `json:"document_id"`
		DocumentStatus string  `json:"document_status"`
		DecisionNote   *string `json:"decision_note"`
	}
	if err := json.Unmarshal(conflictRecorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode conflict response: %v", err)
	}
	if response.Error != "approval cannot be decided in the current state" {
		t.Fatalf("unexpected conflict error: %+v", response)
	}
	if response.ApprovalID != approval.ID || response.Status != "approved" || response.DocumentStatus != string(documents.StatusApproved) {
		t.Fatalf("unexpected conflict state response: %+v", response)
	}
	if response.DecisionNote == nil || *response.DecisionNote != "Looks correct." {
		t.Fatalf("unexpected conflict decision note: %+v", response)
	}
}

func seedOrgAndUserInOrg(t *testing.T, ctx context.Context, db *sql.DB, roleCode, orgID string) (string, string) {
	t.Helper()

	var userID string
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.users (email, display_name) VALUES ($1, 'Example User') RETURNING id`,
		"user-"+time.Now().UTC().Format("150405.000000000")+"@example.com",
	).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO identityaccess.memberships (org_id, user_id, role_code) VALUES ($1, $2, $3)`,
		orgID,
		userID,
		roleCode,
	); err != nil {
		t.Fatalf("insert membership: %v", err)
	}

	if err := identityaccess.NewService(db).SetUserPassword(ctx, identityaccess.SetUserPasswordInput{
		UserID:    userID,
		Password:  testLoginPassword,
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("set test user password: %v", err)
	}

	return orgID, userID
}

func loadOrgSlugAndUserEmail(t *testing.T, ctx context.Context, db *sql.DB, orgID, userID string) (string, string) {
	t.Helper()

	var orgSlug string
	if err := db.QueryRowContext(ctx, `SELECT slug FROM identityaccess.orgs WHERE id = $1`, orgID).Scan(&orgSlug); err != nil {
		t.Fatalf("load org slug: %v", err)
	}

	var userEmail string
	if err := db.QueryRowContext(ctx, `SELECT email FROM identityaccess.users WHERE id = $1`, userID).Scan(&userEmail); err != nil {
		t.Fatalf("load user email: %v", err)
	}

	return orgSlug, userEmail
}

func loadMembershipIDForUser(t *testing.T, ctx context.Context, db *sql.DB, orgID, userID string) string {
	t.Helper()

	var membershipID string
	if err := db.QueryRowContext(
		ctx,
		`SELECT id FROM identityaccess.memberships WHERE org_id = $1 AND user_id = $2`,
		orgID,
		userID,
	).Scan(&membershipID); err != nil {
		t.Fatalf("load membership id: %v", err)
	}
	return membershipID
}

func issueBrowserSessionCookies(t *testing.T, ctx context.Context, db *sql.DB, handler http.Handler, orgID, userID string) []*http.Cookie {
	t.Helper()

	orgSlug, userEmail := loadOrgSlugAndUserEmail(t, ctx, db, orgID, userID)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/session/login", bytes.NewBufferString(`{
		"org_slug":"`+orgSlug+`",
		"email":"`+userEmail+`",
		"password":"`+testLoginPassword+`",
		"device_label":"integration-browser"
	}`))
	loginReq.Header.Set("Content-Type", "application/json")

	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, loginReq)
	if loginRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected login status: got %d body=%s", loginRecorder.Code, loginRecorder.Body.String())
	}

	return loginRecorder.Result().Cookies()
}

func issueBearerAccessToken(t *testing.T, ctx context.Context, db *sql.DB, handler http.Handler, orgID, userID string) string {
	t.Helper()

	orgSlug, userEmail := loadOrgSlugAndUserEmail(t, ctx, db, orgID, userID)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/session/token", bytes.NewBufferString(`{
		"org_slug":"`+orgSlug+`",
		"email":"`+userEmail+`",
		"password":"`+testLoginPassword+`",
		"device_label":"integration-token"
	}`))
	loginReq.Header.Set("Content-Type", "application/json")

	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, loginReq)
	if loginRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected token login status: got %d body=%s", loginRecorder.Code, loginRecorder.Body.String())
	}

	var response struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(loginRecorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode token login response: %v", err)
	}
	if strings.TrimSpace(response.AccessToken) == "" {
		t.Fatalf("expected access token in response: %s", loginRecorder.Body.String())
	}
	return response.AccessToken
}

func applyResponseCookies(req *http.Request, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
}

func applyBearer(req *http.Request, accessToken string) {
	req.Header.Set("Authorization", "Bearer "+accessToken)
}

type multipartUpload struct {
	FileName    string
	ContentType string
	Content     []byte
}

func newMultipartRequest(t *testing.T, method, target string, fields map[string]string, files map[string]multipartUpload) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write multipart field %s: %v", key, err)
		}
	}
	for fieldName, upload := range files {
		part, err := writer.CreateFormFile(fieldName, upload.FileName)
		if err != nil {
			t.Fatalf("create multipart file %s: %v", fieldName, err)
		}
		if _, err := io.Copy(part, bytes.NewReader(upload.Content)); err != nil {
			t.Fatalf("write multipart content %s: %v", fieldName, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(method, target, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func requireContains(t *testing.T, body, want string) {
	t.Helper()
	if !strings.Contains(body, want) {
		t.Fatalf("expected response to contain %q, body=%s", want, body)
	}
}

func requireNotContains(t *testing.T, body, unwanted string) {
	t.Helper()
	if strings.Contains(body, unwanted) {
		t.Fatalf("expected response not to contain %q, body=%s", unwanted, body)
	}
}

func createPendingApproval(t *testing.T, ctx context.Context, documentService *documents.Service, workflowService *workflow.Service, actor identityaccess.Actor) (workflow.Approval, documents.Document) {
	t.Helper()

	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: "invoice",
		Title:    "Approval-backed invoice",
		Actor:    actor,
	})
	if err != nil {
		t.Fatalf("create draft document: %v", err)
	}

	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      actor,
	})
	if err != nil {
		t.Fatalf("submit document: %v", err)
	}

	approval, err := workflowService.RequestApproval(ctx, workflow.RequestApprovalInput{
		DocumentID: doc.ID,
		QueueCode:  "finance",
		Reason:     "needs approval",
		Actor:      actor,
	})
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}

	return approval, doc
}

func postApprovedGSTInvoice(t *testing.T, ctx context.Context, accountingService *accounting.Service, documentService *documents.Service, workflowService *workflow.Service, operator, approver identityaccess.Actor) {
	t.Helper()

	doc, _, err := accountingService.CreateInvoice(ctx, accounting.CreateInvoiceInput{
		Title:          "Posted GST invoice",
		InvoiceRole:    accounting.InvoiceRoleSales,
		CurrencyCode:   "INR",
		ReferenceValue: "INV-TEST-1001",
		Summary:        "Browser accounting review test invoice",
		Actor:          operator,
	})
	if err != nil {
		t.Fatalf("create invoice draft: %v", err)
	}
	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("submit invoice draft: %v", err)
	}
	approval, err := workflowService.RequestApproval(ctx, workflow.RequestApprovalInput{
		DocumentID: doc.ID,
		QueueCode:  "finance",
		Reason:     "post invoice",
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}
	if _, _, err := workflowService.DecideApproval(ctx, workflow.DecideApprovalInput{
		ApprovalID: approval.ID,
		Decision:   "approved",
		Actor:      approver,
	}); err != nil {
		t.Fatalf("approve invoice: %v", err)
	}

	receivable := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "1100",
		Name:                "Accounts Receivable",
		AccountClass:        accounting.AccountClassAsset,
		ControlType:         accounting.ControlTypeReceivable,
		AllowsDirectPosting: false,
		Actor:               operator,
	})
	gstOutput := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "2101",
		Name:                "GST Output",
		AccountClass:        accounting.AccountClassLiability,
		ControlType:         accounting.ControlTypeGSTOutput,
		AllowsDirectPosting: false,
		Actor:               operator,
	})
	revenue := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "4000",
		Name:         "Service Revenue",
		AccountClass: accounting.AccountClassRevenue,
		Actor:        operator,
	})
	gst18 := createTaxCode(t, ctx, accountingService, accounting.CreateTaxCodeInput{
		Code:             "GST18",
		Name:             "GST Output 18%",
		TaxType:          accounting.TaxTypeGST,
		RateBasisPoints:  1800,
		PayableAccountID: gstOutput.ID,
		Actor:            operator,
	})

	if _, _, _, err := accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   doc.ID,
		Summary:      "Post approved invoice with GST",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeGST,
		EffectiveOn:  time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC),
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 177000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 150000},
			{AccountID: gstOutput.ID, Description: "GST payable", CreditMinor: 27000, TaxCode: gst18.Code},
		},
		Actor: operator,
	}); err != nil {
		t.Fatalf("post invoice: %v", err)
	}
}

func seedBrowserReviewData(t *testing.T, ctx context.Context, documentService *documents.Service, workflowService *workflow.Service, accountingService *accounting.Service, inventoryService *inventoryops.Service, workOrderService *workorders.Service, workforceService *workforce.Service, operator, approver identityaccess.Actor) workorders.WorkOrder {
	t.Helper()

	workOrderResult, err := workOrderService.CreateWorkOrder(ctx, workorders.CreateWorkOrderInput{
		WorkOrderCode: "WO-RPT-1001",
		Title:         "Review execution chain",
		Summary:       "Browser reporting coverage",
		Actor:         operator,
	})
	if err != nil {
		t.Fatalf("create work order: %v", err)
	}

	worker := createWorker(t, ctx, workforceService, workforce.CreateWorkerInput{
		WorkerCode:             "TECH-RPT-1",
		DisplayName:            "Reporting Technician",
		DefaultHourlyCostMinor: 3600,
		CostCurrencyCode:       "INR",
		Actor:                  operator,
	})
	task, err := workflowService.CreateTask(ctx, workflow.CreateTaskInput{
		ContextType:         "work_order",
		ContextID:           workOrderResult.WorkOrder.ID,
		Title:               "Inspect and post",
		QueueCode:           "dispatch",
		AccountableWorkerID: worker.ID,
		Actor:               operator,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if _, err := workflowService.UpdateTaskStatus(ctx, workflow.UpdateTaskStatusInput{
		TaskID: task.ID,
		Status: "completed",
		Actor:  operator,
	}); err != nil {
		t.Fatalf("complete task: %v", err)
	}

	item := createItem(t, ctx, inventoryService, inventoryops.CreateItemInput{
		SKU:          "RPT-MAT-1",
		Name:         "Reporting Material",
		ItemRole:     inventoryops.ItemRoleServiceMaterial,
		TrackingMode: inventoryops.TrackingModeNone,
		Actor:        operator,
	})
	warehouse := createLocation(t, ctx, inventoryService, inventoryops.CreateLocationInput{
		Code:         "RPT-WH-1",
		Name:         "Reporting Warehouse",
		LocationRole: inventoryops.LocationRoleWarehouse,
		Actor:        operator,
	})

	receiptDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "inventory_receipt", "Inventory receipt")
	if _, err := inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: receiptDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{{
			ItemID:                item.ID,
			MovementPurpose:       inventoryops.MovementPurposeServiceConsumption,
			UsageClassification:   inventoryops.UsageBillable,
			DestinationLocationID: warehouse.ID,
			QuantityMilli:         5000,
		}},
		Actor: operator,
	}); err != nil {
		t.Fatalf("capture receipt: %v", err)
	}

	issueDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "inventory_issue", "Inventory issue")
	if _, err := inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: issueDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{{
			ItemID:               item.ID,
			MovementPurpose:      inventoryops.MovementPurposeServiceConsumption,
			UsageClassification:  inventoryops.UsageBillable,
			SourceLocationID:     warehouse.ID,
			QuantityMilli:        2000,
			CostMinor:            5400,
			CostCurrencyCode:     "INR",
			AccountingHandoff:    true,
			ExecutionContextType: inventoryops.ExecutionContextWorkOrder,
			ExecutionContextID:   workOrderResult.WorkOrder.WorkOrderCode,
		}},
		Actor: operator,
	}); err != nil {
		t.Fatalf("capture issue: %v", err)
	}

	if _, err := workOrderService.SyncInventoryUsage(ctx, workorders.SyncInventoryUsageInput{
		WorkOrderID: workOrderResult.WorkOrder.ID,
		Actor:       operator,
	}); err != nil {
		t.Fatalf("sync inventory usage: %v", err)
	}

	startedAt := time.Date(2026, 3, 21, 9, 0, 0, 0, time.UTC)
	laborEntry, err := workforceService.RecordLabor(ctx, workforce.RecordLaborInput{
		WorkerID:    worker.ID,
		WorkOrderID: workOrderResult.WorkOrder.ID,
		TaskID:      task.ID,
		StartedAt:   startedAt,
		EndedAt:     startedAt.Add(2 * time.Hour),
		Note:        "Execution review labor",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("record labor: %v", err)
	}

	materialExpense := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "5101",
		Name:         "Material Expense",
		AccountClass: accounting.AccountClassExpense,
		Actor:        operator,
	})
	laborExpense := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "5102",
		Name:         "Labor Expense",
		AccountClass: accounting.AccountClassExpense,
		Actor:        operator,
	})
	accruedOffset := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "2201",
		Name:         "Accrued Costs",
		AccountClass: accounting.AccountClassLiability,
		Actor:        operator,
	})

	laborJournalDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "journal", "Labor posting")
	if _, err := accountingService.PostWorkOrderLabor(ctx, accounting.PostWorkOrderLaborInput{
		DocumentID:       laborJournalDoc.ID,
		WorkOrderID:      workOrderResult.WorkOrder.ID,
		ExpenseAccountID: laborExpense.ID,
		OffsetAccountID:  accruedOffset.ID,
		Summary:          "Post labor review costs",
		EffectiveOn:      startedAt,
		Actor:            operator,
	}); err != nil {
		t.Fatalf("post work order labor: %v", err)
	}

	materialJournalDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "journal", "Material posting")
	if _, err := accountingService.PostWorkOrderInventory(ctx, accounting.PostWorkOrderInventoryInput{
		DocumentID:       materialJournalDoc.ID,
		WorkOrderID:      workOrderResult.WorkOrder.ID,
		ExpenseAccountID: materialExpense.ID,
		OffsetAccountID:  accruedOffset.ID,
		Summary:          "Post material review costs",
		EffectiveOn:      startedAt,
		Actor:            operator,
	}); err != nil {
		t.Fatalf("post work order inventory: %v", err)
	}

	if laborEntry.ID == "" {
		t.Fatal("expected labor entry id")
	}
	return workOrderResult.WorkOrder
}

func prepareApprovedDocumentOfType(t *testing.T, ctx context.Context, documentService *documents.Service, workflowService *workflow.Service, operator, approver identityaccess.Actor, typeCode, title string) documents.Document {
	t.Helper()

	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: typeCode,
		Title:    title,
		Actor:    operator,
	})
	if err != nil {
		t.Fatalf("create draft %s document: %v", typeCode, err)
	}
	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("submit %s document: %v", typeCode, err)
	}
	approval, err := workflowService.RequestApproval(ctx, workflow.RequestApprovalInput{
		DocumentID: doc.ID,
		QueueCode:  "operations",
		Reason:     "prepare review data",
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("request %s approval: %v", typeCode, err)
	}
	_, doc, err = workflowService.DecideApproval(ctx, workflow.DecideApprovalInput{
		ApprovalID: approval.ID,
		Decision:   "approved",
		Actor:      approver,
	})
	if err != nil {
		t.Fatalf("approve %s document: %v", typeCode, err)
	}
	return doc
}

func createItem(t *testing.T, ctx context.Context, service *inventoryops.Service, input inventoryops.CreateItemInput) inventoryops.Item {
	t.Helper()

	item, err := service.CreateItem(ctx, input)
	if err != nil {
		t.Fatalf("create item %s: %v", input.SKU, err)
	}
	return item
}

func createLocation(t *testing.T, ctx context.Context, service *inventoryops.Service, input inventoryops.CreateLocationInput) inventoryops.Location {
	t.Helper()

	location, err := service.CreateLocation(ctx, input)
	if err != nil {
		t.Fatalf("create location %s: %v", input.Code, err)
	}
	return location
}

func createWorker(t *testing.T, ctx context.Context, service *workforce.Service, input workforce.CreateWorkerInput) workforce.Worker {
	t.Helper()

	worker, err := service.CreateWorker(ctx, input)
	if err != nil {
		t.Fatalf("create worker %s: %v", input.WorkerCode, err)
	}
	return worker
}

func createLedgerAccount(t *testing.T, ctx context.Context, service *accounting.Service, input accounting.CreateLedgerAccountInput) accounting.LedgerAccount {
	t.Helper()

	account, err := service.CreateLedgerAccount(ctx, input)
	if err != nil {
		t.Fatalf("create ledger account %s: %v", input.Code, err)
	}
	return account
}

func createTaxCode(t *testing.T, ctx context.Context, service *accounting.Service, input accounting.CreateTaxCodeInput) accounting.TaxCode {
	t.Helper()

	code, err := service.CreateTaxCode(ctx, input)
	if err != nil {
		t.Fatalf("create tax code %s: %v", input.Code, err)
	}
	return code
}
