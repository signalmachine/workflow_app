# workflow_app Execution Plan

Date: 2026-03-27
Status: Draft canonical execution plan
Purpose: define the execution path from the completed thin-v1 foundation through the active v2 implementation phase for the `workflow_app` codebase.

## 1. Milestone 0: New repo bootstrap

Goal:

1. create the `workflow_app` repository with the correct thin-v1 shape from day one

Outputs:

1. Go module
2. migration runner
3. first migrations for org, users, memberships, audit, idempotency
4. canonical `new_app_docs` planning set moved into the new repo
5. initial repository naming and product identity established as `workflow_app`

## 2. Milestone 1: Control boundary and document kernel

Goal:

1. establish the first safe business-control layer

Scope:

1. identity and auth
2. audit
3. idempotency
4. attachments
5. approvals and approval queue
6. AI run history, tool policy, and bounded coordinator-to-specialist routing
7. document kernel with named minimum document families
8. persisted inbound request intake plus queue-oriented AI processing seams needed for later browser and mobile clients

Exit criteria:

1. AI tools operate only through domain services
2. approvals exist for sensitive actions
3. auditable writes are transaction-safe
4. supported document families have one canonical identity model
5. the v1 document kernel explicitly supports work-order, invoice, payment or receipt, inventory receipt, issue, and adjustment, journal, and AI-draft document families
6. the multi-agent model is observable through durable run, step, artifact, recommendation, approval, and delegation records
7. project-linked inventory consumption in v1 reuses supported inventory issue or adjustment document flows with execution-context linkage rather than introducing a separate project-document family
8. inbound user requests can persist durably before AI processing, so later asynchronous AI operation does not depend on synchronous request-response handling

Current implementation checkpoint:

1. identity memberships, sessions, and role-aware service authorization are implemented
2. audit and idempotency foundations are implemented from Milestone 0
3. the shared document kernel now has central document identity, lifecycle state, and document-family registration
4. approvals, approval queue entries, and approval decisions are implemented with transactional audit writes
5. AI tool registry, tool policy, run history, artifacts, recommendations, approval linkage, and delegation traces are implemented for bounded coordinator-to-specialist routing
6. persisted inbound requests now exist with durable draft, queued, processing, processed, completed, failed, and cancelled status handling plus explicit queue-claim seams
7. PostgreSQL-backed attachment metadata, request-message linkage, and transcription-derived text records now exist with AI-run linkage preserved against the originating request
8. AI runs can now link back to persisted inbound requests, and reporting can review that request -> run -> recommendation -> approval -> document chain without raw SQL
9. adopted payload ownership is now implemented for work-order, invoice, and payment or receipt document families with shared party/contact support records reused where applicable rather than duplicated into document-local truth
10. persisted inbound requests now allocate a durable user-visible `REQ-...` reference at draft creation time and preserve it through queue submission so acknowledgments and review surfaces do not depend on raw UUIDs
11. draft inbound requests now support editing and hard deletion, while queued or cancelled pre-processing requests can return to `draft` for amendment and later resubmission without changing the stable request reference
12. inbound-request list filtering, request-detail lookup, and processed-proposal lookup now all support the stable `REQ-...` request reference, with reference resolution staying inside the authorized reporting read path
13. the current intake and review foundation is implemented at the service and reporting-read-model level, providing the backend base that later provider-backed AI execution and the promoted web application layer will build on
14. this milestone is now complete in its main control-boundary foundation, and the next v1 slice should focus on remaining reporting polish on top of that stable request-reference model

Remediation planning note:

1. detailed remediation for adopted document ownership is captured in `thin_v1_archive/adopted_document_ownership_remediation_plan.md`
2. detailed remediation for persist-first inbound request and attachment support is captured in `thin_v1_archive/inbound_request_and_attachment_foundation_plan.md`
3. the recommended implementation order is adopted document ownership first, then inbound request and attachment foundations

## 3. Milestone 2: Accounting and tax foundation

Goal:

1. make financial truth real early

Scope:

