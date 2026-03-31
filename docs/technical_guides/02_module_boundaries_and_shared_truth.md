# Module Boundaries And Shared Truth

Date: 2026-03-31
Status: Active technical guide
Purpose: explain which package owns which truth, how cross-module ownership works, and why `workflow_app` avoids duplicate business models.

## 1. Core rule

Every meaningful business record should have one owner.

That means:

1. one package owns the write path
2. one package owns the invariants
3. other packages can read or link to the record
4. other packages must not silently re-own the same business truth

This is one of the main guardrails against CRM-style drift and duplicated module-local truth.

## 2. First-class modules

The thin-v1 first-class modules are the code packages under `internal/`:

1. `identityaccess`
2. `ai`
3. `documents`
4. `workflow`
5. `accounting`
6. `inventory_ops` in the planning vocabulary, implemented in code as `internal/inventoryops`
7. `workforce`
8. `work_orders` in the planning vocabulary, implemented in code as `internal/workorders`
9. `attachments`
10. `reporting`

Each package is narrow, but it is not shallow. The implementation goal is small breadth with strong ownership discipline.

## 3. What each module owns

### 3.1 `identityaccess`

Owns orgs, users, memberships, sessions, and password-backed authentication.

### 3.2 `documents`

Owns the central document record, document numbering, supported document families, and lifecycle participation.

### 3.3 `workflow`

Owns approval truth, approval queues, approval decisions, and non-posted review orchestration.

### 3.4 `accounting`

Owns ledger accounts, journal entries, posting invariants, reversals, and control-account truth.

### 3.5 `inventory_ops`

Owns item and location truth, inventory movements, source-destination semantics, and inventory usage classification.

### 3.6 `work_orders`

Owns work-order execution truth, execution status history, and work-order-specific linkage for labor and material usage.

### 3.7 `attachments`

Owns attachment bytes, media-type validation, and links from attachments into request messages or other supported records.

### 3.8 `reporting`

Owns operator-facing read models and derived review views.

### 3.9 `ai`

Owns AI runs, steps, artifacts, tool policy, recommendations, and delegation traces. It does not become a second approval system.

## 4. Support records versus primary modules

Some data exists to support thin-v1 workflows without becoming a product center:

1. parties
2. contacts
3. tax foundation records
4. notifications where needed for workflow visibility

These records are allowed only because they support a stronger workflow or a stronger ownership boundary.

## 5. How cross-module composition works

Cross-module flows are composed through explicit identifiers and handoff contracts.

Examples:

1. inbound request to AI run
2. AI recommendation to approval request
3. approval decision to document posting
4. inventory execution handoff to work-order consumption
5. work-order and labor truth into centralized accounting

The key pattern is that the next module receives an explicit input record. The previous module does not reach into the next module's tables and write its own version of truth.

## 6. Why shared identity matters

The codebase relies on canonical identifiers to keep joins reliable.

Example:

```go
type DocumentReview struct {
	DocumentID       string
	ApprovalID       sql.NullString
	RequestReference sql.NullString
	RecommendationID sql.NullString
	RunID            sql.NullString
}
```

That shape means the browser can move from a document to its originating request, recommendation, approval, or run without inventing a second lookup model.

## 7. What not to do

Do not:

1. add a duplicate module-local copy of a central business record
2. let the browser layer become a second write owner
3. let reporting become the place where business truth is mutated
4. add a primary CRM module just because some supporting records look relational
5. widen a support record into a de facto primary domain without updating the canonical planning docs first

## 8. Change policy

When a change touches module boundaries, ask three questions:

1. which package owns the truth
2. which package owns the write path
3. which package should only read or link

If those answers are unclear, the design likely needs tightening before implementation.
