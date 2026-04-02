# workflow_app Application Workflow Catalog

Date: 2026-04-02
Status: Active durable workflow catalog updated for the grouped landing pages at `/app/operations`, `/app/review`, and `/app/inventory`, the searchable route catalog at `/app/routes`, the utility surfaces at `/app/settings` plus access-scoped `/app/admin`, the admin accounting setup surface at `/app/admin/accounting`, and the role-aware operator home on `/app`
Purpose: capture the application workflows and related feature continuity in one durable reference document for implementation review, testing, onboarding, and later user-guide preparation.

## 1. How to read this document

This document is workflow-reference material, not the live planning tracker.

Catalog rule:

1. every meaningful application feature, surface, or control seam should map to one or more workflows in this catalog or in a later durable workflow-reference document
2. if something materially affects operator behavior or persisted business control but cannot be mapped to a workflow, that gap should be treated as a documentation or design issue rather than ignored

Scope rule:

1. this document is not limited to thin v1
2. it should represent the workflows supported by the application at a given point in time
3. as new workflows land in later phases, extend this catalog rather than creating version-fragmented replacements unless the document becomes too large to manage cleanly

Status meanings used here:

1. `implemented`: supported in the current codebase
2. `repo_verified`: covered by repository build or test verification
3. `live_validated`: exercised on the real `/app` plus `/api/...` seam and recorded in the active validation docs
4. `pending_live_validation`: implemented, but still awaiting explicit live workflow validation

For live planning and next steps, use:

1. `new_app_docs/new_app_tracker.md`
2. `workflow_validation_track.md`

Workflow-review policy:

1. this catalog should support bounded end-to-end review and testing, not only feature listing
2. when operator readiness or user-testing readiness is in question, use this catalog together with the workflow checklist instead of relying only on unit or package verification
3. each workflow here should be reviewable in terms of persistence continuity, control-boundary behavior, review-surface visibility, and upstream or downstream linkage

## 2. Current workflow catalog

### 2.1 Browser session login and active-session continuity

Purpose:
allow an operator to start a password-backed browser-scoped session, load `/app`, and continue through the shared backend seam that later non-browser and mobile clients will also reuse.

Entry points:

1. `POST /api/session/login`
2. `/app/login`
3. `GET /api/session`
4. `/app`

Current status:

1. implemented
2. repo_verified
3. exact request-continuity on process failure is repo_verified through both `/api/...` and `/app/...`
4. pending_live_validation

Primary continuity surfaces:

1. role-aware home
2. operations landing
3. review landing
4. inventory landing
5. dedicated request-submission page
6. route catalog
7. settings utility surface
8. access-scoped admin utility surface
9. admin accounting setup surface
10. session introspection
11. subsequent browser-authenticated `/api/...` writes and review reads

### 2.1.1 Admin accounting setup maintenance

Purpose:
allow an admin actor to create and browse foundational ledger accounts, tax codes, and accounting periods, and to close accounting periods, through one bounded maintenance seam that stays separate from posted-truth accounting review.

Entry points:

1. `GET /api/admin/accounting/ledger-accounts`
2. `POST /api/admin/accounting/ledger-accounts`
3. `GET /api/admin/accounting/tax-codes`
4. `POST /api/admin/accounting/tax-codes`
5. `GET /api/admin/accounting/periods`
6. `POST /api/admin/accounting/periods`
7. `POST /api/admin/accounting/periods/{period_id}/close`
8. `/app/admin/accounting`

Expected outputs:

1. bounded admin-only master-data creation on the shared accounting service seam
2. visible browser continuity between the admin maintenance hub and the accounting setup page
3. durable audit-visible setup writes for ledger accounts, tax codes, and accounting periods
4. bounded period-close control without widening posted-truth accounting review into generic editing

Current status:

1. implemented
2. repo_verified
3. pending_live_validation

### 2.2 Inbound request submit and queue processing

Purpose:
capture a request durably, process it through the AI seam, and surface the resulting proposal and execution trail.

Entry points:

1. `POST /api/inbound-requests`
2. `/app/submit-inbound-request`
3. `POST /app/inbound-requests`
4. `POST /api/agent/process-next-queued-inbound-request`
5. `/app/agent/process-next-queued-inbound-request`

Expected outputs:

1. persisted `REQ-...` request reference
2. inbound request lifecycle transitions
3. AI run, step, artifact, recommendation, and optional delegation records
4. request-detail and proposal continuity in `/api/review/...` and `/app/...`
5. provider-backed review can use bounded request-scoped detail and processed-proposal continuity tools in addition to secondary org-level status summary context

Current status:

1. implemented
2. repo_verified
3. dedicated browser submission-page continuity is repo_verified
4. live_validated

### 2.3 Draft save, amend, queue, cancel, and delete lifecycle

Purpose:
allow pre-processing request parking, revision, queueing, cancellation, amendment back to draft, and hard deletion of unprocessed drafts.