1. ledger accounts
2. journal entries and lines
3. balanced posting
4. posting lifecycle
5. reversal path
6. GST and TDS baseline seams
7. receivable and payable control-account readiness
8. period and numbering control seams

Exit criteria:

1. unbalanced posting cannot persist
2. posted truth is append-only
3. posting is explicit and idempotent
4. accounting posting occurs only through centralized posting paths from supported documents
5. proposer, submitter, approver, poster, and timestamps remain reconstructible for audit review
6. approval and posting remain distinct control boundaries, with approver-versus-poster separation left policy-configurable rather than hard-coded globally

Current implementation checkpoint:

1. ledger accounts now exist as a first-class `accounting` foundation record
2. append-only journal entries and journal lines are implemented with database-backed balance enforcement
3. approved documents can post through one centralized accounting service with duplicate-post protection
4. reversals create explicit reversal entries rather than mutating posted truth
5. GST and TDS foundation records plus tax-aware posting validation are now implemented
6. accounting periods, effective-date control, journal review queries, and control-account balance views are now implemented
7. this milestone is complete and now serves as the accounting base for the remaining thin-v1 support-record and adopted-document work

## 4. Milestone 3: Inventory foundation

Goal:

1. make stock truth real early

Scope:

1. items
2. locations
3. movements
4. receipt and issue flows
5. movement-purpose classification
6. source and destination modeling
7. baseline billable and non-billable material usage classification
8. explicit support for both trading resale flows and execution-consumption flows
9. schema room for serialized, lot-tracked, or installed-equipment traceability where the delivery case requires it

Exit criteria:

1. stock is derived from movements
2. movements are append-only
3. service-material and resale-stock flows are explicitly distinguishable
4. inventory documents can feed both execution context and accounting outcomes through explicit handoff paths
5. inventory foundation supports both buy-sell trading and service or project execution consumption on one shared movement model
6. traceable equipment classes can preserve identity-level linkage where the business flow requires it

Current implementation checkpoint:

1. `inventory_ops` now exists as a first-class module with item and location master records
2. append-only inventory movement truth is implemented with database-backed mutation blocking
3. stock is now derived from movements rather than stored as mutable truth
4. source and destination semantics, movement-purpose classification, and billable/non-billable usage classification are enforced in the first inventory service layer
5. approved inventory document references can now validate compatible receipt, issue, and adjustment movement recording paths
6. inventory document payload ownership now exists through inventory-owned document header and line records with one-to-one `document_id` linkage back to the shared `documents` kernel
7. pending accounting handoff rows and pending execution-context links now make downstream inventory outcomes explicit without shifting ownership away from `accounting` or future execution modules
8. this milestone is complete and now serves as the inventory base for the remaining thin-v1 support-record and adopted-document work

## 5. Milestone 4: Execution foundation

Goal:

1. make work execution a first-class truth layer

Scope:

1. work orders
2. tasks
3. assignment
4. labor capture
5. execution history
6. material-usage linkage
7. support for inventory consumption against work orders and other minimal supported execution contexts

Exit criteria:

1. work orders are the primary execution record
2. tasks have one clear accountable owner
3. labor and material context can attach to work execution cleanly
4. execution records link back to source documents and forward to accounting or inventory outcomes where applicable
5. work orders are primary where present, but the model can also attach inventory consumption to project execution without requiring a broad projects module

Current implementation checkpoint:

1. `work_orders` now exists as a first-class module with work-order identity, code, title, summary, and lifecycle status
2. execution status history is now append-only and recorded transactionally with work-order state changes
3. pending `inventory_ops.execution_links` for `work_order` context can now be consumed transactionally into first-class `work_orders.material_usages`
4. inventory execution-link consumption now marks the originating inventory linkage as `linked` without shifting ownership away from `inventory_ops`
5. `workflow` now owns shared work-order task records with one clear accountable worker and append-only task lifecycle participation
6. `workforce` now owns worker master records plus first labor-entry capture with cost-rate snapshots linked to work orders and optional work-order tasks
7. `workforce` now creates pending labor-accounting handoffs from append-only labor truth without writing ledger state directly
8. `accounting` can now consume approved journal documents plus pending labor handoffs into centralized work-order labor postings while preserving idempotent posting ownership
9. `inventory_ops` accounting handoffs now carry explicit cost snapshots so `accounting` can consume pending work-order-linked inventory handoffs into centralized material-cost postings without weakening posting ownership
10. remaining thin-v1 execution completion includes linking adopted work-order payload truth back to the shared document kernel with one-to-one ownership semantics
11. this milestone is complete in its first execution slice, but thin-v1 still requires adopted work-order document ownership completion

