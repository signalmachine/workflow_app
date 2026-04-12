# 18 — Code Review and Improvement Plan

Date: 2026-04-12  
Validated: 2026-04-12 (code-verified against live source)  
Status: Active — pending implementation  
Priority: P1 items should be resolved before broad production use. P2–P3 items should be tracked in the implementation plan and resolved progressively. P4–P5 items are lower-urgency improvements.

This document captures the findings from the April 2026 full-application code review and prescribes concrete fixes for each issue. The scope covers the Go backend (`internal/`, `cmd/`), the Svelte frontend (`web/`), module configuration, and security posture.

---

## How to Use This Document

Issues are grouped by priority tier:

- **P1 — Defects:** Incorrect behavior. Fix before production or broad user-testing expansion.
- **P2 — Structural issues:** Maintainability and correctness risks that accumulate with growth. Fix as part of the next planned refactor milestone.
- **P3 — Design improvements:** Correctness edge cases, performance considerations, and patterns that should be hardened. Fix opportunistically or as part of a bounded hardening pass.
- **P4 — Frontend improvements:** Svelte-layer issues that affect type safety, error handling, and long-term maintainability.
- **P5 — Module and configuration issues:** One-time setup or configuration corrections with broad impact.

> **Validation note:** Every issue in this document was verified against the live source files before being recorded. Line numbers, function names, and code excerpts are cross-referenced with the actual codebase as of 2026-04-12.

Each issue has: a description, affected files, root cause, and a concrete fix with code.

---

## P1 — Defects

### P1-1 · Authentication errors returned as HTTP 400 instead of 401

**Affected files:**
- `internal/app/api_review_handlers.go` — 25+ call sites (all review handlers)
- `internal/app/api_inbound_handlers.go` — lines 97, 198, 374
- `internal/app/api_approval_handlers.go` — lines 44, 101

**Not affected** (already correct): all navigation handlers (`api_navigation_handlers.go`) call `actorFromRequest` and correctly return `http.StatusUnauthorized`; all admin handlers use `adminActorFromRequest` via `writeAdminActorError`; `handleProcessNextQueuedInboundRequest` also handles this correctly.

**Description:**

`actorFromRequest()` can return `identityaccess.ErrUnauthorized` when no valid session cookie or bearer token is present. At review and inbound handler call sites, the error is handled with a generic 400 response that also leaks `err.Error()` directly to the client:

```go
// current — WRONG at review and inbound call sites
actor, err := h.actorFromRequest(r)
if err != nil {
    writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
    return
}
```

`http.StatusBadRequest` is semantically wrong — a missing or invalid session is an authentication failure, not a bad client request. The leaked error string (e.g. from `actorFromHeaders`: `"missing required authentication headers"`, `"authentication headers must be UUIDs"`, or `"unauthorized"`) reveals internal implementation detail.

The Svelte client's 401-redirect flow depends on receiving a real 401, so 400 responses cause silent failures instead of driving the user to the login page.

The correct pattern already exists in `handleProcessNextQueuedInboundRequest` and in the admin handler's `writeAdminActorError` helper (`api_admin_handlers.go` lines 15–21):

```go
// writeAdminActorError — the correct shape, used by admin handlers
func writeAdminActorError(w http.ResponseWriter, err error) {
    if errors.Is(err, identityaccess.ErrUnauthorized) {
        writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
        return
    }
    writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
}
```

Note: `writeAdminActorError` still leaks `err.Error()` on the non-unauthorized branch, which is a minor residual issue for the admin surface (lower severity because admin is a privileged role accessed only by known operators). The `writeActorError` below should not replicate that leak.

**Fix:**

Add a shared `writeActorError` helper in `api.go` (or `api_errors.go` after the refactor in P2-1) and replace all incorrect call sites in review and inbound handlers:

```go
// Add this helper alongside writeAdminActorError
func writeActorError(w http.ResponseWriter, err error) {
    // actorFromRequest only errors when authentication is missing or invalid.
    // All error cases map to 401 — do not leak internal error strings.
    _ = err
    writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
}
```

Replace every pattern of the form:

```go
actor, err := h.actorFromRequest(r)
if err != nil {
    writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
    return
}
```

with:

```go
actor, err := h.actorFromRequest(r)
if err != nil {
    writeActorError(w, err)
    return
}
```

This applies to all handlers in:
- `api_review_handlers.go` (all handlers — confirmed 25+ sites)
- `api_inbound_handlers.go` (`handleSubmitInboundRequest` line 97, `handleInboundRequestAction` line 198, `handleDownloadAttachment` line 374)
- `api_approval_handlers.go` (lines 44 and 101)

Do **not** change navigation handlers — they already use `http.StatusUnauthorized` directly and correctly.

---

### P1-2 · GET-only detail handlers return HTTP 404 for wrong method instead of 405

**Affected file:** `internal/app/api_review_handlers.go`

**Confirmed affected handlers:**
- `handleGetInboundRequestDetail` (line 51–53)
- `handleGetProcessedProposalDetail` (line 170–172)
- `handleGetApprovalQueueDetail` (line 291–293)
- `handleGetDocumentReview` (line 371–373)
- `handleGetJournalEntryDetail` (line 447–449)
- `handleGetInventoryMovementDetail` (line 746–748)
- `handleGetWorkOrderReview` (line 886–888)
- `handleGetAuditEventDetail` (line 961–963)

**Description:**

These GET-only detail handlers check the HTTP method and respond with `http.NotFound` for a method mismatch:

```go
// current — WRONG (example from handleGetInboundRequestDetail line 51)
if r.Method != http.MethodGet {
    http.NotFound(w, r) // returns 404 HTML page, not 405 JSON
    return
}
```

A client sending `POST /api/review/inbound-requests/{id}` receives a 404 HTML page, which is misleading and inconsistent with all other handlers that correctly use 405. Note that `http.NotFound` also writes an HTML body, not JSON, breaking the API contract.

**Fix:**

Replace the `http.NotFound` with the standard 405 response in all eight handlers:

```go
// correct
if r.Method != http.MethodGet {
    writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
    return
}
```

---

### P1-3 · `GET /api/session` leaks internal error on non-unauthorized failures

**Affected file:** `internal/app/api_session_handlers.go` — lines 127–133

**Description:**

In `handleCurrentSession`, when `sessionContextFromRequest` returns an error that is not `ErrUnauthorized`, the handler responds with `400 Bad Request` and passes `err.Error()` directly to the client (line 132):

```go
// current — line 127–133 of api_session_handlers.go
context, err := h.sessionContextFromRequest(r)
if err != nil {
    if errors.Is(err, identityaccess.ErrUnauthorized) {
        writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
        return
    }
    writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()}) // leaks internal error string
    return
}
```

