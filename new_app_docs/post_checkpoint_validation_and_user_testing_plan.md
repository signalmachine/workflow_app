# workflow_app Post-Checkpoint Validation And User Testing Plan

Date: 2026-03-29
Status: Paused after partial validation; Milestone 9 is complete, and the next step is a full Milestone 9 review before the remaining live workflows resume
Purpose: define the first explicit post-checkpoint validation step, record the bounded partial result reached before pause, and preserve the exact work that should resume after the Milestone 9 implementation is reviewed against the plan docs.

## 1. Why this plan exists

Thin-v1 is now checkpoint complete at its planned foundation depth.

That does not yet answer a narrower operational question:

1. is the live provider-backed AI path ready for supervised user testing
2. are the key operator workflows reliable enough for guided end-to-end testing
3. where do real defects still exist in the request -> AI -> review -> approval -> downstream inspection chain

The next step should therefore be validation-led rather than breadth-led.

As of 2026-03-29, that validation pass remains intentionally paused after partial completion. The Milestone 9 readiness-hardening prerequisite is complete, but the next active step should first be a full review of the Milestone 9 implementation against the milestone plan and related canonical docs before the remaining live workflows resume.

## 2. Planning decision

The first post-checkpoint slice was:

1. a bounded post-checkpoint validation and live-workflow hardening pass

The next active slice is now:

1. review the completed Milestone 9 implementation against its plan docs
2. then resume the paused post-checkpoint live workflow validation from the remaining workflows captured below

The paused validation slice is not:

1. a broad new feature milestone
2. silent reopening of thin-v1 scope
3. a v2 breadth promotion

The new active milestone is also not:

1. a broad product expansion
2. a restart of thin-v1 foundation work
3. a substitute for the remaining live workflows captured below

## 3. Pause decision

The validation slice is not being marked complete.

It is being paused deliberately with explicit carry-forward work.

Reason:

1. the slice already produced useful signal through the first real provider-backed seam pass
2. that signal exposed bounded readiness gaps that are better addressed before further deep live-workflow testing
3. continuing the remaining live workflows immediately would produce lower-value testing signal while those known gaps remain open

Pause outcome:

1. treat this document as the canonical record of completed validation evidence and remaining workflow work
2. the Milestone 9 prerequisite is now satisfied, but the next session should first confirm implementation-versus-plan alignment before the remaining work resumes from Step 2 through Step 5
3. resume from the remaining workflows rather than restarting the whole validation phase later

## 4. Objective

Drive `workflow_app` to supervised AI-backed user-testing readiness by:

1. fixing live-path defects that block real provider-backed execution
2. exercising the shared backend and browser layer through canonical end-to-end workflows
3. repeating focused review plus structured testing until the canonical user-testing flows have no known blocking defects
4. recording any remaining non-blocking defects or gaps explicitly instead of inferring readiness from repository tests alone

## 5. Validation stance

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

## 6. Planned execution order

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

1. start with a complete review of the Milestone 9 implementation against `milestone_9_user_testing_readiness_hardening_plan.md` and the related canonical planning docs
2. if that review is clean, continue with Step 2 on the real `/app` plus `/api/...` seam rather than reopening readiness hardening
3. run `set -a; source .env; set +a; go run ./cmd/app`
4. execute the canonical browser-facing workflow set in Step 3 and record concrete pass or fail results in this document and the tracker

Current result on 2026-03-28:

1. `set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 go run ./cmd/app` started successfully against the configured development database
2. the real shared seam was exercised successfully through browser-session login, active-session lookup, dashboard load, inbound-request submission, queued-request processing, inbound-request review list and exact detail reads, processed-proposal review, approval-queue review, and browser request-detail rendering
3. the shared seam therefore appears operational outside direct service calls and outside repository-only tests

Next-session Step 2 start order:

1. rerun `set -a; source .env; set +a; go run ./cmd/verify-agent` first so the live provider seam is reconfirmed before wider browser checks continue
2. run `set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 go run ./cmd/app`
3. use the real shared `/app` plus `/api/...` seam rather than direct service calls for the remaining workflows unless a live blocker forces lower-level diagnosis

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
6. a second shared-backend continuity gap is now also closed inside the same validation slice: processed-proposal review now keeps document and suggested-queue continuity visible before approval exists by deriving those fields from recommendation payload when needed, and operators can now request workflow approval directly from processed proposals through both `/api/review/processed-proposals/{recommendation_id}/request-approval` and `/app/review/proposals/{recommendation_id}/request-approval` with atomic recommendation-link persistence
7. after that approval-request slice, `go build ./...`, `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./...`, and targeted browser-rendering coverage all passed
8. workflows 2 through 4 still remain to be executed explicitly in the live environment, but workflow 3 no longer has a known shared-backend blocker at the proposal-to-approval transition

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

Current interim result on 2026-03-29:

1. not ready yet for supervised AI-backed user testing
2. the first blocking defect on the live provider path is now cleared: the final persisted brief is request-centered again on the live OpenAI path, and stale queue-status wording is rejected before persistence
3. additional workflow coverage is still required for draft-amend continuity, approval-producing flows, and failed-provider or failed-processing visibility before readiness can be stated
4. the proposal-to-approval shared seam is now available for that remaining workflow coverage, so the remaining live workflows are deferred rather than blocked by a missing backend capability
5. Milestone 9 is now complete, so the next active work should be a full Milestone 9 review followed by workflows 2 through 4 in the order listed below

Deferred resume order for workflows 2 through 4 after Milestone 9:

1. workflow 2: save a draft, continue editing it, queue it, process it, and verify the resulting request plus proposal continuity in both `/api/review/...` and `/app/...`
2. workflow 3: start from a processed proposal that identifies a submitted document, request approval through the shared seam, decide that approval, and verify downstream approval plus document-review continuity
3. workflow 4: force or reproduce one failed provider or failed-processing path, then verify failure reason, timestamps, and troubleshooting continuity across exact request detail, filtered review, and any linked proposal or run views
4. after each workflow, record explicit pass or fail evidence in this document and `new_app_tracker.md` before moving to the next workflow
5. if the workflow support reference or reusable live checklist changes materially, update `docs/workflows/application_workflow_catalog.md` and `docs/workflows/end_to_end_validation_checklist.md` in the same change

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

## 7. Resume order after Milestone 9

With Milestone 9 now complete, resume this paused validation slice in this order:

1. review the completed Milestone 9 implementation against `milestone_9_user_testing_readiness_hardening_plan.md` and the related canonical planning docs
2. rerun `set -a; source .env; set +a; go run ./cmd/verify-agent`
3. run `set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 go run ./cmd/app`
4. execute workflow 2 for draft save or edit -> queue -> process continuity
5. execute workflow 3 for processed proposal -> request approval -> approval decision continuity
6. execute workflow 4 for failed provider or failed processing visibility
7. update this document and `new_app_tracker.md` with explicit pass or fail evidence after each workflow
8. end with one explicit readiness result rather than leaving user-testing readiness implicit

## 8. Guardrails

During this slice:

1. keep the work bounded to validation, blocker fixes, and narrow hardening that directly supports the live workflow
2. do not broaden scope into general UX polish, mobile-product work, or new v2 features
3. do not reopen closed milestones except to record a concrete regression or blocker found during validation
4. prefer fixing the real shared backend or provider seam over adding testing-only workarounds
5. keep the testing centered on the real operator chain rather than synthetic isolated demos

## 9. Exit criteria

This paused slice should only be marked complete when:

1. `go build ./...` passes
2. `set -a; source .env; set +a; go test -p 1 ./...` passes
3. `set -a; source .env; set +a; go run ./cmd/verify-agent` passes against the configured live provider path
4. the canonical end-to-end workflows above have been executed and their results documented
5. those workflows have no known blocking defects that would invalidate supervised user testing
6. the repository has one explicit readiness statement for supervised AI-backed user testing

## 10. Result handling

If the slice succeeds:

1. record supervised AI-backed user testing as the next operational step

If the slice fails:

1. record the blocking defects in the canonical tracker
2. promote only the narrowest bounded follow-up slice needed to clear those blockers
