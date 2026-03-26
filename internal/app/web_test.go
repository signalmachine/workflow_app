package app

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/reporting"
)

func TestHandleWebDocumentDetailFallsBackToDocumentScopedAccountingLink(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			getDocumentReview: func(context.Context, reporting.GetDocumentReviewInput) (reporting.DocumentReview, error) {
				return reporting.DocumentReview{
					DocumentID:         "doc-123",
					TypeCode:           "invoice",
					Title:              "Posted invoice",
					Status:             "posted",
					CreatedByUserID:    "user-123",
					CreatedAt:          time.Date(2026, 3, 26, 10, 0, 0, 0, time.UTC),
					UpdatedAt:          time.Date(2026, 3, 26, 11, 0, 0, 0, time.UTC),
					JournalEntryNumber: sql.NullInt64{Int64: 42, Valid: true},
					JournalEntryPostedAt: sql.NullTime{
						Time:  time.Date(2026, 3, 26, 11, 30, 0, 0, time.UTC),
						Valid: true,
					},
				}, nil
			},
		},
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				return identityaccess.SessionContext{
					Actor: identityaccess.Actor{
						OrgID:     "org-123",
						UserID:    "user-123",
						SessionID: "00000000-0000-4000-8000-000000000123",
					},
					RoleCode:  identityaccess.RoleOperator,
					OrgSlug:   "acme",
					UserEmail: "operator@example.com",
					Session: identityaccess.Session{
						ID:        "00000000-0000-4000-8000-000000000123",
						OrgID:     "org-123",
						UserID:    "user-123",
						ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
					},
				}, nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/review/documents/doc-123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `/app/review/accounting?document_id=doc-123">Entry #42</a>`) {
		t.Fatalf("expected document-scoped accounting fallback link, body=%s", body)
	}
	if strings.Contains(body, `/app/review/accounting">Entry #42</a>`) {
		t.Fatalf("unexpected generic accounting fallback link, body=%s", body)
	}
}

func TestHandleWebInventoryAddsStockContinuityLinks(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInventoryStock: func(context.Context, reporting.ListInventoryStockInput) ([]reporting.InventoryStockItem, error) {
				return []reporting.InventoryStockItem{
					{
						ItemID:       "item-123",
						ItemSKU:      "RPT-MAT-1",
						ItemName:     "Reporting material",
						ItemRole:     "material",
						LocationID:   "loc-123",
						LocationCode: "MAIN",
						LocationName: "Main store",
						LocationRole: "warehouse",
						OnHandMilli:  1200,
					},
				}, nil
			},
		},
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				return testSessionContext(), nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/review/inventory", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `/app/review/inventory?item_id=item-123&amp;location_id=loc-123#stock-balances`) {
		t.Fatalf("expected stock-balance anchor link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory?item_id=item-123&amp;location_id=loc-123#movement-history`) {
		t.Fatalf("expected movement-history link from stock row, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory?item_id=item-123#reconciliation`) {
		t.Fatalf("expected reconciliation link from stock row, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory?location_id=loc-123#movement-history`) {
		t.Fatalf("expected location movement link from stock row, body=%s", body)
	}
}

func TestHandleWebAuditDetailLinksInventoryEntitiesToStockBalances(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listAuditEvents: func(context.Context, reporting.LookupAuditEventsInput) ([]reporting.AuditEvent, error) {
				return []reporting.AuditEvent{
					{
						ID:         "audit-123",
						EventType:  "inventory_ops.item_reviewed",
						EntityType: "inventory_ops.item",
						EntityID:   "item-123",
						OccurredAt: time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
					},
				}, nil
			},
		},
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				return testSessionContext(), nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/review/audit/audit-123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `/app/review/inventory?item_id=item-123#stock-balances`) {
		t.Fatalf("expected item-scoped stock-balance link, body=%s", body)
	}
	if !strings.Contains(body, `Open inventory item review`) {
		t.Fatalf("expected updated inventory audit label, body=%s", body)
	}
}

func testSessionContext() identityaccess.SessionContext {
	return identityaccess.SessionContext{
		Actor: identityaccess.Actor{
			OrgID:     "org-123",
			UserID:    "user-123",
			SessionID: "00000000-0000-4000-8000-000000000123",
		},
		RoleCode:  identityaccess.RoleOperator,
		OrgSlug:   "acme",
		UserEmail: "operator@example.com",
		Session: identityaccess.Session{
			ID:        "00000000-0000-4000-8000-000000000123",
			OrgID:     "org-123",
			UserID:    "user-123",
			ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		},
	}
}

