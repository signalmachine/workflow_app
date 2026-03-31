# Accounting Journal, Control Accounts, And Reversals

Date: 2026-03-31
Status: Active technical guide
Purpose: explain the accounting ownership model, how journal entries are posted, and how reversals and control accounts fit into the system.

## 1. What accounting owns

`internal/accounting` owns the financial truth:

1. ledger accounts
2. accounting periods
3. tax codes
4. journal entries
5. journal lines
6. centralized document posting
7. reversals
8. control-account balances

The package is intentionally strict. It should be harder to write accounting truth incorrectly than to write it correctly.

## 2. Ledger accounts and control types

Ledger accounts carry class and control metadata such as:

1. asset
2. liability
3. equity
4. revenue
5. expense

Control types identify special treatment for receivable, payable, GST, and TDS roles.

That matters because review and posting logic need to know when an account participates in a control boundary rather than ordinary direct posting.

## 3. Posting documents

Posting takes a central document and turns it into balanced journal truth.

```go
entry, lines, document, err := s.PostDocument(ctx, accounting.PostDocumentInput{
	DocumentID:   document.ID,
	Summary:      "posted from approved source document",
	CurrencyCode: "USD",
	TaxScopeCode: accounting.TaxScopeGST,
	Lines:        lines,
	Actor:        actor,
})
```

The key rule is that the journal entry is append-only and the document receives the posting outcome through the document package.

## 4. Reversals

Reversals are separate from ordinary posting.

```go
entry, lines, document, err := s.ReverseDocument(ctx, accounting.ReverseDocumentInput{
	DocumentID:  document.ID,
	Reason:      "posted in error",
	EffectiveOn: time.Now().UTC(),
	Actor:       actor,
})
```

That separation is important because a reversal is a controlled corrective action, not an in-place edit of a posted journal.

## 5. Balancing rules

Journal entries must be balanced.

That means the accounting layer must not accept an entry where debits and credits do not reconcile. The service layer enforces the invariant before the write becomes durable.

## 6. Control account reviews

Control-account balances are derived review data, not arbitrary mutable state.

Operators use them to inspect things like:

1. receivables
2. payables
3. tax control accounts

Those balances are read from the reporting layer, but the accounting package owns the underlying financial truth that makes them meaningful.

## 7. Why posting is centralized

Centralized posting avoids duplicate implementations across document families.

If every module posted its own financial truth directly, the application would drift into inconsistent accounting behavior. Central posting keeps the ledger rules in one place.

## 8. What to keep stable

Be careful with:

1. balance enforcement
2. reversal constraints
3. source-document linkage
4. posting fingerprints
5. accounting-period checks
6. control-account semantics