`sessionContextFromRequest` wraps calls to `authService.AuthenticateSession` and `authService.AuthenticateAccessToken`. A database connectivity error, unexpected panic-recover, or auth provider internal error would produce a 400 with the raw Go error string visible to the client. Additionally, 400 is the wrong status — this is an authentication failure or a server-side fault, neither of which is a client request error.

**Fix:**

```go
// Replace lines 127–133 with:
context, err := h.sessionContextFromRequest(r)
if err != nil {
    if errors.Is(err, identityaccess.ErrUnauthorized) {
        writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
        return
    }
    // Any other error is an internal failure, not a client request error.
    // Do not expose the raw error string.
    writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to load session"})
    return
}
```

---

### P1-4 · Inventory landing API handler is entirely unimplemented

**Affected area:** `internal/app/api_navigation_handlers.go`, `internal/app/api.go` (route registration)

**Description:**

`GetInventoryLandingSnapshot` is defined in:
1. The `reporting.Service` implementation (`internal/reporting/service.go` line 953)
2. The `InventoryLandingSnapshot` struct (`internal/reporting/service.go` line 260)
3. The `operatorReviewReader` interface (`internal/app/api.go` line 174)

However, there is **no HTTP handler** for it. Grepping for `inventoryLandingPath`, `navigationInventoryPath`, or `GetInventoryLandingSnapshot` in `api_navigation_handlers.go` returns no results. The mux registration loop in `newAgentAPIHandlerWithDependencies` (lines 579–637) has no entry for an inventory landing navigation endpoint.

The route catalog (`api_navigation_handlers.go` line 236) advertises `webInventoryHubPath` (`/app/inventory`) to the Svelte frontend, but there is no corresponding `/api/navigation/inventory` endpoint to back it with live data. The Svelte frontend also has no reference to any inventory hub API call.

This means the inventory hub page in the Svelte app either loads without data or is unreachable. It is a functional gap.

**Fix:**

1. Add a path constant:
```go
navigationInventoryPath = "/api/navigation/inventory"
```

2. Implement the handler in `api_navigation_handlers.go`:
```go
func (h *AgentAPIHandler) handleGetNavigationInventory(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != navigationInventoryPath {
        http.NotFound(w, r)
        return
    }
    if r.Method != http.MethodGet {
        writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
        return
    }
    if h.reviewService == nil {
        writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
        return
    }

    actor, err := h.actorFromRequest(r)
    if err != nil {
        writeActorError(w, err) // use new helper from P1-1
        return
    }

    snapshot, err := h.reviewService.GetInventoryLandingSnapshot(r.Context(), actor, 20)
    if err != nil {
        handleReviewError(w, err, "failed to load inventory landing")
        return
    }

    // Map snapshot fields to existing inventory response types
    stock := make([]inventoryStockResponse, 0, len(snapshot.Stock))
    for _, item := range snapshot.Stock {
        stock = append(stock, mapInventoryStock(item))
    }
    movements := make([]inventoryMovementResponse, 0, len(snapshot.Movements))
    for _, item := range snapshot.Movements {
        movements = append(movements, mapInventoryMovement(item))
    }
    recon := make([]inventoryReconciliationResponse, 0, len(snapshot.Reconciliation))
    for _, item := range snapshot.Reconciliation {
        recon = append(recon, mapInventoryReconciliation(item))
    }

    writeJSON(w, http.StatusOK, struct {
        Stock          []inventoryStockResponse          `json:"stock"`
        Movements      []inventoryMovementResponse       `json:"movements"`
        Reconciliation []inventoryReconciliationResponse `json:"reconciliation"`
    }{
        Stock:          stock,
        Movements:      movements,
        Reconciliation: recon,
    })
}
```

3. Register the route in `newAgentAPIHandlerWithDependencies` (after line 621):
```go
mux.HandleFunc(navigationInventoryPath, handler.handleGetNavigationInventory)
```

4. Add the corresponding TypeScript interface to `web/src/lib/api/types.ts` (see P4-4) and implement the Svelte inventory hub page data loader.

---

## P2 — Structural Issues

### P2-1 · `api.go` is a 2317-line God file

**Affected file:** `internal/app/api.go`

**Description:**

`api.go` currently contains: all route constants, all request/response DTO structs, the `AgentAPIHandler` struct definition, all service interfaces, all handler constructor variants, all utility helpers (JSON encode/decode, cookie helpers, path parsers, primitives), and all mapping functions. This makes the file ~105 KB and the dominant navigation surface for the entire `app` package.

The handler dispatch has already been partially split across `api_*_handlers.go` files, which is the right direction. What remains in `api.go` is everything else — a mix of orthogonal concerns.

**Fix — Recommended file decomposition:**

| New file | Content to move |
|---|---|
| `internal/app/api_constants.go` | All `const` path/header/cookie names and durations |
| `internal/app/api_interfaces.go` | All service interface definitions (`inboundRequestSubmitter`, `operatorReviewReader`, etc.) |
| `internal/app/api_types.go` | All request/response DTO structs (`submitInboundRequestRequest`, `ledgerAccountResponse`, etc.) |
| `internal/app/api_constructor.go` | All `New*` constructors, DI wiring, mux registration |
| `internal/app/api_helpers.go` | `decodeJSONBody`, `writeJSON`, `writeJSONBodyError`, `parseLimit`, `parseOptionalDate`, `parseChildPath`, `parseChildActionPath`, `parseAttachmentContentPath`, `parseApprovalDecisionPath`, `contentDisposition`, `bearerTokenFromRequest`, `sessionCookiesFromRequest`, `cookieValue`, `setSessionCookies`, `clearSessionCookies`, `sessionCookiesShouldBeSecure` |
| `internal/app/api_mappers.go` | All `map*` functions (`mapLedgerAccount`, `mapParty`, `mapInboundRequestDetail`, etc.) and primitive pointer helpers (`stringPtr`, `timePtr`, `int64Ptr`, `nullableString`, `nullStringPtr`) |
| `internal/app/api_errors.go` | All `handle*Error` functions (`handleReviewError`, `handleAccountingAdminError`, `handlePartyAdminError`, `handleInventoryAdminError`, `handleAccessAdminError`, `writeActorError`, `writeAdminActorError`) |
| `internal/app/api.go` | Keep only: `AgentAPIHandler` struct, `actorFromRequest`, `adminActorFromRequest`, `sessionContextFromRequest`, `actorFromHeaders`, `handleRoot` |

