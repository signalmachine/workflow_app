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

func TestHandleWebWorkOrdersPassesExactWorkOrderFilter(t *testing.T) {
	var captured reporting.ListWorkOrdersInput
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listWorkOrders: func(_ context.Context, input reporting.ListWorkOrdersInput) ([]reporting.WorkOrderReview, error) {
				captured = input
				return []reporting.WorkOrderReview{{
					WorkOrderID:   "work-order-123",
					WorkOrderCode: "WO-123",
					Title:         "Filtered work order",
					DocumentID:    "doc-123",
					Status:        "open",
				}}, nil
			},
		},
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				return testSessionContext(), nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/review/work-orders?work_order_id=work-order-123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if captured.WorkOrderID != "work-order-123" {
		t.Fatalf("expected exact work-order filter, got %+v", captured)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `name="work_order_id" value="work-order-123"`) {
		t.Fatalf("expected work-order filter value in form, body=%s", body)
	}
}

func TestHandleWebWorkOrderDetailAddsFilteredReviewLink(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			getWorkOrderReview: func(context.Context, reporting.GetWorkOrderReviewInput) (reporting.WorkOrderReview, error) {
				return reporting.WorkOrderReview{
					WorkOrderID:    "work-order-123",
					WorkOrderCode:  "WO-123",
					Title:          "Filtered work order",
					Summary:        "Review continuity",
					Status:         "in_progress",
					DocumentID:     "doc-123",
					DocumentStatus: "approved",
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/work-orders/work-order-123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `/app/review/work-orders?work_order_id=work-order-123">Filtered list view</a>`) {
		t.Fatalf("expected filtered work-order review link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/accounting?document_id=doc-123">Accounting review</a>`) {
		t.Fatalf("expected accounting review link, body=%s", body)
	}
}

func TestHandleWebDocumentsAddsExactApprovalLink(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listDocuments: func(context.Context, reporting.ListDocumentsInput) ([]reporting.DocumentReview, error) {
				return []reporting.DocumentReview{
					{
						DocumentID:        "doc-123",
						TypeCode:          "invoice",
						Title:             "Reviewable invoice",
						Status:            "submitted",
						ApprovalID:        sql.NullString{String: "approval-123", Valid: true},
						ApprovalStatus:    sql.NullString{String: "pending", Valid: true},
						ApprovalQueueCode: sql.NullString{String: "finance_review", Valid: true},
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/documents", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `/app/review/approvals/approval-123">finance_review</a>`) {
		t.Fatalf("expected exact approval link in document review, body=%s", body)
	}
}

func TestHandleWebProposalsAddsSummaryAndExactLinks(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listProcessedProposalStatusSummary: func(context.Context, identityaccess.Actor) ([]reporting.ProcessedProposalStatusSummary, error) {
				return []reporting.ProcessedProposalStatusSummary{
					{
						RecommendationStatus: "approval_requested",
						ProposalCount:        2,
						RequestCount:         1,
						DocumentCount:        1,
						LatestCreatedAt:      time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
					},
				}, nil
			},
			listProcessedProposals: func(context.Context, reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error) {
				return []reporting.ProcessedProposalReview{
					{
						RequestReference:     "REQ-000123",
						RequestStatus:        "processed",
						RecommendationID:     "rec-123",
						RecommendationStatus: "approval_requested",
						Summary:              "Review continuity",
						ApprovalID:           sql.NullString{String: "approval-123", Valid: true},
						ApprovalStatus:       sql.NullString{String: "pending", Valid: true},
						ApprovalQueueCode:    sql.NullString{String: "finance_review", Valid: true},
						DocumentID:           sql.NullString{String: "doc-123", Valid: true},
						DocumentTitle:        sql.NullString{String: "Invoice proposal", Valid: true},
						DocumentTypeCode:     sql.NullString{String: "invoice", Valid: true},
						DocumentStatus:       sql.NullString{String: "submitted", Valid: true},
						CreatedAt:            time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/proposals", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `/app/review/proposals?status=approval_requested">Open approval_requested</a>`) {
		t.Fatalf("expected status-summary filter link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/proposals/rec-123">Open exact proposal</a>`) {
		t.Fatalf("expected exact proposal link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/approvals/approval-123">Open exact approval</a>`) {
		t.Fatalf("expected exact approval link, body=%s", body)
	}
}

func TestHandleWebAppDashboardAddsExactProposalAndApprovalLinks(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listProcessedProposals: func(context.Context, reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error) {
				return []reporting.ProcessedProposalReview{
					{
						RequestReference:     "REQ-000123",
						RecommendationID:     "rec-123",
						RecommendationStatus: "approval_requested",
						Summary:              "Dashboard continuity",
						ApprovalID:           sql.NullString{String: "approval-123", Valid: true},
						ApprovalStatus:       sql.NullString{String: "pending", Valid: true},
						ApprovalQueueCode:    sql.NullString{String: "finance_review", Valid: true},
						DocumentID:           sql.NullString{String: "doc-123", Valid: true},
						DocumentTitle:        sql.NullString{String: "Invoice proposal", Valid: true},
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

	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `/app/review/proposals/rec-123">Open exact proposal</a>`) {
		t.Fatalf("expected dashboard exact proposal link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/approvals/approval-123">finance_review</a>`) {
		t.Fatalf("expected dashboard exact approval link, body=%s", body)
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
	listApprovalQueue                  func(context.Context, reporting.ListApprovalQueueInput) ([]reporting.ApprovalQueueEntry, error)
	listDocuments                      func(context.Context, reporting.ListDocumentsInput) ([]reporting.DocumentReview, error)
	getDocumentReview                  func(context.Context, reporting.GetDocumentReviewInput) (reporting.DocumentReview, error)
	listInventoryStock                 func(context.Context, reporting.ListInventoryStockInput) ([]reporting.InventoryStockItem, error)
	listAuditEvents                    func(context.Context, reporting.LookupAuditEventsInput) ([]reporting.AuditEvent, error)
	listWorkOrders                     func(context.Context, reporting.ListWorkOrdersInput) ([]reporting.WorkOrderReview, error)
	getWorkOrderReview                 func(context.Context, reporting.GetWorkOrderReviewInput) (reporting.WorkOrderReview, error)
	listProcessedProposals             func(context.Context, reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error)
	listProcessedProposalStatusSummary func(context.Context, identityaccess.Actor) ([]reporting.ProcessedProposalStatusSummary, error)
}

func (s stubOperatorReviewReader) ListApprovalQueue(ctx context.Context, input reporting.ListApprovalQueueInput) ([]reporting.ApprovalQueueEntry, error) {
	if s.listApprovalQueue != nil {
		return s.listApprovalQueue(ctx, input)
	}
	return nil, nil
}

func (s stubOperatorReviewReader) ListDocuments(ctx context.Context, input reporting.ListDocumentsInput) ([]reporting.DocumentReview, error) {
	if s.listDocuments != nil {
		return s.listDocuments(ctx, input)
	}
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

func (s stubOperatorReviewReader) ListWorkOrders(ctx context.Context, input reporting.ListWorkOrdersInput) ([]reporting.WorkOrderReview, error) {
	if s.listWorkOrders != nil {
		return s.listWorkOrders(ctx, input)
	}
	return nil, nil
}

func (s stubOperatorReviewReader) GetWorkOrderReview(ctx context.Context, input reporting.GetWorkOrderReviewInput) (reporting.WorkOrderReview, error) {
	if s.getWorkOrderReview != nil {
		return s.getWorkOrderReview(ctx, input)
	}
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

func (s stubOperatorReviewReader) ListProcessedProposals(ctx context.Context, input reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error) {
	if s.listProcessedProposals != nil {
		return s.listProcessedProposals(ctx, input)
	}
	return nil, nil
}

func (s stubOperatorReviewReader) ListProcessedProposalStatusSummary(ctx context.Context, actor identityaccess.Actor) ([]reporting.ProcessedProposalStatusSummary, error) {
	if s.listProcessedProposalStatusSummary != nil {
		return s.listProcessedProposalStatusSummary(ctx, actor)
	}
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
