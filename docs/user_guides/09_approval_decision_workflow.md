# Approval Decision Workflow

Date: 2026-04-12
Status: Active
Purpose: explain how to review a pending approval and make the final approve or reject decision.

## 1. Open the approval review surface

Use one of these entry points:

1. `/app/review/approvals`
2. `/app/review/approvals/{approval_id}`

Open the exact approval record when you need to make the decision.

## 2. Make the decision

Use the decision action from the approval detail page:

`/app/approvals/{approval_id}/decision`

This is a POST form action, not a standalone GET page. From the approval detail page, choose the approval outcome that matches the review.

Example:

An approver opens a pending approval for a vendor invoice generated from `REQ-000123`. The approver checks the proposal and document links, approves the request if the document is correct, then confirms the approval detail shows the decision and the document moved into the expected downstream state.

## 3. Verify downstream continuity

After the decision lands, confirm:

1. the approval status changed as expected
2. the linked document moved to the correct downstream state
3. the original proposal still traces to the approval
4. the originating request remains visible in the chain

## 4. Troubleshooting

If the decision page does not accept the action:

1. confirm the approval ID is correct
2. confirm the approval is still pending
3. reload the approval detail page and retry once

If the downstream state does not update:

1. reload the approval detail page
2. check the linked document page
3. verify the proposal and request records still match the approval chain
