# workflow_app Tracker

Date: 2026-03-27
Status: Draft reset tracker
Purpose: track the `workflow_app` plan and guard against scope drift during bootstrap and implementation.

## 1. Current status

| Item | Status | Notes |
| --- | --- | --- |
| New thin-v1 reset accepted | in_progress | Planning reset is now active inside the `workflow_app` repository and implementation is underway |
| New canonical module map | done | `workflow_app` should start without a primary CRM module |
| New thin-v1 scope boundary | done | Thin-v1 scope has been restated for `workflow_app` |
| New execution path | done | Milestone order exists for `workflow_app` |
| Thin-v1 quality rule | done | `workflow_app` is explicitly foundation-heavy and rigor-heavy even though breadth stays narrow |
| Current-codebase v1 gap review | done | Main missing foundation gaps versus the new thin-v1 plan have been documented |
| V2 breadth parking lot | done | Deferred v2 capabilities now have their own plan folder so they do not leak into v1 |
| Foundational document and posting bridge clarified | done | Minimum document families and cross-module posting path are explicit for v1 |
| Multi-agent stance clarified | done | Multi-agent architecture remains in v1 at bounded foundation depth; advanced autonomy is deferred to v2 |
| Implementation defaults captured | done | Locked defaults now exist as a canonical active doc for implementation decisions |
| Foundation coverage checklist captured | done | V1 completion now has an explicit foundation-complete checklist |
| Milestone 0 bootstrap | done | Go module, migration runner, env template, and control-boundary migrations are implemented and verified against primary and test databases |
| Milestone 1 document and approval kernel | done | Shared document identity, approvals, approval queue, decisions, sessions, role-aware service boundaries, and the AI run, tool-policy, artifact, recommendation, delegation trace, inbound-request, and attachment foundations are implemented with queue-oriented request processing seams and reporting-visible causation |
| Milestone 2 accounting foundation | done | Ledger accounts, append-only journal entries and lines, document-linked centralized posting, reversal entries, GST/TDS tax foundation records, accounting periods, effective-date posting control, journal review queries, and control-account balance views are implemented and covered by integration tests |
| Milestone 3 inventory foundation | done | The inventory foundation now includes `inventory_ops` items, locations, movement numbering, append-only movements, derived stock balances, inventory-owned document payload and line records, explicit execution and accounting handoffs, and costed inventory-accounting handoffs consumed through centralized journal posting covered by integration tests |
| Milestone 4 execution foundation | done | `work_orders` now includes first-class work-order records, append-only execution status history, transactional consumption of pending inventory execution links into work-order material-usage truth, workflow-owned work-order tasks with one accountable worker, workforce-owned labor capture with cost snapshots, and centralized accounting consumption of both labor and work-order-linked inventory handoffs covered by integration tests |
| Milestone 5 review and reporting surfaces | done | `reporting` now exposes approval queue, document, accounting journal review, control-account balance review, GST/TDS tax summaries, inventory stock, inventory movement review, inventory reconciliation review, work-order, audit lookup, inbound-request, and processed-proposal review surfaces covered by integration tests; stable inbound-request references now exist for operator tracking and submission acknowledgments, inbound-request list filtering now supports exact `REQ-...` reference lookup, request detail and processed-proposal reads resolve by stable `REQ-...` reference inside the authorized reporting read path instead of depending on raw UUID-only lookup, inbound-request review now also surfaces persisted cancellation and failure reasons with their timestamps for operator troubleshooting plus submitter, session, metadata, attachment provenance, AI step and delegation detail, AI artifact detail, and recommendation payload context, and queue-oriented reporting summaries now provide status-count rollups for inbound requests and processed proposals, so remaining v1 work has moved from reporting polish to provider-backed AI execution and the web layer |
| Milestone 6 provider-backed AI execution | done | `internal/ai` now includes optional OpenAI provider configuration loading, the official OpenAI Go SDK, a Responses-API-backed provider adapter, and a coordinator flow that can claim one queued inbound request, assemble request, attachment, and derived-text context, run a hard-capped tool loop, enforce per-capability tool policy, auto-execute the first reporting read tool when policy allows, optionally route the result through one allowlisted specialist capability with a durable child run and delegation record, persist the resulting coordinator and specialist steps with tool-execution metadata, write a provider brief artifact and operator-review recommendation, and mark the request `processed` or `failed` without bypassing existing control boundaries. `internal/app` now exposes a shared backend seam over that path plus submission, attachment transport, operator review, approval decisions, and browser-usable session auth: `POST /api/session/login` now starts an org-scoped browser session from org slug plus user email and sets the session cookies, `GET /api/session` resolves the active browser session, `POST /api/session/logout` revokes that session, `POST /api/agent/process-next-queued-inbound-request` processes the next queued request, `POST /api/inbound-requests` persists the initial request message plus optional inline attachments and queues the request in one backend workflow, `GET /api/attachments/{attachment_id}/content` serves persisted attachment bytes back through the same auth boundary, `GET /api/review/inbound-requests`, `GET /api/review/inbound-request-status-summary`, `GET /api/review/inbound-requests/{request_reference_or_id}`, `GET /api/review/processed-proposals`, `GET /api/review/processed-proposal-status-summary`, and `GET /api/review/approval-queue` surface browser-usable operator review reads backed by `reporting`, and `POST /api/approvals/{approval_id}/decision` routes approval actions through the existing workflow control boundary. `cmd/verify-agent` still provides opt-in live-provider verification on the same seam, `cmd/app` serves the widened runnable API surface, and provider-gated plus API integration coverage now exists; see `ai_provider_execution_plan.md` |
| Milestone 7 usable web application layer | done | the `/app` browser surface now covers the first operator loop plus downstream review continuity on the same shared backend seam. In addition to the already-landed sign-in, request submission, queue processing, approval actions, and document or accounting or inventory or work-order or audit review surfaces, the final closeout sweep is now complete: operators can save new requests as drafts, continue draft editing, add draft attachments, queue a draft from the browser, cancel queued pre-processing requests, return queued or cancelled requests to draft for amendment, hard-delete unprocessed drafts, and use the dashboard plus full inbound-request review as strong browser entry points for parked, failed, cancelled, in-flight, processed, and completed requests. The closeout sweep also fixed the last browser continuity gap by carrying persisted cancellation and failure timestamps plus reasons through exact inbound-request detail and filtered inbound-request review, and browser integration coverage now exercises both parked-request lifecycle management and full request-status visibility. Milestone 8 is now the active next implementation target. |
| Minimum thin-v1 party and contact support depth | done | `parties` support records now cover external party identity plus support-depth contacts with tenant-safe service boundaries and integration tests |
| Remaining thin-v1 adopted-document gaps | done | thin v1 adopted document-family ownership is now implemented for work-order, invoice, and payment or receipt document families through module-owned one-to-one payload bridges keyed by `document_id`; see `adopted_document_ownership_remediation_plan.md` |
| Minimum thin-v1 inbound-request and browser-ingress foundation | done | persist-first inbound requests, request messages, PostgreSQL-backed attachments, transcription derivatives, queue claim and status transitions, stable `REQ-...` references, draft editing and hard deletion, queued-request amend-back-to-draft support, AI run causation, and reporting-visible inbound-request and processed-proposal review now exist for thin-v1 browser testing at the service and reporting-read-model level; see `inbound_request_and_attachment_foundation_plan.md` |