## 6. Milestone 5: Reports and review

Goal:

1. make the system usable for human review without broad UI scope

Scope:

1. approval queue
2. document lists
3. accounting views
4. inventory views
5. work-order views
6. audit lookup
7. baseline GST and TDS summary views

Exit criteria:

1. humans can inspect current truth without raw database access
2. reports reconcile to source documents and ledgers
3. thin v1 is operationally reviewable
4. humans can review the document -> approval -> posting chain without reconstructing it manually

Current implementation checkpoint:

1. `reporting` now exists as a first-class read-only module for operator-facing inspection surfaces
2. approval queue review now joins queue, approval, document, and posting state in one reporting read path
3. document review now exposes the document -> approval -> posting chain in one read model
4. accounting journal review and control-account balance review now sit behind `reporting` read surfaces rather than only domain-local queries
5. baseline GST and TDS summary views now exist as first-class reporting outputs with tax-code and control-account linkage
6. inventory stock review now exposes item and location labels on derived stock truth without requiring raw SQL
7. inventory movement review now exposes movement history with source and destination context plus linked document metadata in one reporting read path
8. inventory reconciliation review now exposes document-line accounting and execution handoff state, linked work-order context, and posted journal linkage for inventory review without raw SQL
9. work-order review now exposes task, labor, material-usage, and posted-cost rollups in one inspection surface
10. audit lookup now exists as a coherent reporting read path scoped to tenant and entity filters
11. minimum thin-v1 party and contact support depth now exists through tenant-safe `parties` support records and support-depth contacts
12. inbound-request list and detail review plus processed-proposal review now expose the persist-first request -> AI -> approval -> document chain needed for the promoted v1 application layer through service and reporting-read-model seams
13. inbound-request reporting now includes persisted cancellation and failure reasons with their timestamps plus submitter, session, metadata, attachment provenance, AI step and delegation detail, AI artifact detail, recommendation payloads, and processed-proposal document context so operator troubleshooting does not require raw table reads
14. queue-oriented reporting now also includes inbound-request status summaries plus processed-proposal status summaries so later web surfaces can show request and proposal health without reconstructing counts client-side
15. remaining thin-v1 completion is now concentrated around final operator-facing reporting polish rather than missing inbound-request, request-reference, or adopted-document foundation coverage
16. the next implementation target is to finish the remaining reporting polish before any v2 breadth work begins

## 7. Milestone 6: Provider-backed AI execution

Goal:

1. make the AI-agent-first operating model usable through a real provider-backed execution path

Scope:

1. OpenAI Go SDK integration in `internal/ai`
2. Responses API-based agent execution
3. environment wiring for `OPENAI_API_KEY` and bounded model selection
4. provider-backed coordinator execution over persisted inbound requests
5. bounded coordinator-to-specialist delegation on top of the existing durable run model
6. tool-loop and tool-policy enforcement for real provider-backed runs
7. structured-output and validation boundaries where provider output drives downstream proposals
8. provider timeout, retry, and failure handling that preserves business-state safety
9. modern workflow AI agent architecture patterns including persisted intake, explicit specialist routing, loop control, and structured execution traces
10. opt-in live-provider verification and integration tests

Exit criteria:

1. the active codebase uses the OpenAI Go SDK for the first real v1 agent path
2. provider-backed AI can process at least one queued inbound request through the existing request -> run -> artifact or recommendation -> approval -> document chain
3. AI writes still route only through normal domain services, approvals, and policy checks
4. `.env.example` documents the required provider configuration
5. the provider-backed path includes the core reliability and validation foundations needed for safe real use rather than only a thin happy-path demo
6. provider-backed tests are opt-in and do not make default repository verification depend on external credentials

