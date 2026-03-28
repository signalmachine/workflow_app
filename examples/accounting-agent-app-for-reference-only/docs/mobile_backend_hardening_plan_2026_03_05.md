# Mobile Backend Hardening Plan (Flutter Client)

Date: 2026-03-05  
Status: Planning only (no implementation in this document)

## 1. Objective

Prepare the backend for a production-grade Flutter mobile client used by field workers, while preserving current web behavior.

Primary outcomes:
- Stable and explicit mobile API contracts.
- Mobile-safe authentication and session model.
- Complete field-critical endpoints (especially inventory).
- Predictable error handling, versioning, and compatibility policy.
- Operational safety for multi-instance deployments.

## 2. Scope and Non-Goals

In scope:
- Backend API/auth hardening.
- API contract normalization and documentation.
- Missing endpoint implementation for mobile MVP.
- Reliability improvements for chat confirmation flow.
- Observability, rate limiting, and compatibility guarantees.

Out of scope:
- Flutter UI implementation.
- New business workflows beyond existing domain capabilities.
- Major domain model redesign.

## 3. Current State Summary (Why Hardening Is Needed)

Observed constraints:
- Auth middleware reads only `auth_token` cookie.
- Cookie is `Secure` + `SameSiteStrict`, which is good for browser security, but awkward for native mobile API calls.
- Some inventory routes still return `501 Not Implemented`.
- JSON payload shapes are inconsistent (mix of raw arrays, wrapped objects, and structs without JSON tags).
- README claims bearer auth support, but middleware behavior is cookie-only.
- Chat pending confirmation tokens are in-memory only, not durable across restarts or horizontal scaling.

## 4. Design Principles

- Backward compatible first; web UI must not break.
- Explicit contracts over implicit struct serialization.
- One canonical API behavior for web + mobile.
- Idempotent writes where retries are expected (mobile networks are unstable).
- Prefer additive changes, then deprecate old behavior with a clear timeline.

## 5. Target API/Auth Model

### 5.1 Authentication Strategy

Adopt dual-mode auth:
- Browser: keep cookie-based auth as-is.
- Mobile/API clients: support `Authorization: Bearer <JWT>`.

Required behavior:
- `/api/auth/login` can issue:
  - cookie session (current behavior), and
  - JWT in response body for mobile mode.
- `RequireAuth` accepts either valid bearer token or cookie.
- Role and company checks remain unchanged.

Security requirements:
- Explicit token TTL and refresh policy.
- Rotation-friendly signing secret strategy.
- Optional refresh token endpoint (recommended for long-lived mobile sessions).

### 5.2 API Contract Standardization

Introduce canonical response envelope for all JSON endpoints:
- Success:
  - `data`
  - `meta` (optional: pagination, request_id, warnings)
- Error:
  - `error.code`
  - `error.message`
  - `error.request_id`
  - `error.details` (optional)

Normalize conventions:
- `snake_case` JSON fields.
- RFC3339 timestamps or `YYYY-MM-DD` only where domain-specific.
- Decimal amounts serialized as strings consistently.

### 5.3 Versioning

Add versioned API namespace for hardened contracts:
- `/api/v1/...`

Compatibility policy:
- Keep existing unversioned routes during migration.
- Mark old routes as deprecated once `/api/v1` is stable.
- Remove only after agreed deprecation window.

## 6. Implementation Phases

## Phase 0: Contract and Migration Baseline

Deliverables:
- OpenAPI spec for existing endpoints (as-is behavior).
- Contract gap matrix: current vs target payloads.
- Endpoint inventory tagged with `ready`, `partial`, `not_implemented`.

Acceptance criteria:
- Every current JSON route documented with request/response examples.
- Breaking differences identified and mapped to `/api/v1`.

## Phase 1: Auth Hardening for Mobile

Deliverables:
- Bearer token support in auth middleware.
- Login response supports mobile token issuance.
- Optional refresh token flow (recommended).
- Logout semantics for both cookie and token flows.

Acceptance criteria:
- Existing browser login/logout continues to work unchanged.
- Mobile clients can authenticate without cookie jar management.
- Role-based and company-based authorization parity across auth modes.

## Phase 2: `/api/v1` Foundation + Envelope Normalization

Deliverables:
- New `/api/v1` router group.
- Standard success/error response structures.
- Stable typed DTOs for all v1 handlers (no raw domain structs directly serialized).

Acceptance criteria:
- All v1 endpoints return normalized response shapes.
- Request validation errors are deterministic and machine-readable.
- `request_id` present on all error responses.

## Phase 3: Mobile MVP Endpoint Completion

Priority endpoint groups:
- Auth/session: login, me, logout, refresh (if implemented).
- Sales: customers, products, orders, order lifecycle actions.
- Purchases: vendors, purchase orders, PO lifecycle actions.
- Inventory: warehouses, stock list, stock receive (currently missing).
- Accounting lite: trial balance, account statement, P&L snapshot.