## 2. Immediate next steps

1. start Milestone 8 client-neutral backend hardening now that Milestone 7 browser closeout is complete
2. keep Milestone 8 centered on shared `/api/...` contract discipline, request-status semantics, review-read consistency, approval-action semantics, and attachment behavior that later lightweight mobile clients will inherit
3. keep widening or correcting the shared backend only in client-neutral slices that strengthen correctness, continuity, or reuse rather than creating a browser-specific versus mobile-specific split
4. if Milestone 8 work exposes a browser-layer regression or a newly discovered residual Milestone 7 blocker, document it explicitly and fix it narrowly rather than reopening broad browser-surface expansion
5. keep the codebase centered on the approved first-class modules while allowing support-depth records such as `parties` and `contacts` where the canonical module-boundary doc explicitly permits them
6. add attachments only where they support approval evidence, document support flows, or persisted inbound request intake
7. use `new_app_v1_gap_review_from_current_codebase.md` as historical context only, not as the live list of remaining missing foundation areas
8. use `new_app_implementation_defaults.md` as the default-rules reference during implementation
9. use `new_app_foundation_coverage.md` as the v1 completion checklist and foundation coverage control
10. keep the thin-v1 web stack unchanged during Milestone 8 unless the canonical planning set explicitly changes it

## 2.1 Planned next implementation order

