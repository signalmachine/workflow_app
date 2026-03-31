# Document Lifecycle And Posting Boundaries

Date: 2026-03-31
Status: Active technical guide
Purpose: explain how central document truth works, how document states change, and where posting and approval boundaries begin and end.

## 1. Why documents matter

`workflow_app` uses a central document record as the bridge between review, approval, posting, and downstream execution.

The document is not just a label. It is the durable business object that downstream modules reference when they need shared lifecycle truth.

## 2. Document states

The document package currently models states such as:

1. `draft`
2. `submitted`
3. `approved`
4. `rejected`
5. `posted`
6. `reversed`
7. `voided`

Those states are controlled and not all transitions are legal.

## 3. Ownership

`internal/documents` owns:

1. document creation
2. lifecycle transitions
3. numbering fields
4. source-document linkage
5. approval outcome application
6. posting outcome application

The central rule is that documents remain the shared truth, while domain-specific payloads attach to them.

## 4. Draft to submitted

The normal document flow starts with a draft, then moves to submitted.

```go
doc, err := s.CreateDraft(ctx, documents.CreateDraftInput{
	TypeCode: input.TypeCode,
	Title:    input.Title,
	Actor:    input.Actor,
})

doc, err = s.Submit(ctx, documents.SubmitInput{
	DocumentID: doc.ID,
	Actor:      input.Actor,
})
```

The document service checks state before changing it. That prevents a random caller from "submitting" a document that is already approved, posted, or reversed.

## 5. Approval and posting boundaries

Document approval and document posting are different concerns.

1. approval answers "may this continue"
2. posting answers "has the financial truth been committed"

The workflow package applies approval outcomes. The accounting package applies posting outcomes. The document package stores the shared state transitions that tie those steps together.

That separation matters because it keeps accounting from silently becoming the approval system.

## 6. Source documents

Documents may reference a `SourceDocumentID`.

This is how downstream family records can inherit continuity from an originating document without duplicating document truth.

Examples include:

1. derivative business records
2. follow-on operational documents
3. reversal or voiding continuations where needed

## 7. What to keep stable

The document layer is sensitive to:

1. invalid state transitions
2. numbering consistency
3. approval outcome handling
4. posting outcome handling
5. source-document continuity

If those change, the downstream workflow graph can become inconsistent even if the rest of the app still compiles.

