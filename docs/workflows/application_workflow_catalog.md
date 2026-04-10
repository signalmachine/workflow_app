# workflow_app Application Workflow Catalog

Date: 2026-04-10
Status: Active durable workflow catalog updated for the current served Svelte runtime at `/app`, including the contextual-navigation shell, the grouped landing pages at `/app/operations`, `/app/review`, and `/app/inventory`, the searchable route catalog at `/app/routes`, the utility surfaces at `/app/settings` plus access-scoped `/app/admin`, the grouped admin directory routes at `/app/admin/master-data` and `/app/admin/lists`, the admin accounting, party, access-control, and inventory setup surfaces at `/app/admin/accounting`, `/app/admin/parties`, `/app/admin/access`, and `/app/admin/inventory`, the accounting report directory at `/app/review/accounting`, the dedicated accounting report destinations under `/app/review/accounting/journal-entries`, `/app/review/accounting/control-balances`, `/app/review/accounting/tax-summaries`, `/app/review/accounting/trial-balance`, `/app/review/accounting/balance-sheet`, and `/app/review/accounting/income-statement`, the role-aware operator home on `/app`, the `/app/review/inbound-requests` list route, and the exact `/app/inbound-requests/{request_reference_or_id}` detail route with parked-request lifecycle controls
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

1. `new_app_docs/new_app_tracker_v2.md`
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
9. admin master-data directory
10. admin lists directory
11. admin accounting setup surface
12. admin party setup surface
13. admin access-control surface
14. admin inventory setup surface
15. accounting report directory and dedicated journal, control-balance, tax-summary, trial-balance, balance-sheet, and income-statement destinations
16. session introspection
17. subsequent browser-authenticated `/api/...` writes and review reads

### 2.1.1 Admin accounting setup maintenance

Purpose:
allow an admin actor to create and browse foundational ledger accounts, tax codes, and accounting periods, to close accounting periods, and to mark ledger accounts or tax codes active or inactive through one bounded maintenance seam that stays separate from posted-truth accounting review.

Entry points:

1. `GET /api/admin/accounting/ledger-accounts`
2. `POST /api/admin/accounting/ledger-accounts`
3. `GET /api/admin/accounting/tax-codes`
4. `POST /api/admin/accounting/tax-codes`
5. `GET /api/admin/accounting/periods`
6. `POST /api/admin/accounting/periods`
7. `POST /api/admin/accounting/periods/{period_id}/close`
8. `/app/admin/master-data`
9. `/app/admin/lists`
10. `/app/admin/accounting`

Expected outputs:

1. bounded admin-only master-data creation on the shared accounting service seam
2. visible browser continuity between the admin maintenance hub, grouped admin directory pages, and the accounting setup page
3. durable audit-visible setup writes for ledger accounts, tax codes, and accounting periods
4. bounded active or inactive status governance for ledger accounts and tax codes without widening into generic edit-heavy CRUD
5. bounded period-close control without widening posted-truth accounting review into generic editing

Current status:

1. implemented
2. repo_verified
3. pending_live_validation

### 2.1.2 Admin party setup maintenance

Purpose:
allow an admin actor to create and browse bounded customer and vendor support records, to open exact party detail with current contact visibility, and to mark a party active or inactive through one maintenance seam that stays separate from workflow review pages and avoids CRM-first drift.

Entry points:

1. `GET /api/admin/parties`
2. `POST /api/admin/parties`
3. `GET /api/admin/parties/{party_id}`
4. `/app/admin/parties`
5. `/app/admin/parties/{party_id}`

Expected outputs:

1. bounded admin-only customer and vendor support-record creation on the shared `parties` service seam
2. visible browser continuity between the admin maintenance hub and the party setup page
3. exact party-detail continuity with current contact visibility before downstream document or accounting work depends on the record
4. bounded active or inactive party governance on the same shared truth model
5. shared API reuse for later non-browser admin maintenance without introducing browser-local truth

Current status:

1. implemented
2. repo_verified
3. pending_live_validation

### 2.1.3 Admin access maintenance