This decomposition results in ~8 focused files of ~100–400 lines each rather than one 2300-line file. No behavioral changes are needed — this is a pure mechanical reorganization.

---

### P2-2 · `accounting/service.go` is a 3358-line God file

**Affected file:** `internal/accounting/service.go`

**Description:**

The entire accounting domain — error sentinels, domain types, all input/output structs, and every service method — lives in a single 90 KB file. This includes: ledger account management, tax codes, accounting periods, journal posting, reversal, invoice document lifecycle, payment/receipt lifecycle, work order labor and inventory posting, and financial statement generation.

**Fix — Recommended file decomposition:**

| New file | Content |
|---|---|
| `internal/accounting/errors.go` | All `Err*` sentinel vars and `StatusActive`/`StatusInactive` constants |
| `internal/accounting/types.go` | All domain struct types (`LedgerAccount`, `JournalEntry`, `TaxCode`, etc.) and all `*Input` structs |
| `internal/accounting/service.go` | Keep only: `Service` struct, `NewService`, shared private helpers (`isValidSetupStatus`, `nullableString`, `hashPostingFingerprint`, etc.) |
| `internal/accounting/ledger.go` | `CreateLedgerAccount`, `UpdateLedgerAccountStatus`, `ListLedgerAccounts`, `createLedgerAccountTx`, `scanLedgerAccount` |
| `internal/accounting/taxcode.go` | `CreateTaxCode`, `UpdateTaxCodeStatus`, `ListTaxCodes`, `createTaxCodeTx`, `scanTaxCode` |
| `internal/accounting/period.go` | `CreateAccountingPeriod`, `CloseAccountingPeriod`, `ListAccountingPeriods`, `createAccountingPeriodTx`, `closeAccountingPeriodTx`, `scanAccountingPeriod` |
| `internal/accounting/journal.go` | `PostDocument`, `ReverseDocument`, `ListJournalEntries`, related `Tx` helpers, `postDocumentTx`, `reverseDocumentTx`, `scanJournalEntry`, `scanJournalLine` |
| `internal/accounting/invoice.go` | `CreateInvoice`, `createInvoiceTx`, `scanInvoiceDocument` |
| `internal/accounting/payment.go` | `CreatePaymentReceipt`, `createPaymentReceiptTx`, `scanPaymentReceiptDocument` |
| `internal/accounting/workorder.go` | `PostWorkOrderLabor`, `PostWorkOrderInventory`, `postWorkOrderLaborTx`, `postWorkOrderInventoryTx`, labor/inventory handoff scanning |
| `internal/accounting/reports.go` | `GetTrialBalance`, `GetBalanceSheet`, `GetIncomeStatement`, `ListControlAccountBalances`, `ListTaxSummaries`, and all report scanning/mapping |

This is a same-package reorganization — no export changes or interface changes needed. Run `go build ./internal/accounting/...` after each file move to confirm nothing breaks.

---

### P2-3 · Six public factory constructors — two exact duplicates, four with inconsistent signatures

**Affected file:** `internal/app/api.go` — lines 515–548

**Description:**

There are six public constructors for `AgentAPIHandler`. Two are exact duplicates. The other four have inconsistent signatures that make it unclear which to use for which purpose:

```go
// Lines 515–521 — EXACT DUPLICATE: identical bodies
func NewAgentAPIHandler(db *sql.DB) http.Handler { return newAgentAPIHandler(db) }
func NewServedAgentAPIHandler(db *sql.DB) http.Handler { return newAgentAPIHandler(db) }

// Lines 534–540 — narrow test/partial-dep constructors
func NewAgentAPIHandlerWithProcessorLoader(loader ...) http.Handler
func NewAgentAPIHandlerWithServices(loader ..., submissionService ...) http.Handler

// Lines 542–548 — NEAR DUPLICATE: same signature, same body
func NewAgentAPIHandlerWithDependencies(loader, submission, review, approval, auth) http.Handler {
    return newAgentAPIHandlerWithDependencies(loader, submission, review, approval, nil, nil, auth)
}
func NewServedAgentAPIHandlerWithDependencies(loader, submission, review, approval, auth) http.Handler {
    return newAgentAPIHandlerWithDependencies(loader, submission, review, approval, nil, nil, auth)
}
```

Problems:
1. `NewAgentAPIHandler` and `NewServedAgentAPIHandler` are literally the same function.
2. `NewAgentAPIHandlerWithDependencies` and `NewServedAgentAPIHandlerWithDependencies` take the same parameters and produce the same result.
3. All four `WithDependencies` variants pass `nil` for `proposalApproval` and `accountingAdmin`, which means they produce a handler with reduced functionality — the accounting admin and proposal approval seams are not wired. This is only appropriate for tests.
4. The actual production constructor is the private `newAgentAPIHandler`, wired via `NewAgentAPIHandler`. The public `WithDependencies` variants are test-facing but are not marked as such.

**Fix:**

1. Delete `NewServedAgentAPIHandler` — it is identical to `NewAgentAPIHandler`.
2. Delete `NewServedAgentAPIHandlerWithDependencies` — it is identical to `NewAgentAPIHandlerWithDependencies`.
3. Keep `NewAgentAPIHandlerWithProcessorLoader` and `NewAgentAPIHandlerWithServices` for test use — mark them with a `// For testing only.` comment.
4. Keep `NewAgentAPIHandlerWithDependencies` for integration tests — add a `// For testing only.` comment and document that `proposalApproval` and `accountingAdmin` are nil in this path.
5. Add a brief comment on `NewAgentAPIHandler` noting it is the canonical production entry point.
6. Grep `cmd/` and test files to confirm which public constructors are actually used before deleting.

---

### P2-4 · `...any` variadic injection for optional services in handler constructor

**Affected file:** `internal/app/api.go` — lines 550–560

**Description:**

`newAgentAPIHandlerWithDependencies` uses an `...any` variadic to thread in `partiesAdminService` and `inventoryAdminService` (confirmed at lines 550–560):

```go
func newAgentAPIHandlerWithDependencies(
    loader queuedInboundRequestProcessorLoader,
    submissionService inboundRequestSubmitter,
    reviewService operatorReviewReader,
    approvalService approvalDecisionService,
    proposalApproval proposalApprovalService,
    accountingAdmin accountingAdminService,
    authService browserSessionService,
    optionalServices ...any,     // ← type-unsafe escape hatch
) http.Handler {
    var partyAdminService partiesAdminService
    var inventoryAdmin inventoryAdminService
    for _, svc := range optionalServices {
        switch typed := svc.(type) {
        case partiesAdminService:
            partyAdminService = typed
        case inventoryAdminService:
            inventoryAdmin = typed
        }
    }
```

