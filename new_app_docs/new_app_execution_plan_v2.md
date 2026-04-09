# workflow_app Execution Plan V2

Date: 2026-04-09
Status: Active execution order after archive reset
Purpose: define the current execution order without carrying the full completed milestone narrative in the default context.

## 1. Completed baseline

1. thin-v1 foundation is complete
2. Milestone 10 browser rebuild history is complete and archived
3. Milestone 11 shell and navigation history is complete and archived
4. Milestone 12 admin-maintenance planning history is archived after implementation progress moved beyond it
5. Milestone 13 Slice 1 and Slice 2 are implemented

## 2. Active execution order

1. complete the remaining Milestone 13 Slice 3 implementation work
2. verify the resulting Svelte cutover and shared backend continuity
3. update durable workflow-validation material in `docs/workflows/`
4. then promote one next bounded v2 milestone based on real remaining product or architecture need

## 2.1 Milestone 13 Slice 3 active checkpoint

The implemented Slice 3 baseline already includes:

1. Svelte continuity for `/app/settings`
2. Svelte admin continuity for `/app/admin`, `/app/admin/accounting`, `/app/admin/parties`, `/app/admin/parties/{party_id}`, `/app/admin/access`, and `/app/admin/inventory`
3. exact Svelte detail routes for inbound requests, approvals, proposals, documents, accounting entries, inventory movements, work orders, and audit events
4. direct detail-route continuity from the migrated list, landing, home, and coordinator-chat surfaces where exact IDs are already known
5. direct downstream accounting follow-through from exact request and proposal detail where the linked document already exists
6. Go serving of the embedded Svelte runtime at `/app`
7. retirement of the old template-based `/app` serving path and its compatibility branch

Remaining Slice 3 closeout is intentionally narrow:

1. bounded real-seam workflow validation on the current `/app` plus `/api/...` runtime
2. workflow-checklist and evidence updates in `docs/workflows/`
3. one grouped corrective slice only if that real-seam sweep exposes a real defect or missing support seam

## 3. Promotion rule for the next milestone

1. do not reopen completed milestone buckets broadly
2. if a real defect is found in completed work, handle it as one bounded corrective slice
3. choose the next milestone based on the strongest remaining production-shape need, not on historical sequence inertia

## 4. Verification rule

1. do not treat implementation as complete without running the required verification or recording an explicit blocker
2. use `../docs/technical_guides/07_testing_and_verification.md` for exact verification commands and workflow
3. for Milestone 13 closeout, verification must include frontend checks, canonical Go verification, and bounded end-to-end validation on the real served Svelte runtime
4. browser-serving changes must be checked for real asset serving, correct SPA fallback behavior, and one bounded browser smoke pass on `/app`