Planned implementation checkpoint:

1. Milestone 6 is now active after review confirmed Milestone 5 reporting coverage is complete enough for thin-v1 review and browser-ready read seams
2. `internal/ai` now loads and validates OpenAI provider configuration from `OPENAI_API_KEY` and `OPENAI_MODEL`
3. `internal/ai` now uses the official OpenAI Go SDK and the Responses API for the first real provider-backed path
4. the first coordinator flow can now claim one queued inbound request, assemble request, attachment, and derived-text context, call the provider, persist the resulting coordinator run and step, write a provider brief artifact plus operator-review recommendation, and transition the request to `processed` or `failed`
5. the provider-backed coordinator path now includes a hard-capped Responses tool loop, per-capability tool-policy enforcement, and the first reporting read tool for inbound-request status summaries, with tool-execution metadata persisted in the coordinator step, artifact, and recommendation payloads
6. the coordinator can now optionally route one allowlisted specialist delegation through a durable child run and delegation record, with the final provider-backed artifact and recommendation persisting on that specialist run while the coordinator run remains the bounded parent
6. provider configuration remains optional so the default local build and database-backed test flow does not require external credentials
6. `.env.example` now documents the OpenAI variables needed for later live-provider slices
7. detailed sequencing and constraints are captured in `thin_v1_archive/ai_provider_execution_plan.md`
8. `internal/app` now provides shared backend seams over the provider-backed coordinator path, request submission, and attachment download, and `cmd/verify-agent` now exercises the live-processing seam for opt-in provider verification
9. remaining Milestone 6 work is now the operator-review and browser-usable auth contracts that will exercise the live AI path on top of that shared seam
10. this milestone should continue as a sequence of narrow vertical slices rather than one monolithic delivery

## 8. Milestone 7: Usable web application layer

Goal:

1. make the application usable through a real web layer on top of the shared backend foundations

Scope:

1. browser-usable auth and active-org handling
2. inbound request submission and tracking
3. attachment upload and download contracts
4. approval queue and decision surfaces
5. request, proposal, document, accounting, inventory, work-order, and audit review surfaces through the web layer
6. enough page flow, navigation, and operator continuity that v1 is usable without direct service or database tooling
7. shared backend contracts that a later v2 mobile client can also reuse
8. a Go-native web implementation approach with server-rendered HTML as the baseline and lightweight progressive enhancement rather than a separate frontend build system

Exit criteria:

1. the application has a real web layer rather than only service or API seams
2. the web layer can drive and inspect the live provider-backed AI path
3. the web and later mobile clients are planned around one backend foundation rather than separate backends
4. the web layer remains aligned with approval, audit, and domain-service boundaries

Current implementation checkpoint:

1. the first Milestone 7 browser slice now exists as a server-rendered `/app` surface on top of the shared backend seam
2. operators can now sign in with browser-session auth, submit inbound requests with file attachments, process the next queued request, review recent requests and pending approvals, open inbound-request detail, inspect attachments plus AI runs, artifacts, recommendations, and proposals, and decide approvals without dropping back to bespoke scripts
3. the next Milestone 7 browser slice widened that same application surface into downstream document and accounting review: `/app/review/documents` exposes document review, and `/app/review/accounting` exposes journal-entry review, control-account balance review, and tax-summary review on the same browser session and reporting-read foundation
4. the next Milestone 7 browser slice now also widens that same application surface into `/app/review/inventory`, `/app/review/work-orders`, `/app/review/work-orders/{work_order_id}`, and `/app/review/audit`, so operators can continue into stock, movement, reconciliation, work-order rollup, and audit inspection without leaving the browser layer
5. the shared backend seam now also exposes `GET /api/review/documents`, `GET /api/review/accounting/journal-entries`, `GET /api/review/accounting/control-account-balances`, `GET /api/review/accounting/tax-summaries`, `GET /api/review/processed-proposals`, `GET /api/review/processed-proposal-status-summary`, `GET /api/review/inventory/stock`, `GET /api/review/inventory/movements`, `GET /api/review/inventory/reconciliation`, `GET /api/review/work-orders`, `GET /api/review/work-orders/{work_order_id}`, and `GET /api/review/audit-events`, all reachable through the same browser session-cookie auth path as the rest of the browser flow
6. the latest continuity slice now also adds `/app/review/proposals`, giving operators a dedicated browser review surface for processed proposals with proposal-status summary, request-reference filtering, and direct continuation back into request detail and forward into downstream documents
7. accounting review now also supports exact source-`document_id` journal drill-down through both `/api/review/accounting/journal-entries?document_id=...` and `/app/review/accounting?document_id=...`, and the browser templates now link document, inventory-reconciliation, and work-order accounting context into that filtered accounting surface so operators can continue the downstream control chain without reopening broad journal lists
8. inventory review now also supports exact `movement_id` drill-down through both `/api/review/inventory/movements?movement_id=...` and `/app/review/inventory?movement_id=...`, and audit-page movement links now continue into that filtered inventory surface so operators can move from audit events back into the exact movement and reconciliation context instead of looping on the audit page
9. proposal and approval review now also support exact `recommendation_id` and exact `approval_id` drill-down through both `/api/review/processed-proposals?recommendation_id=...`, `/app/review/proposals?recommendation_id=...`, `/api/review/approval-queue?approval_id=...`, and `/app/review/approvals?approval_id=...`, and audit-page entity links now continue directly into exact inbound-request detail, exact approval review, and exact proposal review instead of leaving those audit results as dead ends
10. accounting review now also supports exact `entry_id` drill-down through both `/api/review/accounting/journal-entries?entry_id=...` and `/app/review/accounting?entry_id=...`, and the browser layer now includes a dedicated `/app/review/accounting/{entry_id}` page plus direct journal-entry links from document, approval, inventory-reconciliation, accounting-list, and audit surfaces so operators can continue into one exact posting record instead of reopening broader accounting lists
11. accounting review now also supports exact `tax_code` drill-down through `/api/review/accounting/tax-summaries?tax_code=...`, and the browser layer now includes dedicated `/app/review/accounting/control-accounts/{account_id}` and `/app/review/accounting/tax-summaries/{tax_code}` pages plus cross-links between control-account and tax-summary review so those accounting summaries become real operator stops instead of passive tables
12. inbound-request detail now also links request-level AI recommendations and downstream proposals into exact proposal review, exact approval review, filtered request review, and direct inbound-request or recommendation audit lookup so the intake-review page can continue operators into downstream control decisions instead of ending at evidence inspection
13. work-order review now also supports exact `work_order_id` drill-down through both `/api/review/work-orders?work_order_id=...` and `/app/review/work-orders?work_order_id=...`, and the browser layer now links `/app/review/work-orders/{work_order_id}` back into focused work-order review plus direct accounting review so execution-detail inspection remains inside one operator flow instead of ending on an isolated detail page
14. inventory movement detail now also links exact movement review into focused item stock, item movement history, item reconciliation, source and destination location movement history, source-document reconciliation, and source-document accounting review so exact inventory inspection remains inside one browser review path instead of ending on a standalone detail page
15. inventory item and location review now also have dedicated browser detail stops at `/app/review/inventory/items/{item_id}` and `/app/review/inventory/locations/{location_id}` on top of the already-landed stock and movement seams, and stock rows, movement detail, plus audit-page item or location entities now continue into those exact item and location pages instead of stopping at anchored list filters
16. exact inbound-request detail now also resolves `run:<agent-run-id>` and `delegation:<delegation-id>` through the already-landed `/app/inbound-requests/{request_reference_or_id}` and `/api/review/inbound-requests/{request_reference_or_id}` seams, and audit review now links `ai.agent_run` plus `ai.agent_delegation` entities back into the exact inbound-request execution trail so provider-execution audit events no longer dead-end on generic audit results
17. exact inbound-request detail now also resolves `step:<agent-step-id>` through those same shared browser and API seams, and audit review now links `ai.agent_run_step` plus `ai.agent_step` entities back into the exact inbound-request step block so step-level execution audit can land on the precise persisted step rather than only the broader request page
18. the earlier Milestone 7 slices for browser inbound-request lifecycle management and downstream provenance continuity for exact accounting and inventory stops are now landed on the shared backend seam plus `/app`
19. the latest dashboard refinement slice now turns `/app` into a stronger operator starting surface for parked, failed, cancelled, and in-flight requests by adding status-specific entry cues, operator-priority summary ordering, richer failure and cancellation context in recent-request rows, and tighter exact draft or lifecycle or execution continuations on the same shared backend seam
20. the final Milestone 7 closeout sweep is now complete, and it found one real late-stage browser continuity blocker: exact inbound-request detail and filtered inbound-request review were not carrying persisted cancellation and failure reasons forward even though the dashboard already surfaced them
21. that blocker is now fixed, and browser integration coverage now exercises parked-request lifecycle management plus browser visibility for draft, queued, processing, failed, cancelled, processed, and completed request states
22. this milestone is therefore complete, and the next active implementation target was Milestone 8 client-neutral backend hardening for later lightweight mobile reuse
23. detailed sequencing, slice scope, and control rules are captured in `thin_v1_archive/web_application_layer_plan.md`
24. residual browser work should now be treated as regression fixes or later UX refinement rather than as an active milestone plan
25. the active thin-v1 web-stack direction remains explicit: keep Go server-rendered HTML as the primary rendering model, prefer `htmx` where partial updates materially improve operator continuity, use `Alpine.js` only for small local state, and avoid introducing a Node toolchain unless the canonical planning set changes

