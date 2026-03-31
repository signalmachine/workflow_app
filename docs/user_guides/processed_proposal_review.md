# Processed Proposal Review

Date: 2026-03-31
Status: Active
Purpose: explain how to review a processed proposal and confirm continuity back to the originating request.

## 1. Open the proposal review surface

Open `/app/review/proposals`.

Use this page to find processed proposals that need inspection.

You can also filter the list by request reference with `/app/review/proposals?request_reference=REQ-...`.

## 2. Review one proposal

Open a specific proposal:

`/app/review/proposals/{recommendation_id}`

Check that the page shows:

1. the originating request reference
2. the AI run and recommendation trail
3. the proposal payload
4. any linked downstream document context
5. the request-approval action when a submitted document is ready for approval routing

## 3. Confirm continuity

The review should let you confirm:

1. the proposal still traces back to the correct `REQ-...` request
2. the recommendation output matches the request you expected
3. the browser review surface shows the same key facts as the API review surface

## 4. Troubleshooting

If the proposal detail page does not load:

1. confirm the recommendation ID is correct
2. confirm the request was processed
3. confirm you are signed into the correct org

If the continuity looks wrong:

1. re-open the parent request detail page
2. compare the request, recommendation, and proposal records
3. check whether you are viewing an older or unrelated recommendation