Recommended sequence:

1. start Milestone 8 with one explicit hardening pass across the shared `/api/...` contracts already exercised by the finished browser layer
2. keep that pass focused on real client-neutral semantics, exit-criteria validation, and narrow hardening fixes rather than new browser-surface expansion
3. keep richer draft-attachment editing beyond the landed additive upload flow as residual only if later evidence proves it materially necessary

Reason:

1. the adopted document-family ownership mismatch is now closed for work-order, invoice, and payment or receipt families
2. inbound request intake, attachment support, queue claim semantics, and reporting-visible AI causation now sit on top of the stabilized document-adoption model
3. the reporting foundation is complete enough for thin-v1 review and browser-ready read seams, and the provider-backed coordinator plus browser-session auth now make the shared backend usable from a real browser client
4. the landed coordinator slice includes a hard-capped tool loop with policy-enforced read-tool execution plus bounded specialist delegation while keeping the default contributor workflow provider-independent
5. shared backend contracts, a focused live-provider verification command, queued-request processing, request submission, attachment transport, operator review, approval action, and browser-usable session auth now exist for driving the live path outside direct service calls, and the landed browser slices plus the final closeout sweep now prove that seam with operator login, intake, parked-request lifecycle management, full request-status visibility, detail review, approval actions, plus downstream document, accounting, inventory, work-order, and audit review
6. the browser milestone is therefore complete enough that the next meaningful work is no longer web-surface expansion; it is shared-backend hardening for later lightweight mobile reuse
7. residual browser work should now be treated as regression fixes or later UX refinement rather than as an active milestone plan

## 2.2 Milestone 8 preview

Milestone 8 should follow Milestone 7 rather than compete with its remaining browser work.

Planned Milestone 8 focus:

1. client-neutral backend hardening for later lightweight mobile reuse on the same backend foundation
2. explicit contract discipline for shared `/api/...` paths already exercised by the web layer
3. standardization of request-status, review-read, approval-action, and attachment semantics where later non-browser clients would otherwise inherit browser-specific assumptions
4. next-step auth-evolution planning for non-browser clients without replacing the active browser-session model prematurely

Milestone 8 guardrails:

1. do not let Milestone 8 erase or defer the still-pending Milestone 7 browser work
2. do not treat Milestone 8 as the mobile-product build milestone
3. keep mobile UX, full mobile auth-product depth, push behavior, offline behavior, and broader multimodal client breadth outside Milestone 8 unless the canonical planning set is later updated explicitly

## 2.3 Remaining Milestone 7 slice analysis

The Milestone 7 closeout sweep is now complete.

Closeout result:

1. the final structured pass against the browser exit criteria found one real late-stage continuity blocker: exact inbound-request detail and filtered inbound-request review were not carrying persisted cancellation and failure reasons forward even though the dashboard already surfaced them
2. that blocker is now fixed, and browser integration coverage now exercises parked-request lifecycle management plus browser visibility for draft, queued, processing, failed, cancelled, processed, and completed request states
3. Milestone 7 is therefore complete, and the next active implementation slice is Milestone 8 client-neutral backend hardening

Planned-slice control rule:

1. treat the active slices above as the current explicit implementation plan
2. if later implementation reveals a concrete browser regression, record it explicitly as residual Milestone 7 follow-up rather than silently reopening Milestone 7 as a broad work bucket
3. do not use the possibility of later residual browser work as a reason to defer Milestone 8

## 3. Scope guardrail

The replacement effort fails if it repeats the old pattern of letting support workflows become the center of gravity.

Do not:

1. add a primary `crm` module
2. add opportunity or estimate breadth as the first workflow center
3. build large human operational UI before review and report surfaces
4. broaden scope because a feature feels commercially attractive

## 4. Success test

The replacement effort is on track only if:

1. documents, approvals, ledgers, inventory, execution, and reports are visibly the center of the model
2. the module list stays narrow
3. every new feature can be justified as foundation, not convenience
