# workflow_app Non-Browser Auth Evolution Plan

Date: 2026-03-27
Status: Approved plan with first auth slice implemented
Purpose: define the next authentication path for lightweight non-browser clients on the shared backend foundation without destabilizing the finished browser-session flow.

## 1. Current auth reality

The current backend already has one credible session foundation, but it is exposed through two uneven client paths:

1. browser users authenticate through `POST /api/session/login`, which resolves org slug plus user email into one active-org session and sets the `workflow_session_id` plus `workflow_refresh_token` cookies
2. browser reads and writes then authenticate through those cookies on the shared `/api/...` seam
3. direct API and integration-test callers can still use `X-Workflow-Org-ID`, `X-Workflow-User-ID`, and `X-Workflow-Session-ID` as a pre-production automation path
4. the durable truth for both paths is still `identityaccess.sessions`, with one session bound to one active org context at a time

That shape is good enough for the landed browser layer and focused developer automation, but it is not the right long-term auth contract for lightweight non-browser clients.

Main gaps for later non-browser use:

1. cookie auth is browser-friendly but awkward for native or embedded clients
2. the UUID actor-header path is not a credible long-term auth mechanism and should not be widened into a public client contract
3. later clients need explicit token lifecycle, session introspection, and revocation behavior on the same backend seam rather than browser-only assumptions

## 2. Planning decision

The next non-browser auth path should be additive, not replacement.

Active decision:

1. keep the current browser-session cookie flow as the active v1 browser auth path
2. keep `identityaccess.sessions` as the one canonical device-session truth
3. add a first-class non-browser token path on top of that same session foundation
4. treat the UUID actor-header path as pre-production automation compatibility, not as the future non-browser auth model
5. do not introduce a second auth subsystem, a mobile-only backend, or a separate tenant-context model

## 3. Target auth shape

The intended next-step auth model for non-browser clients is:

1. a non-browser client starts a device-scoped session on the same org-scoped identity foundation already used by the browser flow
2. the server returns an explicit JSON auth payload rather than relying on cookies
3. authenticated API calls use `Authorization: Bearer <access-token>` on the existing shared `/api/...` routes
4. refresh and revocation stay session-aware and device-scoped
5. browser cookies and bearer tokens both resolve to the same active session truth, authorization rules, and audit context

Recommended token stance:

1. use opaque random bearer tokens rather than JWT-first self-contained tokens
2. store token secrets hashed at rest, matching the existing refresh-token posture
3. keep access tokens short-lived and separately refreshable
4. keep refresh-token rotation explicit and revocation-capable
5. keep token validation anchored to durable session truth so org context, revocation, expiry, and last-seen updates remain database-driven

## 4. Recommended API evolution

The first additive non-browser auth implementation should stay narrow and should reuse the current session semantics.

Recommended contract shape:

1. keep `POST /api/session/login` as the browser-cookie login endpoint
2. add one non-browser session-start endpoint that returns JSON auth material for the same underlying session model
3. add one refresh endpoint that rotates refresh material and issues the next short-lived access token
4. let the existing session-introspection and logout paths accept the additive token auth path once implemented
5. keep all later review, submission, approval, and attachment endpoints on the same shared `/api/...` routes rather than creating client-specific duplicates

Recommended initial non-browser auth response shape:

1. `session_id`
2. active org metadata
3. active user metadata
4. role code
5. short-lived access token plus expiry
6. refresh token plus expiry

This keeps the browser and non-browser clients aligned on one shared session record while still giving non-browser callers an explicit token contract.

## 5. Guardrails

This auth-evolution slice should not become broader than necessary.

Do not:

1. replace the working browser-cookie flow during thin v1
2. build full mobile-product auth depth such as push-device management, biometric UX, or offline credential caches inside this slice
3. widen the UUID actor-header path into the primary supported client contract
4. create web-specific versus mobile-specific authorization logic
5. treat non-browser auth as permission to change tenant or role semantics

Required invariants:

1. one session still carries one active org context at a time
2. authorization still resolves against the active org membership on that session
3. browser and non-browser clients still hit the same domain services, approval boundaries, and reporting reads
4. revocation, expiry, and replacement stay reconstructible from database state

## 6. Recommended implementation order after this plan

The first bounded auth slice from this plan is now implemented on the shared backend seam.

Implemented first auth slice:

1. add the additive non-browser session-start, refresh, and token-authentication path on top of the existing `identityaccess.sessions` foundation
2. teach shared API auth middleware to accept either browser cookies or bearer access tokens without changing handler-level authorization rules
3. keep the UUID actor-header path available only as temporary automation compatibility for pre-production tooling and existing tests
4. add API integration coverage for token issue, token refresh, token-authenticated reads or writes, revoked-session rejection, and org-safe authorization

Recommended follow-up slice after that:

1. decide whether the UUID actor-header path should be narrowed further or removed from general shared-API usage once bearer-session coverage exists
2. only then consider broader API-versioning, pagination, or incremental-sync work if the next milestone explicitly promotes it

## 7. Completion result for Milestone 8

This document completes the fifth planned Milestone 8 slice.

Completion result:

1. the current browser-session model remains the active v1 auth path
2. the next non-browser auth step is now explicit instead of implied
3. later implementation should add bearer-session support on the same backend foundation rather than improvising a second auth model