Entry points:

1. `/app/inbound-requests`
2. `/app/inbound-requests/{request_reference}`
3. `POST /api/inbound-requests`
4. `POST /api/inbound-requests/{request_id}/{action}`

Expected outputs:

1. stable request identity from draft onward
2. correct status transitions across `draft`, `queued`, `cancelled`, and return-to-draft paths
3. preserved request and attachment continuity until deletion

Current status:

1. implemented
2. repo_verified
3. exact draft save and edit -> queue -> process continuity is now repo_verified on 2026-03-29 through `internal/app` integration coverage across both `/api/...` and `/app/...`
4. pending_live_validation

### 2.4 Processed proposal review and continuity

Purpose:
surface the AI-produced proposal with upstream request continuity and downstream document or approval continuity.

Entry points:

1. `GET /api/review/processed-proposals`
2. `/app/review/proposals`
3. `/app/review/proposals/{recommendation_id}`

Expected outputs:

1. request reference and request status continuity
2. recommendation and run continuity
3. approval linkage when present
4. document linkage when present
5. payload-derived document and suggested-queue continuity even before approval exists

Current status:

1. implemented
2. repo_verified
3. exact processed-proposal detail continuity after draft-originated processing is now repo_verified on 2026-03-29 through `internal/app` integration coverage across both `/api/...` and `/app/...`
4. partially live_validated
5. pending_live_validation for the approval-producing live branch

### 2.5 Processed proposal to approval request

Purpose:
turn a processed proposal that identifies a submitted document into a workflow approval request without leaving the shared backend seam.

Entry points:

1. `POST /api/review/processed-proposals/{recommendation_id}/request-approval`
2. `/app/review/proposals/{recommendation_id}/request-approval`

Expected outputs:

1. workflow approval creation
2. atomic recommendation-to-approval linkage
3. preserved proposal continuity into approval review
4. queue-code continuity on both API and browser review surfaces

Current status:

1. implemented
2. repo_verified
3. full proposal -> request-approval continuity is now repo_verified on 2026-03-29 through `internal/app` integration coverage across both `/api/...` and `/app/...`
4. pending_live_validation

### 2.6 Approval decision and downstream continuity

Purpose:
allow an approver to approve or reject a pending approval and continue into downstream review.

Entry points:

1. `POST /api/approvals/{approval_id}/decision`
2. `/app/approvals/{approval_id}/decision`
3. `/app/review/approvals`
4. `/app/review/approvals/{approval_id}`

Expected outputs:

1. approval decision persistence
2. document state transition continuity
3. exact approval review continuity
4. proposal and upstream request continuity where provenance exists

Current status:

1. implemented
2. repo_verified
3. full proposal-requested approval -> approval decision -> exact approval and document continuity is now repo_verified on 2026-03-29 through `internal/app` integration coverage across both `/api/...` and `/app/...`
4. pending_live_validation for the full proposal-requested-approval path

### 2.7 Failed provider or failed processing visibility

Purpose:
make AI-processing failures reviewable and operator-actionable rather than silent.

Entry points:

1. `GET /api/review/inbound-requests`
2. `GET /api/review/inbound-requests/{request_reference_or_id}`
3. `/app/review/inbound-requests`
4. `/app/inbound-requests/{request_reference}`

Expected outputs:

1. failed request state
2. failure reason and failed timestamp
3. AI run and step failure visibility
4. operator troubleshooting continuity from exact request detail

Current status:

1. implemented
2. repo_verified
3. pending_live_validation

### 2.8 Downstream review surfaces

Purpose:
let operators inspect the downstream truth and provenance created or linked by the request -> AI -> approval chain.

Primary surfaces:

1. documents
2. approvals
3. processed proposals
4. accounting review
5. inventory review
6. work-order review
7. audit lookup

Current status:

1. implemented
2. repo_verified
3. partially live_validated

### 2.9 Operations feed and agent-chat continuity

Purpose:
keep durable one-way coordinator or system updates distinct from guidance-oriented coordinator requests while preserving one shared request-truth model.

Entry points:

1. `/app/operations-feed`
2. `/app/agent-chat`
3. `POST /app/inbound-requests` with the dedicated `agent_chat` browser path

Expected outputs:

1. operations feed continuity from recent request, proposal, and approval truth into exact workflow pages
2. dedicated coordinator-chat request submission that persists onto the shared inbound-request foundation rather than inventing a second conversation store
3. exact `REQ-...` continuity from queued chat request into request detail, proposal review, and downstream workflow surfaces

Current status:

1. implemented
2. repo_verified
3. pending_live_validation

## 3. Reference rule

When a workflow changes materially:

1. update the active planning docs in `new_app_docs/`
2. update this catalog when the durable workflow reference has drifted

When a new meaningful feature or support seam is introduced:

1. map it to one or more workflows
2. add or update the relevant workflow-reference documentation if that mapping is not already clear