Acceptance criteria:
- No field-worker-critical endpoint returns `501`.
- All create/update routes have idempotency strategy where needed.
- Full happy-path integration tests pass for mobile MVP workflows.

## Phase 4: Chat/API Reliability for Multi-Instance Deployment

Deliverables:
- Replace in-memory pending action store with persistent backing store (DB/Redis).
- Token TTL and replay/duplicate protections preserved.
- Safe behavior under restart and horizontal scaling.

Acceptance criteria:
- Pending confirm actions survive server restart.
- Confirm/cancel works correctly when request hits different instance.

## Phase 5: Operational Hardening

Deliverables:
- Rate limiting per IP/user for auth and write endpoints.
- Audit/event logs for mobile write operations.
- Structured metrics: latency, error rates, auth failures, rate-limit hits.
- Security headers and CORS policy review for mobile clients.

Acceptance criteria:
- SLO dashboards and alerts defined.
- Abuse and brute-force scenarios throttled.

## 7. Testing and Validation Strategy

Unit and integration:
- Middleware tests for cookie vs bearer precedence and failure modes.
- Handler tests for standardized envelope output.
- Integration tests for core field workflows end-to-end.

Compatibility:
- Regression suite for current web endpoints.
- Contract tests for `/api/v1` schemas.

Suggested local validation gate after implementation:
1. `make test`
2. `go test -p 1 ./...`
3. `go run ./cmd/verify-db-health`
4. Auth smoke tests (cookie and bearer modes)
5. Mobile workflow smoke tests (orders, PO, inventory receive)

## 8. Risks and Mitigations

Risk: Breaking existing web clients during auth/contract refactor.  
Mitigation: Keep legacy routes and cookie flow intact until v1 stabilization.

Risk: Serialization drift from direct domain struct exposure.  
Mitigation: Dedicated API DTO layer for all v1 responses.

Risk: Multi-instance inconsistency in chat confirmation.  
Mitigation: Persistent pending-action storage before mobile chat rollout.

Risk: Ambiguous behavior due to README/documentation mismatch.  
Mitigation: Align docs with actual behavior at each phase; publish OpenAPI as source of truth.

## 9. Suggested Work Breakdown (Issue Backlog)

Epic A: Auth and session hardening
- A1: Bearer token parsing in `RequireAuth`.
- A2: Login endpoint mobile token mode.
- A3: Refresh token endpoint and revocation model (if adopted).

Epic B: API contract standardization
- B1: DTO definitions with full JSON tags.
- B2: Response envelope helpers.
- B3: `/api/v1` route group and migration.

Epic C: Inventory API completion
- C1: List warehouses endpoint.
- C2: Stock listing endpoint.
- C3: Stock receive endpoint.

Epic D: Chat durability
- D1: Persistent pending store schema/backing service.
- D2: Confirm/cancel handlers migrated to durable store.

Epic E: Observability and protection
- E1: Rate limiting middleware.
- E2: Structured metrics and logs.
- E3: Security/CORS tightening for mobile origins.

## 10. Rollout Plan

1. Ship Phase 1 and Phase 2 behind feature flags (or separate v1 routes).
2. Run web regression + mobile integration smoke tests in staging.
3. Enable mobile client against `/api/v1` in staging.
4. Production canary release to a limited user group.
5. Full rollout after monitoring period and error budget check.

## 11. Definition of Done (Program Level)

Backend is considered hardened for mobile when:
- Field-worker mobile MVP workflows are fully API-backed.
- Mobile authentication is first-class (bearer supported and tested).
- `/api/v1` contracts are stable and documented in OpenAPI.
- Chat confirmation works reliably in multi-instance deployments.
- Observability and rate limiting are active in production.

## 12. Execution-Ready Ticket Plan (No Implementation Yet)

This section decomposes the roadmap into implementation tickets with concrete code touchpoints.

Legend:
- Priority: `P0` critical, `P1` high, `P2` medium
- Type: `API`, `Auth`, `Infra`, `Docs`, `Test`

### Phase 0 Tickets: Baseline and Contract Inventory

1. Ticket P0-DOC-001: OpenAPI baseline for current endpoints
- Type: `Docs`
- Outcome: machine-readable spec for current behavior.
- Likely touchpoints:
  - `docs/` new OpenAPI file (`openapi_current.yaml`)
  - `README.md` API section links to spec
- Validation:
  - Spec lints successfully
  - Every current route in `internal/adapters/web/handlers.go` represented

