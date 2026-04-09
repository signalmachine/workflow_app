# Web And API Seams

Date: 2026-04-09
Status: Active technical guide
Purpose: explain how the shared backend seam serves both the browser layer and the JSON API layer.

## 1. The seam model

`workflow_app` intentionally uses one backend for both browser and API access.

That means:

1. the Go binary serves the built Svelte application at `/app` and JSON endpoints under `/api/...`
2. browser pages and JSON endpoints call the same services
3. browser actions and API actions share the same auth model
4. Svelte route loads and browser actions use the same reporting and service seams as the API responses
5. the web layer is not a second business backend

This is one of the key reasons the codebase stays coherent even as the UI grows.

## 2. The handler assembly

The shared handler is assembled in `internal/app/api.go`.

```go
handler := &AgentAPIHandler{
	loadProcessor:     loader,
	submissionService: submissionService,
	reviewService:     reviewService,
	approvalService:   approvalService,
	proposalApproval:  proposalApproval,
	authService:       authService,
}
```

The same handler then registers both browser routes and API routes.

```go
mux.HandleFunc(webAppPath, handler.handleSvelteApp)
mux.HandleFunc(webAppPath+"/", handler.handleSvelteApp)
mux.HandleFunc(reviewInboundRequestsPath, handler.handleListInboundRequests)
mux.HandleFunc(agentProcessNextQueuedPath, handler.handleProcessNextQueuedInboundRequest)
```

The browser-serving path is now a served SPA shell plus JSON endpoints, not a separate set of Go HTML page handlers per route.

## 3. Browser pages versus JSON endpoints

Browser handlers usually:

1. resolve a session
2. serve the Svelte shell or redirect after auth-sensitive entry checks
3. let the browser call `/api/...` or route-level load helpers for data
4. keep route continuity stable under `/app`

JSON handlers usually:

1. resolve an actor
2. validate the request body
3. call the shared service
4. return structured JSON

The important point is that both paths still go through the same service objects.

## 4. Submission and process routes

The most important promoted routes are:

1. `GET /app`
2. `GET /app/submit-inbound-request`
3. `POST /api/inbound-requests`
4. `POST /api/agent/process-next-queued-inbound-request`
5. `GET /api/review/...`
6. `GET /app/review/...`

That set covers the main operator loop: sign in, submit, queue, process, review.

## 5. Shared handler responsibilities

The `AgentAPIHandler` is the orchestration point for:

1. auth extraction
2. JSON decoding
3. browser redirects
4. SPA shell and static-asset serving
5. response mapping
6. request routing

Example actor resolution:

```go
func (h *AgentAPIHandler) actorFromRequest(r *http.Request) (identityaccess.Actor, error) {
	if h.authService == nil {
		if actor, err := actorFromHeaders(r); err == nil {
			return actor, nil
		}
		return identityaccess.Actor{}, fmt.Errorf("unauthorized")
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		return identityaccess.Actor{}, err
	}
	return sessionContext.Actor, nil
}
```

That code shows the intent clearly: authentication is a transport concern, but it resolves into a canonical actor that downstream services understand.

## 6. Review reads drive browser continuity

The Svelte browser review routes are built on `internal/reporting`.

This means the browser can navigate across:

1. inbound requests
2. proposals
3. approvals
4. documents
5. accounting
6. inventory
7. work orders
8. audit events

Those pages are not independent data views. They are continuations of the same persisted workflow graph.

## 7. Why the web layer stays Go-served but not browser-owned

The active web stack is a Go-served Svelte application on the same origin as the JSON API.

That choice keeps the shared backend authoritative while still using a stronger browser runtime than the retired Go-template layer.

The architecture rule is:

1. Go owns business logic, workflow rules, approvals, reporting composition, and durable state
2. Svelte owns display, interaction, route composition, and browser ergonomics
3. `/app` and `/api/...` remain one shared seam on the same auth and backend foundation
4. the browser must not become a second business-logic home merely because it is now richer than the earlier template layer

## 8. What to keep stable

Be careful with:

1. route shapes
2. `/app` shell and static-asset behavior
3. redirect behavior
4. auth behavior
5. review filter semantics
6. exact request-reference continuity
7. response mapping between browser and API paths

If these drift, the browser surface can become inconsistent even if the services remain correct.

## 9. Ongoing architecture guardrail

The main long-term maintenance risk in this layer is allowing `internal/app` to become a second business-logic home.

Guardrail:

1. `internal/app` should stay responsible for transport concerns, orchestration wiring, and presentation mapping
2. domain decisions should stay in shared services
3. operator review composition should stay in `internal/reporting`
4. Svelte route growth should not become a reason to move durable business branching into handler code or client-only state

Planned follow-up activity:

1. perform a bounded review of `internal/app` for presentation or transport logic that is drifting toward business-rule ownership
2. where drift is found, plan narrow follow-up refactors that push those decisions back into shared service contracts or reporting read seams without creating a browser-specific backend
