package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
		if !strings.Contains(recorder.Body.String(), `data-sveltekit-preload-data="hover"`) {
			t.Fatalf("expected svelte shell for %s, got %s", route, recorder.Body.String())
		}
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
	body := recorder.Body.String()
	if !strings.Contains(body, `data-sveltekit-preload-data="hover"`) {
		t.Fatalf("expected svelte app shell body, got %s", body)
	}
	if !strings.Contains(body, `import("/app/_app/immutable/entry/start.`) {
		t.Fatalf("expected app-root entry import with /app base, got %s", body)
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
	if !strings.Contains(recorder.Body.String(), `data-sveltekit-preload-data="hover"`) {
		t.Fatalf("expected svelte shell body, got %s", recorder.Body.String())
	}
}
