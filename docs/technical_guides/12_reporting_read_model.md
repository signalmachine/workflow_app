# Reporting Read Model

Date: 2026-03-31
Status: Active technical guide
Purpose: explain how the reporting package turns module-owned truth into operator-facing review views.

## 1. What reporting is

`internal/reporting` is a read-model package.

It does not own business truth. It assembles durable review views from the underlying module tables so humans and the browser can inspect the workflow graph safely.

The package supports the operator-facing questions:

1. what is waiting
2. what changed
3. what posted
4. what failed
5. what is linked to what

## 2. Why it exists

The reporting layer exists because the browser should not need to manually reconstruct joins across intake, AI, documents, workflow, accounting, inventory, and execution tables.

Instead, reporting provides stable inspection models such as:

1. approval queues
2. document reviews
3. journal entries
4. control account balances
5. tax summaries
6. inventory stock
7. inventory movements
8. inventory reconciliation
9. work orders
10. inbound request review
11. processed proposal review
12. audit lookup

## 3. Read model examples

Example shapes from the package include:

```go
type DocumentReview struct {
	DocumentID        string
	TypeCode          string
	Status            string
	ApprovalID        sql.NullString
	RequestReference  sql.NullString
	RecommendationID  sql.NullString
	RunID             sql.NullString
	JournalEntryID    sql.NullString
}
```

That is a review model, not a write model. It is designed to help an operator trace continuity quickly.

## 4. Filtering and exact lookup

Reporting supports filtered list reads and exact drill-downs.

Examples:

1. exact `REQ-...` request references
2. exact `document_id`
3. exact `approval_id`
4. exact `recommendation_id`
5. exact `entry_id`
6. exact `movement_id`
7. exact `event_id`

That exact lookup model is what lets the browser land on a single record instead of only a broad list page.

## 5. Review semantics

The reporting package is deliberately opinionated.

It validates review filters and rejects malformed inputs instead of leaking database casting errors. This keeps the browser and API surface cleaner and more predictable.

## 6. What reporting should not do

Do not treat reporting as:

1. a write layer
2. a business-rule owner
3. a place to add duplicate state transitions
4. a second approval system

If a change needs a new write behavior, it belongs in the owning module first.

## 7. What to keep stable

Be careful with:

1. filter semantics
2. entity-link continuity
3. exact lookup support
4. derived-count calculations
5. browser-facing review expectations

