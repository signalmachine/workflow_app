# Inbound Request Lifecycle And Queue Processing

Date: 2026-04-10
Status: Active technical guide
Purpose: explain the durable intake model, the request lifecycle states, and how queued requests flow into AI processing.

## 1. The request model

Inbound requests are not documents.

They are persisted intake records with their own lifecycle, identifiers, and audit trail. The business document that may be produced later is a separate record with a separate lifecycle.

The request record is designed to handle both:

1. human-originated intake
2. system-originated intake

That is why the intake layer is persist-first and queue-oriented.

## 2. Lifecycle states

The intake service models request progress through states such as:

1. `draft`
2. `queued`
3. `processing`
4. `processed`
5. `acted_on`
6. `completed`
7. `failed`
8. `cancelled`

Draft requests may be edited before they are queued. Draft requests are not supposed to be processed by AI.

## 3. The core service flow

The `internal/app` submission service creates the user-facing workflow around the intake package.

```go
request, err := s.intakeService.CreateDraft(ctx, intake.CreateDraftInput{
	OriginType: input.OriginType,
	Channel:    input.Channel,
	Metadata:   input.Metadata,
	Actor:      input.Actor,
})

message, err := s.intakeService.AddMessage(ctx, intake.AddMessageInput{
	RequestID:   request.ID,
	MessageRole: input.MessageRole,
	TextContent: input.MessageText,
	Actor:       input.Actor,
})

if queued {
	request, err = s.intakeService.QueueRequest(ctx, intake.QueueRequestInput{
		RequestID: request.ID,
		Actor:     input.Actor,
	})
}
```

That sequence matters because:

1. the request exists before AI processing begins
2. the message exists before queueing
3. queueing is explicit rather than implied
4. every step is auditable

## 4. Request reference and continuity

Requests get a stable `REQ-...` reference for operator use. That reference is important because it gives humans a durable intake identifier that is easier to work with than an internal UUID.

The browser and reporting layers both use that reference for continuity. The request detail page can also resolve related execution or delegation records when the lookup is prefixed with `run:`, `step:`, or `delegation:`.

## 5. Draft save versus submit

The codebase supports both draft saves and immediate queueing.

The two shapes exist because operators sometimes need to park a request without triggering AI, then resume it later.

Important rule:

1. draft requests remain draft
2. queued requests become eligible for processing
3. only queued requests should be claimed by the AI coordinator

## 6. Queue claim and processing

Queued intake is processed through a claim-and-process pattern.

The AI coordinator claims the next queued request and transitions it into processing before provider execution begins.

```go
request, err := c.intakeService.ClaimNextQueued(ctx, intake.ClaimNextQueuedInput{
	Channel: strings.TrimSpace(input.Channel),
	Actor:   input.Actor,
})
```

This avoids multiple processors working the same request at once and keeps the workflow durable.

## 7. Failure handling

The intake flow is designed so that failures do not lose the original request record.

If submission fails after a draft is created, the draft is cleaned up or left in a recoverable state depending on the exact operation. If queue processing fails, the request is marked failed with a reason that can be surfaced through reporting and browser review.

That makes the request lifecycle reconstructible from the database rather than from transient process memory.

## 8. Attachments and derived text

Requests may include attachments. In thin-v1, the original attachment bytes live in PostgreSQL, and derived text such as transcription can also be linked back to the same intake record.

That means the request can carry:

1. raw message text
2. uploaded attachment bytes
3. transcription or derived text
4. provenance links from the derived data back to the original request message

## 9. Browser-facing behavior

The browser layer exposes the intake lifecycle as operator actions:

1. submit a new request
2. save a draft
3. queue a draft
4. cancel a queued request
5. amend a queued or cancelled request back to draft
6. hard-delete an unprocessed draft

The current promoted Svelte route shape is:

1. `/app/submit-inbound-request` for new request creation
2. `/app/review/inbound-requests` for the request list and filtering surface
3. `/app/inbound-requests/{request_reference_or_id}` for exact request detail and parked-request lifecycle actions
4. `/app/operations` for the browser action that processes the next queued request
5. `/api/inbound-requests/{request_id}/{action}` and `/api/agent/process-next-queued-inbound-request` for the backend-owned mutation seams

Those actions are deliberate and bounded. They exist to support a controlled operator workflow, not a generic record editor.

## 10. What to be careful about

When changing request lifecycle code, be careful with:

1. state transitions
2. timestamp updates
3. cancellation and failure reasons
4. audit writes
5. exact `REQ-...` continuity
6. queue claim behavior

These are the edges where a seemingly small edit can create a workflow regression.
