# service_day Refactor Tracker v1

Date: 2026-03-19
Status: Active thin-v1 tracker
Purpose: track thin-v1 planning completion and implementation alignment against the active canonical plan.

## 1. Status legend

1. `not_started`
2. `in_progress`
3. `blocked`
4. `done`

## 2. Tracking rules

1. keep status factual
2. do not mark `done` without evidence
3. if scope changes, update the relevant canonical `plan_docs/` file first
4. this file is the single live implementation-status tracker inside `plan_docs/`
5. use `implementation_plan/implementation_tracker.md` only for historical implementation evidence and older remediation detail

## 3. Current canonical status

| Item | Status | Notes |
| --- | --- | --- |
| Thin-v1 doctrine reset | done | `plan_docs/` is now the active canonical planning set |
| Foundation checklist | in_progress | Core checklist exists, but implementation coverage is still partial |
| Schema and module-boundary canon | done | Reduced thin-v1 ownership and schema baseline now exists in `plan_docs/`, including explicit document, approval, and workforce ownership clarifications |
| Locked implementation defaults | done | Reduced thin-v1 defaults now exist in `plan_docs/`, including labor-capture and document-ownership clarifications |
| Thin-v1 milestone path | done | Active milestone path lives in `service_day_refactor_execution_plan_v1.md` |
| Thin-v1 status tracker | done | Active tracker now lives in this file |
| Legacy plan relabeling | in_progress | Main legacy docs are being relabeled so they stop reading as active canon |
| CRM-first legacy drift reduction | in_progress | Active thin-v1 docs are clear, but shipped CRM-heavy implementation still requires careful maintenance discipline |
| Accounting, inventory, work-order, and reporting implementation | in_progress | `documents` kernel foundation plus workflow-owned shared approvals for AI recommendation acceptance, approval listing, and terminal decision flow now exist, but inventory, work orders, reporting, broader approval orchestration, and broader document families remain open |

## 4. Active implementation implications

1. already-implemented CRM, AI, identity, workflow, attachment, notification, documents, and accounting slices remain in the repo
2. those slices should be maintained where needed, but they do not reset thin-v1 priorities
3. new implementation work should be justified against the thin-v1 milestones, not against the broader legacy roadmap
4. the currently shipped thin-v1 foundation progress includes a first shared `documents` kernel, accounting-journal linkage, and a workflow-owned shared-approval path that now covers AI recommendation acceptance, approval listing, and approval decisioning; broader approval orchestration depth still remains open

## 5. Current documentation cleanup checklist

| Item | Status | Notes |
| --- | --- | --- |
| `README.md` planning handoff aligned to `plan_docs/` | done | Root planning section now points to active docs |
| `AGENTS.md` key-file list aligned to active set | done | Missing active docs added |
| `implementation_plan/README.md` relabeled as legacy | done | Legacy entrypoint no longer presents itself as the canonical default |
| Main legacy docs relabeled from active/complete to legacy | done | Main broad-planning docs now read as legacy reference material |
| Legacy remediation-note posture clarified | in_progress | Focused remediation notes still live under `implementation_plan/`; keep their legacy role explicit and continue tracking cleanup through the linked remediation note |
| Migration map changed from provisional to accepted | done | Thin-v1 reset no longer reads as merely proposed |

## 6. Near-term thin-v1 implementation priority

1. preserve the current control boundary and audit posture
2. deepen accounting, inventory, workforce, work-order, document, and reporting foundations
3. avoid broadening CRM, launch UX, portal, exchange, payroll, or later business-mode scope while calling the result thin v1
4. treat approval-depth, reporting, inventory, workforce, and work-order foundation slices as the next valid areas of work; do not use the existing CRM-heavy implementation footprint as justification to resume sales-workflow breadth

## 7. Documentation remediation follow-up

If the repo still contains stale authority, status, or cross-reference issues after this cleanup wave, track the remaining work in:

1. `plan_docs/documentation_alignment_remediation_2026_03_18.md`

## 8. Evidence rule

Before marking a thin-v1 implementation item `done`, record:

1. what changed
2. what commands or review validated it
3. which active `plan_docs/` files the work aligns with

## 9. Recent implementation evidence

Approval queue depth in `workflow` was extended on 2026-03-19 to add shared approval listing plus pending-to-terminal approval decisioning over the existing `workflow_approvals` table and HTTP routes.

Validation evidence:

1. `go build ./cmd/app ./cmd/migrate`
2. `go test ./...`
3. `/bin/bash -lc 'set -a; source .env; set +a; timeout 120s go test -v -tags integration -count=1 ./internal/workflow'`
4. `/bin/bash -lc 'set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app'`

Plan alignment:

1. `plan_docs/service_day_refactor_execution_plan_v1.md`
2. `plan_docs/service_day_foundation_coverage_v1.md`
3. `plan_docs/service_day_schema_and_module_boundaries_v1.md`