This is type-unsafe — the compiler cannot enforce what is passed in the variadic. The caller at line 531 passes `partiesService` and `inventoryService` as bare interface values and relies on runtime type switching to assign them. A mistake in ordering or type would compile and silently fail.

**Fix:**

Since this is a private function called only from `newAgentAPIHandler` (line 529–531) and the test-facing public constructors, expand its signature to include explicit named parameters:

```go
func newAgentAPIHandlerWithDependencies(
    loader          queuedInboundRequestProcessorLoader,
    submissionSvc   inboundRequestSubmitter,
    reviewSvc       operatorReviewReader,
    approvalSvc     approvalDecisionService,
    proposalApprSvc proposalApprovalService,
    accountingAdmin accountingAdminService,
    authSvc         browserSessionService,
    partiesAdmin    partiesAdminService,    // was in ...any
    inventoryAdmin  inventoryAdminService,  // was in ...any
) http.Handler {
```

Update the one production call site (`newAgentAPIHandler` line 529–531) and the test-facing public constructors. The `accessAdminService` extraction from `authService` via type assertion (lines 561–565) is a deliberate interface composition and should remain.

---

### P2-5 · Duplicate constant `adminPartyContactsPath == adminPartiesPath`

**Affected file:** `internal/app/api.go` — lines 113–114

**Description:**

```go
adminPartiesPath       = "/api/admin/parties"   // line 113
adminPartyContactsPath = "/api/admin/parties"   // line 114 — exact same value
```

These two constants hold the same string value. In `handleAdminPartyDetail` (line 621), the POST branch uses `adminPartyContactsPath` in `parseChildActionPath`, which is functionally identical to using `adminPartiesPath`. The alias creates the false impression that contacts are served from a different URL prefix than parties.

**Context:** The naming makes sense architecturally — contact creation is logically `POST /api/admin/parties/{id}/contacts` and status updates are `POST /api/admin/parties/{id}/status` — both resolved by `parseChildActionPath(adminPartiesPath, ...)`. The constant alias was likely added for readability intent but ends up being misleading rather than clarifying.

**Fix:**

Remove `adminPartyContactsPath` and use `adminPartiesPath` directly. Update the single call site in `handleAdminPartyDetail`:

```go
// api_admin_handlers.go line 621 — was adminPartyContactsPath
partyID, action, ok := parseChildActionPath(adminPartiesPath, r.URL.Path)
```

Then delete the constant from `api.go` line 114.

---

### P2-6 · Read-only list operations open write transactions with `FOR UPDATE`

**Affected files:**
- `internal/accounting/service.go` — `ListLedgerAccounts` (line 798), `ListTaxCodes` (line 821), `ListAccountingPeriods` (line 844), `ListJournalEntries` (line 867), `ListControlAccountBalances` (line 890), and others
- `internal/reporting/service.go` — `beginAuthorizedRead` helper (line 3920) used by all 20+ reporting methods

**Description:**

All read-only operations route their authorization through `identityaccess.AuthorizeTx`, which acquires a `FOR UPDATE` lock on the session and membership rows:

```sql
-- inside AuthorizeTx (used by accounting and reporting read paths)
SELECT m.role_code
FROM identityaccess.sessions s
JOIN identityaccess.memberships m ON m.id = s.membership_id
WHERE s.id = $1 ...
FOR UPDATE;   -- ← write lock on every read-path call
```

This is correct for mutation operations (prevents concurrent state from changing under the operation). For read-only list operations it is an unnecessary write lock.

The `reporting.Service` has a `beginAuthorizedRead` helper (line 3920) that centralizes this for all 20+ reporting queries — which is an excellent hook for the fix:

```go
// reporting/service.go line 3920 — current
func (s *Service) beginAuthorizedRead(ctx context.Context, actor identityaccess.Actor) (*sql.Tx, error) {
    tx, err := s.db.BeginTx(ctx, nil) // nil options = read-write tx
    if err != nil {
        return nil, fmt.Errorf("begin reporting read: %w", err)
    }
    if err := identityaccess.AuthorizeTx(ctx, tx, actor, ...roles); err != nil {
        _ = tx.Rollback()
        return nil, err
    }
    return tx, nil
}
```

**Fix:**

1. Add `AuthorizeReadOnlyTx` to `internal/identityaccess/auth.go` — same logic as `AuthorizeTx` but without `FOR UPDATE`:

```go
// AuthorizeReadOnlyTx validates the actor's session and role without a write lock.
// Use for read-only operations only.
func AuthorizeReadOnlyTx(ctx context.Context, tx *sql.Tx, actor Actor, allowedRoles ...string) error {
    const query = `
SELECT m.role_code
FROM identityaccess.sessions s
JOIN identityaccess.memberships m ON m.id = s.membership_id
WHERE s.id = $1
  AND s.org_id = $2
  AND s.user_id = $3
  AND s.status = 'active'
  AND s.expires_at > NOW()
  AND m.status = 'active'`
    // ... same scan and role-check logic, no FOR UPDATE
}
```

2. In `reporting/service.go`, update `beginAuthorizedRead` to use `ReadOnly: true` and `AuthorizeReadOnlyTx`:

```go
func (s *Service) beginAuthorizedRead(ctx context.Context, actor identityaccess.Actor) (*sql.Tx, error) {
    tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
    if err != nil {
        return nil, fmt.Errorf("begin reporting read: %w", err)
    }
    if err := identityaccess.AuthorizeReadOnlyTx(ctx, tx, actor,
        identityaccess.RoleAdmin, identityaccess.RoleOperator, identityaccess.RoleApprover); err != nil {
        _ = tx.Rollback()
        return nil, err
    }
    return tx, nil
}
```

3. For `accounting.Service` read-only list methods, similarly switch to `AuthorizeReadOnlyTx`. Each of `ListLedgerAccounts`, `ListTaxCodes`, `ListAccountingPeriods` can use `BeginTx(ctx, &sql.TxOptions{ReadOnly: true})`.

---

## P3 — Design Improvements

### P3-1 · No `http.MaxBytesReader` on attachment upload endpoints

**Affected file:** `internal/app/api_inbound_handlers.go`

**Description:**

`handleSubmitInboundRequest` and the `"draft"` action in `handleInboundRequestAction` read base64-encoded attachment content from the request body without a size cap. A client sending an arbitrarily large body can exhaust server memory.

**Fix:**

Wrap the body with `http.MaxBytesReader` before reading. Define a server-level constant for the maximum body size:

```go
// in api_constants.go (after P2-1 refactor)
const maxAttachmentBodyBytes = 25 << 20 // 25 MB

// in each handler that accepts attachments, before decodeJSONBody:
r.Body = http.MaxBytesReader(w, r.Body, maxAttachmentBodyBytes)
defer r.Body.Close()
```

Handle the resulting `http.MaxBytesError` in `writeJSONBodyError`:

```go
import "errors"

func writeJSONBodyError(w http.ResponseWriter, err error) {
    var maxBytesErr *http.MaxBytesError
    switch {
    case errors.As(err, &maxBytesErr):
        writeJSON(w, http.StatusRequestEntityTooLarge, errorResponse{Error: "request body too large"})
    case errors.Is(err, errRequestBodyRequired):
        writeJSON(w, http.StatusBadRequest, errorResponse{Error: "request body is required"})
    default:
        writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON request body"})
    }
}
```

---

### P3-2 · `fs.Sub` called on every static file request

**Affected file:** `internal/app/web_static.go` — line 21

**Description:**

```go
func (h *AgentAPIHandler) handleSvelteApp(w http.ResponseWriter, r *http.Request) {
    distFS, err := fs.Sub(webDistFS, "web_dist") // called on every request
```

`fs.Sub` on an `embed.FS` is lightweight (returns a wrapper, not a copy), but it is needlessly called for every static request. It should be computed once.

**Fix:**

Add `webDistFS fs.FS` to `AgentAPIHandler` and compute it at construction:

```go
// In AgentAPIHandler struct
type AgentAPIHandler struct {
    // ... existing fields ...
    webFS fs.FS
}

// In newAgentAPIHandlerWithDependencies, before building handler:
distFS, err := fs.Sub(webDistFS, "web_dist")
if err != nil {
    // embed.FS.Sub cannot fail for a known embedded path; panic is appropriate here
    panic(fmt.Sprintf("web bundle not embedded correctly: %v", err))
}

handler := &AgentAPIHandler{
    // ... existing fields ...
    webFS: distFS,
}

// In handleSvelteApp:
func (h *AgentAPIHandler) handleSvelteApp(w http.ResponseWriter, r *http.Request) {
    if h.webFS == nil {
        http.Error(w, "web bundle unavailable", http.StatusServiceUnavailable)
        return
    }
    // use h.webFS instead of computing distFS
```

---

### P3-3 · `looksLikeStaticAsset` uses `.` in basename — fragile for dotted path segments

**Affected file:** `internal/app/web_static.go` — lines 49–52

**Description:**

```go
func looksLikeStaticAsset(name string) bool {
    base := path.Base(strings.TrimSpace(name))
    return strings.Contains(base, ".")
}
```

Any URL path segment containing a dot is treated as a static asset request. A route like `/app/review/approvals/REQ-2026.001` or a future UUID collision with dots would return 404 instead of serving the SPA shell.

**Fix:**

Match on known static file extensions explicitly:

```go
var knownStaticExtensions = map[string]bool{
    ".js":    true,
    ".css":   true,
    ".map":   true,
    ".html":  true,
    ".ico":   true,
    ".png":   true,
    ".webp":  true,
    ".svg":   true,
    ".woff":  true,
    ".woff2": true,
    ".ttf":   true,
    ".json":  true, // e.g. manifest.json
    ".txt":   true, // e.g. robots.txt
}

func looksLikeStaticAsset(name string) bool {
    ext := strings.ToLower(path.Ext(path.Base(strings.TrimSpace(name))))
    return ext != "" && knownStaticExtensions[ext]
}
```

---

### P3-4 · Bearer token parsed twice in `handleSessionLogout`

**Affected file:** `internal/app/api_session_handlers.go` — lines 193–195

**Description:**

```go
case bearerTokenFromRequest(r) != "":
    if err := h.authService.RevokeAccessTokenSession(r.Context(), bearerTokenFromRequest(r)); err != nil {
```

`bearerTokenFromRequest(r)` is called twice — once in the `case` expression and once as an argument to `RevokeAccessTokenSession`. While not expensive, it is inconsistent style.

**Fix:**

```go
if token := bearerTokenFromRequest(r); token != "" {
    if err := h.authService.RevokeAccessTokenSession(r.Context(), token); err != nil {
        // ...
    }
} else {
    // cookie path
}
```

Restructure the `switch` as a simple `if/else`:

```go
token := bearerTokenFromRequest(r)
if token != "" {
    if err := h.authService.RevokeAccessTokenSession(r.Context(), token); err != nil {
        if errors.Is(err, identityaccess.ErrUnauthorized) || errors.Is(err, identityaccess.ErrSessionNotActive) {
            writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
            return
        }
        writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to revoke session"})
        return
    }
} else {
    sessionID, refreshToken, ok := sessionCookiesFromRequest(r)
    if !ok {
        writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
        return
    }
    if err := h.authService.RevokeAuthenticatedSession(r.Context(), sessionID, refreshToken); err != nil {
        if errors.Is(err, identityaccess.ErrUnauthorized) || errors.Is(err, identityaccess.ErrSessionNotActive) {
            writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
            return
        }
        writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to revoke session"})
        return
    }
    clearSessionCookies(w, sessionCookiesShouldBeSecure(r))
}

writeJSON(w, http.StatusOK, struct {
    Revoked bool `json:"revoked"`
}{Revoked: true})
```

---

### P3-5 · Trailing-data rejection in `decodeJSONBody` is non-standard

**Affected file:** `internal/app/api.go` — lines 770–778

**Description:**

After successfully decoding the primary JSON value, `decodeJSONBody` attempts a second decode and rejects the request if any trailing data (even whitespace) is found:

```go
var trailing any
if err := decoder.Decode(&trailing); err != nil {
    if errors.Is(err, io.EOF) {
        return nil  // clean end: OK
    }
    return err      // trailing non-EOF: reject
}
return errors.New("invalid JSON request body")
```

Standard HTTP JSON APIs do not reject trailing data. Some clients (proxies, middleware) may add a trailing newline or carriage return. This can cause valid requests to fail with a generic `400` with no actionable error message.

**Fix:**

Remove the trailing-data check entirely. The first `decoder.Decode(dst)` is sufficient:

```go
func decodeJSONBody(r *http.Request, dst any, allowEmpty bool) error {
    if r == nil || r.Body == nil {
        if allowEmpty {
            return nil
        }
        return errRequestBodyRequired
    }

    decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields()
    if err := decoder.Decode(dst); err != nil {
        if allowEmpty && errors.Is(err, io.EOF) {
            return nil
        }
        if errors.Is(err, io.EOF) {
            return errRequestBodyRequired
        }
        return err
    }
    return nil
}
```

---

### P3-6 · No `AttachmentStore` abstraction — object-storage migration will be invasive

