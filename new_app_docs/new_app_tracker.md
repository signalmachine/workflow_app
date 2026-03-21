# workflow_app Tracker

Date: 2026-03-21
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
| Milestone 6 provider-backed AI execution | in_progress | Milestone 6 is now active. The first slice adds `internal/ai` provider configuration loading and validation for `OPENAI_API_KEY` plus `OPENAI_MODEL`, documents those variables in `.env.example`, and keeps the default build and test flow provider-independent while the live OpenAI-backed execution adapter, tool loop, and verification commands remain to be implemented; see `ai_provider_execution_plan.md` |
| Milestone 7 usable web application layer | planned | after provider-backed AI execution, v1 should land a usable web layer for auth, request submission, attachment transport, approval actions, and operator review on top of the shared backend foundation that a later v2 mobile client will also use; execute this as multiple narrow vertical slices rather than one monolithic delivery; see `web_application_layer_plan.md` |
| Minimum thin-v1 party and contact support depth | done | `parties` support records now cover external party identity plus support-depth contacts with tenant-safe service boundaries and integration tests |
| Remaining thin-v1 adopted-document gaps | done | thin v1 adopted document-family ownership is now implemented for work-order, invoice, and payment or receipt document families through module-owned one-to-one payload bridges keyed by `document_id`; see `adopted_document_ownership_remediation_plan.md` |
| Minimum thin-v1 inbound-request and browser-ingress foundation | done | persist-first inbound requests, request messages, PostgreSQL-backed attachments, transcription derivatives, queue claim and status transitions, stable `REQ-...` references, draft editing and hard deletion, queued-request amend-back-to-draft support, AI run causation, and reporting-visible inbound-request and processed-proposal review now exist for thin-v1 browser testing at the service and reporting-read-model level; see `inbound_request_and_attachment_foundation_plan.md` |

## 2. Immediate next steps

1. continue Milestone 6 from the now-landed provider-configuration slice by adding the OpenAI Go SDK dependency and the first provider-backed adapter in `internal/ai`
2. wire the first coordinator execution path from queued inbound requests into provider-backed run creation, artifact persistence, and recommendation persistence without bypassing approval or posting boundaries
3. add provider-gated verification coverage so default `go test ./...` remains provider-independent while live OpenAI checks stay opt-in
4. after Milestone 6, implement the usable web application layer on the same backend foundation that a later v2 mobile client will reuse
5. keep the codebase centered on the approved first-class modules while allowing support-depth records such as `parties` and `contacts` where the canonical module-boundary doc explicitly permits them
6. add attachments only where they support approval evidence, document support flows, or persisted inbound request intake
7. use `new_app_v1_gap_review_from_current_codebase.md` as historical context only, not as the live list of remaining missing foundation areas
8. use `new_app_implementation_defaults.md` as the default-rules reference during implementation
9. use `new_app_foundation_coverage.md` as the v1 completion checklist and foundation coverage control

## 2.1 Planned next implementation order

Recommended sequence:

1. continue Milestone 6 from the now-landed provider-configuration slice by adding the first OpenAI-backed coordinator execution path
2. implement the usable web application layer after Milestone 6 so operators can work through the browser on the same backend contracts that later mobile will reuse
3. execute Milestones 6 and 7 through small end-to-end slices rather than broad monolithic pushes so implementation stays controllable and reviewable

Reason:

1. the adopted document-family ownership mismatch is now closed for work-order, invoice, and payment or receipt families
2. inbound request intake, attachment support, queue claim semantics, and reporting-visible AI causation now sit on top of the stabilized document-adoption model
3. the reporting foundation is now complete enough for thin-v1 review and browser-ready read seams, so the next major remaining v1 gaps are live provider-backed AI execution and the usable web layer needed to make the application operable through the browser
4. the provider-config slice is a narrow Milestone 6 start that keeps the default contributor workflow provider-independent while the live execution path is built incrementally
5. both remaining milestones are substantial enough that they should be decomposed into narrow vertical slices to avoid schedule drag and architecture sprawl

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
