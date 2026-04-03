package app

import (
	"bytes"
	"context"
	"database/sql"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"workflow_app/internal/accounting"
	"workflow_app/internal/attachments"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/inventoryops"
	"workflow_app/internal/parties"
	"workflow_app/internal/reporting"
)

func TestRenderWebPageRejectsUnmappedTemplateData(t *testing.T) {
	handler := &AgentAPIHandler{}
	recorder := httptest.NewRecorder()

	handler.renderWebPage(recorder, webPageData{Title: "workflow_app"})

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusInternalServerError, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "web page template not configured") {
		t.Fatalf("expected unmapped-template error body, got %s", recorder.Body.String())
	}
}

func TestRegisterWebRoutesSvelteModeServesSPAFallback(t *testing.T) {
	handler := &AgentAPIHandler{webFrontend: webFrontendSvelte}
	mux := http.NewServeMux()
	registerWebRoutes(mux, handler)

	req := httptest.NewRequest(http.MethodGet, "/app/login", nil)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "workflow_app SPA fallback placeholder") {
		t.Fatalf("expected SPA fallback placeholder body, got %s", recorder.Body.String())
	}
}

func TestHandleSvelteAppServesIndexAtAppRoot(t *testing.T) {
	handler := &AgentAPIHandler{webFrontend: webFrontendSvelte}
	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	recorder := httptest.NewRecorder()

	handler.handleSvelteApp(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "workflow_app web bundle placeholder") {
		t.Fatalf("expected bundle placeholder body, got %s", recorder.Body.String())
	}
}

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

