# workflow_app Operator Communication And Intake Surfaces Plan

Date: 2026-03-30
Status: Partially implemented post-refresh implementation slice; dashboard-only home and dedicated inbound-request submission page are landed, operations feed and agent chat remain planned
Purpose: define the next bounded browser-surface product changes after the visual-refresh pass: a pure dashboard home page, a dedicated inbound-request submission page, a one-way operations feed, and a separate two-way agent chat surface.

## 1. Why this plan exists

The current promoted web layer has one important product-shape weakness:

1. the home page is carrying too many different jobs at once
2. inbound-request submission currently competes with dashboard responsibilities
3. there is no dedicated durable coordinator or system communication surface
4. there is no separate interactive coordinator-chat surface for non-workflow conversational help

Those concerns should be separated before later workflow validation and broader user-facing refinement continue.

## 2. Planning decision

Current decision:

1. keep `Home` primarily as a dashboard
2. move `Submit inbound request` to a dedicated page
3. add `Operations feed` as a durable one-way coordinator or system communication page
4. add `Agent chat` as a separate two-way coordinator communication page

This is intentionally different from:

1. keeping intake embedded on the home page
2. treating feed and chat as the same surface
3. turning the application into a generic chat-first product

## 3. Surface roles

### 3.1 Home

`Home` should be a dashboard first.

It should emphasize:

1. request status summaries
2. recent requests
3. pending approvals
4. important alerts
5. quick links into the dedicated pages for intake, feed, chat, and review

It should not remain the primary form-entry surface for new inbound requests.

### 3.2 Submit inbound request

`Submit inbound request` should become its own dedicated page.

Required behavior:

1. focused intake form with the current request-message and attachment behavior
2. after submit, clearly show success or failure state
3. on success, show the exact `REQ-...` reference
4. offer clear next-step links such as request detail, dashboard, and the operations feed when appropriate

### 3.3 Operations feed

`Operations feed` should be a durable one-way communication surface.

Purpose:

1. coordinator or system messages inform the operator about important events
2. request status changes and notable operational activity become visible in one place
3. the surface stays event-driven and durable rather than conversational

Key rule:

1. this is primarily one-way communication from the system or coordinator to the user
2. user response should normally happen through linked workflow pages or through the separate chat surface

### 3.4 Agent chat

`Agent chat` should be a separate two-way communication surface.

Purpose:

1. let users ask the coordinator for clarification, explanation, or guidance
2. keep conversational interaction distinct from durable operational status updates
3. avoid overloading the operations feed with freeform back-and-forth

Key rule:

1. `Operations feed` and `Agent chat` must remain distinct
2. feed is durable one-way status or event communication
3. chat is interactive two-way coordinator communication

## 4. Implementation order

Implementation status:

1. `Home` is now a pure dashboard
2. `Submit inbound request` now has a dedicated page with clear result messaging and exact `REQ-...` continuity
3. `Operations feed` remains the next planned durable one-way coordinator or system communication page
4. `Agent chat` remains the following planned separate two-way coordinator communication page

## 5. Architecture guardrails

Do not:

1. collapse request intake into a generic chat surface
2. make the operations feed the primary place to perform workflow actions
3. turn coordinator chat into a broad autonomy surface that bypasses approvals, review, or database truth
4. create a second truth owner for request status or coordinator communication

Required invariants:

1. request intake remains persist-first and queue-oriented
2. durable request references remain separate from coordinator messages
3. feed items and chat interactions should tie back to workflow entities where appropriate
4. the browser and any later clients should continue sharing the same backend foundation

## 6. Validation expectations

Required verification for these slices:

1. `gopls` diagnostics on edited Go files
2. `go build ./cmd/... ./internal/...`
3. `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`
4. bounded browser review of the dashboard, dedicated submission page, operations feed, and agent chat surfaces
5. workflow-reference updates in `docs/workflows/` when those surfaces materially change supported operator paths
