# Web And API Seams

Date: 2026-03-31
Status: Active technical guide
Purpose: explain how the shared backend seam serves both the browser layer and the JSON API layer.

## 1. The seam model

`workflow_app` intentionally uses one backend for both browser and API access.

That means:

1. browser pages and JSON endpoints call the same services
2. browser actions and API actions share the same auth model
3. review pages are rendered from the same reporting reads as the API responses
4. the web layer is not a second business backend

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
mux.HandleFunc(webAppPath, handler.handleWebAppDashboard)
mux.HandleFunc(webSubmitInboundPath, handler.handleWebSubmitInboundRequest)
mux.HandleFunc(reviewInboundRequestsPath, handler.handleListInboundRequests)
mux.HandleFunc(agentProcessNextQueuedPath, handler.handleProcessNextQueuedInboundRequest)
```

## 3. Browser pages versus JSON endpoints

Browser handlers usually:

1. resolve a session
2. load review or submission data
3. render HTML
4. redirect after mutations

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
4. HTML rendering
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

The browser review pages are built on `internal/reporting`.

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

## 7. Why the web layer stays server-rendered

The preferred thin-v1 stack is Go server rendering with standard HTML behavior.

That choice keeps the shared backend authoritative and avoids introducing an unnecessary frontend toolchain. It also keeps the browser and API contracts easier to reason about during workflow-critical changes.

## 8. What to keep stable

Be careful with:

1. route shapes
2. redirect behavior
3. auth behavior
4. review filter semantics
5. exact request-reference continuity
6. response mapping between browser and API paths

If these drift, the browser surface can become inconsistent even if the services remain correct.

## 9. Ongoing architecture guardrail

The main long-term maintenance risk in this layer is allowing `internal/app` to become a second business-logic home.

Guardrail:

1. `internal/app` should stay responsible for transport concerns, orchestration wiring, and presentation mapping
2. domain decisions should stay in shared services
3. operator review composition should stay in `internal/reporting`
4. browser-template growth should not become a reason to move durable business branching into handler code

Planned follow-up activity:

1. perform a bounded review of `internal/app` for presentation or transport logic that is drifting toward business-rule ownership
2. where drift is found, plan narrow follow-up refactors that push those decisions back into shared service contracts or reporting read seams without creating a browser-specific backend