type stubOperatorReviewReader struct {
	getDocumentReview  func(context.Context, reporting.GetDocumentReviewInput) (reporting.DocumentReview, error)
	listInventoryStock func(context.Context, reporting.ListInventoryStockInput) ([]reporting.InventoryStockItem, error)
	listAuditEvents    func(context.Context, reporting.LookupAuditEventsInput) ([]reporting.AuditEvent, error)
}

func (s stubOperatorReviewReader) ListApprovalQueue(context.Context, reporting.ListApprovalQueueInput) ([]reporting.ApprovalQueueEntry, error) {
	return nil, nil
}

func (s stubOperatorReviewReader) ListDocuments(context.Context, reporting.ListDocumentsInput) ([]reporting.DocumentReview, error) {
	return nil, nil
}

func (s stubOperatorReviewReader) GetDocumentReview(ctx context.Context, input reporting.GetDocumentReviewInput) (reporting.DocumentReview, error) {
	if s.getDocumentReview != nil {
		return s.getDocumentReview(ctx, input)
	}
	return reporting.DocumentReview{}, nil
}

func (s stubOperatorReviewReader) ListJournalEntries(context.Context, reporting.ListJournalEntriesInput) ([]reporting.JournalEntryReview, error) {
	return nil, nil
}

func (s stubOperatorReviewReader) ListControlAccountBalances(context.Context, reporting.ListControlAccountBalancesInput) ([]reporting.ControlAccountBalance, error) {
	return nil, nil
}

func (s stubOperatorReviewReader) ListTaxSummaries(context.Context, reporting.ListTaxSummariesInput) ([]reporting.TaxSummary, error) {
	return nil, nil
}

func (s stubOperatorReviewReader) ListInventoryStock(ctx context.Context, input reporting.ListInventoryStockInput) ([]reporting.InventoryStockItem, error) {
	if s.listInventoryStock != nil {
		return s.listInventoryStock(ctx, input)
	}
	return nil, nil
}

func (s stubOperatorReviewReader) ListInventoryMovements(context.Context, reporting.ListInventoryMovementsInput) ([]reporting.InventoryMovementReview, error) {
	return nil, nil
}

func (s stubOperatorReviewReader) ListInventoryReconciliation(context.Context, reporting.ListInventoryReconciliationInput) ([]reporting.InventoryReconciliationItem, error) {
	return nil, nil
}

func (s stubOperatorReviewReader) ListWorkOrders(context.Context, reporting.ListWorkOrdersInput) ([]reporting.WorkOrderReview, error) {
	return nil, nil
}

func (s stubOperatorReviewReader) GetWorkOrderReview(context.Context, reporting.GetWorkOrderReviewInput) (reporting.WorkOrderReview, error) {
	return reporting.WorkOrderReview{}, nil
}

func (s stubOperatorReviewReader) LookupAuditEvents(ctx context.Context, input reporting.LookupAuditEventsInput) ([]reporting.AuditEvent, error) {
	if s.listAuditEvents != nil {
		return s.listAuditEvents(ctx, input)
	}
	return nil, nil
}

func (s stubOperatorReviewReader) ListInboundRequests(context.Context, reporting.ListInboundRequestsInput) ([]reporting.InboundRequestReview, error) {
	return nil, nil
}

func (s stubOperatorReviewReader) GetInboundRequestDetail(context.Context, reporting.GetInboundRequestDetailInput) (reporting.InboundRequestDetail, error) {
	return reporting.InboundRequestDetail{}, nil
}

func (s stubOperatorReviewReader) ListInboundRequestStatusSummary(context.Context, identityaccess.Actor) ([]reporting.InboundRequestStatusSummary, error) {
	return nil, nil
}

func (s stubOperatorReviewReader) ListProcessedProposals(context.Context, reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error) {
	return nil, nil
}

func (s stubOperatorReviewReader) ListProcessedProposalStatusSummary(context.Context, identityaccess.Actor) ([]reporting.ProcessedProposalStatusSummary, error) {
	return nil, nil
}

type stubBrowserSessionService struct {
	authenticateSession func(context.Context, string, string) (identityaccess.SessionContext, error)
}

func (s stubBrowserSessionService) StartBrowserSession(context.Context, identityaccess.StartBrowserSessionInput) (identityaccess.BrowserSession, error) {
	return identityaccess.BrowserSession{}, nil
}

func (s stubBrowserSessionService) AuthenticateSession(ctx context.Context, sessionID, refreshToken string) (identityaccess.SessionContext, error) {
	if s.authenticateSession != nil {
		return s.authenticateSession(ctx, sessionID, refreshToken)
	}
	return identityaccess.SessionContext{}, identityaccess.ErrUnauthorized
}

func (s stubBrowserSessionService) RevokeAuthenticatedSession(context.Context, string, string) error {
	return nil
}