**Affected files:** `internal/attachments/` package

**Description:**

Attachment content is stored as raw bytes in a PostgreSQL column. The technical guide (`13_attachments_and_derived_text.md`) documents the intent to preserve a clean migration path to external object storage. However, there is no interface or abstraction at the storage boundary — the attachment service writes and reads bytes directly against the database.

When the migration to S3 or equivalent object storage is required, it will require touching the attachment service internals, every caller, and every test that seeds attachment content.

**Fix:**

Introduce a minimal `AttachmentStore` interface inside the `attachments` package:

```go
// internal/attachments/store.go

package attachments

import "context"

// AttachmentStore abstracts the storage backend for attachment binary content.
// The initial implementation uses PostgreSQL. Future implementations may use
// an external object store such as S3 or GCS.
type AttachmentStore interface {
    // Put stores attachment binary content under the given attachment ID.
    // It is idempotent — re-storing the same ID replaces the content.
    Put(ctx context.Context, attachmentID string, content []byte) error

    // Get retrieves the binary content for an attachment.
    // Returns ErrAttachmentNotFound if no content exists for the ID.
    Get(ctx context.Context, attachmentID string) ([]byte, error)
}

// PostgresAttachmentStore implements AttachmentStore using a PostgreSQL column.
type PostgresAttachmentStore struct {
    db *sql.DB
}

func NewPostgresAttachmentStore(db *sql.DB) *PostgresAttachmentStore {
    return &PostgresAttachmentStore{db: db}
}

func (s *PostgresAttachmentStore) Put(ctx context.Context, attachmentID string, content []byte) error {
    // existing UPDATE logic
}

func (s *PostgresAttachmentStore) Get(ctx context.Context, attachmentID string) ([]byte, error) {
    // existing SELECT logic
}
```

Inject `AttachmentStore` into the attachment service rather than embedding `*sql.DB`-level queries. Future migration is then a single swap of the injected implementation.

---

### P3-7 · `last_seen_at` update on every authenticated request creates a write hotspot

**Affected file:** `internal/identityaccess/auth.go` — `authenticateSessionTx` lines 689–698

**Description:**

Every authenticated request that uses a browser cookie calls `AuthenticateSession`, which executes:

1. `BEGIN`
2. `FOR UPDATE` lock on the session row
3. Refresh token hash comparison
4. `UPDATE sessions SET last_seen_at = NOW() WHERE id = $1`
5. `COMMIT`

At scale, this creates a serialization bottleneck: concurrent requests from the same session queue behind each other to acquire the `FOR UPDATE` lock. The `last_seen_at` value changes on every request, even for millisecond-apart requests.

**Fix — rate-limit the update:**

Conditionally update `last_seen_at` only if it has not been updated recently (e.g., within the last 5 minutes). Add a condition to the UPDATE statement:

```sql
UPDATE identityaccess.sessions
SET last_seen_at = NOW()
WHERE id = $1
  AND (last_seen_at IS NULL OR last_seen_at < NOW() - INTERVAL '5 minutes')
RETURNING last_seen_at;
```

If the RETURNING clause returns no rows (update skipped because last_seen_at is recent), use the existing `session.LastSeenAt` value. This reduces write frequency by ~100× for active sessions without losing session activity data.

Alternatively, consider updating `last_seen_at` only on the `GET /api/session` endpoint (where the user is explicitly checking their session) rather than on every API call.

---

## P4 — Frontend Improvements

### P4-1 · `Content-Type` header detection in `apiRequest` is implicit

**Affected file:** `web/src/lib/api/client.ts` — lines 36–37

**Description:**

```ts
...(init?.body ? { 'Content-Type': 'application/json' } : {}),
```

If a caller passes a non-JSON body (e.g., a `FormData` object or a `Blob`), this check will incorrectly set `Content-Type: application/json`. The browser would normally set the correct multipart boundary for `FormData` automatically, but this explicit override breaks that.

**Fix:**

Only set `Content-Type: application/json` when the body is a string (the standard pattern for JSON payloads):

```ts
export async function apiRequest<T>(
    input: RequestInfo | URL,
    init?: RequestInit,
    fetcher: typeof fetch = fetch
): Promise<T> {
    const isJsonBody = typeof init?.body === 'string';
    const response = await fetcher(input, {
        credentials: 'same-origin',
        headers: {
            Accept: 'application/json',
            ...(isJsonBody ? { 'Content-Type': 'application/json' } : {}),
            ...(init?.headers ?? {})
        },
        ...init
    });
    return parseResponse<T>(response);
}
```

Callers that pass JSON should already be serializing via `JSON.stringify(payload)` which produces a string. This change enforces that assumption explicitly.

---

### P4-2 · No 401-triggered session refresh or redirect logic in the API client

**Affected file:** `web/src/lib/api/client.ts`

**Description:**

When the backend returns a 401, `apiRequest` throws `APIClientError` with `status: 401`. There is no interceptor or retry logic to transparently redirect to `/app/login`. Each individual page or component that calls an API function must handle the 401 itself.

For browser-session-based auth (current model with 24-hour sessions), this is acceptable in user testing. For token-based sessions with short-lived access tokens (`/api/session/refresh`), this becomes a correctness issue — an expired token causes every API call to fail without a recovery path.

**Fix (phased):**

Phase 1 — Global 401 handler in the Svelte layout:

In `web/src/routes/(app)/+layout.ts` (or `+layout.svelte`), catch 401 errors from data loading and redirect to login:

```ts
// +layout.ts
export async function load({ fetch, url }) {
    try {
        const session = await getSessionContext(fetch);
        return { session };
    } catch (err) {
        if (err instanceof APIClientError && err.status === 401) {
            throw redirect(302, `/app/login?next=${encodeURIComponent(url.pathname)}`);
        }
        throw err;
    }
}
```

Phase 2 — Token refresh interceptor (when token sessions are used):

Extend `apiRequest` to attempt one token refresh on 401 before propagating the error:

```ts
export async function apiRequest<T>(
    input: RequestInfo | URL,
    init?: RequestInit,
    fetcher: typeof fetch = fetch
): Promise<T> {
    const response = await fetcher(input, buildInit(init));
    if (response.status === 401) {
        const refreshed = await tryRefreshSession(fetcher);
        if (refreshed) {
            const retryResponse = await fetcher(input, buildInit(init));
            return parseResponse<T>(retryResponse);
        }
    }
    return parseResponse<T>(response);
}
```

---

### P4-3 · `types.ts` should be split by domain as it grows

**Affected file:** `web/src/lib/api/types.ts` (752 lines)

**Description:**

