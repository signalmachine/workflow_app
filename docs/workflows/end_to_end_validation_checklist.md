# workflow_app End-to-End Validation Checklist

Date: 2026-04-08
Status: Durable checklist with a pre-validation browser-review precheck for the current Milestone 13 served Svelte runtime at `/app`, including the contextual-navigation shell, the promoted workflow, utility, admin, and detail-route families, before broader live workflow validation resumes
Purpose: provide a reusable bounded checklist for live review and testing of application end-to-end workflows on the real `/app` plus `/api/...` seam.

## 1. Use of this checklist

Use this checklist for:

1. post-checkpoint live workflow validation
2. supervised user-testing preparation
3. regression review after workflow-affecting changes

This checklist complements, but does not replace:

1. `docs/workflows/workflow_validation_track.md`
2. repository verification commands

Policy:

1. this checklist exists to prevent broad exploratory manual testing without a documented workflow list and explicit assertions
2. use it when the real testing question is workflow reliability on the actual shared operator seam, not merely compile success or isolated service correctness

## 2. Session-start checks

1. review `new_app_docs/new_app_tracker.md`
2. review `docs/workflows/workflow_validation_track.md`
3. review `docs/workflows/application_workflow_catalog.md`
4. rerun `set -a; source .env; set +a; go run ./cmd/verify-agent`
5. run `set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 go run ./cmd/app`
6. if the served Svelte runtime, the contextual-navigation shell, or any promoted workflow, utility, admin, or detail-route family is newly landed and not yet closed, review `/app/login`, `/app`, `/app/routes`, `/app/settings`, `/app/admin` for an admin actor, `/app/admin/accounting` for an admin actor, `/app/admin/parties` for an admin actor, `/app/admin/parties/{party_id}` for exact-detail contact creation, `/app/admin/access` for an admin actor, `/app/admin/inventory` for an admin actor, `/app/operations`, `/app/review`, `/app/inventory`, `/app/submit-inbound-request`, `/app/operations-feed`, and `/app/agent-chat`, plus `/app/inbound-requests/{request_reference_or_id}`, `/app/review/inbound-requests`, `/app/review/approvals`, `/app/review/proposals`, `/app/review/documents`, `/app/review/accounting`, `/app/review/inventory`, `/app/review/work-orders`, and `/app/review/audit` on desktop and a narrow-width viewport and record pass or blocker evidence before resuming live workflow validation

## 2.1 Milestone 13 post-cutover precheck

Before broader end-to-end workflow validation resumes, use this bounded post-cutover precheck:

1. confirm `/app/login`, `/app`, `/app/routes`, `/app/settings`, `/app/admin` for an admin actor, `/app/admin/accounting` for an admin actor, `/app/admin/parties` for an admin actor, `/app/admin/parties/{party_id}` including exact-detail contact creation, `/app/admin/access` for an admin actor, `/app/admin/inventory` for an admin actor, `/app/operations`, `/app/review`, `/app/inventory`, `/app/submit-inbound-request`, `/app/operations-feed`, and `/app/agent-chat` render cleanly and preserve their primary navigation actions, including multi-term route-catalog searches such as `pending approvals` or `failed requests`, plus visible active or inactive status controls on the promoted admin master-data pages
2. confirm `/app/inbound-requests/{request_reference_or_id}` renders request controls, evidence, execution trace, and downstream continuity links
3. confirm `/app/review/inbound-requests`, `/app/review/approvals`, `/app/review/proposals`, `/app/review/documents`, `/app/review/accounting`, `/app/review/inventory`, `/app/review/work-orders`, and `/app/review/audit` render cleanly with filters, contained tables, and exact drill-down links
4. confirm one exact drill-down chain across request -> proposal -> approval -> document
5. confirm one exact drill-down chain from request or proposal into accounting or inventory or work-order detail
6. confirm the served runtime returns real static assets under `/app/_app/...` and returns `404` for missing asset requests instead of silently falling back to the SPA shell
7. record pass or blocker evidence in `workflow_validation_track.md` before treating the Milestone 13 browser-validation closeout as complete
8. confirm the promoted shell now reads as a thin blue-gray operator application with a major-area sidebar, contextual section tabs, route-directory landing pages instead of hero-card mosaics, simpler login, and single-column workflow pages where those surfaces are the active default

## 2.2 Browser-review runbook

When section 2.1 applies, run the browser review in this order.

### 2.2.1 Setup

