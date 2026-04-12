# Audit Lookup

Date: 2026-04-12
Status: Active
Purpose: explain how to use the audit lookup surface to trace workflow history across linked records.

## 1. Open the audit lookup surface

Open the audit lookup page from the browser navigation or from a linked workflow record:

1. `/app/review/audit`
2. `/app/review/audit/{event_id}`
3. `/app/review/audit?event_id={event_id}`
4. `/app/review/audit?entity_type={entity_type}&entity_id={entity_id}`

Use this surface when you need a trace-oriented view of workflow activity rather than a single business record.

## 2. Review the audit trail

Check that the audit page shows:

1. the event or record identity you expected
2. the action history
3. the linked request, proposal, approval, or downstream record
4. any state change details needed to explain the event

Example:

If a request unexpectedly moved from queued back to draft, open `/app/review/audit?entity_type=inbound_request&entity_id={request_id}` and compare the audit event with the request detail page. The audit event should explain who or what performed the state change.

## 3. Confirm continuity

The important checks are:

1. the audit entry matches the workflow event you are investigating
2. the browser page and API read agree on the same audit facts
3. the linked records still trace back to the same chain of actions

## 4. Troubleshooting

If the audit trail is not clear enough:

1. reopen the originating workflow record
2. compare the audit entry with the request, proposal, or approval page
3. verify you are looking at the right org and record identity
