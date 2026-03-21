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
| Milestone 1 document and approval kernel | done | Shared document identity, approvals, approval queue, decisions, sessions, role-aware service boundaries, and the AI run, tool-policy, artifact, recommendation, and delegation trace foundation are implemented and covered by integration tests |
| Milestone 2 accounting foundation | done | Ledger accounts, append-only journal entries and lines, document-linked centralized posting, reversal entries, GST/TDS tax foundation records, accounting periods, effective-date posting control, journal review queries, and control-account balance views are implemented and covered by integration tests |
| Milestone 3 inventory foundation | done | The inventory foundation now includes `inventory_ops` items, locations, movement numbering, append-only movements, derived stock balances, inventory-owned document payload and line records, explicit execution and accounting handoffs, and costed inventory-accounting handoffs consumed through centralized journal posting covered by integration tests |
| Milestone 4 execution foundation | done | `work_orders` now includes first-class work-order records, append-only execution status history, transactional consumption of pending inventory execution links into work-order material-usage truth, workflow-owned work-order tasks with one accountable worker, workforce-owned labor capture with cost snapshots, and centralized accounting consumption of both labor and work-order-linked inventory handoffs covered by integration tests |
| Milestone 5 review and reporting surfaces | in_progress | `reporting` now exposes approval queue, document, accounting journal review, control-account balance review, GST/TDS tax summaries, inventory stock, inventory movement review, inventory reconciliation review, work-order, and audit lookup surfaces covered by integration tests |
| Minimum thin-v1 party and contact support depth | done | `parties` support records now cover external party identity plus support-depth contacts with tenant-safe service boundaries and integration tests |
| Remaining thin-v1 adopted-document gaps | in_progress | thin v1 still requires owning payload completion for work-order, invoice, and payment or receipt document families before the foundation can be considered complete |
| Minimum thin-v1 inbound-request and browser-ingress foundation | in_progress | thin v1 now needs persist-first inbound request intake, attachment references, queue-oriented AI processing, and review visibility for browser-based user testing without promoting full mobile-product depth |

## 2. Immediate next steps

1. complete adopted document-family ownership for work-order, invoice, and payment or receipt payloads with one-to-one linkage back to the shared `documents` kernel, using shared support-record identities where applicable rather than module-local duplicates
2. implement minimum persist-first inbound request intake, attachment references, queue-oriented AI processing, and browser-usable review visibility for thin-v1 user testing
3. finish the remaining Milestone 5 reporting polish after those foundation gaps land
4. keep the codebase limited to the approved first-class modules while thin-v1 foundation completion continues
5. add attachments only where they support approval evidence, document support flows, or persisted inbound request intake
6. use `new_app_v1_gap_review_from_current_codebase.md` as the reference list of remaining missing foundation areas
7. use `new_app_implementation_defaults.md` as the default-rules reference during implementation
8. use `new_app_foundation_coverage.md` as the v1 completion checklist and foundation coverage control

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
