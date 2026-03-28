# workflow_app Post-Checkpoint Validation And User Testing Plan

Date: 2026-03-28
Status: Active validation slice
Purpose: define the first explicit post-checkpoint implementation and validation step so the next session starts from a deliberate plan instead of reopening thin-v1 by momentum.

## 1. Why this plan exists

Thin-v1 is now checkpoint complete at its planned foundation depth.

That does not yet answer a narrower operational question:

1. is the live provider-backed AI path ready for supervised user testing
2. are the key operator workflows reliable enough for guided end-to-end testing
3. where do real defects still exist in the request -> AI -> review -> approval -> downstream inspection chain

The next step should therefore be validation-led rather than breadth-led.

## 2. Planning decision

The next active slice is:

1. a bounded post-checkpoint validation and live-workflow hardening pass

It is not:

1. a broad new feature milestone
2. silent reopening of thin-v1 scope
3. a v2 breadth promotion

## 3. Objective

Drive `workflow_app` to supervised AI-backed user-testing readiness by:

1. fixing live-path defects that block real provider-backed execution
2. exercising the shared backend and browser layer through canonical end-to-end workflows
3. repeating focused review plus structured testing until the canonical user-testing flows have no known blocking defects
4. recording any remaining non-blocking defects or gaps explicitly instead of inferring readiness from repository tests alone

## 4. Validation stance

Do not choose between code review and testing as if they are substitutes.

Use them in sequence:

1. focused code review on the currently failing or high-risk live path
2. targeted fixes for real blockers
3. stepwise end-to-end workflow testing on the real application seam
4. repeat that review -> fix -> test loop until the canonical workflows are stable
5. only then broader supervised operator testing

Reason:

1. code review is the fastest way to isolate live integration defects and weak contracts
2. end-to-end testing is the only credible way to assess readiness for an AI-agent-first operator workflow
3. testing without targeted review can become slow and noisy when a live-path defect blocks the whole chain

## 5. Planned execution order

### 5.1 Step 1: restore live-provider verification

Start with the focused live verification path:

1. run `set -a; source .env; set +a; go run ./cmd/verify-agent`
2. treat any failure before successful provider-backed request processing as a blocker to user-testing readiness
3. fix the blocker before moving into broader workflow testing

Current known blocker entering this plan:

1. resolved on 2026-03-28: the live OpenAI Responses integration no longer uses dotted function-tool names, and the bounded tool loop now continues statelessly with `store: false` instead of depending on `previous_response_id`
2. current Step 1 result: `go build ./...`, `set -a; source .env; set +a; go test -p 1 ./...`, and `set -a; source .env; set +a; go run ./cmd/verify-agent` all passed on 2026-03-28
3. remaining work in this plan is now Step 2 through Step 5 browser and end-to-end workflow validation rather than restoration of the live provider seam

### 5.2 Step 2: validate the core application seam

After the focused live verification passes:

1. run the browser-facing application with the configured environment
2. exercise the shared `/app` plus `/api/...` seam against the real database and real provider path
3. confirm that the application remains operable outside direct service calls and outside repository-only tests

Next-session start point:

1. start with Step 2 on the real `/app` plus `/api/...` seam rather than reopening Step 1
2. run `set -a; source .env; set +a; go run ./cmd/app`
3. execute the canonical browser-facing workflow set in Step 3 and record concrete pass or fail results in this document and the tracker

### 5.3 Step 3: execute canonical end-to-end workflows

Run a small, explicit workflow set rather than broad exploratory testing first.

Minimum canonical workflows:

1. login -> submit inbound request -> queue processing -> AI review result visible in browser review
2. draft request -> continue editing -> queue -> processing -> downstream request and proposal continuity
3. request that produces an approval need -> approval action -> downstream document or review continuity
4. failed provider or processing path -> failure visibility -> operator troubleshooting continuity
5. request or proposal or approval or downstream review cross-link continuity back to the originating `REQ-...` request and AI execution trail

### 5.4 Step 4: assert each workflow boundary explicitly

For each canonical workflow, verify:

1. inbound request persistence and lifecycle status transitions
2. AI run, step, artifact, recommendation, and delegation trace persistence where expected
3. approval creation and approval-decision behavior where expected
4. downstream review visibility through reporting and browser seams
5. operator-visible continuity across exact detail and cross-linked review pages

Do not mark a workflow passed only because the final page renders.

### 5.5 Step 5: summarize readiness

End the slice with one explicit result:

1. ready for supervised AI-backed user testing
2. not ready, with the blocking defects listed explicitly

Do not leave readiness implicit.

Readiness bar:

1. no known blocker remains in the live provider-backed path
2. no unresolved high-severity correctness, control-boundary, or operator-continuity defect remains in the canonical workflows
3. any residual low-severity defects are explicitly recorded and consciously accepted rather than ignored

## 6. Guardrails

During this slice:

1. keep the work bounded to validation, blocker fixes, and narrow hardening that directly supports the live workflow
2. do not broaden scope into general UX polish, mobile-product work, or new v2 features
3. do not reopen closed milestones except to record a concrete regression or blocker found during validation
4. prefer fixing the real shared backend or provider seam over adding testing-only workarounds
5. keep the testing centered on the real operator chain rather than synthetic isolated demos

## 7. Exit criteria

This slice is complete only when:

1. `go build ./...` passes
2. `set -a; source .env; set +a; go test -p 1 ./...` passes
3. `set -a; source .env; set +a; go run ./cmd/verify-agent` passes against the configured live provider path
4. the canonical end-to-end workflows above have been executed and their results documented
5. those workflows have no known blocking defects that would invalidate supervised user testing
6. the repository has one explicit readiness statement for supervised AI-backed user testing

## 8. Result handling

If the slice succeeds:

1. record supervised AI-backed user testing as the next operational step

If the slice fails:

1. record the blocking defects in the canonical tracker
2. promote only the narrowest bounded follow-up slice needed to clear those blockers