2. Ticket P0-DOC-002: Endpoint readiness matrix
- Type: `Docs`
- Outcome: `ready/partial/not_implemented` inventory for mobile.
- Likely touchpoints:
  - `docs/mobile_backend_hardening_plan_2026_03_05.md` (appendix table)
  - `internal/adapters/web/handlers.go` as source mapping
- Validation:
  - Matrix includes auth, sales, purchase, inventory, accounting, chat

### Phase 1 Tickets: Auth Hardening for Mobile

1. Ticket P0-AUTH-101: Bearer token support in `RequireAuth`
- Type: `Auth`
- Outcome: middleware accepts either cookie or `Authorization: Bearer`.
- Likely touchpoints:
  - `internal/adapters/web/auth.go` (`RequireAuth`, shared token parsing helper)
  - `internal/adapters/web/middleware.go` (if auth header/cors adjustments needed)
- Validation:
  - Cookie flow unchanged
  - Bearer flow authenticated equivalently
  - Role/company checks unchanged

2. Ticket P0-AUTH-102: Login endpoint mobile token mode
- Type: `Auth`
- Outcome: `/api/auth/login` supports explicit mobile mode response body token.
- Likely touchpoints:
  - `internal/adapters/web/auth.go` (`login` handler request/response DTO)
  - `internal/app/result_types.go` (if session payload shape formalized)
- Validation:
  - Web mode still sets cookie
  - Mobile mode returns token metadata (expiry, token_type)

3. Ticket P1-AUTH-103: Refresh token flow (recommended)
- Type: `Auth`
- Outcome: long-lived mobile sessions without frequent password re-login.
- Likely touchpoints:
  - `internal/adapters/web/auth.go` (refresh endpoint)
  - `internal/adapters/web/handlers.go` (route registration)
  - `migrations/` new table for refresh tokens or session records
  - `internal/app/app_service.go` (service methods if needed)
- Validation:
  - Token rotation works
  - Revocation behavior documented and tested

4. Ticket P1-AUTH-104: README/auth docs alignment
- Type: `Docs`
- Outcome: docs match real auth behavior.
- Likely touchpoints:
  - `README.md`
  - `docs/deployment/*.md`
- Validation:
  - No contradiction between docs and middleware behavior

### Phase 2 Tickets: `/api/v1` and Contract Normalization

1. Ticket P0-API-201: Introduce v1 router namespace
- Type: `API`
- Outcome: `/api/v1/...` route group for hardened contracts.
- Likely touchpoints:
  - `internal/adapters/web/handlers.go` (new v1 groups)
- Validation:
  - Existing unversioned routes remain functional
  - v1 routes reachable and tested

2. Ticket P0-API-202: Standard response envelope helpers
- Type: `API`
- Outcome: consistent success/error JSON for v1.
- Likely touchpoints:
  - `internal/adapters/web/errors.go` (new helper variants for v1)
  - `internal/adapters/web/*` handler files (v1 handlers)
- Validation:
  - All v1 responses follow envelope contract
  - Error includes `request_id` and `code`

3. Ticket P0-API-203: DTO layer for v1 responses/requests
- Type: `API`
- Outcome: explicit JSON tags and stable field naming.
- Likely touchpoints:
  - `internal/adapters/web/` new DTO file(s), e.g. `dto_v1.go`
  - Optional refactor away from direct domain serialization in handlers:
    - `orders.go`, `vendors.go`, `accounting.go`, `users.go`, `chat.go`
- Validation:
  - `snake_case` keys only in v1 payloads
  - No accidental PascalCase serialization in v1

4. Ticket P1-API-204: Deprecation header on legacy endpoints
- Type: `API`
- Outcome: migration signaling for clients.
- Likely touchpoints:
  - `internal/adapters/web/handlers.go` / middleware for legacy routes
- Validation:
  - Legacy endpoints return deprecation metadata

### Phase 3 Tickets: Mobile MVP Endpoint Completion

1. Ticket P0-API-301: Implement warehouses endpoint
- Type: `API`
- Outcome: replace `501` on warehouses listing.
- Likely touchpoints:
  - `internal/adapters/web/handlers.go` (route binding from `notImplemented`)
  - `internal/adapters/web/orders.go` or new `inventory_api.go`
  - `internal/app/service.go` (`ListWarehouses` already exists)
  - `internal/app/app_service.go` (`ListWarehouses` already exists)
- Validation:
  - Returns company-scoped warehouse list

2. Ticket P0-API-302: Implement stock endpoint
- Type: `API`
- Outcome: replace `501` for stock levels.
- Likely touchpoints:
  - `internal/adapters/web/handlers.go`
  - `internal/adapters/web/orders.go` or new `inventory_api.go`
  - `internal/app/service.go` (`GetStockLevels` exists)
- Validation:
  - Returns stock for requested company with expected shape

