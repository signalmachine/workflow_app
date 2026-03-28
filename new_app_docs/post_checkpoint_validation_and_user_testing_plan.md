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

Current result on 2026-03-28:

1. `set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 go run ./cmd/app` started successfully against the configured development database
2. the real shared seam was exercised successfully through browser-session login, active-session lookup, dashboard load, inbound-request submission, queued-request processing, inbound-request review list and exact detail reads, processed-proposal review, approval-queue review, and browser request-detail rendering
3. the shared seam therefore appears operational outside direct service calls and outside repository-only tests

### 5.3 Step 3: execute canonical end-to-end workflows

Run a small, explicit workflow set rather than broad exploratory testing first.

Minimum canonical workflows:

1. login -> submit inbound request -> queue processing -> AI review result visible in browser review
2. draft request -> continue editing -> queue -> processing -> downstream request and proposal continuity
3. request that produces an approval need -> approval action -> downstream document or review continuity
4. failed provider or processing path -> failure visibility -> operator troubleshooting continuity
5. request or proposal or approval or downstream review cross-link continuity back to the originating `REQ-...` request and AI execution trail

Current workflow result on 2026-03-28:

1. workflow 1 passed structurally on the live seam: login -> submit inbound request -> queue processing -> AI review result visible in API and browser review
2. workflow 5 also passed for the same live request: exact request detail, AI run, step, artifact, and processed-proposal continuity remained visible by stable `REQ-...` reference through both `/api/review/...` and `/app/...`
3. that first live result exposed a concrete blocker: the provider-backed recommendation and artifact were generic and stale because they described the request as merely being in `processing` based on the queue-status summary tool output rather than centering the actual request message
4. that blocker is now cleared on 2026-03-28: `internal/ai` was hardened so request evidence is primary, queue summary is explicitly secondary, stale `processing`-style briefs fail validation, and the OpenAI provider gets one bounded repair turn before failing when the first structured brief stays too generic
5. after that hardening, `go build ./...`, `set -a; source .env; set +a; timeout 600s go test -p 1 ./...`, and `set -a; source .env; set +a; go run ./cmd/verify-agent` all passed, and the live verification summary returned a request-specific warehouse-pump inspection brief
6. workflows 2 through 4 remain unexecuted in the live environment and must still be run explicitly

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

Current interim result on 2026-03-28:

1. not ready yet for supervised AI-backed user testing
2. the first blocking defect on the live provider path is now cleared: the final persisted brief is request-centered again on the live OpenAI path, and stale queue-status wording is rejected before persistence
3. additional workflow coverage is still required for draft-amend continuity, approval-producing flows, and failed-provider or failed-processing visibility before readiness can be stated

### 5.6 Immediate blocker-remediation slice

Before continuing the remaining live workflows, clear the first known provider-quality blocker explicitly.

Bounded remediation objective:

1. make the first provider-backed coordinator brief request-centered, materially specific, and lifecycle-correct on the final persisted artifact and recommendation

Implementation order:

1. review the current coordinator prompt, tool-registration set, and final artifact or recommendation persistence path in `internal/ai`
2. tighten the coordinator instructions so the request message, attachments, and derived texts remain the primary evidence and org-level queue summary context is secondary only
3. narrow or rebalance the first read-tool path so the queue-status summary tool cannot dominate simple single-request reviews when it adds little value
4. add or strengthen tests in `internal/ai` and `internal/app` that fail when the final recommendation ignores the request content, repeats transient `processing`-style lifecycle wording after completion, or otherwise produces a generic status-only brief
5. rerun `go build ./...`, `set -a; source .env; set +a; go test -p 1 ./...`, and `set -a; source .env; set +a; go run ./cmd/verify-agent`
6. rerun the same live `/app` plus `/api/...` seam workflow used on 2026-03-28 and confirm that the final artifact and recommendation now stay centered on the actual request content
7. only after that blocker is cleared, resume workflows 2 through 4 from Step 5.3

Stop rule:

1. treat this as one bounded correctness slice, not as a broad AI-feature expansion
2. if the blocker cannot be cleared without introducing broader prompt or tool-surface redesign, record that explicitly before widening scope

Result on 2026-03-28:

1. the blocker-remediation slice succeeded without widening scope into a broader AI redesign
2. the landed hardening includes stronger coordinator instructions, request-centered validation in the shared coordinator contract, tighter request-evidence prompting, and one bounded OpenAI repair turn for generic first-pass structured output
3. repository verification and live provider verification both passed after that change
4. the next session should resume workflows 2 through 4 from section 5.3 rather than reopening this blocker-remediation slice

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
