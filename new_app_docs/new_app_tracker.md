# New App Tracker

Date: 2026-03-19
Status: Draft reset tracker
Purpose: track the replacement-codebase plan and guard against scope drift before implementation begins.

## 1. Current status

| Item | Status | Notes |
| --- | --- | --- |
| New thin-v1 reset accepted | in_progress | Planning reset is being prepared before the new repository is created |
| New canonical module map | done | The replacement codebase should start without a primary CRM module |
| New thin-v1 scope boundary | done | Thin-v1 scope has been restated for the replacement codebase |
| New execution path | done | Milestone order exists for the replacement codebase |
| Thin-v1 quality rule | done | The replacement app is explicitly foundation-heavy and rigor-heavy even though breadth stays narrow |
| Current-codebase v1 gap review | done | Main missing foundation gaps versus the new thin-v1 plan have been documented |
| V2 breadth parking lot | done | Deferred v2 capabilities now have their own plan folder so they do not leak into v1 |
| New repository bootstrap | not_started | Repository not created yet |

## 2. Immediate next steps

1. create the new repository
2. move `new_app_docs/` into that repository
3. scaffold Go module, migration runner, and Milestone 0 tables
4. keep the initial codebase limited to the approved first-class modules
5. use `new_app_v1_gap_review_from_current_codebase.md` as the reference list of missing foundation areas to prioritize in the new repo

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
