# Document Review

Date: 2026-04-12
Status: Active
Purpose: explain how to review a document and confirm its continuity through the downstream workflow chain.

## 1. Open the document surface

Open the document review page from the browser navigation or from a linked workflow record:

1. `/app/review/documents`
2. `/app/review/documents/{document_id}`
3. `/app/review/documents?document_id={document_id}`

Use this surface when you need to inspect the current truth for a single document.

## 2. Review the document

Check that the document page shows:

1. the correct document identity
2. the current document state
3. any linked upstream request or proposal
4. any linked downstream approval or posting context

Example:

When a proposal creates a submitted invoice document, open `/app/review/documents/{document_id}` from the proposal or approval page. Confirm the document state, source request, proposal, and approval context before treating the document as ready for accounting review.

## 3. Confirm continuity

The important checks are:

1. the document still matches the originating request or approval chain
2. the browser page and API read agree on the same document facts
3. the linked workflow records still point to the same underlying document

## 4. Troubleshooting

If the document page is missing expected context:

1. reopen the originating request or approval record
2. confirm the document was actually created or linked
3. verify you are in the correct org session
