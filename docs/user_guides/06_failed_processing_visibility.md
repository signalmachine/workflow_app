# Failed Processing Visibility

Date: 2026-04-12
Status: Active
Purpose: explain how to review a failed provider or failed processing path and trace the failure back to the exact request.

## 1. Open the failure review surface

Use one of these entry points:

1. `/app/review/inbound-requests`
2. `/app/inbound-requests/{request_reference}`
3. `/app/inbound-requests/run:<agent-run-id>`
4. `/app/inbound-requests/step:<agent-step-id>`
5. `/app/inbound-requests/delegation:<delegation-id>`

Use the list page when you are looking for failed requests. Use the request detail page when you already know the exact request reference.

## 2. Review the failed request

Check that the request page shows:

1. the failed request state
2. the failure reason
3. the failed timestamp
4. the AI run and step history around the failure

Example:

If queue processing fails for `REQ-000124`, open `/app/inbound-requests/REQ-000124` and review the failed timestamp plus run and step entries. If the provider failed before a proposal was created, use the failure state to decide whether to retry, amend the request back to draft, or collect more evidence for a defect report.

## 3. Confirm troubleshooting continuity

The key check is that the failure is still tied to the exact `REQ-...` record and not just to a transient process log.

Verify:

1. the request reference matches the item you expected
2. the failure reason is visible on the request detail page
3. the browser page and API review path agree on the failure state
4. the visible run or step history is enough to explain the failure
5. if needed, the request detail page can be reached directly from an AI run, step, or delegation lookup

## 4. Troubleshooting

If the request does not appear failed:

1. confirm you opened the correct request reference
2. confirm the request actually reached processing
3. refresh the failure review page and check the status again

If the failure details are too sparse:

1. inspect the run and step entries for the provider error
2. compare the request detail page with the review list page
3. verify you are signed into the correct org
