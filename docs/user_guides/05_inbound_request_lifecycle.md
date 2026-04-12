# Inbound Request Lifecycle

Date: 2026-04-12
Status: Active
Purpose: explain how to create, park, edit, queue, cancel, delete, and process inbound requests from the browser surfaces.

## 1. What this guide covers

This guide follows the active request workflow from the workflow catalog:

1. draft request creation
2. draft editing
3. queueing
4. cancellation
5. hard deletion of unprocessed drafts
6. queued processing and downstream continuity

## 2. Create a request

Open `/app/submit-inbound-request`.

From that page, you can:

1. enter the request text
2. attach supporting files when needed
3. save the request as a draft
4. submit the request into the processing queue

Use draft save when the request is not ready for processing yet. Use submit when the request is ready to be queued.

Example:

An operator receives a vendor email that says `Please approve the attached repair-material purchase for Harbor Retail job 42`. The operator opens `/app/submit-inbound-request`, writes a short request summary, attaches the vendor quote, and saves it as a draft if the job number or vendor contact still needs confirmation. When the details are ready, the operator queues the same request instead of creating a second request.

## 3. Continue editing a draft

Use `/app/review/inbound-requests` when you need to find a parked, queued, failed, or processed request from the browser list.

Open the exact request detail page for the stable `REQ-...` reference, for example:

`/app/inbound-requests/{request_reference_or_id}`

From the draft detail view, you should be able to:

1. continue editing the same draft
2. add more draft content and attachments from the same detail page
3. keep the request parked until it is ready

The request reference should stay stable while the request remains in draft, queued, or processed states.

Example:

If the saved draft is `REQ-000123`, use `/app/inbound-requests/REQ-000123` to add the missing vendor contact. The reference should remain `REQ-000123` after the edit and after queueing.

## 4. Queue, cancel, or delete

From the request detail page:

1. queue the draft when it is ready for AI processing
2. cancel the request if it should not be processed
3. amend a queued, cancelled, or failed pre-processing request back to draft when the UI offers that control
4. delete the request only if it is still an unprocessed draft and the UI allows hard deletion

If a request has already moved into processing or has been processed, do not expect the draft-only delete action to remain available.

Example:

If `REQ-000123` was queued by mistake before processing started, cancel it or amend it back to draft when the page offers that control. If it is still only an unprocessed draft and should be discarded entirely, use delete from the exact detail page.

## 5. Process the queue

Use the processing action when you want the next queued request handled by the coordinator:

1. browser surface: `/app/operations`, using the `Process next queued request` action
2. API path: `POST /api/agent/process-next-queued-inbound-request`

After processing, check the request record for:

1. the request lifecycle status
2. AI run and step records
3. artifact or brief output
4. recommendation output
5. any delegation or specialist follow-up records

Example:

After the operations page processes `REQ-000123`, open the request detail again and follow the proposal link. If the coordinator created a submitted document, continue from the proposal into document review or approval request rather than relying on the operations page alone.

## 6. Verify continuity

After a request is processed, use the request detail and review pages to verify continuity:

1. the exact `REQ-...` reference should still identify the request
2. the request should link into proposal review when a recommendation exists
3. the review surfaces should show the same upstream request facts
4. any failure should show a reason and failed timestamp

## 7. Troubleshooting

If a saved draft does not queue:

1. confirm the request is still a draft
2. confirm you are using the right org session
3. confirm the page shows the current saved state before retrying

If processing reports a failure:

1. check the failure reason on the request detail page
2. check the failed timestamp
3. inspect the run and step history for the provider or processing error