All API response and request types are defined in a single file. At 752 lines this is already large, and it will grow with every new API surface. Scrolling to find a specific type is slow, and merge conflicts on this file will become common.

**Fix — Recommended split:**

```
web/src/lib/api/types/
    index.ts          — re-exports from all domain files
    session.ts        — SessionContext, SessionLoginRequest, TokenSession
    intake.ts         — SubmitInboundRequestPayload, InboundRequestReview, InboundRequestDetail, SaveInboundDraftPayload
    proposals.ts      — ProcessedProposalReview, ProcessedProposalStatusSummary
    approvals.ts      — ApprovalQueueEntry, DecideApprovalRequest
    documents.ts      — DocumentReview
    accounting.ts     — LedgerAccount, TaxCode, AccountingPeriod, JournalEntryReview, ControlAccountBalance, TaxSummary, TrialBalanceReport, BalanceSheetReport, IncomeStatementReport
    inventory.ts      — InventoryItem, InventoryLocation, InventoryStockItem, InventoryMovementReview, InventoryReconciliationItem
    parties.ts        — Party, Contact, PartyDetailResponse
    access.ts         — OrgUserMembership
    workorders.ts     — WorkOrderReview
    audit.ts          — AuditEvent
    navigation.ts     — DashboardSnapshot, OperationsSnapshot, ReviewLandingSnapshot, AgentChatSnapshot, RouteCatalogEntry
    common.ts         — APIError, HomeAction, OperationsFeedItem
```

The barrel `index.ts` re-exports everything so existing imports (`import type { ... } from '$lib/api/types'`) continue to work without change:

```ts
// web/src/lib/api/types/index.ts
export * from './common';
export * from './session';
export * from './intake';
// ... etc
```

---

### P4-4 · `InventoryLandingSnapshot` type is missing from `types.ts` and the API client

**Affected files:** `web/src/lib/api/types.ts`, `web/src/lib/api/client.ts` (or equivalent API module)

**Description:**

This issue is related to P1-4 (the missing backend handler). Once the backend handler `handleGetNavigationInventory` is implemented, the Svelte frontend needs:
1. A TypeScript type for the response.
2. An API client function to call the endpoint.
3. A SvelteKit data loader in the inventory hub page (`+page.ts` or `+layout.ts`).

The `reporting.InventoryLandingSnapshot` struct (confirmed at `internal/reporting/service.go` line 260) has these fields:
```go
type InventoryLandingSnapshot struct {
    Stock          []InventoryStockItem
    Movements      []InventoryMovementReview
    Reconciliation []InventoryReconciliationItem
    RecentLimit    int
}
```

These types (`InventoryStockItem`, `InventoryMovementReview`, `InventoryReconciliationItem`) are already mapped in `api_review_handlers.go` and their response shapes are defined in `api.go` and typed in `types.ts`.

**Fix:**

After implementing the backend handler (P1-4), add to `types.ts`:

```ts
export interface InventoryLandingSnapshot {
    stock: InventoryStockItem[];          // already typed in types.ts
    movements: InventoryMovementReview[]; // already typed in types.ts
    reconciliation: InventoryReconciliationItem[]; // already typed in types.ts
}
```

Add an API client function:
```ts
export async function getInventoryLanding(fetcher = fetch): Promise<InventoryLandingSnapshot> {
    return apiRequest<InventoryLandingSnapshot>('/api/navigation/inventory', undefined, fetcher);
}
```

Implement the inventory hub page data loader to call this function.

---

## P5 — Module and Configuration Issues

### P5-1 · Go module name is not a valid module path

**Affected file:** `go.mod` — line 1

**Description:**

```
module workflow_app
```

`workflow_app` is not a valid Go module path for a publishable module. Valid module paths must be URL-like (e.g., `github.com/org/workflow_app`). While this works for a private local build, it:
- Is unconventional and may cause issues with tooling
- Would require a broad rename if the module is ever imported externally or split
- Makes the import paths non-standard (`workflow_app/internal/...`)

**Fix:**

Choose a canonical module path and perform a search-and-replace across all Go files. If the repository is private:

```
module github.com/your-org/workflow_app
```

Steps:
1. Update `go.mod` line 1.
2. Run `grep -r '"workflow_app/' --include='*.go' -l` to find all import paths.
3. Replace `"workflow_app/` with `"github.com/your-org/workflow_app/` everywhere.
4. Run `go build ./... && go test ./...` to confirm no broken imports.

This is a one-time, low-risk change if done as a single committed operation.

---

### P5-2 · No CORS middleware — relies entirely on SameSite cookies

**Affected file:** `internal/app/api.go` — `newAgentAPIHandlerWithDependencies` route registration

**Description:**

The application has no explicit Cross-Origin Resource Sharing (CORS) middleware. The current protection relies on:
1. `SameSite=Lax` cookies (prevents CSRF for state-changing requests)
2. The Svelte SPA being served from the same origin as the API

This is correct for the current deployment model, but there is no defense-in-depth. If the application is ever placed behind a CDN or API gateway that injects CORS headers, or if the API is accessed from a different origin for any reason, the backend has no explicit assertion of allowed origins.

**Fix:**

Add a minimal CORS middleware that:
1. Explicitly rejects credentialed cross-origin requests from non-allowlisted origins with `403 Forbidden`
2. Sets appropriate `Vary: Origin` headers for caching correctness
3. Handles preflight `OPTIONS` requests

