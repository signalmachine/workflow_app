# Svelte Web Runtime

Date: 2026-04-12
Status: Active technical guide
Purpose: describe the current promoted Svelte web UI runtime as implemented, including ownership boundaries, route shape, build output, and verification expectations.

## 1. Current runtime

The promoted browser UI is a SvelteKit application in `web/`.

It is no longer a Go `html/template` UI. The earlier template-based browser layer has been retired from the active runtime.

Current runtime shape:

1. SvelteKit owns the browser route tree, layouts, component composition, client-side interaction, and same-origin API calls
2. Go serves the built Svelte static bundle under `/app`
3. Go also serves JSON API endpoints under `/api/...`
4. both browser and API calls use the same session and actor model
5. business logic, workflow rules, reporting composition, approvals, posting, and persistence remain backend-owned

The core rule is that Svelte is the display and interaction layer. It must not become a browser-local business backend.

## 2. Build and serving model

The Svelte build is configured in `web/svelte.config.js`:

1. `@sveltejs/adapter-static` writes pages and assets to `../internal/app/web_dist`
2. the SvelteKit base path is `/app`
3. the static fallback file is `200.html`

The Go embed layer lives in `internal/app/web_static.go`:

1. `//go:embed all:web_dist` embeds the built frontend artifact
2. `handleSvelteApp` serves `200.html` for `/app` and non-asset `/app/...` routes
3. static asset requests with filename extensions are served from the embedded bundle when present
4. missing static asset requests return `404` rather than falling back to the SPA shell

This means frontend source changes under `web/` are not visible to the Go-served runtime until the frontend build has been rerun and `internal/app/web_dist` has been refreshed.

## 3. Route structure

The Svelte route tree uses route groups:

1. `web/src/routes/(public)/login` for browser sign-in
2. `web/src/routes/(app)/...` for authenticated application routes
3. `web/src/routes/(app)/+layout.ts` for session loading and unauthenticated redirect handling
4. `web/src/routes/(app)/+layout.svelte` for the shared application shell

The promoted route families include:

1. `/app`
2. `/app/routes`
3. `/app/settings`
4. `/app/admin`
5. `/app/admin/master-data`
6. `/app/admin/lists`
7. `/app/admin/accounting`
8. `/app/admin/parties`
9. `/app/admin/parties/{party_id}`
10. `/app/admin/access`
11. `/app/admin/inventory`
12. `/app/submit-inbound-request`
13. `/app/operations`
14. `/app/operations-feed`
15. `/app/agent-chat`
16. `/app/inventory`
17. `/app/inbound-requests/{request_reference_or_id}`
18. `/app/review`
19. `/app/review/inbound-requests`
20. `/app/review/proposals`
21. `/app/review/proposals/{recommendation_id}`
22. `/app/review/approvals`
23. `/app/review/approvals/{approval_id}`
24. `/app/review/documents`
25. `/app/review/documents/{document_id}`
26. `/app/review/accounting`
27. `/app/review/accounting/journal-entries`
28. `/app/review/accounting/control-balances`
29. `/app/review/accounting/tax-summaries`
30. `/app/review/accounting/trial-balance`
31. `/app/review/accounting/balance-sheet`
32. `/app/review/accounting/income-statement`
33. `/app/review/accounting/{entry_id}`
34. `/app/review/inventory`
35. `/app/review/inventory/{movement_id}`
36. `/app/review/work-orders`
37. `/app/review/work-orders/{work_order_id}`
38. `/app/review/audit`
39. `/app/review/audit/{event_id}`

The list above should stay aligned with `docs/workflows/application_workflow_catalog.md` and the served-route tests in `internal/app`.

## 4. Data and auth flow

The Svelte API client lives under `web/src/lib/api/`.

Current pattern:

1. `apiRequest` sends same-origin requests with `credentials: 'same-origin'`
2. browser session cookies stay `HttpOnly` and are not copied into browser storage
3. API errors are normalized into `APIClientError`
4. route loads and page actions call the shared `/api/...` seam rather than reaching into database or domain state directly
5. `web/src/lib/api/types.ts` mirrors the JSON contracts returned by the Go API handlers

The protected app layout calls `GET /api/session` through `getCurrentSession`. If that request returns `401`, the route redirects to the login page with the original destination encoded as `next`.

## 5. Shell and navigation model

The shared shell is implemented under `web/src/lib/components/shell/`.

Current shape:

1. `AppShell.svelte` composes the authenticated operator shell
2. `TopBar.svelte` owns the top application bar and session controls
3. `SideNav.svelte` owns major-area navigation
4. `ContextTabs.svelte` owns the local section tabs for the active major area
5. desktop sidebar collapse preference is read through `web/src/lib/utils/shell.ts`

The promoted model is desktop-first:

1. the top bar stays above both the sidebar and content
2. the sidebar represents durable major areas such as Agent, Accounting, Inventory, Operations, Admin, and Settings
3. contextual tabs begin over the main content column, not over the sidebar
4. grouped directory pages such as Admin `Master Data`, Admin `Lists`, and Accounting report directory routes keep crowded areas from becoming one large mixed workspace

## 6. Ownership boundaries

Svelte may own:

1. route composition
2. page layout
3. form state and feedback
4. table and filter presentation
5. local navigation state
6. API client convenience wrappers

Svelte must not own:

1. workflow lifecycle rules
2. approval decision rules
3. accounting report composition
4. posting or reversal rules
5. inventory or work-order business invariants
6. authorization decisions beyond redirecting unauthenticated users
7. durable truth that belongs in PostgreSQL-backed services

When a page needs richer behavior, prefer adding or refining a shared `/api/...` contract in Go over duplicating domain branching in Svelte.

## 7. Verification expectations

For frontend-only changes:

```bash
npm --prefix web run check
npm --prefix web run build
```

Run focused Vitest route or component tests when the changed surface has test coverage.

For changes that affect served runtime behavior, also verify the Go serving seam:

```bash
go build ./cmd/... ./internal/...
set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...
```

For browser-critical workflow changes:

1. rebuild the frontend artifact before diagnosing served-browser behavior
2. restart `cmd/app` after rebuilding
3. use Playwright or the bounded workflow checklist when the real question is rendered `/app` continuity
4. verify that real static assets under `/app/_app/...` load and missing assets return `404`
5. record workflow evidence in `docs/workflows/workflow_validation_track.md` when supported workflow behavior changes

## 8. Documentation maintenance

This guide is the durable current-state technical guide for the Svelte browser runtime.

Do not use implementation-era migration material as the canonical current-state source. If the Svelte runtime, route model, serving model, or frontend ownership boundaries change, update this guide and then update downstream user or workflow docs as needed.
