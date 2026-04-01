# workflow_app End-to-End Validation Checklist

Date: 2026-04-01
Status: Durable checklist with pre-validation browser-review precheck for the rebuilt Milestone 10 route family plus the Milestone 11 Slice 2 landing pages before broader live workflow validation resumes
Purpose: provide a reusable bounded checklist for live review and testing of application end-to-end workflows on the real `/app` plus `/api/...` seam.

## 1. Use of this checklist

Use this checklist for:

1. post-checkpoint live workflow validation
2. supervised user-testing preparation
3. regression review after workflow-affecting changes

This checklist complements, but does not replace:

1. `new_app_docs/post_checkpoint_validation_and_user_testing_plan.md`
2. repository verification commands

Policy:

1. this checklist exists to prevent broad exploratory manual testing without a documented workflow list and explicit assertions
2. use it when the real testing question is workflow reliability on the actual shared operator seam, not merely compile success or isolated service correctness

## 2. Session-start checks

1. review `new_app_docs/new_app_tracker.md`
2. review `new_app_docs/post_checkpoint_validation_and_user_testing_plan.md`
3. review `docs/workflows/application_workflow_catalog.md`
4. rerun `set -a; source .env; set +a; go run ./cmd/verify-agent`
5. run `set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 go run ./cmd/app`
6. if the rebuilt Milestone 10 browser family or the Milestone 11 landing-page shell changes are newly landed and not yet closed, review `/app/login`, `/app`, `/app/operations`, `/app/review`, `/app/inventory`, `/app/submit-inbound-request`, `/app/operations-feed`, `/app/agent-chat`, `/app/inbound-requests/{request_reference_or_id}`, `/app/review/inbound-requests`, `/app/review/approvals`, `/app/review/proposals`, `/app/review/documents`, `/app/review/accounting`, `/app/review/inventory`, `/app/review/work-orders`, and `/app/review/audit` on desktop and a narrow-width viewport and record pass or blocker evidence before resuming live workflow validation

## 2.1 Milestone 10 closeout precheck

Before broader end-to-end workflow validation resumes, use this bounded Milestone 10 closeout precheck:

1. confirm `/app/login`, `/app`, `/app/operations`, `/app/review`, `/app/inventory`, `/app/submit-inbound-request`, `/app/operations-feed`, and `/app/agent-chat` render cleanly and preserve their primary navigation actions
2. confirm `/app/inbound-requests/{request_reference_or_id}` renders request controls, evidence, execution trace, and downstream continuity links
3. confirm `/app/review/inbound-requests`, `/app/review/approvals`, `/app/review/proposals`, `/app/review/documents`, `/app/review/accounting`, `/app/review/inventory`, `/app/review/work-orders`, and `/app/review/audit` render cleanly with filters, contained tables, and exact drill-down links
4. confirm one exact drill-down chain across request -> proposal -> approval -> document
5. confirm one exact drill-down chain from request or proposal into accounting or inventory or work-order detail
6. record pass or blocker evidence in `workflow_validation_track.md` before treating Milestone 10 as complete

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
4. update `new_app_docs/post_checkpoint_validation_and_user_testing_plan.md` with workflow pass or fail evidence
5. update `docs/workflows/application_workflow_catalog.md` if durable workflow status or support depth changed