```go
// internal/app/cors.go

package app

import (
    "net/http"
    "strings"
)

// corsMiddleware wraps an http.Handler with CORS protection.
// allowedOrigins should include only the origins that are permitted to make
// credentialed requests (e.g., "https://app.yourdomain.com").
// Pass an empty slice to deny all cross-origin credentialed requests.
func corsMiddleware(allowedOrigins []string, next http.Handler) http.Handler {
    allowed := make(map[string]bool, len(allowedOrigins))
    for _, o := range allowedOrigins {
        allowed[strings.ToLower(strings.TrimSpace(o))] = true
    }

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := strings.TrimSpace(r.Header.Get("Origin"))
        w.Header().Add("Vary", "Origin")

        if origin == "" {
            // Same-origin request; no CORS headers needed.
            next.ServeHTTP(w, r)
            return
        }

        if !allowed[strings.ToLower(origin)] {
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }
            http.Error(w, "cross-origin requests are not permitted", http.StatusForbidden)
            return
        }

        w.Header().Set("Access-Control-Allow-Origin", origin)
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == http.MethodOptions {
            w.Header().Set("Access-Control-Max-Age", "3600")
            w.WriteHeader(http.StatusNoContent)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

Wire it into the mux in `newAgentAPIHandlerWithDependencies`:

```go
// Read allowed origins from environment variable (comma-separated)
// Leave empty for same-origin-only operation (no cross-origin allowed).
allowedOrigins := parseAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
return corsMiddleware(allowedOrigins, mux)
```

For the current same-origin SPA deployment, pass an empty allowed-origins list. The middleware then correctly passes same-origin requests through and rejects any cross-origin attempt.

---

### P5-3 · No structured error boundary or telemetry in the Svelte frontend

**Affected files:** `web/src/routes/+error.svelte`, `web/src/routes/(app)/+layout.svelte`

**Description:**

The root `+error.svelte` (637 bytes) provides a basic SvelteKit error boundary, but it does not:
- Emit structured client-side error telemetry
- Distinguish retriable network errors from backend errors
- Provide the user with actionable recovery paths (retry, go home, contact support)

For user testing this is acceptable. For production it means silent client-side failures with no observability.

**Fix (phased):**

Phase 1 — Improve `+error.svelte` UX:

```svelte
<!-- web/src/routes/+error.svelte -->
<script>
    import { page } from '$app/stores';
    import { goto } from '$app/navigation';

    const statusMessages: Record<number, string> = {
        401: 'You need to log in to access this page.',
        403: 'You do not have permission to view this page.',
        404: 'This page could not be found.',
        500: 'An unexpected error occurred on our end.',
    };

    $: statusText = statusMessages[$page.status] ?? 'Something went wrong.';
</script>

<div class="error-page">
    <h1>Error {$page.status}</h1>
    <p>{statusText}</p>
    {#if $page.status === 401}
        <a href="/app/login">Log in</a>
    {:else}
        <button onclick={() => goto('/app')}>Go to Home</button>
    {/if}
</div>
```

Phase 2 — Structured error logging (when observability tooling is chosen):

Add an error logger utility that captures unhandled client errors with context:

```ts
// web/src/lib/utils/error_logger.ts
export function logClientError(context: string, err: unknown): void {
    const message = err instanceof Error ? err.message : String(err);
    // Replace with Sentry, Datadog, or custom endpoint as required
    console.error(`[${context}]`, message, err);
}
```

Call `logClientError` in `parseResponse` for unexpected (non-4xx) failures:

```ts
} catch {
    logClientError('api_response_parse', `HTTP ${response.status} — ${input}`);
    throw new APIClientError(response.status, `Request failed with status ${response.status}`);
}
```

---

## Implementation Sequencing

### Immediate (before broad production use)

| # | Issue | Effort | Risk |
|---|---|---|---|
| **P1-1** | Auth errors as 400 across review/inbound handlers | Low | Low |
| **P1-2** | Wrong HTTP 404 (HTML) for method mismatch in 8 detail handlers | Low | Low |
| **P1-3** | Session GET leaks internal error string on non-auth failures | Low | Low |
| **P1-4** | Inventory landing API handler entirely unimplemented | Medium | Low |
| **P3-1** | No `MaxBytesReader` on attachment upload endpoints | Low | Low |

### Next refactor milestone (can be batched)

| # | Issue | Effort | Risk |
|---|---|---|---|
| P2-3 | Remove/clarify 6 factory constructors (2 exact duplicates) | Low | Low |
| P2-4 | Remove `...any` variadic injection | Low | Low |
| P2-5 | Remove duplicate `adminPartyContactsPath` constant | Very low | Very low |
| P3-4 | Bearer token parsed twice in logout handler | Very low | Very low |
| P3-5 | Remove non-standard trailing-data rejection in JSON decoder | Low | Low |
| P3-2 | Compute `fs.Sub` once at startup instead of per-request | Low | Low |
| P3-3 | Replace dot-presence check with extension allowlist | Low | Low |
| P4-1 | Fix implicit `Content-Type` detection in `apiRequest` | Low | Low |
| P4-4 | Add `InventoryLandingSnapshot` type + API client function | Low | Low |

### Planned refactor (larger scope, own slice)

| # | Issue | Effort | Risk |
|---|---|---|---|
| P2-1 | Split `api.go` God file (2317 lines) by concern | High | Medium |
| P2-2 | Split `accounting/service.go` God file (3358 lines) by concern | High | Medium |
| P2-6 | Add `AuthorizeReadOnlyTx` + update `beginAuthorizedRead` | Medium | Low |
| P3-6 | Introduce `AttachmentStore` interface | Medium | Low |
| P3-7 | Rate-limit `last_seen_at` writes on authenticated requests | Medium | Low |
| P4-2 | Global 401 handler + optional token refresh interceptor | Medium | Low |
| P4-3 | Split `types.ts` into domain-scoped files | Medium | Low |

### Configuration (one-time, anytime)

| # | Issue | Effort | Risk |
|---|---|---|---|
| P5-1 | Fix Go module name to proper VCS-prefixed path | Medium | Low (but broad) |
| P5-2 | Add explicit CORS middleware | Medium | Low |
| P5-3 | Improve error boundary UX + add structured error logging | Medium | Low |

---

## Verification Checklist

After each fix, run the following before marking it complete:

```bash
# Go build — no regressions
go build ./cmd/... ./internal/...

# Go vet
go vet ./cmd/... ./internal/...

# Focused package verification (example: if api.go or api_review_handlers.go touched)
set -a; source .env; set +a
go test -count=1 -race ./internal/app/...

# Full serialized suite with race detection (for P2 God-file splits)
set -a; source .env; set +a
go test -p 1 -count=1 -race ./cmd/... ./internal/...

# Svelte type check
npm --prefix web run check

# Svelte build (confirms no import errors or type gaps)
npm --prefix web run build

# gopls diagnostics on changed files (in editor or via CLI)
```

### P1 fix verification

After applying P1-1 through P1-4, manually verify:
1. `curl -X GET http://localhost:8080/api/review/inbound-requests` without cookies → confirm `401`, not `400`.
2. `curl -X POST http://localhost:8080/api/review/inbound-requests/{valid-uuid}` → confirm `405 Method Not Allowed`, not `404`.
3. `curl -X GET http://localhost:8080/api/session` with an expired or missing cookie → confirm `401`, not `400`.
4. `curl -X GET http://localhost:8080/api/navigation/inventory` with valid session → confirm the inventory landing response is returned.
5. Integration tests in `internal/app/` should confirm the 401 behavior for unauthenticated review requests.

### P1-4 (inventory landing) verification

After implementing the handler:
1. Check that `GET /api/navigation/inventory` is registered in the mux by hitting it with credentials.
2. Confirm the Svelte inventory hub page (`/app/inventory`) loads stock, movements, and reconciliation data.
