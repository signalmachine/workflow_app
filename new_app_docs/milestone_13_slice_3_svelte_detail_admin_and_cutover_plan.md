# workflow_app Milestone 13 Slice 3 Plan

Date: 2026-04-03
Status: Planned
Purpose: define the third Milestone 13 implementation slice so detail surfaces, admin surfaces, parity closeout, cutover, and legacy browser retirement happen together instead of being left as an indefinite cleanup tail.

## 1. Slice role

This slice finishes the migration.

It should:

1. migrate the promoted detail-route family
2. migrate settings and admin-maintenance surfaces needed for current operator continuity
3. complete bounded parity review against the existing browser route family
4. switch Go serving from the old template-based `/app` to the built Svelte frontend
5. retire superseded Go-template browser code and tests

## 2. Scope

In scope:

1. exact detail routes across inbound requests, approvals, proposals, documents, accounting, inventory, work orders, and audit where they remain part of the promoted browser family
2. `/app/settings`
3. `/app/admin`
4. current bounded admin-maintenance surfaces that already exist on the shared backend
5. Go cutover work for serving the Svelte build under `/app`
6. deletion or retirement of old template-based browser code that becomes dead after cutover
7. end-to-end validation and workflow-checklist updates tied to the new browser surface

Out of scope:

1. new workflow breadth unrelated to current browser parity
2. reopening completed backend milestones under the label of migration
3. a long-lived production dual-stack browser model

## 3. Required outcomes

This slice is complete only when:

1. the promoted detail-route family is available in Svelte
2. current admin and settings continuity required by the product is available in Svelte
3. the Go binary serves the Svelte build at `/app`
4. the earlier Go-template browser implementation is no longer the active serving path
5. workflow validation on the new browser surface is recorded on the `docs/workflows/` track

## 4. Cutover rule

1. switch serving only after the bounded parity checklist for the promoted route family passes
2. once serving switches, prefer immediate dead-code retirement for the old browser path rather than preserving a silent fallback
3. if one small fallback seam must remain temporarily, document it explicitly and give it a bounded removal plan

## 5. Verification

Before closing this slice:

1. run frontend verification on all migrated detail and admin routes
2. run canonical Go verification for the cutover and any backend changes
3. run bounded end-to-end workflow validation on the real `/app` plus `/api/...` seam after cutover
4. update workflow checklists and route expectations under `docs/workflows/`