func TestHandleWebDocumentDetailAddsUpstreamProposalContinuity(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			getDocumentReview: func(context.Context, reporting.GetDocumentReviewInput) (reporting.DocumentReview, error) {
				return reporting.DocumentReview{
					DocumentID:           "doc-123",
					TypeCode:             "invoice",
					Title:                "Posted invoice",
					Status:               "posted",
					CreatedByUserID:      "user-123",
					CreatedAt:            time.Date(2026, 3, 27, 8, 0, 0, 0, time.UTC),
					UpdatedAt:            time.Date(2026, 3, 27, 8, 30, 0, 0, time.UTC),
					RequestReference:     sql.NullString{String: "REQ-000123", Valid: true},
					RecommendationID:     sql.NullString{String: "rec-123", Valid: true},
					RecommendationStatus: sql.NullString{String: "approval_requested", Valid: true},
					RunID:                sql.NullString{String: "run-123", Valid: true},
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/documents/doc-123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `/app/inbound-requests/REQ-000123">REQ-000123</a>`) {
		t.Fatalf("expected request continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/proposals/rec-123">Proposal</a>`) {
		t.Fatalf("expected proposal continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/run:run-123#run-run-123">AI run</a>`) {
		t.Fatalf("expected AI run continuity link, body=%s", body)
	}
}

func TestHandleWebAccountingDetailAddsUpstreamContinuityLinks(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listJournalEntries: func(context.Context, reporting.ListJournalEntriesInput) ([]reporting.JournalEntryReview, error) {
				return []reporting.JournalEntryReview{{
					EntryID:              "entry-123",
					EntryNumber:          42,
					EntryKind:            "posting",
					SourceDocumentID:     sql.NullString{String: "doc-123", Valid: true},
					CurrencyCode:         "INR",
					TaxScopeCode:         "gst",
					Summary:              "Posted inventory issue",
					PostedByUserID:       "user-123",
					EffectiveOn:          time.Date(2026, 3, 27, 9, 0, 0, 0, time.UTC),
					PostedAt:             time.Date(2026, 3, 27, 9, 5, 0, 0, time.UTC),
					CreatedAt:            time.Date(2026, 3, 27, 9, 0, 0, 0, time.UTC),
					DocumentTypeCode:     sql.NullString{String: "inventory_issue", Valid: true},
					DocumentNumber:       sql.NullString{String: "INV-123", Valid: true},
					DocumentStatus:       sql.NullString{String: "posted", Valid: true},
					ApprovalID:           sql.NullString{String: "approval-123", Valid: true},
					ApprovalStatus:       sql.NullString{String: "approved", Valid: true},
					ApprovalQueueCode:    sql.NullString{String: "inventory_review", Valid: true},
					RequestReference:     sql.NullString{String: "REQ-000123", Valid: true},
					RecommendationID:     sql.NullString{String: "rec-123", Valid: true},
					RecommendationStatus: sql.NullString{String: "approved", Valid: true},
					RunID:                sql.NullString{String: "run-123", Valid: true},
					LineCount:            2,
					TotalDebitMinor:      1000,
					TotalCreditMinor:     1000,
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/accounting/entry-123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `/app/inbound-requests/REQ-000123">REQ-000123</a>`) {
		t.Fatalf("expected request continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/proposals/rec-123">Proposal</a>`) {
		t.Fatalf("expected proposal continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/approvals/approval-123">inventory_review</a>`) {
		t.Fatalf("expected approval continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/run:run-123#run-run-123">AI run</a>`) {
		t.Fatalf("expected AI run continuity link, body=%s", body)
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
	if !strings.Contains(body, `/app/review/inventory/items/item-123`) {
		t.Fatalf("expected exact inventory item link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory?item_id=item-123&amp;location_id=loc-123#movement-history`) {
		t.Fatalf("expected movement-history link from stock row, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory?item_id=item-123#reconciliation`) {
		t.Fatalf("expected reconciliation link from stock row, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory/locations/loc-123`) {
		t.Fatalf("expected exact inventory location link, body=%s", body)
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
	if !strings.Contains(body, `/app/review/inventory/items/item-123`) {
		t.Fatalf("expected item detail link, body=%s", body)
	}
	if !strings.Contains(body, `Open inventory item review`) {
		t.Fatalf("expected updated inventory audit label, body=%s", body)
	}
}

func TestHandleWebAuditDetailLinksAIRunEntitiesToInboundRequestDetail(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listAuditEvents: func(context.Context, reporting.LookupAuditEventsInput) ([]reporting.AuditEvent, error) {
				return []reporting.AuditEvent{
					{
						ID:         "audit-123",
						EventType:  "ai.run_completed",
						EntityType: "ai.agent_run",
						EntityID:   "run-123",
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
	if !strings.Contains(body, `/app/inbound-requests/run:run-123#run-run-123`) {
		t.Fatalf("expected AI run inbound-request detail link, body=%s", body)
	}
	if !strings.Contains(body, `Open inbound request execution detail`) {
		t.Fatalf("expected AI run audit label, body=%s", body)
	}
}

func TestHandleWebAuditDetailLinksAIStepEntitiesToExactInboundRequestSection(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listAuditEvents: func(context.Context, reporting.LookupAuditEventsInput) ([]reporting.AuditEvent, error) {
				return []reporting.AuditEvent{
					{
						ID:         "audit-123",
						EventType:  "ai.step_completed",
						EntityType: "ai.agent_run_step",
						EntityID:   "step-123",
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
	if !strings.Contains(body, `/app/inbound-requests/step:step-123#step-step-123`) {
		t.Fatalf("expected AI step inbound-request detail link, body=%s", body)
	}
	if !strings.Contains(body, `Open inbound request step detail`) {
		t.Fatalf("expected AI step audit label, body=%s", body)
	}
}

func TestHandleWebAuditDetailLinksAIDelegationEntitiesToExactInboundRequestSection(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listAuditEvents: func(context.Context, reporting.LookupAuditEventsInput) ([]reporting.AuditEvent, error) {
				return []reporting.AuditEvent{
					{
						ID:         "audit-123",
						EventType:  "ai.delegation_completed",
						EntityType: "ai.agent_delegation",
						EntityID:   "delegation-123",
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
	if !strings.Contains(body, `/app/inbound-requests/delegation:delegation-123#delegation-delegation-123`) {
		t.Fatalf("expected AI delegation inbound-request detail link, body=%s", body)
	}
	if !strings.Contains(body, `Open inbound request delegation detail`) {
		t.Fatalf("expected AI delegation audit label, body=%s", body)
	}
}

func TestHandleWebInboundRequestDetailResolvesAIRunLookup(t *testing.T) {
	var captured reporting.GetInboundRequestDetailInput
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			getInboundRequestDetail: func(_ context.Context, input reporting.GetInboundRequestDetailInput) (reporting.InboundRequestDetail, error) {
				captured = input
				return reporting.InboundRequestDetail{
					Request: reporting.InboundRequestReview{
						RequestID:        "request-123",
						RequestReference: "REQ-000123",
						Status:           "processed",
						ReceivedAt:       time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
						CreatedAt:        time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
						UpdatedAt:        time.Date(2026, 3, 26, 12, 5, 0, 0, time.UTC),
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

	req := httptest.NewRequest(http.MethodGet, "/app/inbound-requests/run:run-123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if captured.RunID != "run-123" || captured.RequestID != "" || captured.RequestReference != "" || captured.DelegationID != "" {
		t.Fatalf("unexpected inbound request detail lookup: %+v", captured)
	}
}

func TestHandleWebInboundRequestDetailResolvesAIStepLookup(t *testing.T) {
	var captured reporting.GetInboundRequestDetailInput
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			getInboundRequestDetail: func(_ context.Context, input reporting.GetInboundRequestDetailInput) (reporting.InboundRequestDetail, error) {
				captured = input
				return reporting.InboundRequestDetail{
					Request: reporting.InboundRequestReview{
						RequestID:        "request-123",
						RequestReference: "REQ-000123",
						Status:           "processed",
						ReceivedAt:       time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
						CreatedAt:        time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
						UpdatedAt:        time.Date(2026, 3, 26, 12, 5, 0, 0, time.UTC),
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

	req := httptest.NewRequest(http.MethodGet, "/app/inbound-requests/step:step-123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if captured.StepID != "step-123" || captured.RequestID != "" || captured.RequestReference != "" || captured.RunID != "" || captured.DelegationID != "" {
		t.Fatalf("unexpected inbound request detail lookup: %+v", captured)
	}
}

func TestHandleWebInboundRequestDetailAddsAnchoredExecutionSections(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			getInboundRequestDetail: func(_ context.Context, input reporting.GetInboundRequestDetailInput) (reporting.InboundRequestDetail, error) {
				return reporting.InboundRequestDetail{
					Request: reporting.InboundRequestReview{
						RequestID:        "request-123",
						RequestReference: "REQ-000123",
						Status:           "processed",
						ReceivedAt:       time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
						CreatedAt:        time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
						UpdatedAt:        time.Date(2026, 3, 26, 12, 5, 0, 0, time.UTC),
					},
					Runs: []reporting.AIRunReview{{
						RunID:          "run-123",
						AgentRole:      "coordinator",
						CapabilityCode: "intake.process",
						Status:         "completed",
						Summary:        "Completed request processing",
						StartedAt:      time.Date(2026, 3, 26, 12, 0, 30, 0, time.UTC),
						CompletedAt:    sql.NullTime{Time: time.Date(2026, 3, 26, 12, 2, 0, 0, time.UTC), Valid: true},
					}, {
						RunID:          "run-456",
						AgentRole:      "specialist",
						CapabilityCode: "reporting.read",
						Status:         "completed",
						Summary:        "Loaded reporting context",
						StartedAt:      time.Date(2026, 3, 26, 12, 1, 0, 0, time.UTC),
						CompletedAt:    sql.NullTime{Time: time.Date(2026, 3, 26, 12, 1, 30, 0, time.UTC), Valid: true},
					}},
					Steps: []reporting.AIStepReview{{
						StepID:       "step-123",
						RunID:        "run-123",
						StepIndex:    1,
						StepType:     "tool_call",
						StepTitle:    "Load intake context",
						Status:       "completed",
						InputPayload: []byte(`{"tool":"load_request"}`),
						CreatedAt:    time.Date(2026, 3, 26, 12, 0, 45, 0, time.UTC),
					}},
					Delegations: []reporting.AIDelegationReview{{
						DelegationID:        "delegation-123",
						ParentRunID:         "run-123",
						ChildRunID:          "run-456",
						RequestedByStepID:   sql.NullString{String: "step-123", Valid: true},
						CapabilityCode:      "reporting.read",
						Reason:              "Read downstream context",
						ChildAgentRole:      "specialist",
						ChildCapabilityCode: "reporting.read",
						ChildRunStatus:      "completed",
						CreatedAt:           time.Date(2026, 3, 26, 12, 1, 0, 0, time.UTC),
					}},
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

	req := httptest.NewRequest(http.MethodGet, "/app/inbound-requests/REQ-000123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `id="run-run-123"`) {
		t.Fatalf("expected anchored run section, body=%s", body)
	}
	if !strings.Contains(body, `id="step-step-123"`) {
		t.Fatalf("expected anchored step section, body=%s", body)
	}
	if !strings.Contains(body, `id="delegation-delegation-123"`) {
		t.Fatalf("expected anchored delegation section, body=%s", body)
	}
	if !strings.Contains(body, `href="#run-run-123">Run run-123</a>`) {
		t.Fatalf("expected step link back to run section, body=%s", body)
	}
	if !strings.Contains(body, `href="#step-step-123">step-123</a>`) {
		t.Fatalf("expected delegation link back to requesting step, body=%s", body)
	}
	if !strings.Contains(body, `href="#run-run-456">run-456</a>`) {
		t.Fatalf("expected delegation child run link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/audit?entity_type=ai.agent_run&amp;entity_id=run-123`) {
		t.Fatalf("expected run audit link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/audit?entity_type=ai.agent_run_step&amp;entity_id=step-123`) {
		t.Fatalf("expected step audit link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/audit?entity_type=ai.agent_delegation&amp;entity_id=delegation-123`) {
		t.Fatalf("expected delegation audit link, body=%s", body)
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

func TestHandleWebWorkOrdersAddsUpstreamContinuityLinks(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listWorkOrders: func(_ context.Context, input reporting.ListWorkOrdersInput) ([]reporting.WorkOrderReview, error) {
				return []reporting.WorkOrderReview{{
					WorkOrderID:      "work-order-123",
					WorkOrderCode:    "WO-123",
					Title:            "Linked work order",
					DocumentID:       "doc-123",
					Status:           "open",
					RequestReference: sql.NullString{String: "REQ-000123", Valid: true},
					RecommendationID: sql.NullString{String: "rec-123", Valid: true},
					RecommendationStatus: sql.NullString{
						String: "approval_requested",
						Valid:  true,
					},
					ApprovalID:     sql.NullString{String: "approval-123", Valid: true},
					ApprovalStatus: sql.NullString{String: "pending", Valid: true},
					RunID:          sql.NullString{String: "run-123", Valid: true},
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/work-orders", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `/app/inbound-requests/REQ-000123">REQ-000123</a>`) {
		t.Fatalf("expected request continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/proposals/rec-123">approval_requested</a>`) {
		t.Fatalf("expected proposal continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/approvals/approval-123">pending</a>`) {
		t.Fatalf("expected approval continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/run:run-123#run-run-123">AI run</a>`) {
		t.Fatalf("expected AI run continuity link, body=%s", body)
	}
}

func TestHandleWebWorkOrderDetailAddsUpstreamContinuityLinks(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			getWorkOrderReview: func(context.Context, reporting.GetWorkOrderReviewInput) (reporting.WorkOrderReview, error) {
				return reporting.WorkOrderReview{
					WorkOrderID:          "work-order-123",
					WorkOrderCode:        "WO-123",
					Title:                "Linked work order",
					Summary:              "Execution review continuity",
					Status:               "in_progress",
					DocumentID:           "doc-123",
					DocumentStatus:       "approved",
					RequestReference:     sql.NullString{String: "REQ-000123", Valid: true},
					RecommendationID:     sql.NullString{String: "rec-123", Valid: true},
					RecommendationStatus: sql.NullString{String: "approval_requested", Valid: true},
					ApprovalID:           sql.NullString{String: "approval-123", Valid: true},
					ApprovalStatus:       sql.NullString{String: "pending", Valid: true},
					RunID:                sql.NullString{String: "run-123", Valid: true},
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
	if !strings.Contains(body, `/app/inbound-requests/REQ-000123">REQ-000123</a>`) {
		t.Fatalf("expected request continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/proposals/rec-123">approval_requested</a>`) {
		t.Fatalf("expected proposal continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/approvals/approval-123">pending</a>`) {
		t.Fatalf("expected approval continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/run:run-123#run-run-123">AI run</a>`) {
		t.Fatalf("expected AI run continuity link, body=%s", body)
	}
}

func TestHandleWebInventoryDetailAddsFocusedContinuityLinks(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInventoryMovements: func(context.Context, reporting.ListInventoryMovementsInput) ([]reporting.InventoryMovementReview, error) {
				return []reporting.InventoryMovementReview{
					{
						MovementID:              "movement-123",
						MovementNumber:          42,
						DocumentID:              sql.NullString{String: "doc-123", Valid: true},
						DocumentTitle:           sql.NullString{String: "Inventory issue", Valid: true},
						DocumentNumber:          sql.NullString{String: "INV-42", Valid: true},
						DocumentStatus:          sql.NullString{String: "posted", Valid: true},
						ApprovalID:              sql.NullString{String: "approval-123", Valid: true},
						ApprovalStatus:          sql.NullString{String: "approved", Valid: true},
						ApprovalQueueCode:       sql.NullString{String: "inventory_review", Valid: true},
						RequestReference:        sql.NullString{String: "REQ-000123", Valid: true},
						RecommendationID:        sql.NullString{String: "rec-123", Valid: true},
						RecommendationStatus:    sql.NullString{String: "approved", Valid: true},
						RunID:                   sql.NullString{String: "run-123", Valid: true},
						ItemID:                  "item-123",
						ItemSKU:                 "MAT-123",
						ItemName:                "Copper pipe",
						ItemRole:                "material",
						MovementType:            "issue",
						MovementPurpose:         "execution",
						UsageClassification:     "billable",
						SourceLocationID:        sql.NullString{String: "loc-src", Valid: true},
						SourceLocationCode:      sql.NullString{String: "MAIN", Valid: true},
						SourceLocationName:      sql.NullString{String: "Main store", Valid: true},
						DestinationLocationID:   sql.NullString{String: "loc-dst", Valid: true},
						DestinationLocationCode: sql.NullString{String: "VAN-1", Valid: true},
						DestinationLocationName: sql.NullString{String: "Truck stock", Valid: true},
						QuantityMilli:           500,
						CreatedByUserID:         "user-123",
						CreatedAt:               time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
						ReferenceNote:           "WO issue",
					},
				}, nil
			},
			listInventoryReconciliation: func(context.Context, reporting.ListInventoryReconciliationInput) ([]reporting.InventoryReconciliationItem, error) {
				return []reporting.InventoryReconciliationItem{
					{
						DocumentID:              "doc-123",
						DocumentTypeCode:        "inventory_issue",
						DocumentTitle:           "Inventory issue",
						DocumentStatus:          "posted",
						ApprovalID:              sql.NullString{String: "approval-123", Valid: true},
						ApprovalStatus:          sql.NullString{String: "approved", Valid: true},
						ApprovalQueueCode:       sql.NullString{String: "inventory_review", Valid: true},
						RequestReference:        sql.NullString{String: "REQ-000123", Valid: true},
						RecommendationID:        sql.NullString{String: "rec-123", Valid: true},
						RecommendationStatus:    sql.NullString{String: "approved", Valid: true},
						RunID:                   sql.NullString{String: "run-123", Valid: true},
						LineNumber:              1,
						MovementID:              "movement-123",
						MovementNumber:          42,
						ItemID:                  "item-123",
						ItemSKU:                 "MAT-123",
						ItemName:                "Copper pipe",
						WorkOrderID:             sql.NullString{String: "work-order-123", Valid: true},
						WorkOrderCode:           sql.NullString{String: "WO-123", Valid: true},
						ExecutionLinkStatus:     sql.NullString{String: "linked", Valid: true},
						JournalEntryID:          sql.NullString{String: "entry-123", Valid: true},
						JournalEntryNumber:      sql.NullInt64{Int64: 91, Valid: true},
						AccountingHandoffStatus: sql.NullString{String: "posted", Valid: true},
						MovementCreatedAt:       time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/inventory/movement-123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `/app/review/inventory/items/item-123">Open item review</a>`) {
		t.Fatalf("expected exact item review link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory?item_id=item-123#movement-history">Item movement history</a>`) {
		t.Fatalf("expected item movement history link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory?item_id=item-123#stock-balances">Stock balances</a>`) {
		t.Fatalf("expected item stock-balance link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory?location_id=loc-src#movement-history">Location movements</a>`) {
		t.Fatalf("expected source location movement link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory?location_id=loc-dst#movement-history">Location movements</a>`) {
		t.Fatalf("expected destination location movement link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory/locations/loc-src`) {
		t.Fatalf("expected source location detail link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory/locations/loc-dst`) {
		t.Fatalf("expected destination location detail link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory?document_id=doc-123#reconciliation">Document reconciliation</a>`) {
		t.Fatalf("expected document reconciliation link, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/REQ-000123">REQ-000123</a>`) {
		t.Fatalf("expected request continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/proposals/rec-123">Proposal</a>`) {
		t.Fatalf("expected proposal continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/approvals/approval-123">inventory_review</a>`) {
		t.Fatalf("expected approval continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/run:run-123#run-run-123">AI run</a>`) {
		t.Fatalf("expected AI run continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/accounting?document_id=doc-123">Accounting review</a>`) {
		t.Fatalf("expected accounting review link from source document block, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/work-orders/work-order-123">WO-123</a>`) {
		t.Fatalf("expected work-order link in reconciliation rows, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/accounting/entry-123">Entry #91</a>`) {
		t.Fatalf("expected accounting entry link in reconciliation rows, body=%s", body)
	}
}

func TestHandleWebInventoryItemDetailRendersExactItemReviewStop(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInventoryStock: func(context.Context, reporting.ListInventoryStockInput) ([]reporting.InventoryStockItem, error) {
				return []reporting.InventoryStockItem{{
					ItemID:       "item-123",
					ItemSKU:      "MAT-123",
					ItemName:     "Copper pipe",
					ItemRole:     "material",
					LocationID:   "loc-123",
					LocationCode: "MAIN",
					LocationName: "Main store",
					LocationRole: "warehouse",
					OnHandMilli:  1200,
				}}, nil
			},
			listInventoryMovements: func(context.Context, reporting.ListInventoryMovementsInput) ([]reporting.InventoryMovementReview, error) {
				return []reporting.InventoryMovementReview{{
					MovementID:              "movement-123",
					MovementNumber:          42,
					ItemID:                  "item-123",
					ItemSKU:                 "MAT-123",
					ItemName:                "Copper pipe",
					ItemRole:                "material",
					MovementType:            "issue",
					SourceLocationID:        sql.NullString{String: "loc-123", Valid: true},
					SourceLocationCode:      sql.NullString{String: "MAIN", Valid: true},
					DestinationLocationID:   sql.NullString{String: "loc-456", Valid: true},
					DestinationLocationCode: sql.NullString{String: "VAN-1", Valid: true},
					QuantityMilli:           300,
					CreatedByUserID:         "user-123",
					CreatedAt:               time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
				}}, nil
			},
			listInventoryReconciliation: func(context.Context, reporting.ListInventoryReconciliationInput) ([]reporting.InventoryReconciliationItem, error) {
				return []reporting.InventoryReconciliationItem{{
					DocumentID:         "doc-123",
					DocumentTitle:      "Inventory issue",
					DocumentTypeCode:   "inventory_issue",
					DocumentStatus:     "posted",
					LineNumber:         1,
					MovementID:         "movement-123",
					MovementNumber:     42,
					WorkOrderID:        sql.NullString{String: "work-order-123", Valid: true},
					WorkOrderCode:      sql.NullString{String: "WO-123", Valid: true},
					JournalEntryID:     sql.NullString{String: "entry-123", Valid: true},
					JournalEntryNumber: sql.NullInt64{Int64: 91, Valid: true},
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/inventory/items/item-123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `/app/review/inventory?item_id=item-123#stock-balances`) {
		t.Fatalf("expected filtered inventory back-link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory/locations/loc-123`) {
		t.Fatalf("expected location continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory/movement-123`) {
		t.Fatalf("expected movement continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/accounting/entry-123">Entry #91</a>`) {
		t.Fatalf("expected accounting continuity link, body=%s", body)
	}
}

func TestHandleWebInventoryLocationDetailRendersExactLocationReviewStop(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInventoryStock: func(context.Context, reporting.ListInventoryStockInput) ([]reporting.InventoryStockItem, error) {
				return []reporting.InventoryStockItem{{
					ItemID:       "item-123",
					ItemSKU:      "MAT-123",
					ItemName:     "Copper pipe",
					ItemRole:     "material",
					LocationID:   "loc-123",
					LocationCode: "MAIN",
					LocationName: "Main store",
					LocationRole: "warehouse",
					OnHandMilli:  1200,
				}}, nil
			},
			listInventoryMovements: func(context.Context, reporting.ListInventoryMovementsInput) ([]reporting.InventoryMovementReview, error) {
				return []reporting.InventoryMovementReview{{
					MovementID:              "movement-123",
					MovementNumber:          42,
					ItemID:                  "item-123",
					ItemSKU:                 "MAT-123",
					ItemName:                "Copper pipe",
					ItemRole:                "material",
					MovementType:            "issue",
					SourceLocationCode:      sql.NullString{String: "MAIN", Valid: true},
					DestinationLocationCode: sql.NullString{String: "VAN-1", Valid: true},
					QuantityMilli:           300,
					CreatedByUserID:         "user-123",
					CreatedAt:               time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/inventory/locations/loc-123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `/app/review/inventory?location_id=loc-123#stock-balances`) {
		t.Fatalf("expected filtered inventory back-link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory/items/item-123`) {
		t.Fatalf("expected item continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory/movement-123`) {
		t.Fatalf("expected movement continuity link, body=%s", body)
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

func TestHandleWebDocumentsAddUpstreamRequestAndProposalLinks(t *testing.T) {
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
						RequestReference:  sql.NullString{String: "REQ-000123", Valid: true},
						RecommendationID:  sql.NullString{String: "rec-123", Valid: true},
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
	if !strings.Contains(body, `/app/inbound-requests/REQ-000123">REQ-000123</a>`) {
		t.Fatalf("expected request link in document review, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/proposals/rec-123">Proposal</a>`) {
		t.Fatalf("expected proposal link in document review, body=%s", body)
	}
}

func TestHandleWebApprovalDetailAddsUpstreamProposalContinuity(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listApprovalQueue: func(context.Context, reporting.ListApprovalQueueInput) ([]reporting.ApprovalQueueEntry, error) {
				return []reporting.ApprovalQueueEntry{
					{
						ApprovalID:           "approval-123",
						QueueCode:            "finance_review",
						QueueStatus:          "pending",
						ApprovalStatus:       "pending",
						RequestedAt:          time.Date(2026, 3, 27, 11, 0, 0, 0, time.UTC),
						RequestedByUserID:    "user-123",
						DocumentID:           "doc-123",
						DocumentTypeCode:     "invoice",
						DocumentTitle:        "Invoice proposal",
						DocumentStatus:       "submitted",
						RequestReference:     sql.NullString{String: "REQ-000123", Valid: true},
						RecommendationID:     sql.NullString{String: "rec-123", Valid: true},
						RecommendationStatus: sql.NullString{String: "approval_requested", Valid: true},
						RunID:                sql.NullString{String: "run-123", Valid: true},
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/approvals/approval-123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `/app/inbound-requests/REQ-000123">REQ-000123</a>`) {
		t.Fatalf("expected request continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/proposals/rec-123">Proposal</a>`) {
		t.Fatalf("expected proposal continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/run:run-123#run-run-123">AI run</a>`) {
		t.Fatalf("expected AI run continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/proposals/rec-123">approval_requested</a>`) {
		t.Fatalf("expected proposal status link in linked-record table, body=%s", body)
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

func TestHandleWebProposalDetailShowsRequestApprovalFormWhenDocumentExists(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listProcessedProposals: func(context.Context, reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error) {
				return []reporting.ProcessedProposalReview{
					{
						RequestReference:     "REQ-000123",
						RequestStatus:        "processed",
						RecommendationID:     "rec-123",
						RunID:                "run-123",
						RecommendationType:   "request_approval",
						RecommendationStatus: "proposed",
						Summary:              "Request finance approval for the invoice.",
						SuggestedQueueCode:   sql.NullString{String: "finance_review", Valid: true},
						DocumentID:           sql.NullString{String: "doc-123", Valid: true},
						DocumentTitle:        sql.NullString{String: "Invoice proposal", Valid: true},
						DocumentTypeCode:     sql.NullString{String: "invoice", Valid: true},
						DocumentStatus:       sql.NullString{String: "submitted", Valid: true},
						CreatedAt:            time.Date(2026, 3, 27, 11, 0, 0, 0, time.UTC),
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/proposals/rec-123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `/app/review/proposals/rec-123/request-approval`) {
		t.Fatalf("expected request-approval form action, body=%s", body)
	}
	if !strings.Contains(body, `name="queue_code" value="finance_review"`) {
		t.Fatalf("expected suggested queue value, body=%s", body)
	}
	if !strings.Contains(body, `Open document`) {
		t.Fatalf("expected document continuity action, body=%s", body)
	}
}

func TestHandleWebAppDashboardAddsExactProposalAndApprovalLinks(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listApprovalQueue: func(context.Context, reporting.ListApprovalQueueInput) ([]reporting.ApprovalQueueEntry, error) {
				return []reporting.ApprovalQueueEntry{
					{
						ApprovalID:     "approval-123",
						QueueCode:      "finance_review",
						QueueStatus:    "pending",
						ApprovalStatus: "pending",
						DocumentID:     "doc-123",
						DocumentTitle:  "Invoice proposal",
					},
				}, nil
			},
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
	if !strings.Contains(body, `/app/review/approvals?queue_code=finance_review&amp;status=pending">finance_review</a>`) {
		t.Fatalf("expected dashboard filtered approval queue link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/approvals/approval-123">Open exact approval</a>`) {
		t.Fatalf("expected dashboard approval-detail continuation link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/audit?entity_type=documents.document&amp;entity_id=doc-123">Audit trail</a>`) {
		t.Fatalf("expected dashboard document audit continuity link, body=%s", body)
	}
}

func TestHandleWebAppDashboardAddsInboundStatusAndRunContinuityLinks(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInboundRequestStatusSummary: func(context.Context, identityaccess.Actor) ([]reporting.InboundRequestStatusSummary, error) {
				return []reporting.InboundRequestStatusSummary{
					{
						Status:          "queued",
						RequestCount:    3,
						MessageCount:    4,
						AttachmentCount: 2,
						LatestUpdatedAt: time.Date(2026, 3, 27, 9, 0, 0, 0, time.UTC),
					},
				}, nil
			},
			listInboundRequests: func(context.Context, reporting.ListInboundRequestsInput) ([]reporting.InboundRequestReview, error) {
				return []reporting.InboundRequestReview{
					{
						RequestID:        "request-123",
						RequestReference: "REQ-000123",
						Status:           "processed",
						Channel:          "browser",
						MessageCount:     2,
						AttachmentCount:  1,
						LastRunID:        sql.NullString{String: "run-123", Valid: true},
						UpdatedAt:        time.Date(2026, 3, 27, 9, 5, 0, 0, time.UTC),
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
	if !strings.Contains(body, `/app/review/inbound-requests?status=queued">Queued requests</a>`) {
		t.Fatalf("expected dashboard status-summary continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/run:run-123#run-run-123">Open latest run</a>`) {
		t.Fatalf("expected dashboard latest-run continuity link, body=%s", body)
	}
}

func TestHandleWebAppDashboardUsesSharedDashboardSnapshot(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			getDashboardSnapshot: func(context.Context, identityaccess.Actor, int, int, int) (reporting.DashboardSnapshot, error) {
				return reporting.DashboardSnapshot{
					Navigation: reporting.WorkflowNavigationSnapshot{
						InboundSummary: []reporting.InboundRequestStatusSummary{
							{
								Status:          "queued",
								RequestCount:    2,
								LatestUpdatedAt: time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC),
							},
						},
					},
					InboundRequests: []reporting.InboundRequestReview{
						{
							RequestReference: "REQ-009900",
							Status:           "queued",
							UpdatedAt:        time.Date(2026, 4, 1, 9, 5, 0, 0, time.UTC),
						},
					},
				}, nil
			},
			listInboundRequests: func(context.Context, reporting.ListInboundRequestsInput) ([]reporting.InboundRequestReview, error) {
				t.Fatal("dashboard handler should not compose recent requests directly")
				return nil, nil
			},
			listProcessedProposals: func(context.Context, reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error) {
				t.Fatal("dashboard handler should not compose proposals directly")
				return nil, nil
			},
			listInboundRequestStatusSummary: func(context.Context, identityaccess.Actor) ([]reporting.InboundRequestStatusSummary, error) {
				t.Fatal("dashboard handler should not compose navigation summary directly")
				return nil, nil
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
	if !strings.Contains(recorder.Body.String(), `REQ-009900`) {
		t.Fatalf("expected dashboard snapshot content, body=%s", recorder.Body.String())
	}
}

func TestHandleWebAppDashboardAddsStatusSpecificEntryPointActions(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInboundRequestStatusSummary: func(context.Context, identityaccess.Actor) ([]reporting.InboundRequestStatusSummary, error) {
				return []reporting.InboundRequestStatusSummary{
					{
						Status:          "failed",
						RequestCount:    1,
						MessageCount:    1,
						AttachmentCount: 0,
						LatestUpdatedAt: time.Date(2026, 3, 27, 9, 0, 0, 0, time.UTC),
					},
					{
						Status:          "draft",
						RequestCount:    2,
						MessageCount:    3,
						AttachmentCount: 1,
						LatestUpdatedAt: time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
					},
					{
						Status:          "cancelled",
						RequestCount:    1,
						MessageCount:    1,
						AttachmentCount: 0,
						LatestUpdatedAt: time.Date(2026, 3, 27, 8, 0, 0, 0, time.UTC),
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
	if !strings.Contains(body, `Resume parked drafts before they enter the queue.`) {
		t.Fatalf("expected draft entry-point blurb, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inbound-requests?status=draft`) {
		t.Fatalf("expected draft entry-point action, body=%s", body)
	}
	if !strings.Contains(body, `Inspect failed requests, understand the break, and restart follow-up work.`) {
		t.Fatalf("expected failed entry-point blurb, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inbound-requests?status=failed`) {
		t.Fatalf("expected failed entry-point action, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inbound-requests?status=cancelled`) {
		t.Fatalf("expected cancelled entry-point action, body=%s", body)
	}
	if strings.Index(body, `Continue drafts`) > strings.Index(body, `Review failures`) {
		t.Fatalf("expected draft status link to sort ahead of failed status link, body=%s", body)
	}
}

func TestHandleWebAppDashboardRendersRefreshedEnterpriseShell(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInboundRequestStatusSummary: func(context.Context, identityaccess.Actor) ([]reporting.InboundRequestStatusSummary, error) {
				return []reporting.InboundRequestStatusSummary{
					{
						Status:          "queued",
						RequestCount:    2,
						MessageCount:    3,
						AttachmentCount: 1,
						LatestUpdatedAt: time.Date(2026, 3, 30, 9, 0, 0, 0, time.UTC),
					},
				}, nil
			},
			listProcessedProposalStatusSummary: func(context.Context, identityaccess.Actor) ([]reporting.ProcessedProposalStatusSummary, error) {
				return []reporting.ProcessedProposalStatusSummary{
					{RecommendationStatus: "approval_requested", ProposalCount: 1},
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
	if !strings.Contains(body, `class="brand-mark">WA</span>`) {
		t.Fatalf("expected thin top-bar brand mark, body=%s", body)
	}
	if !strings.Contains(body, `Workflow App`) {
		t.Fatalf("expected workflow app branding, body=%s", body)
	}
	if !strings.Contains(body, `Workflow destinations`) {
		t.Fatalf("expected workflow destination navigation band, body=%s", body)
	}
	if !strings.Contains(body, `>User</summary>`) {
		t.Fatalf("expected compact user utility menu, body=%s", body)
	}
	if !strings.Contains(body, `Search routes`) {
		t.Fatalf("expected route-catalog search from shell utility menu, body=%s", body)
	}
	if !strings.Contains(body, `/app/settings" class="nav-link">Settings</a>`) {
		t.Fatalf("expected settings utility link, body=%s", body)
	}
	if !strings.Contains(body, `Persist-first operator shell`) {
		t.Fatalf("expected thin shell posture copy, body=%s", body)
	}
	if !strings.Contains(body, `/app/operations" class="nav-link">Operations</a>`) {
		t.Fatalf("expected operations landing in shell navigation, body=%s", body)
	}
	if !strings.Contains(body, `/app/review" class="nav-link">Review</a>`) {
		t.Fatalf("expected review landing in shell navigation, body=%s", body)
	}
	if !strings.Contains(body, `/app/inventory" class="nav-link">Inventory</a>`) {
		t.Fatalf("expected inventory landing in shell navigation, body=%s", body)
	}
	if !strings.Contains(body, `Operator home`) {
		t.Fatalf("expected role-aware home headline, body=%s", body)
	}
	if !strings.Contains(body, `/app/submit-inbound-request">Start a new request</a>`) {
		t.Fatalf("expected role-aware intake action, body=%s", body)
	}
	if !strings.Contains(body, `/app/routes">Open route catalog</a>`) {
		t.Fatalf("expected secondary route-catalog action, body=%s", body)
	}
	if !strings.Contains(body, `Route directory`) {
		t.Fatalf("expected plain route-directory section on dashboard, body=%s", body)
	}
	if strings.Contains(body, `action="/app/inbound-requests" enctype="multipart/form-data"`) {
		t.Fatalf("expected dashboard to stop embedding the inbound request form, body=%s", body)
	}
}

func TestHandleWebRouteCatalogFiltersRoutesByQueryAndRole(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				return testSessionContext(), nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/routes?q=approval", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `Searchable route discovery`) {
		t.Fatalf("expected query-specific heading, body=%s", body)
	}
	if !strings.Contains(body, `Approval review`) {
		t.Fatalf("expected approval review route in search results, body=%s", body)
	}
	if strings.Contains(body, `/app/admin`) {
		t.Fatalf("expected non-admin route catalog to hide admin route, body=%s", body)
	}
}

func TestHandleWebRouteCatalogMatchesOperatorIntentTerms(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				return testSessionContext(), nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/routes?q=pending+approvals", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `Searchable route discovery`) {
		t.Fatalf("expected multi-term heading, body=%s", body)
	}
	if !strings.Contains(body, `Approval review`) {
		t.Fatalf("expected approval review result for operator-intent query, body=%s", body)
	}
}

func TestHandleWebSettingsShowsRoleAwareHomeActions(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInboundRequestStatusSummary: func(context.Context, identityaccess.Actor) ([]reporting.InboundRequestStatusSummary, error) {
				return []reporting.InboundRequestStatusSummary{
					{Status: "pending", RequestCount: 0},
					{Status: "queued", RequestCount: 3},
				}, nil
			},
			listProcessedProposalStatusSummary: func(context.Context, identityaccess.Actor) ([]reporting.ProcessedProposalStatusSummary, error) {
				return []reporting.ProcessedProposalStatusSummary{
					{RecommendationStatus: "approval_requested", ProposalCount: 2},
				}, nil
			},
			listApprovalQueue: func(context.Context, reporting.ListApprovalQueueInput) ([]reporting.ApprovalQueueEntry, error) {
				return []reporting.ApprovalQueueEntry{
					{ApprovalID: "approval-1"},
					{ApprovalID: "approval-2"},
				}, nil
			},
		},
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				ctx := testSessionContext()
				ctx.RoleCode = identityaccess.RoleApprover
				return ctx, nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/settings", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `User-scoped settings and continuity`) {
		t.Fatalf("expected settings heading, body=%s", body)
	}
	if !strings.Contains(body, `Settings stays user-scoped`) {
		t.Fatalf("expected explicit settings ownership copy, body=%s", body)
	}
	if !strings.Contains(body, `Review pending approvals (2)`) {
		t.Fatalf("expected role-aware approval shortcut, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/proposals?status=approval_requested`) {
		t.Fatalf("expected approval-ready proposal shortcut, body=%s", body)
	}
	if strings.Contains(body, `Open admin maintenance hub`) {
		t.Fatalf("expected non-admin settings page to avoid admin continuation, body=%s", body)
	}
}

func TestHandleWebAdminRequiresAdminRole(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				return testSessionContext(), nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/admin", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if location := recorder.Header().Get("Location"); !strings.Contains(location, "admin+surface+requires+admin+role") {
		t.Fatalf("expected admin-role redirect, got %s", location)
	}
}

func TestHandleWebAdminShowsMaintenanceHubForAdmin(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				ctx := testSessionContext()
				ctx.RoleCode = identityaccess.RoleAdmin
				return ctx, nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/admin", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `Privileged maintenance hub`) {
		t.Fatalf("expected admin hub heading, body=%s", body)
	}
	if !strings.Contains(body, `Maintenance families`) {
		t.Fatalf("expected maintenance families section, body=%s", body)
	}
	if !strings.Contains(body, `Accounting setup`) {
		t.Fatalf("expected accounting setup family, body=%s", body)
	}
	if !strings.Contains(body, `Current slice: admin-only browser and API maintenance now expose bounded list, create, and period-close controls`) {
		t.Fatalf("expected current-slice accounting setup copy, body=%s", body)
	}
	if !strings.Contains(body, `Party setup`) {
		t.Fatalf("expected party setup family, body=%s", body)
	}
	if !strings.Contains(body, `Inventory setup`) {
		t.Fatalf("expected inventory setup family, body=%s", body)
	}
}

func TestHandleWebAdminAccountingShowsSetupForms(t *testing.T) {
	handler := newAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		nil,
		stubAccountingAdminService{
			listLedgerAccounts: func(context.Context, accounting.ListLedgerAccountsInput) ([]accounting.LedgerAccount, error) {
				return []accounting.LedgerAccount{{ID: "acct-1", Code: "1100", Name: "Accounts Receivable", AccountClass: accounting.AccountClassAsset, ControlType: accounting.ControlTypeReceivable, Status: "active"}}, nil
			},
			listTaxCodes: func(context.Context, accounting.ListTaxCodesInput) ([]accounting.TaxCode, error) {
				return []accounting.TaxCode{{ID: "tax-1", Code: "GST18", Name: "GST 18%", TaxType: accounting.TaxTypeGST, RateBasisPoints: 1800, Status: "active"}}, nil
			},
			listAccountingPeriods: func(context.Context, accounting.ListAccountingPeriodsInput) ([]accounting.AccountingPeriod, error) {
				return []accounting.AccountingPeriod{{ID: "period-1", PeriodCode: "FY2026-04", StartOn: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), EndOn: time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC), Status: "open"}}, nil
			},
		},
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				ctx := testSessionContext()
				ctx.RoleCode = identityaccess.RoleAdmin
				return ctx, nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/admin/accounting", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `Accounting setup maintenance`) {
		t.Fatalf("expected accounting setup heading, body=%s", body)
	}
	if !strings.Contains(body, `Create ledger account`) {
		t.Fatalf("expected ledger account form, body=%s", body)
	}
	if !strings.Contains(body, `Create tax code`) {
		t.Fatalf("expected tax code form, body=%s", body)
	}
	if !strings.Contains(body, `Create accounting period`) {
		t.Fatalf("expected accounting period form, body=%s", body)
	}
	if !strings.Contains(body, `Accounts Receivable`) || !strings.Contains(body, `GST18`) || !strings.Contains(body, `FY2026-04`) {
		t.Fatalf("expected maintenance lists to render seeded data, body=%s", body)
	}
	if !strings.Contains(body, `Mark inactive`) {
		t.Fatalf("expected status controls to render, body=%s", body)
	}
}

func TestHandleWebCreateLedgerAccountRedirectsWithNotice(t *testing.T) {
	var captured accounting.CreateLedgerAccountInput
	handler := newAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		nil,
		stubAccountingAdminService{
			createLedgerAccount: func(_ context.Context, input accounting.CreateLedgerAccountInput) (accounting.LedgerAccount, error) {
				captured = input
				return accounting.LedgerAccount{ID: "acct-1", Code: input.Code, Name: input.Name}, nil
			},
			listLedgerAccounts: func(context.Context, accounting.ListLedgerAccountsInput) ([]accounting.LedgerAccount, error) {
				return nil, nil
			},
			listTaxCodes: func(context.Context, accounting.ListTaxCodesInput) ([]accounting.TaxCode, error) { return nil, nil },
			listAccountingPeriods: func(context.Context, accounting.ListAccountingPeriodsInput) ([]accounting.AccountingPeriod, error) {
				return nil, nil
			},
		},
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				ctx := testSessionContext()
				ctx.RoleCode = identityaccess.RoleAdmin
				return ctx, nil
			},
		},
	)

	form := url.Values{
		"code":                  {"AR1000"},
		"name":                  {"Accounts Receivable"},
		"account_class":         {accounting.AccountClassAsset},
		"control_type":          {accounting.ControlTypeReceivable},
		"allows_direct_posting": {"true"},
	}
	req := httptest.NewRequest(http.MethodPost, "/app/admin/accounting/ledger-accounts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if location := recorder.Header().Get("Location"); !strings.Contains(location, "/app/admin/accounting?notice=Ledger+account+created.") {
		t.Fatalf("expected success redirect, got %s", location)
	}
	if captured.Code != "AR1000" || captured.ControlType != accounting.ControlTypeReceivable || !captured.AllowsDirectPosting {
		t.Fatalf("unexpected captured input: %+v", captured)
	}
}

func TestHandleWebLedgerAccountStatusRedirectsWithNotice(t *testing.T) {
	var captured accounting.UpdateLedgerAccountStatusInput
	handler := newAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		nil,
		stubAccountingAdminService{
			updateLedgerStatus: func(_ context.Context, input accounting.UpdateLedgerAccountStatusInput) (accounting.LedgerAccount, error) {
				captured = input
				return accounting.LedgerAccount{ID: input.AccountID, Status: input.Status}, nil
			},
		},
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				ctx := testSessionContext()
				ctx.RoleCode = identityaccess.RoleAdmin
				return ctx, nil
			},
		},
	)

	form := url.Values{"status": {accounting.StatusInactive}}
	req := httptest.NewRequest(http.MethodPost, "/app/admin/accounting/ledger-accounts/acct-1/status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if location := recorder.Header().Get("Location"); !strings.Contains(location, "/app/admin/accounting?notice=Ledger+account+marked+inactive.") {
		t.Fatalf("expected success redirect, got %s", location)
	}
	if captured.AccountID != "acct-1" || captured.Status != accounting.StatusInactive {
		t.Fatalf("unexpected captured input: %+v", captured)
	}
}

func TestHandleWebAdminPartiesShowsSetupAndDetail(t *testing.T) {
	handler := newAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		nil,
		stubAccountingAdminService{},
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				ctx := testSessionContext()
				ctx.RoleCode = identityaccess.RoleAdmin
				return ctx, nil
			},
		},
		stubPartiesAdminService{
			listParties: func(context.Context, parties.ListPartiesInput) ([]parties.Party, error) {
				return []parties.Party{{ID: "party-1", PartyCode: "CUST-100", DisplayName: "Northwind Service", PartyKind: parties.PartyKindCustomer, Status: parties.StatusActive}}, nil
			},
			getParty: func(context.Context, parties.GetPartyInput) (parties.Party, error) {
				return parties.Party{ID: "party-1", PartyCode: "CUST-100", DisplayName: "Northwind Service", LegalName: "Northwind Service Pvt Ltd", PartyKind: parties.PartyKindCustomer, Status: parties.StatusActive}, nil
			},
			listContacts: func(context.Context, parties.ListContactsInput) ([]parties.Contact, error) {
				return []parties.Contact{{ID: "contact-1", PartyID: "party-1", FullName: "Asha Nair", IsPrimary: true}}, nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/admin/parties/party-1", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `Customer and party setup`) {
		t.Fatalf("expected party setup heading, body=%s", body)
	}
	if !strings.Contains(body, `Create party`) {
		t.Fatalf("expected create party form, body=%s", body)
	}
	if !strings.Contains(body, `Create contact`) {
		t.Fatalf("expected contact creation form, body=%s", body)
	}
	if !strings.Contains(body, `Northwind Service`) || !strings.Contains(body, `Asha Nair`) {
		t.Fatalf("expected party detail and contact list, body=%s", body)
	}
	if !strings.Contains(body, `Mark party inactive`) {
		t.Fatalf("expected party status control, body=%s", body)
	}
}

func TestHandleWebAdminPartiesCreateRedirectsWithNotice(t *testing.T) {
	var captured parties.CreatePartyInput
	handler := newAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		nil,
		stubAccountingAdminService{},
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				ctx := testSessionContext()
				ctx.RoleCode = identityaccess.RoleAdmin
				return ctx, nil
			},
		},
		stubPartiesAdminService{
			createParty: func(_ context.Context, input parties.CreatePartyInput) (parties.Party, error) {
				captured = input
				return parties.Party{ID: "party-1", PartyCode: input.PartyCode, DisplayName: input.DisplayName}, nil
			},
		},
	)

	form := url.Values{
		"party_code":   {"CUST-100"},
		"display_name": {"Northwind Service"},
		"legal_name":   {"Northwind Service Pvt Ltd"},
		"party_kind":   {parties.PartyKindCustomer},
	}
	req := httptest.NewRequest(http.MethodPost, "/app/admin/parties", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if location := recorder.Header().Get("Location"); !strings.Contains(location, "/app/admin/parties?notice=Party+created.") {
		t.Fatalf("expected success redirect, got %s", location)
	}
	if captured.PartyCode != "CUST-100" || captured.PartyKind != parties.PartyKindCustomer {
		t.Fatalf("unexpected captured input: %+v", captured)
	}
}

func TestHandleWebAdminPartyContactCreateRedirectsWithNotice(t *testing.T) {
	var captured parties.CreateContactInput
	handler := newAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		nil,
		stubAccountingAdminService{},
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				ctx := testSessionContext()
				ctx.RoleCode = identityaccess.RoleAdmin
				return ctx, nil
			},
		},
		stubPartiesAdminService{
			createContact: func(_ context.Context, input parties.CreateContactInput) (parties.Contact, error) {
				captured = input
				return parties.Contact{ID: "contact-1", PartyID: input.PartyID, FullName: input.FullName, IsPrimary: input.IsPrimary}, nil
			},
		},
	)

	form := url.Values{
		"full_name":  {"Asha Nair"},
		"role_title": {"Accounts"},
		"email":      {"asha@example.com"},
		"is_primary": {"true"},
	}
	req := httptest.NewRequest(http.MethodPost, "/app/admin/parties/party-1/contacts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if location := recorder.Header().Get("Location"); !strings.Contains(location, "/app/admin/parties/party-1?notice=Party+contact+created.") {
		t.Fatalf("expected success redirect, got %s", location)
	}
	if captured.PartyID != "party-1" || captured.FullName != "Asha Nair" || captured.Email != "asha@example.com" || !captured.IsPrimary {
		t.Fatalf("unexpected captured input: %+v", captured)
	}
}

func TestHandleWebAdminPartyStatusRedirectsWithNotice(t *testing.T) {
	var captured parties.UpdatePartyStatusInput
	handler := newAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		nil,
		stubAccountingAdminService{},
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				ctx := testSessionContext()
				ctx.RoleCode = identityaccess.RoleAdmin
				return ctx, nil
			},
		},
		stubPartiesAdminService{
			updateStatus: func(_ context.Context, input parties.UpdatePartyStatusInput) (parties.Party, error) {
				captured = input
				return parties.Party{ID: input.PartyID, Status: input.Status}, nil
			},
		},
	)

	form := url.Values{"status": {parties.StatusInactive}}
	req := httptest.NewRequest(http.MethodPost, "/app/admin/parties/party-1/status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if location := recorder.Header().Get("Location"); !strings.Contains(location, "/app/admin/parties/party-1?notice=Party+marked+inactive.") {
		t.Fatalf("expected success redirect, got %s", location)
	}
	if captured.PartyID != "party-1" || captured.Status != parties.StatusInactive {
		t.Fatalf("unexpected captured input: %+v", captured)
	}
}

func TestHandleWebAdminAccessShowsMembershipControls(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				ctx := testSessionContext()
				ctx.RoleCode = identityaccess.RoleAdmin
				return ctx, nil
			},
			listOrgUsers: func(context.Context, identityaccess.ListOrgUsersInput) ([]identityaccess.OrgUserMembership, error) {
				return []identityaccess.OrgUserMembership{{
					MembershipID:     "membership-1",
					UserID:           "user-2",
					UserEmail:        "operator@example.com",
					UserDisplayName:  "Operator One",
					UserStatus:       "active",
					RoleCode:         identityaccess.RoleOperator,
					MembershipStatus: "active",
				}}, nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/admin/access", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `User and role controls`) {
		t.Fatalf("expected access heading, body=%s", body)
	}
	if !strings.Contains(body, `operator@example.com`) || !strings.Contains(body, `Update role`) {
		t.Fatalf("expected membership table and role controls, body=%s", body)
	}
}

func TestHandleWebAdminAccessRoleUpdateRedirectsWithNotice(t *testing.T) {
	var captured identityaccess.UpdateMembershipRoleInput
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				ctx := testSessionContext()
				ctx.RoleCode = identityaccess.RoleAdmin
				return ctx, nil
			},
			updateMembershipRole: func(_ context.Context, input identityaccess.UpdateMembershipRoleInput) (identityaccess.OrgUserMembership, error) {
				captured = input
				return identityaccess.OrgUserMembership{MembershipID: input.MembershipID, RoleCode: input.RoleCode}, nil
			},
		},
	)

	form := url.Values{"role_code": {identityaccess.RoleApprover}}
	req := httptest.NewRequest(http.MethodPost, "/app/admin/access/users/membership-1/role", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if location := recorder.Header().Get("Location"); !strings.Contains(location, "/app/admin/access?notice=Membership+role+updated.") {
		t.Fatalf("expected success redirect, got %s", location)
	}
	if captured.MembershipID != "membership-1" || captured.RoleCode != identityaccess.RoleApprover {
		t.Fatalf("unexpected captured input: %+v", captured)
	}
}

func TestHandleWebAdminInventoryShowsSetupForms(t *testing.T) {
	handler := newAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		nil,
		stubAccountingAdminService{},
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				ctx := testSessionContext()
				ctx.RoleCode = identityaccess.RoleAdmin
				return ctx, nil
			},
		},
		stubInventoryAdminService{
			listItems: func(context.Context, inventoryops.ListItemsInput) ([]inventoryops.Item, error) {
				return []inventoryops.Item{{ID: "item-1", SKU: "PUMP-100", Name: "Warehouse Pump", ItemRole: inventoryops.ItemRoleTraceableEquipment, TrackingMode: inventoryops.TrackingModeSerial, Status: "active"}}, nil
			},
			listLocations: func(context.Context, inventoryops.ListLocationsInput) ([]inventoryops.Location, error) {
				return []inventoryops.Location{{ID: "loc-1", Code: "WH-A", Name: "Main Warehouse", LocationRole: inventoryops.LocationRoleWarehouse, Status: "active"}}, nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/admin/inventory", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `Inventory setup maintenance`) {
		t.Fatalf("expected inventory setup heading, body=%s", body)
	}
	if !strings.Contains(body, `Create inventory item`) || !strings.Contains(body, `Create inventory location`) {
		t.Fatalf("expected inventory setup forms, body=%s", body)
	}
	if !strings.Contains(body, `Warehouse Pump`) || !strings.Contains(body, `Main Warehouse`) {
		t.Fatalf("expected inventory master data to render, body=%s", body)
	}
	if !strings.Contains(body, `Mark inactive`) {
		t.Fatalf("expected inventory status controls, body=%s", body)
	}
}

func TestHandleWebAdminInventoryCreateRedirectsWithNotice(t *testing.T) {
	var capturedItem inventoryops.CreateItemInput
	var capturedLocation inventoryops.CreateLocationInput
	handler := newAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		nil,
		stubAccountingAdminService{},
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				ctx := testSessionContext()
				ctx.RoleCode = identityaccess.RoleAdmin
				return ctx, nil
			},
		},
		stubInventoryAdminService{
			createItem: func(_ context.Context, input inventoryops.CreateItemInput) (inventoryops.Item, error) {
				capturedItem = input
				return inventoryops.Item{ID: "item-1", SKU: input.SKU, Name: input.Name}, nil
			},
			createLocation: func(_ context.Context, input inventoryops.CreateLocationInput) (inventoryops.Location, error) {
				capturedLocation = input
				return inventoryops.Location{ID: "loc-1", Code: input.Code, Name: input.Name}, nil
			},
		},
	)

	itemForm := url.Values{
		"sku":           {"PUMP-100"},
		"name":          {"Warehouse Pump"},
		"item_role":     {inventoryops.ItemRoleTraceableEquipment},
		"tracking_mode": {inventoryops.TrackingModeSerial},
	}
	itemReq := httptest.NewRequest(http.MethodPost, "/app/admin/inventory/items", strings.NewReader(itemForm.Encode()))
	itemReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	itemReq.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	itemReq.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	itemRecorder := httptest.NewRecorder()
	handler.ServeHTTP(itemRecorder, itemReq)

	if itemRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected item status: got %d body=%s", itemRecorder.Code, itemRecorder.Body.String())
	}
	if location := itemRecorder.Header().Get("Location"); !strings.Contains(location, "/app/admin/inventory?notice=Inventory+item+created.") {
		t.Fatalf("expected item success redirect, got %s", location)
	}
	if capturedItem.SKU != "PUMP-100" || capturedItem.ItemRole != inventoryops.ItemRoleTraceableEquipment {
		t.Fatalf("unexpected captured item input: %+v", capturedItem)
	}

	locationForm := url.Values{
		"code":          {"WH-A"},
		"name":          {"Main Warehouse"},
		"location_role": {inventoryops.LocationRoleWarehouse},
	}
	locationReq := httptest.NewRequest(http.MethodPost, "/app/admin/inventory/locations", strings.NewReader(locationForm.Encode()))
	locationReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	locationReq.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	locationReq.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	locationRecorder := httptest.NewRecorder()
	handler.ServeHTTP(locationRecorder, locationReq)

	if locationRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected location status: got %d body=%s", locationRecorder.Code, locationRecorder.Body.String())
	}
	if location := locationRecorder.Header().Get("Location"); !strings.Contains(location, "/app/admin/inventory?notice=Inventory+location+created.") {
		t.Fatalf("expected location success redirect, got %s", location)
	}
	if capturedLocation.Code != "WH-A" || capturedLocation.LocationRole != inventoryops.LocationRoleWarehouse {
		t.Fatalf("unexpected captured location input: %+v", capturedLocation)
	}
}

func TestHandleWebAdminInventoryStatusRedirectsWithNotice(t *testing.T) {
	var capturedItem inventoryops.UpdateItemStatusInput
	var capturedLocation inventoryops.UpdateLocationStatusInput
	handler := newAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		nil,
		stubAccountingAdminService{},
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				ctx := testSessionContext()
				ctx.RoleCode = identityaccess.RoleAdmin
				return ctx, nil
			},
		},
		stubInventoryAdminService{
			updateItem: func(_ context.Context, input inventoryops.UpdateItemStatusInput) (inventoryops.Item, error) {
				capturedItem = input
				return inventoryops.Item{ID: input.ItemID, Status: input.Status}, nil
			},
			updateLocation: func(_ context.Context, input inventoryops.UpdateLocationStatusInput) (inventoryops.Location, error) {
				capturedLocation = input
				return inventoryops.Location{ID: input.LocationID, Status: input.Status}, nil
			},
		},
	)

	itemForm := url.Values{"status": {inventoryops.StatusInactive}}
	itemReq := httptest.NewRequest(http.MethodPost, "/app/admin/inventory/items/item-1/status", strings.NewReader(itemForm.Encode()))
	itemReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	itemReq.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	itemReq.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	itemRecorder := httptest.NewRecorder()
	handler.ServeHTTP(itemRecorder, itemReq)

	if itemRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected item status: got %d body=%s", itemRecorder.Code, itemRecorder.Body.String())
	}
	if location := itemRecorder.Header().Get("Location"); !strings.Contains(location, "/app/admin/inventory?notice=Inventory+item+marked+inactive.") {
		t.Fatalf("expected item success redirect, got %s", location)
	}
	if capturedItem.ItemID != "item-1" || capturedItem.Status != inventoryops.StatusInactive {
		t.Fatalf("unexpected captured item status input: %+v", capturedItem)
	}

	locationForm := url.Values{"status": {inventoryops.StatusInactive}}
	locationReq := httptest.NewRequest(http.MethodPost, "/app/admin/inventory/locations/loc-1/status", strings.NewReader(locationForm.Encode()))
	locationReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	locationReq.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	locationReq.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	locationRecorder := httptest.NewRecorder()
	handler.ServeHTTP(locationRecorder, locationReq)

	if locationRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected location status: got %d body=%s", locationRecorder.Code, locationRecorder.Body.String())
	}
	if location := locationRecorder.Header().Get("Location"); !strings.Contains(location, "/app/admin/inventory?notice=Inventory+location+marked+inactive.") {
		t.Fatalf("expected location success redirect, got %s", location)
	}
	if capturedLocation.LocationID != "loc-1" || capturedLocation.Status != inventoryops.StatusInactive {
		t.Fatalf("unexpected captured location status input: %+v", capturedLocation)
	}
}

func TestHandleWebReviewLandingGroupsRouteFamilies(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInboundRequestStatusSummary: func(context.Context, identityaccess.Actor) ([]reporting.InboundRequestStatusSummary, error) {
				return []reporting.InboundRequestStatusSummary{
					{Status: "queued", RequestCount: 2},
					{Status: "processed", RequestCount: 3},
				}, nil
			},
			listProcessedProposalStatusSummary: func(context.Context, identityaccess.Actor) ([]reporting.ProcessedProposalStatusSummary, error) {
				return []reporting.ProcessedProposalStatusSummary{
					{RecommendationStatus: "approval_requested", ProposalCount: 2},
				}, nil
			},
			listApprovalQueue: func(context.Context, reporting.ListApprovalQueueInput) ([]reporting.ApprovalQueueEntry, error) {
				return []reporting.ApprovalQueueEntry{
					{
						ApprovalID:       "approval-123",
						QueueCode:        "finance_review",
						DocumentID:       "doc-123",
						DocumentTitle:    "Pump invoice",
						RequestReference: sql.NullString{String: "REQ-000123", Valid: true},
						RecommendationID: sql.NullString{String: "rec-123", Valid: true},
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

	req := httptest.NewRequest(http.MethodGet, "/app/review", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "Review route families") {
		t.Fatalf("expected review landing heading, body=%s", body)
	}
	if !strings.Contains(body, `/app/review" class="nav-link is-active">Review</a>`) {
		t.Fatalf("expected review landing to activate the grouped review nav item, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inbound-requests`) || !strings.Contains(body, `/app/review/documents`) {
		t.Fatalf("expected landing links into grouped review routes, body=%s", body)
	}
	if !strings.Contains(body, `/app/inventory">Inventory</a>`) {
		t.Fatalf("expected review landing to point to the inventory domain landing, body=%s", body)
	}
}

func TestHandleWebOperationsLandingShowsFeedAndChatBundles(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInboundRequestStatusSummary: func(context.Context, identityaccess.Actor) ([]reporting.InboundRequestStatusSummary, error) {
				return []reporting.InboundRequestStatusSummary{{Status: "queued", RequestCount: 4}}, nil
			},
			listInboundRequests: func(context.Context, reporting.ListInboundRequestsInput) ([]reporting.InboundRequestReview, error) {
				return []reporting.InboundRequestReview{{
					RequestReference: "REQ-000123",
					Status:           "queued",
					Channel:          "browser",
					MessageCount:     2,
					AttachmentCount:  1,
					UpdatedAt:        time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC),
				}}, nil
			},
			listProcessedProposals: func(context.Context, reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error) {
				return []reporting.ProcessedProposalReview{{
					RecommendationID:     "rec-123",
					RequestReference:     "REQ-000123",
					RecommendationStatus: "approval_requested",
					Summary:              "Urgent warehouse pump review",
					CreatedAt:            time.Date(2026, 4, 1, 8, 0, 0, 0, time.UTC),
				}}, nil
			},
			listApprovalQueue: func(context.Context, reporting.ListApprovalQueueInput) ([]reporting.ApprovalQueueEntry, error) {
				return []reporting.ApprovalQueueEntry{{
					ApprovalID:     "approval-123",
					QueueCode:      "finance_review",
					DocumentID:     "doc-123",
					DocumentTitle:  "Pump invoice",
					ApprovalStatus: "pending",
					RequestedAt:    time.Date(2026, 4, 1, 7, 0, 0, 0, time.UTC),
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

	req := httptest.NewRequest(http.MethodGet, "/app/operations", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "Queue-driven operations") {
		t.Fatalf("expected operations landing heading, body=%s", body)
	}
	if !strings.Contains(body, `/app/operations" class="nav-link is-active">Operations</a>`) {
		t.Fatalf("expected operations landing nav activation, body=%s", body)
	}
	if !strings.Contains(body, `/app/operations-feed">Durable feed</a>`) {
		t.Fatalf("expected landing link to durable feed, body=%s", body)
	}
	if !strings.Contains(body, `/app/agent-chat">Coordinator chat</a>`) {
		t.Fatalf("expected landing link to agent chat, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inbound-requests?status=queued">Queued requests</a>`) {
		t.Fatalf("expected queued-request route on operations landing, body=%s", body)
	}
}

func TestHandleWebInventoryLandingShowsDomainBundles(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInventoryStock: func(context.Context, reporting.ListInventoryStockInput) ([]reporting.InventoryStockItem, error) {
				return []reporting.InventoryStockItem{{
					ItemID:       "item-123",
					ItemSKU:      "RPT-MAT-1",
					ItemName:     "Reporting material",
					LocationID:   "loc-123",
					LocationCode: "MAIN",
					LocationName: "Main store",
					OnHandMilli:  1200,
				}}, nil
			},
			listInventoryMovements: func(context.Context, reporting.ListInventoryMovementsInput) ([]reporting.InventoryMovementReview, error) {
				return []reporting.InventoryMovementReview{{MovementID: "move-123", MovementNumber: 123}}, nil
			},
			listInventoryReconciliation: func(context.Context, reporting.ListInventoryReconciliationInput) ([]reporting.InventoryReconciliationItem, error) {
				return []reporting.InventoryReconciliationItem{{
					MovementID:              "move-123",
					MovementNumber:          123,
					DocumentID:              "doc-123",
					DocumentNumber:          sql.NullString{String: "INV-123", Valid: true},
					ExecutionLinkStatus:     sql.NullString{String: "pending", Valid: true},
					AccountingHandoffStatus: sql.NullString{String: "pending", Valid: true},
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

	req := httptest.NewRequest(http.MethodGet, "/app/inventory", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "Inventory review bundle") {
		t.Fatalf("expected inventory landing heading, body=%s", body)
	}
	if !strings.Contains(body, `/app/inventory" class="nav-link is-active">Inventory</a>`) {
		t.Fatalf("expected inventory landing nav activation, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory">Inventory review</a>`) {
		t.Fatalf("expected landing link to inventory review, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory/items/item-123`) || !strings.Contains(body, `/app/review/inventory/locations/loc-123`) {
		t.Fatalf("expected domain landing continuity links for exact item and location review, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inventory/move-123`) {
		t.Fatalf("expected exact movement continuity link, body=%s", body)
	}
}

func TestHandleWebAppDashboardAddsRecoveryActionsForRecentRequests(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInboundRequests: func(context.Context, reporting.ListInboundRequestsInput) ([]reporting.InboundRequestReview, error) {
				return []reporting.InboundRequestReview{
					{
						RequestID:          "req-cancelled",
						RequestReference:   "REQ-000200",
						Status:             "cancelled",
						Channel:            "browser",
						MessageCount:       1,
						AttachmentCount:    0,
						CancellationReason: "operator paused request",
						CancelledAt:        sql.NullTime{Time: time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC), Valid: true},
						UpdatedAt:          time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
					},
					{
						RequestID:        "req-failed",
						RequestReference: "REQ-000201",
						Status:           "failed",
						Channel:          "browser",
						MessageCount:     2,
						AttachmentCount:  1,
						FailureReason:    "provider-backed coordinator execution failed",
						FailedAt:         sql.NullTime{Time: time.Date(2026, 3, 27, 10, 5, 0, 0, time.UTC), Valid: true},
						LastRunID:        sql.NullString{String: "run-201", Valid: true},
						UpdatedAt:        time.Date(2026, 3, 27, 10, 5, 0, 0, time.UTC),
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
	if !strings.Contains(body, `operator paused request`) {
		t.Fatalf("expected cancellation reason on dashboard, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/REQ-000200`) {
		t.Fatalf("expected cancelled request recovery link, body=%s", body)
	}
	if !strings.Contains(body, `provider-backed coordinator execution failed`) {
		t.Fatalf("expected failure reason on dashboard, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/REQ-000201`) {
		t.Fatalf("expected failed request action link, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/run:run-201#run-run-201">Open latest run</a>`) {
		t.Fatalf("expected failed request latest-run link, body=%s", body)
	}
}

func TestHandleWebInboundRequestsAddsAIRunAndProposalLinks(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInboundRequests: func(context.Context, reporting.ListInboundRequestsInput) ([]reporting.InboundRequestReview, error) {
				return []reporting.InboundRequestReview{
					{
						RequestID:                "request-123",
						RequestReference:         "REQ-000123",
						Status:                   "processed",
						Channel:                  "browser",
						OriginType:               "human",
						MessageCount:             2,
						AttachmentCount:          1,
						LastRunID:                sql.NullString{String: "run-123", Valid: true},
						LastRunStatus:            sql.NullString{String: "completed", Valid: true},
						LastRecommendationID:     sql.NullString{String: "rec-123", Valid: true},
						LastRecommendationStatus: sql.NullString{String: "approval_requested", Valid: true},
						UpdatedAt:                time.Date(2026, 3, 27, 9, 5, 0, 0, time.UTC),
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/inbound-requests", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `/app/inbound-requests/run:run-123#run-run-123`) {
		t.Fatalf("expected inbound-request AI run link, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/proposals/rec-123">approval_requested</a>`) {
		t.Fatalf("expected inbound-request proposal continuity link, body=%s", body)
	}
}

func TestHandleWebInboundRequestsRendersRefreshedFilterLayout(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInboundRequestStatusSummary: func(context.Context, identityaccess.Actor) ([]reporting.InboundRequestStatusSummary, error) {
				return []reporting.InboundRequestStatusSummary{
					{
						Status:          "queued",
						RequestCount:    1,
						MessageCount:    2,
						AttachmentCount: 0,
						LatestUpdatedAt: time.Date(2026, 3, 30, 9, 0, 0, 0, time.UTC),
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/inbound-requests?status=queued&request_reference=REQ-000123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `Filter by request state or exact`) {
		t.Fatalf("expected refreshed inbound review intro copy, body=%s", body)
	}
	if !strings.Contains(body, `class="filter-grid"`) {
		t.Fatalf("expected refreshed inbound review filter layout, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inbound-requests" class="pill-link">Clear filters</a>`) {
		t.Fatalf("expected inbound review clear-filters action, body=%s", body)
	}
}

func TestHandleWebProposalsRendersRefreshedFilterLayout(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listProcessedProposalStatusSummary: func(context.Context, identityaccess.Actor) ([]reporting.ProcessedProposalStatusSummary, error) {
				return []reporting.ProcessedProposalStatusSummary{
					{
						RecommendationStatus: "approval_requested",
						ProposalCount:        1,
						RequestCount:         1,
						DocumentCount:        1,
						LatestCreatedAt:      time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC),
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/proposals?status=approval_requested", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `Track the coordinator handoff from exact inbound request reference`) {
		t.Fatalf("expected refreshed proposal review intro copy, body=%s", body)
	}
	if !strings.Contains(body, `class="filter-grid"`) {
		t.Fatalf("expected refreshed proposal review filter layout, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/proposals" class="pill-link">Clear filters</a>`) {
		t.Fatalf("expected proposal review clear-filters action, body=%s", body)
	}
}

func TestHandleWebInboundRequestDetailShowsDraftLifecycleActions(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			getInboundRequestDetail: func(context.Context, reporting.GetInboundRequestDetailInput) (reporting.InboundRequestDetail, error) {
				return reporting.InboundRequestDetail{
					Request: reporting.InboundRequestReview{
						RequestID:        "req-123",
						RequestReference: "REQ-000123",
						Status:           intake.StatusDraft,
						Channel:          "browser",
						OriginType:       intake.OriginHuman,
						ReceivedAt:       time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
						Metadata:         []byte(`{"submitter_label":"front desk"}`),
					},
					Messages: []reporting.InboundRequestMessageReview{
						{
							MessageID:    "msg-123",
							MessageIndex: 1,
							MessageRole:  intake.MessageRoleRequest,
							TextContent:  "Draft message",
							CreatedAt:    time.Date(2026, 3, 27, 10, 1, 0, 0, time.UTC),
						},
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

	req := httptest.NewRequest(http.MethodGet, "/app/inbound-requests/REQ-000123", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `name="intent" value="save_draft"`) {
		t.Fatalf("expected save-draft action, body=%s", body)
	}
	if !strings.Contains(body, `name="intent" value="queue"`) {
		t.Fatalf("expected queue action, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/req-123/delete`) {
		t.Fatalf("expected delete-draft action, body=%s", body)
	}
}

func TestHandleWebSubmitInboundRequestSaveDraftRedirectsToDetail(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		stubSubmissionService{
			saveInboundDraft: func(context.Context, SaveInboundDraftInput) (SaveInboundDraftResult, error) {
				return SaveInboundDraftResult{
					Request: intake.InboundRequest{
						ID:               "req-123",
						RequestReference: "REQ-000123",
						Status:           intake.StatusDraft,
					},
					Message: intake.Message{ID: "msg-123"},
				}, nil
			},
		},
		nil,
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				return testSessionContext(), nil
			},
		},
	)

	form := url.Values{}
	form.Set("submitter_label", "front desk")
	form.Set("message_text", "Draft this request")
	form.Set("intent", "save_draft")
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, values := range form {
		for _, value := range values {
			if err := writer.WriteField(key, value); err != nil {
				t.Fatalf("write multipart field: %v", err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/app/inbound-requests", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	location := recorder.Header().Get("Location")
	if !strings.Contains(location, "/app/inbound-requests/REQ-000123?notice=Draft+saved.") {
		t.Fatalf("unexpected redirect location: %s", location)
	}
}

func TestHandleWebSubmitInboundRequestPageRendersDedicatedFormAndResultState(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				return testSessionContext(), nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/submit-inbound-request?notice=Inbound+request+submitted.&request_reference=REQ-000123&request_status=queued", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `Dedicated intake`) {
		t.Fatalf("expected dedicated intake hero copy, body=%s", body)
	}
	if !strings.Contains(body, `name="return_to" value="/app/submit-inbound-request"`) {
		t.Fatalf("expected dedicated submission form return target, body=%s", body)
	}
	if !strings.Contains(body, `REQ-000123`) {
		t.Fatalf("expected exact request reference result card, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/REQ-000123" class="pill-link">Open exact request detail</a>`) {
		t.Fatalf("expected exact request detail continuation link, body=%s", body)
	}
}

func TestHandleWebSubmitInboundRequestFromDedicatedPageRedirectsBackWithReference(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		stubSubmissionService{
			submitInboundRequest: func(context.Context, SubmitInboundRequestInput) (SubmitInboundRequestResult, error) {
				return SubmitInboundRequestResult{
					Request: intake.InboundRequest{
						ID:               "req-123",
						RequestReference: "REQ-000123",
						Status:           intake.StatusQueued,
					},
					Message: intake.Message{ID: "msg-123"},
				}, nil
			},
		},
		nil,
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				return testSessionContext(), nil
			},
		},
	)

	form := url.Values{}
	form.Set("submitter_label", "front desk")
	form.Set("message_text", "Queue this request from the dedicated intake page")
	form.Set("return_to", "/app/submit-inbound-request")
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, values := range form {
		for _, value := range values {
			if err := writer.WriteField(key, value); err != nil {
				t.Fatalf("write multipart field: %v", err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/app/inbound-requests", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	location := recorder.Header().Get("Location")
	if !strings.Contains(location, "/app/submit-inbound-request?notice=Inbound+request+submitted.") {
		t.Fatalf("unexpected redirect location: %s", location)
	}
	if !strings.Contains(location, "request_reference=REQ-000123") {
		t.Fatalf("expected request reference in redirect location: %s", location)
	}
	if !strings.Contains(location, "request_status=queued") {
		t.Fatalf("expected request status in redirect location: %s", location)
	}
}

func TestHandleWebOperationsFeedRendersDurableEventItems(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInboundRequests: func(context.Context, reporting.ListInboundRequestsInput) ([]reporting.InboundRequestReview, error) {
				return []reporting.InboundRequestReview{{
					RequestID:            "req-123",
					RequestReference:     "REQ-000123",
					Channel:              "browser",
					Status:               "failed",
					MessageCount:         1,
					AttachmentCount:      0,
					FailureReason:        "provider timeout",
					LastRecommendationID: sql.NullString{String: "rec-123", Valid: true},
					UpdatedAt:            time.Date(2026, 3, 30, 14, 0, 0, 0, time.UTC),
				}}, nil
			},
			listProcessedProposals: func(context.Context, reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error) {
				return []reporting.ProcessedProposalReview{{
					RequestReference:     "REQ-000124",
					RecommendationID:     "rec-124",
					RecommendationStatus: "approval_requested",
					Summary:              "Draft invoice ready for review",
					CreatedAt:            time.Date(2026, 3, 30, 13, 0, 0, 0, time.UTC),
				}}, nil
			},
			listApprovalQueue: func(context.Context, reporting.ListApprovalQueueInput) ([]reporting.ApprovalQueueEntry, error) {
				return []reporting.ApprovalQueueEntry{{
					ApprovalID:     "approval-123",
					QueueCode:      "finance_review",
					ApprovalStatus: "pending",
					DocumentID:     "doc-123",
					DocumentTitle:  "Invoice draft",
					RequestedAt:    time.Date(2026, 3, 30, 12, 30, 0, 0, time.UTC),
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

	req := httptest.NewRequest(http.MethodGet, "/app/operations-feed", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `Durable operations feed`) {
		t.Fatalf("expected operations feed hero copy, body=%s", body)
	}
	if !strings.Contains(body, `REQ-000123 moved through failed`) {
		t.Fatalf("expected request-derived feed item, body=%s", body)
	}
	if !strings.Contains(body, `provider timeout`) {
		t.Fatalf("expected request failure summary in feed, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/proposals/rec-124" class="pill-link">Open proposal</a>`) {
		t.Fatalf("expected proposal continuity link in feed, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/approvals/approval-123" class="pill-link">Open approval</a>`) {
		t.Fatalf("expected approval continuity link in feed, body=%s", body)
	}
}

func TestHandleWebAgentChatFiltersChatRequestsAndProposalContinuity(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listInboundRequests: func(context.Context, reporting.ListInboundRequestsInput) ([]reporting.InboundRequestReview, error) {
				return []reporting.InboundRequestReview{
					{
						RequestID:            "req-chat",
						RequestReference:     "REQ-000200",
						Channel:              inboundRequestChannelAgentChat,
						Status:               "processed",
						MessageCount:         1,
						AttachmentCount:      0,
						LastRecommendationID: sql.NullString{String: "rec-chat", Valid: true},
						UpdatedAt:            time.Date(2026, 3, 30, 15, 0, 0, 0, time.UTC),
					},
					{
						RequestID:        "req-browser",
						RequestReference: "REQ-000201",
						Channel:          inboundRequestChannelBrowser,
						Status:           "queued",
						MessageCount:     1,
						AttachmentCount:  0,
						UpdatedAt:        time.Date(2026, 3, 30, 14, 0, 0, 0, time.UTC),
					},
				}, nil
			},
			listProcessedProposals: func(context.Context, reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error) {
				return []reporting.ProcessedProposalReview{
					{
						RequestReference:     "REQ-000200",
						RecommendationID:     "rec-chat",
						RecommendationStatus: "approval_requested",
						Summary:              "Chat follow-up recommendation",
						CreatedAt:            time.Date(2026, 3, 30, 15, 5, 0, 0, time.UTC),
					},
					{
						RequestReference:     "REQ-000201",
						RecommendationID:     "rec-browser",
						RecommendationStatus: "approval_requested",
						Summary:              "Browser submission recommendation",
						CreatedAt:            time.Date(2026, 3, 30, 14, 5, 0, 0, time.UTC),
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

	req := httptest.NewRequest(http.MethodGet, "/app/agent-chat?request_reference=REQ-000200&request_status=queued", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `Coordinator chat`) {
		t.Fatalf("expected agent-chat hero copy, body=%s", body)
	}
	if !strings.Contains(body, `name="channel" value="agent_chat"`) {
		t.Fatalf("expected agent-chat form to submit dedicated channel, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/REQ-000200" class="pill-link">Open exact request detail</a>`) {
		t.Fatalf("expected result continuity card for chat request, body=%s", body)
	}
	if !strings.Contains(body, `REQ-000200`) {
		t.Fatalf("expected chat request to render, body=%s", body)
	}
	if strings.Contains(body, `REQ-000201`) {
		t.Fatalf("expected browser-channel request to stay out of agent chat page, body=%s", body)
	}
	if !strings.Contains(body, `Chat follow-up recommendation`) {
		t.Fatalf("expected chat proposal summary, body=%s", body)
	}
	if strings.Contains(body, `Browser submission recommendation`) {
		t.Fatalf("expected non-chat proposal to stay out of agent chat page, body=%s", body)
	}
}

func TestHandleWebAgentChatUsesSharedAgentChatSnapshot(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			getAgentChatSnapshot: func(context.Context, identityaccess.Actor, int, int) (reporting.AgentChatSnapshot, error) {
				return reporting.AgentChatSnapshot{
					RecentRequests: []reporting.InboundRequestReview{
						{
							RequestReference: "REQ-009901",
							Status:           "queued",
							Channel:          inboundRequestChannelAgentChat,
							UpdatedAt:        time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC),
						},
					},
				}, nil
			},
			listInboundRequests: func(context.Context, reporting.ListInboundRequestsInput) ([]reporting.InboundRequestReview, error) {
				t.Fatal("agent-chat handler should not list requests directly")
				return nil, nil
			},
			listProcessedProposals: func(context.Context, reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error) {
				t.Fatal("agent-chat handler should not list proposals directly")
				return nil, nil
			},
		},
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				return testSessionContext(), nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/agent-chat", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `REQ-009901`) {
		t.Fatalf("expected agent-chat snapshot content, body=%s", recorder.Body.String())
	}
}

func TestHandleWebSubmitInboundRequestFromAgentChatUsesDedicatedChannel(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		stubSubmissionService{
			submitInboundRequest: func(_ context.Context, input SubmitInboundRequestInput) (SubmitInboundRequestResult, error) {
				if input.Channel != inboundRequestChannelAgentChat {
					t.Fatalf("expected agent-chat channel, got %q", input.Channel)
				}
				return SubmitInboundRequestResult{
					Request: intake.InboundRequest{
						ID:               "req-123",
						RequestReference: "REQ-000123",
						Status:           intake.StatusQueued,
					},
					Message: intake.Message{ID: "msg-123"},
				}, nil
			},
		},
		nil,
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				return testSessionContext(), nil
			},
		},
	)

	form := url.Values{}
	form.Set("submitter_label", "dispatch desk")
	form.Set("message_text", "Need coordinator guidance on this issue")
	form.Set("return_to", "/app/agent-chat")
	form.Set("channel", "agent_chat")
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, values := range form {
		for _, value := range values {
			if err := writer.WriteField(key, value); err != nil {
				t.Fatalf("write multipart field: %v", err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/app/inbound-requests", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	location := recorder.Header().Get("Location")
	if !strings.Contains(location, "/app/agent-chat?notice=Inbound+request+submitted.") {
		t.Fatalf("unexpected redirect location: %s", location)
	}
	if !strings.Contains(location, "request_reference=REQ-000123") {
		t.Fatalf("expected request reference in redirect location: %s", location)
	}
	if !strings.Contains(location, "request_status=queued") {
		t.Fatalf("expected request status in redirect location: %s", location)
	}
}

func TestHandleWebAppUnauthenticatedRendersRefreshedLoginSurface(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		stubBrowserSessionService{},
	)

	req := httptest.NewRequest(http.MethodGet, "/app", nil)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `Browser session`) {
		t.Fatalf("expected refreshed login eyebrow, body=%s", body)
	}
	if !strings.Contains(body, `This sign-in surface only issues the browser-session path.`) {
		t.Fatalf("expected refreshed login guidance note, body=%s", body)
	}
	if !strings.Contains(body, `Browser session entry`) {
		t.Fatalf("expected compact public top bar, body=%s", body)
	}
	if !strings.Contains(body, `class="panel login-panel"`) {
		t.Fatalf("expected refreshed login panel class, body=%s", body)
	}
}

func TestHandleWebLoginGetRendersRefreshedLoginSurface(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		stubBrowserSessionService{},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/login?notice=Sign+in+required", nil)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `Sign in`) {
		t.Fatalf("expected login heading, body=%s", body)
	}
	if !strings.Contains(body, `form method="post" action="/app/login"`) {
		t.Fatalf("expected login form action, body=%s", body)
	}
	if !strings.Contains(body, `Sign in required`) {
		t.Fatalf("expected notice to render on login page, body=%s", body)
	}
}

func TestHandleWebLoginGetRedirectsAuthenticatedSessionToDashboard(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		stubBrowserSessionService{
			authenticateSession: func(context.Context, string, string) (identityaccess.SessionContext, error) {
				return testSessionContext(), nil
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/app/login", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if location := recorder.Header().Get("Location"); location != "/app" {
		t.Fatalf("unexpected redirect location: %s", location)
	}
}

func TestHandleWebApprovalDetailUsesContainedTablesOnlyForWrappedLayouts(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		stubOperatorReviewReader{
			listApprovalQueue: func(context.Context, reporting.ListApprovalQueueInput) ([]reporting.ApprovalQueueEntry, error) {
				return []reporting.ApprovalQueueEntry{{
					ApprovalID:        "approval-123",
					QueueCode:         "ops_review",
					QueueStatus:       "pending",
					ApprovalStatus:    "pending",
					DocumentID:        "doc-123",
					DocumentTitle:     "Inbound draft invoice",
					DocumentTypeCode:  "invoice",
					DocumentStatus:    "draft",
					RequestedByUserID: "user-123",
					RequestedAt:       time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC),
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

	req := httptest.NewRequest(http.MethodGet, "/app/review/approvals", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookieName, Value: "00000000-0000-4000-8000-000000000123"})
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-123"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `.table-wrap > table`) {
		t.Fatalf("expected wrapped-table min-width rule, body=%s", body)
	}
	if strings.Contains(body, "table {\n      width: 100%;\n      border-collapse: collapse;\n      font-size: 0.96rem;\n      min-width: 640px;") {
		t.Fatalf("expected global table min-width rule to be removed, body=%s", body)
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
	listJournalEntries                 func(context.Context, reporting.ListJournalEntriesInput) ([]reporting.JournalEntryReview, error)
	listInventoryStock                 func(context.Context, reporting.ListInventoryStockInput) ([]reporting.InventoryStockItem, error)
	listInventoryMovements             func(context.Context, reporting.ListInventoryMovementsInput) ([]reporting.InventoryMovementReview, error)
	listInventoryReconciliation        func(context.Context, reporting.ListInventoryReconciliationInput) ([]reporting.InventoryReconciliationItem, error)
	listAuditEvents                    func(context.Context, reporting.LookupAuditEventsInput) ([]reporting.AuditEvent, error)
	listWorkOrders                     func(context.Context, reporting.ListWorkOrdersInput) ([]reporting.WorkOrderReview, error)
	getWorkOrderReview                 func(context.Context, reporting.GetWorkOrderReviewInput) (reporting.WorkOrderReview, error)
	listInboundRequests                func(context.Context, reporting.ListInboundRequestsInput) ([]reporting.InboundRequestReview, error)
	getInboundRequestDetail            func(context.Context, reporting.GetInboundRequestDetailInput) (reporting.InboundRequestDetail, error)
	listInboundRequestStatusSummary    func(context.Context, identityaccess.Actor) ([]reporting.InboundRequestStatusSummary, error)
	listProcessedProposals             func(context.Context, reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error)
	listProcessedProposalStatusSummary func(context.Context, identityaccess.Actor) ([]reporting.ProcessedProposalStatusSummary, error)
	getOperationsFeedSnapshot          func(context.Context, identityaccess.Actor, int) (reporting.OperationsFeedSnapshot, error)
	getOperationsLandingSnapshot       func(context.Context, identityaccess.Actor, int, int) (reporting.OperationsLandingSnapshot, error)
	getDashboardSnapshot               func(context.Context, identityaccess.Actor, int, int, int) (reporting.DashboardSnapshot, error)
	getAgentChatSnapshot               func(context.Context, identityaccess.Actor, int, int) (reporting.AgentChatSnapshot, error)
	getInventoryLandingSnapshot        func(context.Context, identityaccess.Actor, int) (reporting.InventoryLandingSnapshot, error)
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

func (s stubOperatorReviewReader) ListJournalEntries(ctx context.Context, input reporting.ListJournalEntriesInput) ([]reporting.JournalEntryReview, error) {
	if s.listJournalEntries != nil {
		return s.listJournalEntries(ctx, input)
	}
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

func (s stubOperatorReviewReader) ListInventoryMovements(ctx context.Context, input reporting.ListInventoryMovementsInput) ([]reporting.InventoryMovementReview, error) {
	if s.listInventoryMovements != nil {
		return s.listInventoryMovements(ctx, input)
	}
	return nil, nil
}

func (s stubOperatorReviewReader) ListInventoryReconciliation(ctx context.Context, input reporting.ListInventoryReconciliationInput) ([]reporting.InventoryReconciliationItem, error) {
	if s.listInventoryReconciliation != nil {
		return s.listInventoryReconciliation(ctx, input)
	}
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

func (s stubOperatorReviewReader) ListInboundRequests(ctx context.Context, input reporting.ListInboundRequestsInput) ([]reporting.InboundRequestReview, error) {
	if s.listInboundRequests != nil {
		return s.listInboundRequests(ctx, input)
	}
	return nil, nil
}

func (s stubOperatorReviewReader) GetInboundRequestDetail(ctx context.Context, input reporting.GetInboundRequestDetailInput) (reporting.InboundRequestDetail, error) {
	if s.getInboundRequestDetail != nil {
		return s.getInboundRequestDetail(ctx, input)
	}
	return reporting.InboundRequestDetail{}, nil
}

func (s stubOperatorReviewReader) ListInboundRequestStatusSummary(ctx context.Context, actor identityaccess.Actor) ([]reporting.InboundRequestStatusSummary, error) {
	if s.listInboundRequestStatusSummary != nil {
		return s.listInboundRequestStatusSummary(ctx, actor)
	}
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

func (s stubOperatorReviewReader) GetWorkflowNavigationSnapshot(ctx context.Context, actor identityaccess.Actor, pendingApprovalLimit int) (reporting.WorkflowNavigationSnapshot, error) {
	inboundSummary, err := s.ListInboundRequestStatusSummary(ctx, actor)
	if err != nil {
		return reporting.WorkflowNavigationSnapshot{}, err
	}
	proposalSummary, err := s.ListProcessedProposalStatusSummary(ctx, actor)
	if err != nil {
		return reporting.WorkflowNavigationSnapshot{}, err
	}
	pendingApprovals, err := s.ListApprovalQueue(ctx, reporting.ListApprovalQueueInput{
		Status: "pending",
		Limit:  pendingApprovalLimit,
		Actor:  actor,
	})
	if err != nil {
		return reporting.WorkflowNavigationSnapshot{}, err
	}
	return reporting.WorkflowNavigationSnapshot{
		InboundSummary:    inboundSummary,
		ProposalSummary:   proposalSummary,
		PendingApprovals:  pendingApprovals,
		PendingQueueLimit: pendingApprovalLimit,
	}, nil
}

func (s stubOperatorReviewReader) GetOperationsFeedSnapshot(ctx context.Context, actor identityaccess.Actor, recentLimit int) (reporting.OperationsFeedSnapshot, error) {
	if s.getOperationsFeedSnapshot != nil {
		return s.getOperationsFeedSnapshot(ctx, actor, recentLimit)
	}
	requests, err := s.ListInboundRequests(ctx, reporting.ListInboundRequestsInput{
		Limit: recentLimit,
		Actor: actor,
	})
	if err != nil {
		return reporting.OperationsFeedSnapshot{}, err
	}
	proposals, err := s.ListProcessedProposals(ctx, reporting.ListProcessedProposalsInput{
		Limit: recentLimit,
		Actor: actor,
	})
	if err != nil {
		return reporting.OperationsFeedSnapshot{}, err
	}
	approvals, err := s.ListApprovalQueue(ctx, reporting.ListApprovalQueueInput{
		Limit: recentLimit,
		Actor: actor,
	})
	if err != nil {
		return reporting.OperationsFeedSnapshot{}, err
	}
	return reporting.OperationsFeedSnapshot{
		Requests:    requests,
		Proposals:   proposals,
		Approvals:   approvals,
		RecentLimit: recentLimit,
	}, nil
}

func (s stubOperatorReviewReader) GetOperationsLandingSnapshot(ctx context.Context, actor identityaccess.Actor, pendingApprovalLimit, recentLimit int) (reporting.OperationsLandingSnapshot, error) {
	if s.getOperationsLandingSnapshot != nil {
		return s.getOperationsLandingSnapshot(ctx, actor, pendingApprovalLimit, recentLimit)
	}
	navigation, err := s.GetWorkflowNavigationSnapshot(ctx, actor, pendingApprovalLimit)
	if err != nil {
		return reporting.OperationsLandingSnapshot{}, err
	}
	feed, err := s.GetOperationsFeedSnapshot(ctx, actor, recentLimit)
	if err != nil {
		return reporting.OperationsLandingSnapshot{}, err
	}
	return reporting.OperationsLandingSnapshot{
		Navigation: navigation,
		Feed:       feed,
	}, nil
}

func (s stubOperatorReviewReader) GetDashboardSnapshot(ctx context.Context, actor identityaccess.Actor, pendingApprovalLimit, requestLimit, proposalLimit int) (reporting.DashboardSnapshot, error) {
	if s.getDashboardSnapshot != nil {
		return s.getDashboardSnapshot(ctx, actor, pendingApprovalLimit, requestLimit, proposalLimit)
	}
	navigation, err := s.GetWorkflowNavigationSnapshot(ctx, actor, pendingApprovalLimit)
	if err != nil {
		return reporting.DashboardSnapshot{}, err
	}
	requests, err := s.ListInboundRequests(ctx, reporting.ListInboundRequestsInput{
		Limit: requestLimit,
		Actor: actor,
	})
	if err != nil {
		return reporting.DashboardSnapshot{}, err
	}
	proposals, err := s.ListProcessedProposals(ctx, reporting.ListProcessedProposalsInput{
		Limit: proposalLimit,
		Actor: actor,
	})
	if err != nil {
		return reporting.DashboardSnapshot{}, err
	}
	return reporting.DashboardSnapshot{
		Navigation:      navigation,
		InboundRequests: requests,
		Proposals:       proposals,
		RequestLimit:    requestLimit,
		ProposalLimit:   proposalLimit,
	}, nil
}

func (s stubOperatorReviewReader) GetAgentChatSnapshot(ctx context.Context, actor identityaccess.Actor, requestLimit, proposalLimit int) (reporting.AgentChatSnapshot, error) {
	if s.getAgentChatSnapshot != nil {
		return s.getAgentChatSnapshot(ctx, actor, requestLimit, proposalLimit)
	}
	requests, err := s.ListInboundRequests(ctx, reporting.ListInboundRequestsInput{
		Limit: requestLimit,
		Actor: actor,
	})
	if err != nil {
		return reporting.AgentChatSnapshot{}, err
	}

	chatRequests := make([]reporting.InboundRequestReview, 0, len(requests))
	requestRefs := make(map[string]struct{}, len(requests))
	for _, item := range requests {
		if !strings.EqualFold(strings.TrimSpace(item.Channel), inboundRequestChannelAgentChat) {
			continue
		}
		chatRequests = append(chatRequests, item)
		ref := strings.TrimSpace(item.RequestReference)
		if ref != "" {
			requestRefs[ref] = struct{}{}
		}
	}

	proposals, err := s.ListProcessedProposals(ctx, reporting.ListProcessedProposalsInput{
		Limit: proposalLimit,
		Actor: actor,
	})
	if err != nil {
		return reporting.AgentChatSnapshot{}, err
	}

	chatProposals := make([]reporting.ProcessedProposalReview, 0, len(proposals))
	for _, item := range proposals {
		if _, ok := requestRefs[strings.TrimSpace(item.RequestReference)]; ok {
			chatProposals = append(chatProposals, item)
		}
	}

	return reporting.AgentChatSnapshot{
		RecentRequests:  chatRequests,
		RecentProposals: chatProposals,
		RequestLimit:    requestLimit,
		ProposalLimit:   proposalLimit,
	}, nil
}

func (s stubOperatorReviewReader) GetInventoryLandingSnapshot(ctx context.Context, actor identityaccess.Actor, recentLimit int) (reporting.InventoryLandingSnapshot, error) {
	if s.getInventoryLandingSnapshot != nil {
		return s.getInventoryLandingSnapshot(ctx, actor, recentLimit)
	}
	stock, err := s.ListInventoryStock(ctx, reporting.ListInventoryStockInput{
		Limit: recentLimit,
		Actor: actor,
	})
	if err != nil {
		return reporting.InventoryLandingSnapshot{}, err
	}
	movements, err := s.ListInventoryMovements(ctx, reporting.ListInventoryMovementsInput{
		Limit: recentLimit,
		Actor: actor,
	})
	if err != nil {
		return reporting.InventoryLandingSnapshot{}, err
	}
	reconciliation, err := s.ListInventoryReconciliation(ctx, reporting.ListInventoryReconciliationInput{
		Limit: recentLimit,
		Actor: actor,
	})
	if err != nil {
		return reporting.InventoryLandingSnapshot{}, err
	}
	return reporting.InventoryLandingSnapshot{
		Stock:          stock,
		Movements:      movements,
		Reconciliation: reconciliation,
		RecentLimit:    recentLimit,
	}, nil
}

type stubBrowserSessionService struct {
	authenticateSession     func(context.Context, string, string) (identityaccess.SessionContext, error)
	authenticateAccessToken func(context.Context, string) (identityaccess.SessionContext, error)
	listOrgUsers            func(context.Context, identityaccess.ListOrgUsersInput) ([]identityaccess.OrgUserMembership, error)
	provisionOrgUser        func(context.Context, identityaccess.ProvisionOrgUserInput) (identityaccess.OrgUserMembership, error)
	updateMembershipRole    func(context.Context, identityaccess.UpdateMembershipRoleInput) (identityaccess.OrgUserMembership, error)
}

func (s stubBrowserSessionService) StartBrowserSession(context.Context, identityaccess.StartBrowserSessionInput) (identityaccess.BrowserSession, error) {
	return identityaccess.BrowserSession{}, nil
}

func (s stubBrowserSessionService) StartTokenSession(context.Context, identityaccess.StartTokenSessionInput) (identityaccess.TokenSession, error) {
	return identityaccess.TokenSession{}, nil
}

func (s stubBrowserSessionService) AuthenticateSession(ctx context.Context, sessionID, refreshToken string) (identityaccess.SessionContext, error) {
	if s.authenticateSession != nil {
		return s.authenticateSession(ctx, sessionID, refreshToken)
	}
	return identityaccess.SessionContext{}, identityaccess.ErrUnauthorized
}

func (s stubBrowserSessionService) AuthenticateAccessToken(ctx context.Context, accessToken string) (identityaccess.SessionContext, error) {
	if s.authenticateAccessToken != nil {
		return s.authenticateAccessToken(ctx, accessToken)
	}
	return identityaccess.SessionContext{}, identityaccess.ErrUnauthorized
}

func (s stubBrowserSessionService) RefreshTokenSession(context.Context, string, string, time.Time) (identityaccess.TokenSession, error) {
	return identityaccess.TokenSession{}, identityaccess.ErrUnauthorized
}

func (s stubBrowserSessionService) RevokeAuthenticatedSession(context.Context, string, string) error {
	return nil
}

func (s stubBrowserSessionService) RevokeAccessTokenSession(context.Context, string) error {
	return nil
}

func (s stubBrowserSessionService) ListOrgUsers(ctx context.Context, input identityaccess.ListOrgUsersInput) ([]identityaccess.OrgUserMembership, error) {
	if s.listOrgUsers != nil {
		return s.listOrgUsers(ctx, input)
	}
	return nil, nil
}

func (s stubBrowserSessionService) ProvisionOrgUser(ctx context.Context, input identityaccess.ProvisionOrgUserInput) (identityaccess.OrgUserMembership, error) {
	if s.provisionOrgUser != nil {
		return s.provisionOrgUser(ctx, input)
	}
	return identityaccess.OrgUserMembership{}, nil
}

func (s stubBrowserSessionService) UpdateMembershipRole(ctx context.Context, input identityaccess.UpdateMembershipRoleInput) (identityaccess.OrgUserMembership, error) {
	if s.updateMembershipRole != nil {
		return s.updateMembershipRole(ctx, input)
	}
	return identityaccess.OrgUserMembership{}, nil
}

type stubSubmissionService struct {
	submitInboundRequest func(context.Context, SubmitInboundRequestInput) (SubmitInboundRequestResult, error)
	saveInboundDraft     func(context.Context, SaveInboundDraftInput) (SaveInboundDraftResult, error)
	queueInboundRequest  func(context.Context, QueueInboundRequestInput) (intake.InboundRequest, error)
	cancelInboundRequest func(context.Context, CancelInboundRequestInput) (intake.InboundRequest, error)
	amendInboundRequest  func(context.Context, AmendInboundRequestInput) (intake.InboundRequest, error)
	deleteInboundDraft   func(context.Context, DeleteInboundDraftInput) error
}

func (s stubSubmissionService) SubmitInboundRequest(ctx context.Context, input SubmitInboundRequestInput) (SubmitInboundRequestResult, error) {
	if s.submitInboundRequest != nil {
		return s.submitInboundRequest(ctx, input)
	}
	return SubmitInboundRequestResult{}, nil
}

func (s stubSubmissionService) SaveInboundDraft(ctx context.Context, input SaveInboundDraftInput) (SaveInboundDraftResult, error) {
	if s.saveInboundDraft != nil {
		return s.saveInboundDraft(ctx, input)
	}
	return SaveInboundDraftResult{}, nil
}

func (s stubSubmissionService) QueueInboundRequest(ctx context.Context, input QueueInboundRequestInput) (intake.InboundRequest, error) {
	if s.queueInboundRequest != nil {
		return s.queueInboundRequest(ctx, input)
	}
	return intake.InboundRequest{}, nil
}

func (s stubSubmissionService) CancelInboundRequest(ctx context.Context, input CancelInboundRequestInput) (intake.InboundRequest, error) {
	if s.cancelInboundRequest != nil {
		return s.cancelInboundRequest(ctx, input)
	}
	return intake.InboundRequest{}, nil
}

func (s stubSubmissionService) AmendInboundRequest(ctx context.Context, input AmendInboundRequestInput) (intake.InboundRequest, error) {
	if s.amendInboundRequest != nil {
		return s.amendInboundRequest(ctx, input)
	}
	return intake.InboundRequest{}, nil
}

func (s stubSubmissionService) DeleteInboundDraft(ctx context.Context, input DeleteInboundDraftInput) error {
	if s.deleteInboundDraft != nil {
		return s.deleteInboundDraft(ctx, input)
	}
	return nil
}

func (s stubSubmissionService) DownloadAttachment(context.Context, DownloadAttachmentInput) (attachments.AttachmentContent, error) {
	return attachments.AttachmentContent{}, nil
}

type stubAccountingAdminService struct {
	listLedgerAccounts     func(context.Context, accounting.ListLedgerAccountsInput) ([]accounting.LedgerAccount, error)
	createLedgerAccount    func(context.Context, accounting.CreateLedgerAccountInput) (accounting.LedgerAccount, error)
	updateLedgerStatus     func(context.Context, accounting.UpdateLedgerAccountStatusInput) (accounting.LedgerAccount, error)
	listTaxCodes           func(context.Context, accounting.ListTaxCodesInput) ([]accounting.TaxCode, error)
	createTaxCode          func(context.Context, accounting.CreateTaxCodeInput) (accounting.TaxCode, error)
	updateTaxCodeStatus    func(context.Context, accounting.UpdateTaxCodeStatusInput) (accounting.TaxCode, error)
	listAccountingPeriods  func(context.Context, accounting.ListAccountingPeriodsInput) ([]accounting.AccountingPeriod, error)
	createAccountingPeriod func(context.Context, accounting.CreateAccountingPeriodInput) (accounting.AccountingPeriod, error)
	closeAccountingPeriod  func(context.Context, accounting.CloseAccountingPeriodInput) (accounting.AccountingPeriod, error)
}

func (s stubAccountingAdminService) ListLedgerAccounts(ctx context.Context, input accounting.ListLedgerAccountsInput) ([]accounting.LedgerAccount, error) {
	if s.listLedgerAccounts != nil {
		return s.listLedgerAccounts(ctx, input)
	}
	return nil, nil
}

func (s stubAccountingAdminService) CreateLedgerAccount(ctx context.Context, input accounting.CreateLedgerAccountInput) (accounting.LedgerAccount, error) {
	if s.createLedgerAccount != nil {
		return s.createLedgerAccount(ctx, input)
	}
	return accounting.LedgerAccount{}, nil
}

func (s stubAccountingAdminService) UpdateLedgerAccountStatus(ctx context.Context, input accounting.UpdateLedgerAccountStatusInput) (accounting.LedgerAccount, error) {
	if s.updateLedgerStatus != nil {
		return s.updateLedgerStatus(ctx, input)
	}
	return accounting.LedgerAccount{}, nil
}

func (s stubAccountingAdminService) ListTaxCodes(ctx context.Context, input accounting.ListTaxCodesInput) ([]accounting.TaxCode, error) {
	if s.listTaxCodes != nil {
		return s.listTaxCodes(ctx, input)
	}
	return nil, nil
}

func (s stubAccountingAdminService) CreateTaxCode(ctx context.Context, input accounting.CreateTaxCodeInput) (accounting.TaxCode, error) {
	if s.createTaxCode != nil {
		return s.createTaxCode(ctx, input)
	}
	return accounting.TaxCode{}, nil
}

func (s stubAccountingAdminService) UpdateTaxCodeStatus(ctx context.Context, input accounting.UpdateTaxCodeStatusInput) (accounting.TaxCode, error) {
	if s.updateTaxCodeStatus != nil {
		return s.updateTaxCodeStatus(ctx, input)
	}
	return accounting.TaxCode{}, nil
}

func (s stubAccountingAdminService) ListAccountingPeriods(ctx context.Context, input accounting.ListAccountingPeriodsInput) ([]accounting.AccountingPeriod, error) {
	if s.listAccountingPeriods != nil {
		return s.listAccountingPeriods(ctx, input)
	}
	return nil, nil
}

func (s stubAccountingAdminService) CreateAccountingPeriod(ctx context.Context, input accounting.CreateAccountingPeriodInput) (accounting.AccountingPeriod, error) {
	if s.createAccountingPeriod != nil {
		return s.createAccountingPeriod(ctx, input)
	}
	return accounting.AccountingPeriod{}, nil
}

type stubPartiesAdminService struct {
	listParties   func(context.Context, parties.ListPartiesInput) ([]parties.Party, error)
	getParty      func(context.Context, parties.GetPartyInput) (parties.Party, error)
	createParty   func(context.Context, parties.CreatePartyInput) (parties.Party, error)
	updateStatus  func(context.Context, parties.UpdatePartyStatusInput) (parties.Party, error)
	listContacts  func(context.Context, parties.ListContactsInput) ([]parties.Contact, error)
	createContact func(context.Context, parties.CreateContactInput) (parties.Contact, error)
}

func (s stubPartiesAdminService) ListParties(ctx context.Context, input parties.ListPartiesInput) ([]parties.Party, error) {
	if s.listParties != nil {
		return s.listParties(ctx, input)
	}
	return nil, nil
}

func (s stubPartiesAdminService) GetParty(ctx context.Context, input parties.GetPartyInput) (parties.Party, error) {
	if s.getParty != nil {
		return s.getParty(ctx, input)
	}
	return parties.Party{}, nil
}

func (s stubPartiesAdminService) CreateParty(ctx context.Context, input parties.CreatePartyInput) (parties.Party, error) {
	if s.createParty != nil {
		return s.createParty(ctx, input)
	}
	return parties.Party{}, nil
}

func (s stubPartiesAdminService) UpdatePartyStatus(ctx context.Context, input parties.UpdatePartyStatusInput) (parties.Party, error) {
	if s.updateStatus != nil {
		return s.updateStatus(ctx, input)
	}
	return parties.Party{}, nil
}

func (s stubPartiesAdminService) ListContacts(ctx context.Context, input parties.ListContactsInput) ([]parties.Contact, error) {
	if s.listContacts != nil {
		return s.listContacts(ctx, input)
	}
	return nil, nil
}

func (s stubPartiesAdminService) CreateContact(ctx context.Context, input parties.CreateContactInput) (parties.Contact, error) {
	if s.createContact != nil {
		return s.createContact(ctx, input)
	}
	return parties.Contact{}, nil
}

func (s stubAccountingAdminService) CloseAccountingPeriod(ctx context.Context, input accounting.CloseAccountingPeriodInput) (accounting.AccountingPeriod, error) {
	if s.closeAccountingPeriod != nil {
		return s.closeAccountingPeriod(ctx, input)
	}
	return accounting.AccountingPeriod{}, nil
}

type stubInventoryAdminService struct {
	listItems      func(context.Context, inventoryops.ListItemsInput) ([]inventoryops.Item, error)
	createItem     func(context.Context, inventoryops.CreateItemInput) (inventoryops.Item, error)
	updateItem     func(context.Context, inventoryops.UpdateItemStatusInput) (inventoryops.Item, error)
	listLocations  func(context.Context, inventoryops.ListLocationsInput) ([]inventoryops.Location, error)
	createLocation func(context.Context, inventoryops.CreateLocationInput) (inventoryops.Location, error)
	updateLocation func(context.Context, inventoryops.UpdateLocationStatusInput) (inventoryops.Location, error)
}

func (s stubInventoryAdminService) ListItems(ctx context.Context, input inventoryops.ListItemsInput) ([]inventoryops.Item, error) {
	if s.listItems != nil {
		return s.listItems(ctx, input)
	}
	return nil, nil
}

func (s stubInventoryAdminService) CreateItem(ctx context.Context, input inventoryops.CreateItemInput) (inventoryops.Item, error) {
	if s.createItem != nil {
		return s.createItem(ctx, input)
	}
	return inventoryops.Item{}, nil
}

func (s stubInventoryAdminService) UpdateItemStatus(ctx context.Context, input inventoryops.UpdateItemStatusInput) (inventoryops.Item, error) {
	if s.updateItem != nil {
		return s.updateItem(ctx, input)
	}
	return inventoryops.Item{}, nil
}

func (s stubInventoryAdminService) ListLocations(ctx context.Context, input inventoryops.ListLocationsInput) ([]inventoryops.Location, error) {
	if s.listLocations != nil {
		return s.listLocations(ctx, input)
	}
	return nil, nil
}

func (s stubInventoryAdminService) CreateLocation(ctx context.Context, input inventoryops.CreateLocationInput) (inventoryops.Location, error) {
	if s.createLocation != nil {
		return s.createLocation(ctx, input)
	}
	return inventoryops.Location{}, nil
}

func (s stubInventoryAdminService) UpdateLocationStatus(ctx context.Context, input inventoryops.UpdateLocationStatusInput) (inventoryops.Location, error) {
	if s.updateLocation != nil {
		return s.updateLocation(ctx, input)
	}
	return inventoryops.Location{}, nil
}