3. Ticket P0-API-303: Implement stock receive endpoint
- Type: `API`
- Outcome: replace `501` for inventory receipt workflow.
- Likely touchpoints:
  - `internal/adapters/web/handlers.go`
  - `internal/adapters/web/orders.go` or new `inventory_api.go`
  - `internal/app/request_types.go` (`ReceiveStockRequest` exists)
  - `internal/app/app_service.go` (`ReceiveStock` exists)
- Validation:
  - Successful goods receipt updates stock and books accounting impact

4. Ticket P1-API-304: Idempotency policy for mobile writes
- Type: `API`
- Outcome: safe retries for unstable mobile networks.
- Likely touchpoints:
  - `internal/adapters/web/middleware.go` (idempotency key extraction helper)
  - write handlers in `orders.go`, `vendors.go`, `accounting.go`
  - optional persistence support in DB/migrations
- Validation:
  - Duplicate retries do not create duplicate business transactions

### Phase 4 Tickets: Chat Durability and Multi-Instance Safety

1. Ticket P0-INFRA-401: Persistent pending-action store design
- Type: `Infra`
- Outcome: architecture decision (DB vs Redis).
- Likely touchpoints:
  - `docs/` ADR document
  - schema design notes
- Validation:
  - Decision includes TTL, cleanup, replay safety, and failover behavior

2. Ticket P0-INFRA-402: Persistent pending-action implementation
- Type: `Infra`
- Outcome: replace in-memory map-based pending store.
- Likely touchpoints:
  - `internal/adapters/web/chat.go` (replace `pendingStore` usage)
  - `internal/app/service.go` (optional service abstraction if introduced)
  - `migrations/` new table if DB-backed
- Validation:
  - Pending tokens survive restart
  - Confirm works across instances

3. Ticket P1-TEST-403: Chat durability integration tests
- Type: `Test`
- Outcome: regression safety for token lifecycle.
- Likely touchpoints:
  - `internal/app/*_integration_test.go` (if service-level)
  - `internal/adapters/web/` tests (new)
- Validation:
  - token create/get/expire/confirm/cancel scenarios pass

### Phase 5 Tickets: Operational Hardening

1. Ticket P0-INFRA-501: Rate limiting for auth/write endpoints
- Type: `Infra`
- Outcome: reduce brute force and abuse risk.
- Likely touchpoints:
  - `internal/adapters/web/middleware.go` (rate limiter middleware)
  - `internal/adapters/web/handlers.go` (middleware wiring by route group)
- Validation:
  - Limits enforced with clear error code

2. Ticket P1-INFRA-502: Structured telemetry and metrics
- Type: `Infra`
- Outcome: API observability for mobile rollout.
- Likely touchpoints:
  - `internal/adapters/web/middleware.go` (metrics instrumentation)
  - `cmd/server/main.go` (telemetry bootstrap if needed)
- Validation:
  - Request latency/error/auth-failure metrics visible in dashboards

3. Ticket P1-INFRA-503: CORS policy hardening for mobile clients
- Type: `Infra`
- Outcome: explicit allowed origins/headers per environment.
- Likely touchpoints:
  - `internal/adapters/web/middleware.go` (CORS settings)
  - `.env.example` and deployment docs
- Validation:
  - Only approved origins can call browser-cross-origin APIs

### Cross-Phase Testing Tickets

1. Ticket P0-TEST-901: Auth parity test suite (cookie + bearer)
- Touchpoints:
  - new tests under `internal/adapters/web/`
- Validates:
  - identical authorization decisions for both auth modes

2. Ticket P0-TEST-902: Contract test suite for `/api/v1`
- Touchpoints:
  - contract fixtures in `docs/` or `testdata/`
  - handler tests
- Validates:
  - schema stability and envelope compliance

3. Ticket P1-TEST-903: Field-worker workflow e2e suite
- Touchpoints:
  - integration tests across sales/purchase/inventory
- Validates:
  - core mobile workflows pass in serial DB test mode

## 13. Suggested Sequence and Dependencies

Recommended execution order:
1. Phase 0 (baseline docs/spec)
2. Phase 1 (auth)
3. Phase 2 (v1 contracts)
4. Phase 3 (inventory completion + mobile MVP endpoints)
5. Phase 4 (chat durability)
6. Phase 5 (operational hardening)

Dependency notes:
- Phase 1 is prerequisite for native mobile clients in production.
- Phase 2 should begin before broad mobile development to avoid client churn.
- Phase 4 is required before multi-instance scaling of mobile chat confirmations.
- Phase 5 can partly run in parallel after Phase 2 starts.

## 14. Per-Ticket Definition Template

Use this template when creating implementation tickets:
- Title
- Problem statement
- Scope (in/out)
- Touchpoints (files/modules)
- API contract changes
- Migration impact
- Test plan
- Rollback plan
- Acceptance criteria