1. start the real app with the shared browser seam
2. sign in as an admin actor so `/app/admin` and the broader route family are reachable
3. prepare one desktop viewport around 1280 to 1440 pixels wide
4. prepare one narrow-width viewport around 390 to 430 pixels wide

### 2.2.2 Desktop pass

Review these routes first on desktop:

1. `/app/login`
2. `/app`
3. `/app/routes`
4. `/app/settings`
5. `/app/admin`
6. `/app/admin/accounting`
7. `/app/admin/parties`
8. `/app/admin/parties/{party_id}`
9. `/app/admin/access`
10. `/app/admin/inventory`
11. `/app/operations`
12. `/app/review`
13. `/app/inventory`
14. `/app/submit-inbound-request`
15. `/app/operations-feed`
16. `/app/agent-chat`
17. `/app/inbound-requests/{request_reference_or_id}`
18. `/app/review/inbound-requests`
19. `/app/review/approvals`
20. `/app/review/proposals`
21. `/app/review/documents`
22. `/app/review/accounting`
23. `/app/review/inventory`
24. `/app/review/work-orders`
25. `/app/review/audit`

For each route, check:

1. the shell renders cleanly, the active major area is obvious, and the contextual section tabs match the current page
2. the primary page action or main work surface is visible without excessive decorative framing
3. explanatory copy does not dominate the page above the real work surface
4. tables, filters, grouped route links, and continuity actions remain visually primary

### 2.2.3 Narrow-width pass

Rerun the same route family on a narrow-width viewport.

For each route, check:

1. navigation remains readable and reachable
2. no panel, form, metadata row, or action band overlaps or collapses into unreadable content
3. table overflow stays contained inside intended scroll wrappers
4. continuity links and key actions remain tappable and understandable

### 2.2.4 Continuity pass

After the route review, run these exact continuity chains:

1. request detail -> proposal detail -> approval detail -> document detail
2. request detail or proposal detail -> one downstream accounting or inventory or work-order detail route

For each chain, check:

1. the next exact link is easy to find
2. the destination page preserves the expected identifiers and workflow context
3. returning or continuing deeper does not lose the operator's place in the workflow

### 2.2.5 Evidence recording

Record one short note per route family or continuity chain:

1. `pass: <surface> - <reason>`
2. `blocker: <surface> - <defect> - <promoted fix plan if needed>`

Keep the notes short. The goal is explicit evidence, not a narrative test diary.

## 3. Workflow checklist

### 3.1 Submit and process inbound request

1. log in through the real browser or shared session API
2. submit a new inbound request from `/app/submit-inbound-request`
3. process the next queued inbound request
4. verify request status continuity
5. verify AI run, step, artifact, and recommendation continuity
6. verify exact request and proposal review continuity in both `/api/review/...` and `/app/...`

### 3.2 Draft-amend lifecycle

1. save a new draft
2. continue editing the same draft
3. queue the draft
4. process the queued request
5. verify request continuity from draft through processed state
6. verify proposal continuity after processing

### 3.3 Proposal to approval workflow

1. open a processed proposal that identifies a submitted document
2. request approval from the proposal surface
3. verify approval creation and recommendation linkage
4. decide the approval
5. verify downstream approval and document continuity
6. verify cross-links back to the originating request and AI trail

### 3.4 Failed-processing visibility

1. reproduce or trigger one failed provider or failed-processing path
2. verify failed request state
3. verify failure reason and failed timestamp
4. verify failed run or step visibility
5. verify exact request-detail troubleshooting continuity

## 4. Boundary assertions for every workflow

1. request persistence and lifecycle state are correct
2. AI records are durable and review-visible where expected
3. approval and document control-boundary behavior is correct where expected
4. browser and API review surfaces agree on the important facts
5. exact review pages and cross-links continue correctly across request, proposal, approval, document, and audit surfaces

## 4.1 Review sequence rule

Preferred workflow-critical review and testing sequence:

1. focused code review on the next high-risk workflow
2. narrow fix if a real blocker exists
3. bounded live end-to-end workflow execution
4. explicit pass/fail recording, blocker tracking, and readiness update

## 5. Closeout checks

1. run `go build ./...`
2. run `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./...`
3. update `new_app_docs/new_app_tracker.md` with explicit results
4. update `docs/workflows/workflow_validation_track.md` with workflow pass or fail evidence
5. update `docs/workflows/application_workflow_catalog.md` if durable workflow status or support depth changed
