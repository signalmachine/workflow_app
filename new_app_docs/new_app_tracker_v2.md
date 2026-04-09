# workflow_app Tracker V2

Date: 2026-04-09
Status: Thin-v1 is complete, Milestones 10 through 12 are implemented history, Milestone 13 Slice 1 and Slice 2 are implemented, and the active work is Milestone 13 Slice 3 closeout plus workflow-validation catch-up
Purpose: track the active implementation state, current sequencing, and immediate next steps without carrying full milestone history in the default context.

## 1. Current state

1. thin-v1 foundation is complete and should be treated as closed baseline rather than active open scope
2. the earlier Go-template browser rebuild and follow-on browser correction work are implemented history, not the forward planning surface
3. the promoted web direction is the Svelte-based web application on the shared Go backend
4. Milestone 12 admin-maintenance depth is implemented enough to stop being the active planning focus
5. Milestone 13 Slice 1 foundation and Slice 2 workflow-surface migration are implemented
6. Milestone 13 Slice 3 code checkpoints are largely landed: settings continuity, admin continuity, exact detail-route continuity, served-runtime cutover, and old-template retirement are implemented
7. workflow-validation and browser-review evidence in `docs/workflows/` still needs to catch up to the implemented runtime

## 2. Active implementation order

1. keep this doc cleanup as the completed prerequisite for the next implementation session
2. complete the remaining Milestone 13 Slice 3 closeout work on the shared Svelte plus Go seam
3. run the required verification for the Slice 3 change set
4. bring `docs/workflows/` up to date with explicit workflow-validation evidence and any revised reusable checklists
5. after Milestone 13 closeout, decide the next bounded v2 milestone instead of reopening completed milestone buckets broadly

## 2.1 Active Slice 3 detail now kept in this tracker

The next session should assume this active Slice 3 baseline is already landed:

1. `/app/settings` and the bounded admin family already have promoted Svelte continuity on shared backend seams
2. exact Svelte detail routes already exist for inbound requests, approvals, proposals, documents, accounting, inventory, work orders, and audit
3. migrated list, home, review-landing, and coordinator-chat surfaces already deep-link to exact detail routes where known identifiers exist
4. the served Go runtime already embeds and serves the Svelte frontend at `/app`
5. the retired template-browser `/app` layer has already been removed from the active codebase
6. the main remaining work is real-seam workflow validation and any narrowly grouped corrective follow-up that validation proves necessary

Use these supporting docs for the remaining closeout:

1. `../docs/workflows/end_to_end_validation_checklist.md` for bounded real-seam validation steps
2. `../docs/workflows/workflow_validation_track.md` for explicit validation evidence and blocker tracking

## 3. Current decision gate

The next implementation session should answer this in code and verification, not in more planning expansion:

1. what exact Slice 3 parity or cutover gaps still remain on `/app` and `/api/...`
2. what bounded backend or workflow-support seams are still required to close those gaps cleanly
3. what workflow-validation evidence is still missing once that code lands

## 4. Working rules

1. treat completed milestones as historical context unless a real defect forces a bounded corrective slice
2. keep new planning bounded to one coherent active concern at a time
3. prefer implementation and verification over planning expansion when the next step is already clear
4. keep the default active reading surface limited to the thin v2 docs in the root of `new_app_docs/`
5. if a detail is needed often during active work, keep it in the thin root docs rather than relying on repeated archive lookup
