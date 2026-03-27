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

	"workflow_app/internal/attachments"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
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
	if !strings.Contains(body, `/app/review/inbound-requests?status=queued">Open queued requests</a>`) {
		t.Fatalf("expected dashboard status-summary continuity link, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/run:run-123#run-run-123">Open latest run</a>`) {
		t.Fatalf("expected dashboard latest-run continuity link, body=%s", body)
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
	if !strings.Contains(body, `/app/review/inbound-requests?status=draft">Continue drafts</a>`) {
		t.Fatalf("expected draft entry-point action, body=%s", body)
	}
	if !strings.Contains(body, `Inspect failed requests, understand the break, and restart follow-up work.`) {
		t.Fatalf("expected failed entry-point blurb, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inbound-requests?status=failed">Review failures</a>`) {
		t.Fatalf("expected failed entry-point action, body=%s", body)
	}
	if !strings.Contains(body, `/app/review/inbound-requests?status=cancelled">Recover cancellations</a>`) {
		t.Fatalf("expected cancelled entry-point action, body=%s", body)
	}
	if strings.Index(body, `>draft</span>`) > strings.Index(body, `>failed</span>`) {
		t.Fatalf("expected draft summary card to sort ahead of failed card, body=%s", body)
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
	if !strings.Contains(body, `/app/inbound-requests/REQ-000200">Amend back to draft</a>`) {
		t.Fatalf("expected cancelled request recovery link, body=%s", body)
	}
	if !strings.Contains(body, `provider-backed coordinator execution failed`) {
		t.Fatalf("expected failure reason on dashboard, body=%s", body)
	}
	if !strings.Contains(body, `/app/inbound-requests/REQ-000201">Inspect failure</a>`) {
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

type stubBrowserSessionService struct {
	authenticateSession     func(context.Context, string, string) (identityaccess.SessionContext, error)
	authenticateAccessToken func(context.Context, string) (identityaccess.SessionContext, error)
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