Purpose:
allow an admin actor to provision org-scoped user memberships, attach an existing user to the current org, and update membership roles through one bounded maintenance seam that stays on the shared `identityaccess` truth model.

Entry points:

1. `GET /api/admin/access/users`
2. `POST /api/admin/access/users`
3. `POST /api/admin/access/users/{membership_id}/role`
4. `/app/admin/access`

Expected outputs:

1. bounded admin-only org-user listing on the shared identity and session service seam
2. controlled creation or attachment of org memberships without widening into a broad identity-product console
3. bounded membership-role updates with durable audit visibility
4. protection against the currently signed-in admin accidentally removing their own admin access during this first role-update slice

Current status:

1. implemented
2. repo_verified
3. pending_live_validation

### 2.1.4 Admin inventory setup maintenance

Purpose:
allow an admin actor to create and browse bounded inventory items and inventory locations, and to mark those master records active or inactive, through one maintenance seam that stays on the shared `inventory_ops` truth model and continues directly into the promoted inventory review routes.

Entry points:

1. `GET /api/admin/inventory/items`
2. `POST /api/admin/inventory/items`
3. `GET /api/admin/inventory/locations`
4. `POST /api/admin/inventory/locations`
5. `/app/admin/inventory`

Expected outputs:

1. bounded admin-only inventory item creation on the shared `inventory_ops` service seam
2. bounded admin-only inventory location creation on the same shared inventory foundation
3. bounded active or inactive status governance for inventory items and locations on that same shared truth model
4. visible browser continuity between the admin maintenance hub, the inventory setup page, and the existing exact inventory review routes
5. shared API reuse for later non-browser inventory maintenance without introducing browser-local truth

Current status:

1. implemented
2. repo_verified
3. pending_live_validation

### 2.1.5 Accounting report review

Purpose:
allow operators to inspect posted accounting truth, control balances, tax summaries, and baseline financial statements from the shared reporting seam without moving report composition into the browser runtime.

Entry points:

1. `GET /api/review/accounting/journal-entries`
2. `GET /api/review/accounting/control-account-balances`
3. `GET /api/review/accounting/tax-summaries`
4. `GET /api/review/accounting/trial-balance`
5. `GET /api/review/accounting/balance-sheet`
6. `GET /api/review/accounting/income-statement`
7. `/app/review/accounting`
8. `/app/review/accounting/journal-entries`
9. `/app/review/accounting/control-balances`
10. `/app/review/accounting/tax-summaries`
11. `/app/review/accounting/trial-balance`
12. `/app/review/accounting/balance-sheet`
13. `/app/review/accounting/income-statement`

Expected outputs:

1. exact journal-entry review with upstream document, approval, request, recommendation, and run continuity when that provenance exists
2. control-account balance review with as-of filtering
3. GST and TDS tax-summary review with effective-date range filters
4. trial balance with debit and credit balance totals plus an explicit imbalance total
5. balance sheet with assets, liabilities, equity, current earnings, and an explicit imbalance total
6. income statement with revenue, expense, and net-income totals over a selected effective-date range

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
3. `/app/operations`, using the process-next queued request action
4. `POST /api/agent/process-next-queued-inbound-request`

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

1. `/app/submit-inbound-request`
2. `/app/review/inbound-requests`
3. `/app/inbound-requests/{request_reference_or_id}`
4. `POST /api/inbound-requests`
5. `POST /api/inbound-requests/{request_id}/{action}`

Expected outputs:

1. stable request identity from draft onward
2. correct status transitions across `draft`, `queued`, `cancelled`, and return-to-draft paths
3. preserved request and attachment continuity until deletion

Current status:

1. implemented
2. repo_verified
3. exact draft save and edit -> queue -> process continuity is now repo_verified on 2026-03-29 through `internal/app` integration coverage across both `/api/...` and `/app/...`
4. exact parked-request lifecycle controls on the promoted Svelte request-detail route are repo_verified on 2026-04-10 through focused frontend coverage
5. pending_live_validation

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
3. `POST /api/inbound-requests` from the dedicated `agent_chat` browser path

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
