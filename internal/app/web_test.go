package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func requireSvelteShell(t *testing.T, body string) {
	t.Helper()

	if !strings.Contains(body, `data-sveltekit-preload-data="hover"`) {
		t.Fatalf("expected svelte shell body, got %s", body)
	}
	if !strings.Contains(body, `import("/app/_app/immutable/entry/start.`) {
		t.Fatalf("expected app-root entry import with /app base, got %s", body)
	}
}

func TestRegisterWebRoutesServesSPAFallback(t *testing.T) {
	handler := &AgentAPIHandler{}
	mux := http.NewServeMux()
	registerWebRoutes(mux, handler)

	for _, route := range []string{"/app", "/app/login", "/app/review/accounting/entry-123"} {
		req := httptest.NewRequest(http.MethodGet, route, nil)
		recorder := httptest.NewRecorder()

		mux.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected status for %s: got %d body=%s", route, recorder.Code, recorder.Body.String())
		}
		requireSvelteShell(t, recorder.Body.String())
	}
}

func TestRegisterWebRoutesServesSPAFallbackAcrossPromotedRouteFamilies(t *testing.T) {
	handler := &AgentAPIHandler{}
	mux := http.NewServeMux()
	registerWebRoutes(mux, handler)

	routes := []string{
		webAppPath,
		webLoginPath,
		webRouteCatalogPath,
		webSettingsPath,
		webAdminPath,
		webAdminAccountingPath,
		webAdminPartiesPath,
		webAdminPartiesPath + "/party-123",
		webAdminAccessPath,
		webAdminInventoryPath,
		webOperationsPath,
		webReviewPath,
		webInventoryHubPath,
		webSubmitInboundPagePath,
		webOperationsFeedPath,
		webAgentChatPath,
		"/app/inbound-requests/REQ-000123",
		webInboundRequestsPath,
		webApprovalsPath,
		webApprovalsPath + "/approval-123",
		webProposalsPath,
		webProposalsPath + "/proposal-123",
		webDocumentsPath,
		webDocumentsPath + "/document-123",
		webAccountingPath,
		webAccountingPath + "/entry-123",
		webInventoryPath,
		webInventoryPath + "/movement-123",
		webWorkOrdersPath,
		webWorkOrdersPath + "/work-order-123",
		webAuditPath,
		webAuditPath + "/event-123",
	}

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, route, nil)
			recorder := httptest.NewRecorder()

			mux.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("unexpected status for %s: got %d body=%s", route, recorder.Code, recorder.Body.String())
			}
			requireSvelteShell(t, recorder.Body.String())
		})
	}
}

func TestHandleSvelteAppServesIndexAtAppRoot(t *testing.T) {
	handler := &AgentAPIHandler{}
	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	recorder := httptest.NewRecorder()

	handler.handleSvelteApp(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	requireSvelteShell(t, recorder.Body.String())
}

func TestHandleSvelteAppServesHeadRequests(t *testing.T) {
	handler := &AgentAPIHandler{}
	req := httptest.NewRequest(http.MethodHead, "/app/review/documents/document-123", nil)
	recorder := httptest.NewRecorder()

	handler.handleSvelteApp(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Body.Len(); got != 0 {
		t.Fatalf("expected head response without body, got %d bytes", got)
	}
}

func TestHandleSvelteAppServesEmbeddedJSAsset(t *testing.T) {
	handler := &AgentAPIHandler{}

	indexReq := httptest.NewRequest(http.MethodGet, "/app", nil)
	indexRecorder := httptest.NewRecorder()
	handler.handleSvelteApp(indexRecorder, indexReq)
	if indexRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected index status: got %d body=%s", indexRecorder.Code, indexRecorder.Body.String())
	}

	const assetPrefix = `import("/app/_app/immutable/entry/start.`
	indexBody := indexRecorder.Body.String()
	start := strings.Index(indexBody, assetPrefix)
	if start == -1 {
		t.Fatalf("expected svelte start asset in index body, got %s", indexBody)
	}
	start += len(`import("`)
	end := strings.Index(indexBody[start:], `.js")`)
	if end == -1 {
		t.Fatalf("expected js asset suffix in index body, got %s", indexBody)
	}
	assetPath := indexBody[start : start+end+len(".js")]

	req := httptest.NewRequest(http.MethodGet, assetPath, nil)
	recorder := httptest.NewRecorder()

	handler.handleSvelteApp(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	contentType := recorder.Header().Get("Content-Type")
	if !strings.Contains(contentType, "javascript") && !strings.Contains(contentType, "text/plain") {
		t.Fatalf("expected javascript-ish content type, got %q body=%s", contentType, recorder.Body.String())
	}
	body := recorder.Body.String()
	if strings.Contains(body, "<!doctype html>") {
		t.Fatalf("expected JS asset body, got HTML shell")
	}
	if !strings.Contains(body, "import") && !strings.Contains(body, "export") {
		snippet := body
		if len(snippet) > 120 {
			snippet = snippet[:120]
		}
		t.Fatalf("expected bundled JS module body, got %s", snippet)
	}
}

func TestHandleSvelteAppDoesNotFallbackForMissingStaticAsset(t *testing.T) {
	handler := &AgentAPIHandler{}
	req := httptest.NewRequest(http.MethodGet, "/app/_app/immutable/entry/missing.js", nil)
	recorder := httptest.NewRecorder()

	handler.handleSvelteApp(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected missing asset to return 404, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(strings.ToLower(recorder.Body.String()), "<!doctype html>") {
		t.Fatalf("expected missing asset response to avoid SPA shell fallback, got %s", recorder.Body.String())
	}
}

func TestNewAgentAPIHandlerWithDependenciesServesSvelteShell(t *testing.T) {
	handler := NewAgentAPIHandlerWithDependencies(
		func() (ProcessNextQueuedInboundRequester, error) { return nil, nil },
		nil,
		nil,
		nil,
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/app/review/documents/document-123", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	requireSvelteShell(t, recorder.Body.String())
}
