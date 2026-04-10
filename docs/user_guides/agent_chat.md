# Agent Chat

Date: 2026-04-10
Status: Active
Purpose: explain how to use the coordinator-facing browser chat surface on the persisted inbound-request foundation.

## 1. Open the chat surface

Open `/app/agent-chat`.

Use this page when you want to submit a coordinator-oriented request on the same inbound-request foundation used by the rest of the app.

## 2. Submit a chat request

From the chat surface:

1. enter the guidance or coordination request
2. submit it as an inbound request on the shared queue-oriented path
3. keep the request on the same `REQ-...` foundation as other intake flows

## 3. Confirm continuity

After submission, confirm the request chain by checking:

1. the request detail page
2. the proposal review surface if processing has happened
3. any downstream workflow link that the request produces

The important point is that the chat surface should not create a separate conversation store.

## 4. Troubleshooting

If the chat submission does not persist:

1. confirm the browser session is active
2. confirm the org is correct
3. retry from the same chat page and check the request detail page afterward

If the request does not appear in the normal request flow:

1. check `/app/review/inbound-requests`
2. check `/app/review/proposals`
3. confirm the chat request reached the shared inbound-request queue
