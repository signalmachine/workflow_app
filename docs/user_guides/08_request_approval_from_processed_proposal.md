# Request Approval From a Processed Proposal

Date: 2026-04-12
Status: Active
Purpose: explain how to turn a processed proposal into an approval request while preserving the original workflow chain.

## 1. Start from the proposal

Open the processed proposal detail page:

`/app/review/proposals/{recommendation_id}`

Only request approval when the proposal identifies a submitted document and the workflow actually expects an approval step.

## 2. Create the approval request

Use the approval request action from the proposal detail surface:

1. browser path: `/app/review/proposals/{recommendation_id}/request-approval`
2. API path: `POST /api/review/processed-proposals/{recommendation_id}/request-approval`

The browser path is a POST form action, not a standalone GET page.

After the action completes, confirm that the proposal now links into an approval record.

Example:

A processed proposal created a submitted vendor invoice document and shows that approval is required before posting. Use the request-approval action on the exact proposal detail page, then open the linked approval and confirm it references the same request, recommendation, and document before asking the approver to decide.

## 3. Verify the linkage

Check that the new approval preserves:

1. the originating request reference
2. the recommendation reference
3. the downstream approval identifier
4. the expected queue-code or review continuity on both browser and API paths

## 4. Troubleshooting

If the approval request action is missing:

1. confirm the proposal has a submitted document context
2. confirm the proposal is in the processed state
3. confirm you are on the exact proposal detail page, not the list page

If the created approval does not appear linked:

1. reload the proposal detail page
2. inspect the approval review surface
3. compare the proposal and approval records for the expected reference