## 9. Milestone 8: Client-neutral backend hardening for later mobile reuse

Goal:

1. preserve one backend foundation that the completed web layer and a later lightweight mobile client can both reuse cleanly

Scope:

1. harden client-neutral API contracts already exercised by the web layer
2. standardize async request-processing and review-state semantics where needed for non-browser clients
3. harden attachment upload, download, and transport expectations for later non-browser use
4. define the next auth-evolution path for non-browser clients without replacing the active browser-session model prematurely
5. introduce explicit API-shape and compatibility discipline where later mobile use would otherwise inherit browser-specific assumptions

Exit criteria:

1. operator-critical workflows remain available through shared `/api/...` contracts rather than only browser-form paths
2. browser and later mobile clients still sit on one backend foundation rather than diverging into separate backends
3. shared backend contracts preserve client-neutral request, status, approval, attachment, and review semantics clearly enough that a lightweight mobile client can be added without backend rework
4. any auth-evolution work completed in this milestone remains additive to the browser-session path rather than destabilizing the finished Milestone 7 browser layer

Planned implementation stance:

1. Milestone 8 follows Milestone 7 rather than competing with its remaining browser work
2. the first Milestone 8 slices should focus on backend hardening and contract discipline, not mobile-product build-out
3. dedicated mobile UX, full mobile auth-product depth, push notifications, offline behavior, and broader multimodal client work remain outside this milestone unless a later canonical update promotes them

Current implementation checkpoint:

1. the first four Milestone 8 slices are implemented on the shared `/api/...` seam through lifecycle, review-read, attachment, and approval-action contract hardening
2. the fifth planned slice is now completed in `thin_v1_archive/non_browser_auth_evolution_plan.md`
3. that auth plan keeps browser-session cookies as the active v1 auth path, treats the UUID actor-header path as temporary automation compatibility rather than the long-term client contract, and defines the next additive bearer-session path for lightweight non-browser clients on the same session foundation
4. Milestone 8 is therefore complete as a bounded client-neutral hardening milestone, and the next backend auth work should begin with the first additive token-session implementation slice from that plan rather than reopening Milestone 8 as an open-ended bucket

## 10. Thin-v1 checkpoint closeout

Closeout result:

1. the explicit closeout-and-decision pass is now complete
2. `go build ./...` completed cleanly on 2026-03-27
3. `set -a; source .env; set +a; go test -p 1 ./...` completed cleanly on 2026-03-27
4. review against `new_app_foundation_coverage.md`, `new_app_tracker.md`, and the completed milestone docs found no material missing foundation slice at the current thin-v1 depth
5. the codebase should therefore be treated as thin-v1 checkpoint complete rather than as an implicitly unfinished v1 milestone chain
6. Milestone 10 and later work should therefore be treated as v2-phase implementation on top of the completed thin-v1 foundation

## 11. Post-thin-v1 transition

Thin-v1 is complete.

Milestone 10 is the first active v2 milestone.

V2 in this repository should optimize for:

1. materially stronger operator usability
2. broader capability where it builds on the completed shared foundation
3. production readiness across application, operational, validation, and control layers
4. continued adherence to the same shared backend truth, workflow doctrine, approval boundaries, and database-first discipline

V2 must not be used to justify:

1. reopening foundational modeling that thin-v1 already completed without a concrete defect
2. reviving CRM-first or portal-first gravity
3. splitting the backend into browser-specific versus other-client truth models
4. weakening audit, approval, posting, or persistence boundaries in the name of speed

## 12. Historical post-checkpoint validation note

The next active step is now the bounded post-checkpoint validation and live user-testing readiness slice documented historically in `thin_v1_archive/post_checkpoint_validation_and_user_testing_plan.md`.

That slice should:

1. fix live provider-backed execution blockers first
2. validate the real browser plus shared-backend seam with the configured environment
3. run explicit canonical end-to-end workflows rather than broad exploratory testing first
4. continue focused review plus fix plus test loops until those workflows have no known blocking defects
5. end with one explicit readiness result for supervised AI-backed user testing or with an explicit blocker list

## 13. Milestone 9: User-testing readiness hardening

Goal:

1. strengthen the current implementation baseline before the paused live workflow validation resumes

Scope:

1. auth hardening on top of the existing browser-session and bearer-session foundations
2. bounded coordinator or specialist read-tool expansion inside the current multi-agent architecture
3. shared web or API seam decomposition where the current file concentration materially increases implementation and regression risk

Exit criteria:

1. the active auth path is materially stronger for guided user testing
2. the bounded AI capability surface is stronger without weakening approval and posting boundaries
3. the highest-risk shared web or API file concentration is reduced enough to support safer iteration
4. the repository is ready to resume the paused live workflow validation captured in `thin_v1_archive/post_checkpoint_validation_and_user_testing_plan.md`

Planning note:

1. this milestone exists because the post-checkpoint validation slice produced useful live signal but also identified bounded readiness gaps that are better addressed before further deep workflow testing
2. this milestone should stay bounded to readiness hardening rather than broad product expansion

## 14. Milestone 10: First active v2 milestone

Goal:

1. start the active v2 phase by rebuilding the current browser layer on the same shared backend seam so the web surface becomes modular, scalable, materially easier to evolve, and strong enough to support broader production-readiness work

Scope:

1. modular server-rendered template architecture
2. shared shell and navigation rebuild
3. dashboard and entry-surface rebuild
4. review-page family rebuild
5. detail-page family rebuild
6. parity closeout and removal of the legacy monolithic template structure

Exit criteria:

1. the active browser surface no longer depends on one monolithic shared template
2. the rebuilt shell distinguishes primary workflow actions from secondary review destinations
3. rebuilt dashboard, review, detail, and communication surfaces follow one coherent page model
4. the rebuilt web layer remains aligned with the shared backend truth, approval boundaries, and workflow continuity doctrine
5. the repository can resume browser-led workflow validation on the rebuilt surface instead of continuing piecemeal UI cleanup on the old structure
6. the milestone establishes the first explicit v2 implementation baseline rather than another thin-v1 cleanup pass

Current planning checkpoint:

1. thin-v1 is complete, and this milestone is now the first active v2 implementation milestone
2. the narrower density-cleanup posture from `thin_v1_archive/web_ui_streamlining_plan.md` is now superseded by the explicit full rebuild plan in `milestone_10_web_rebuild_plan.md`
3. implementation should begin only after this milestone is accepted as the next promoted slice
4. implementation should proceed through the three large pre-written slice plans captured in:
5. `milestone_10_slice_1_architecture_and_operator_entry_plan.md`
6. `milestone_10_slice_2_review_workbench_plan.md`
7. `milestone_10_slice_3_detail_and_closeout_plan.md`

## 15. Milestone 11: Operator shell, route taxonomy, and personalized home

Goal:

1. turn the rebuilt browser surface into a calmer and more scalable operator application by replacing heavy global chrome with a lighter primary navigation model, bundle landing pages, searchable route discovery, and a personalized `/app` home

Scope:

1. top-bar primary navigation for the true primary destinations
2. bundle landing pages for grouped route families such as review, operations, and selected domain areas like inventory where justified
3. searchable route catalog or command palette for navigation discovery
4. role-aware and user-aware home-surface composition on `/app`
5. explicit `Settings` and `Admin` surface posture aligned to access level and utility-navigation rules
6. soft-light visual-theme direction for the promoted shell and landing pages
7. minimum preference persistence or shortcut modeling needed to support personalization without creating a second truth model

Exit criteria:

1. the global shell moves to a top-bar model and exposes current workflow breadth cleanly
2. all currently supported workflows are exposed through the web UI either directly from the top bar or through landing pages reached from it
3. top-level navigation uses clear active-state differentiation
4. the top-bar bubble set can wrap across multiple rows without hidden overflow during the broad-exposure phase
5. the promoted shell uses a soft-light palette with dark readable text rather than stark white or dark heavy full-page backgrounds
6. broader route growth can later be streamlined through bundle pages, search, and optional presentation preferences without another shell rewrite
7. `/app` becomes a personalized workflow-centered home surface
8. the resulting operator shell remains aligned with the same shared backend truth and workflow doctrine
9. `Settings` and `Admin` are explicitly placed as role-aware secondary surfaces rather than leaking into the primary workflow shell

Current planning checkpoint:

1. this milestone is now active, with the bounded Slice 1 shell change promoted into implementation while Milestone 10 browser-review closeout remains on the separate workflow-validation track
2. it exists to refine information architecture and operator starting surfaces, not to replace the Milestone 10 modular rebuild
3. the canonical plan for this milestone is `milestone_11_operator_shell_and_navigation_plan.md`
4. Slice 1 is now implemented in code on the modular embedded bundle through a lighter top-bar shell, wrapped bubble navigation, a soft-light palette, and utility-menu placement for future `Settings` plus access-scoped `Admin`
5. Slice 2 landing pages and Slice 3 route-catalog plus personalized-home work remain pending
6. active v2 implementation should continue taking bounded refactoring opportunities where they materially improve modularity, streamline ownership boundaries, or reduce maintenance risk without drifting into unrelated rewrite work

## 16. Execution warning

Do not add CRM breadth, advanced projects, portal work, payroll, broad UI work, or advanced agent-autonomy features during milestones 0 through 5.

Do not treat Milestone 6 as permission to add broad autonomy, broad chat UX, or multi-provider breadth.

Do not treat Milestone 7 as permission to create a second backend for web versus mobile or to turn the product into a broad manual-entry ERP.
During Milestone 7, backend corrections and narrow shared-backend enhancements are still required when the web layer proves a concrete need, but those changes must remain in service of browser-layer integration and operator continuity rather than unrelated new backend feature expansion.

Do not treat Milestone 8 as permission to build the mobile product itself, fork the backend into web-specific versus mobile-specific truth models, or let backend hardening erase the still-required Milestone 7 browser completion criteria.

Do not treat v2 as permission to abandon the completed foundation discipline, duplicate truth ownership, or reintroduce CRM-first product gravity under a broader roadmap label.

## 17. Quality and sophistication rule

`workflow_app` is allowed to be thin in breadth, but it is not allowed to be weak in foundation design.

Implementation rule:

1. solve the foundational modeling problems in v1
2. defer breadth, not rigor
3. let v2 inherit a strong schema and control model rather than a quick MVP patchwork
