# workflow_app Milestone 14 Production Readiness and Workflow Validation Plan

Date: 2026-04-10
Status: Active next milestone
Purpose: define the next bounded v2 milestone after Milestone 13 so the repository closes the implementation and documentation gaps found in the codebase review, strengthens production-readiness verification, resumes deeper workflow validation, and makes the documentation set reflect the current promoted runtime truth.

## 1. Why this milestone exists

Milestone 13 closed the Svelte cutover and real-browser closeout, but the implementation review immediately after that closeout found that the repository still has residual gaps that should be handled as one bounded next milestone rather than as scattered follow-up edits.

The review found eight concrete needs:

1. the promoted Svelte browser layer does not currently expose the full parked-request lifecycle that older docs and archived status still describe as implemented
2. some workflow and user-guide material still points to retired or nonexistent browser paths
3. production-readiness verification is stronger than it was before Milestone 13 closeout, but it still needs broader and more intentional test coverage, workflow-proof coverage, and failure-path validation
4. the canonical and downstream docs need a truth pass so current implementation status, workflow support, and operator guidance are aligned again
5. the shared Svelte shell still needs one bounded desktop layout correction so the sidebar starts directly beneath the top application bar and the contextual tab row begins over the main content column instead of spanning from the far left edge
6. the promoted accounting review surface still lacks baseline operator financial statements such as trial balance, balance sheet, and income statement, and Milestone 14 should also review adjacent reporting gaps and add clearly similar production-shape capability where it materially improves readiness
7. the current contextual-tab and admin information architecture is too noisy and crowded on several promoted pages, and Milestone 14 should restructure it around task-group directory pages that open dedicated destinations for reports, lists, master-data creation, and workflow-specific screens
8. the default demo entity `North Harbor Works` is still under-seeded for realistic user testing and should ship with a standard chart of accounts plus minimum master data needed to exercise the promoted workflows and admin/list/report surfaces

This milestone exists to close those gaps while keeping the repository on the current v2 implementation path instead of reopening historical milestone buckets broadly.

Milestone 14 should also be treated as the milestone that makes the promoted application defensible for bounded user testing on the current runtime instead of keeping that readiness state implicit.

## 2. Milestone objectives

Milestone 14 should achieve the following:

1. fix the concrete implementation and documentation issues uncovered during the codebase review
2. expand and harden verification so the promoted application is closer to a production-ready quality bar
3. execute and record bounded workflow validation against the real served `/app` plus `/api/...` seam
4. update and extend the documentation set so canonical docs, workflow docs, technical guides, and user guides reflect the current codebase truth
5. make the application ready for bounded user testing on the promoted runtime by closing the current implementation, workflow, and documentation mismatches that would otherwise mislead testers
6. add the first baseline accounting reports required for a production-oriented operator surface, starting with trial balance, balance sheet, and income statement on the shared reporting seam
7. review the promoted application for nearby production-shape gaps discovered during Milestone 14 work and add clearly similar supporting capability when doing so materially improves correctness, readiness, or operator continuity
8. restructure the contextual-tab and admin navigation model so tabs lead to focused directory pages and dedicated destination pages rather than overloaded multi-purpose working screens
9. seed the demo entity `North Harbor Works` with a standard chart of accounts and other minimum master data needed for bounded user testing and realistic operator review

## 3. Scope

In scope:

1. correcting the parked-request lifecycle gap on the promoted Svelte browser surface, or explicitly narrowing the supported browser contract if that is the justified production decision
2. correcting stale route references and workflow claims in canonical docs, workflow docs, technical guides, and user guides
3. adding focused frontend, backend, integration, and browser tests where current coverage is too narrow for the real production risk
4. executing bounded live workflow validation for the supported operator flows and recording explicit pass, fail, or blocker evidence
5. reviewing and improving durable guidance in `docs/workflows/`, `docs/user_guides/`, `docs/technical_guides/`, `README.md`, and the active `new_app_docs/` root docs when they drift
6. promoting new workflow-facing docs where the current codebase supports behavior that is not yet documented clearly enough for operators or maintainers
7. making one bounded shared-shell desktop layout correction so the left sidebar touches the top application bar and the contextual tab row begins to the right over the main content area
8. adding baseline accounting reports for trial balance, balance sheet, and income statement on the shared backend seam together with the web review surfaces needed to use them
9. reviewing the application for adjacent production-readiness or parity gaps discovered during Milestone 14 implementation and adding clearly similar features where those additions are coherent, bounded, and justified
10. restructuring contextual tabs in promoted areas such as Admin and Accounting so tabs become grouped directory pages like `Master Data`, `Lists`, `Reports`, or similar task-oriented buckets, and concrete create/list/report destinations open on dedicated pages
11. defining and seeding the minimum demo-data baseline for `North Harbor Works`, including a standard chart of accounts and other essential master data for realistic testing

Out of scope:

1. reopening Milestone 13 browser migration work as a broad architecture program
2. broad new workflow-product expansion unrelated to readiness, validation, or the review findings
3. creating parallel truth between canonical planning docs and downstream guides
4. weakening the shared Go-backend ownership model by shifting business logic into browser-only code
5. turning the milestone into an unbounded feature bucket without tying additions to clear readiness, workflow, reporting, or operator-continuity needs
6. adding navigation depth for its own sake where a direct destination is already the clearer operator path
7. seeding broad fake ERP data without tying it to the minimum realistic baseline needed for supported workflows, reports, and user testing

## 4. Review findings that Milestone 14 must address first

The first bounded corrective slice should start from the review findings already established against the current codebase:

1. the backend still supports draft update, queue, cancel, amend, and delete on `/api/inbound-requests/{request_id}/{action}`, but the promoted Svelte browser runtime currently exposes only new-request creation plus read-only detail continuity
2. the inbound-request lifecycle user guide still points operators to a nonexistent browser path for queue processing instead of the current operations-page action
3. route vocabulary still mixes `/app/inbound-requests` and `/app/review/inbound-requests` in places where only one of those surfaces is actually implemented as a list route on the promoted Svelte runtime
4. archived and downstream documentation still overstates current browser support for parked-request lifecycle flows, so documentation truth is ahead of the current promoted implementation
5. the current shared desktop shell still places the contextual tab row across the far-left edge above the sidebar instead of keeping the sidebar directly under the top app bar and starting contextual tabs over the content column
6. the promoted accounting area still stops short of standard operator financial statements, which leaves accounting review materially underbuilt for a production-oriented business application
7. admin and other promoted contextual-tab families still mix too many actions, forms, and list surfaces onto single pages instead of using cleaner grouped directory pages plus dedicated destination pages
8. the default demo entity still needs a baseline chart of accounts and other minimum master data so user testing does not begin from an unrealistically empty state

Milestone 14 should not treat those as optional polish items. They are the first correctness and readiness gaps to close.

## 5. Delivery slices

Milestone 14 should execute through six bounded slices.

### 5.1 Slice 1: review-gap corrective pass

Goal:

1. remove the concrete implementation and documentation mismatches found in the review

Scope:

1. decide and implement the correct parked-request lifecycle contract on the promoted Svelte surface
2. align inbound-request list, detail, and action continuity on the current `/app` route family
3. update stale route references and overclaimed workflow support in canonical and downstream docs
4. add focused regression coverage for the corrected behavior
5. implement the bounded shared-shell desktop layout correction for sidebar and contextual-tab placement and add focused coverage for the corrected layout contract
6. begin the contextual-tab and destination-page restructuring in the most visibly crowded promoted areas, starting with Admin and Accounting

Stop rule:

1. stop once the promoted browser contract and the documentation contract match again for inbound-request lifecycle support
2. carry the first bounded navigation-structure correction far enough that the promoted shell direction is explicit and ready for the broader Milestone 14 information-architecture pass

Current checkpoint recorded on 2026-04-10:

1. the exact inbound-request detail route now exposes the shared backend-owned parked-request lifecycle controls directly in Svelte: draft save plus queue plus delete, and queued cancel plus amend-back-to-draft
2. the bounded shared-shell desktop layout correction is also landed: the sidebar remains directly under the top app bar while the contextual tabs now begin over the main content column
3. focused frontend verification for the new mutation clients, request-detail lifecycle controls, and shell layout contract passed
4. a follow-up documentation-truth checkpoint on 2026-04-10 realigned the workflow catalog, inbound-request lifecycle guide, agent-chat guide, inbound-request technical guide, and active scope note with the current route contract
5. the grouped navigation checkpoint then added Admin `Master Data` and `Lists` directory pages, changed the Admin contextual tabs to grouped directories plus Access, and changed `/app/review/accounting` into an accounting report directory with dedicated `journal-entries`, `control-balances`, and `tax-summaries` destinations
6. Slice 1 is now closed enough to move the active implementation path to baseline accounting reports in Slice 3 while preserving broader information-architecture cleanup as follow-through when new crowded areas appear

### 5.2 Slice 2: production-readiness test expansion

Goal:

1. expand verification where real production risk still exceeds the current test and browser-proof surface

Scope:

1. review current backend, frontend, integration, and browser coverage for obvious gaps
2. add tests for workflow-critical mutation paths, exact drill-down continuity, failure visibility, and auth or access boundaries where coverage is still weak
3. strengthen focused rerun patterns and verification guidance when the current technical guide is too generic for the actual risk
4. keep test additions aligned with the real promoted runtime rather than with retired template-era behavior

Guardrail:

1. prefer tests that prove business continuity, workflow correctness, or transport contracts over low-signal snapshot or copy-only assertions

Current checkpoint recorded on 2026-04-11:

1. `internal/app` API integration coverage now exercises the promoted inbound-request lifecycle mutation seam end-to-end through browser-session cookies
2. the new coverage proves draft creation without queuing, draft update on the persisted request/message seam, queueing, cross-org cancel rejection without state mutation, same-org cancellation with reason and timestamp, amend-back-to-draft with queue/cancel metadata reset, and final draft deletion
3. focused `internal/app` verification, `go build ./cmd/... ./internal/...`, serialized canonical Go verification, and gopls diagnostics passed for this checkpoint
4. additional approval-decision API coverage now proves that a browser-session approver from another org receives a not-found response for a foreign approval and does not mutate the pending approval status, decision metadata, or submitted document state
5. the focused approval-decision integration test group, full `internal/app` package verification, `go build ./cmd/... ./internal/...`, serialized canonical Go verification, `git diff --check`, and gopls diagnostics passed for that additional checkpoint
6. additional failed-provider API coverage now proves that the browser-session process-next endpoint returns the exact failed request reference and run id and that exact inbound-request review detail exposes the failed request state, failed run state, and failed provider-execution step with the failure payload
7. the focused process-next failure-continuity test, full `internal/app` package verification, `go build ./cmd/... ./internal/...`, serialized canonical Go verification, and gopls diagnostics passed for that additional checkpoint
8. additional proposal-approval API boundary coverage now proves that a browser-session actor from another org receives a not-found response when requesting approval for a foreign processed proposal and does not mutate the recommendation approval link, create an approval row, or change the submitted document state
9. the focused cross-org proposal-approval boundary test, full `internal/app` package verification, `go build ./cmd/... ./internal/...`, serialized canonical Go verification, `git diff --check`, and gopls diagnostics passed for that additional checkpoint
10. additional accounting-report API boundary coverage now proves that trial balance, balance sheet, and income statement reads through browser-session cookies stay scoped to the authenticated org and do not expose another org's posted journal activity
11. the focused accounting-report org-boundary test, full `internal/app` package verification, `go build ./cmd/... ./internal/...`, serialized canonical Go verification, and gopls diagnostics passed for that additional checkpoint
12. Slice 2 should continue with similarly high-value production-readiness tests where remaining workflow, failure-path, auth, or transport-contract risk still exceeds coverage

### 5.3 Slice 3: baseline accounting reports and reporting-gap pass

Goal:

1. add the first standard financial statements required for a production-oriented accounting surface and close the most obvious adjacent reporting gaps discovered during implementation review

Scope:

1. implement trial balance on the shared reporting seam
2. implement balance sheet on the shared reporting seam
3. implement income statement on the shared reporting seam
4. add the promoted web surfaces needed to review those reports without moving accounting logic into Svelte
5. restructure the Accounting contextual tabs and destination pages so reports, lists, and any action-oriented destinations are grouped cleanly instead of crowded onto one page
6. review nearby accounting-reporting gaps exposed during this work and add clearly similar bounded report capability when justified by the same shared reporting foundation
7. add focused tests and documentation for the report contracts, filters, and presentation surfaces

Guardrail:

1. keep financial-statement rules explicit, auditable, and backend-owned, especially account classification, sign conventions, and date or period filtering
2. keep Accounting tabs task-oriented and directory-oriented rather than turning one accounting page into a dense mixed workspace

Current checkpoint recorded on 2026-04-10:

1. trial balance is implemented on the shared reporting seam with active-ledger-account rows, debit and credit balance totals, as-of filtering, and an explicit imbalance total
2. balance sheet is implemented on the shared reporting seam with assets, liabilities, equity, current earnings derived from revenue and expense balances, as-of filtering, and an explicit imbalance total
3. income statement is implemented on the shared reporting seam with revenue, expense, and net-income totals over an effective-date range
4. `/app/review/accounting` now links to dedicated `trial-balance`, `balance-sheet`, and `income-statement` Svelte report destinations without moving report composition rules into browser-only state
5. focused reporting integration tests, focused app route-serving tests, focused Svelte component tests, `npm --prefix web run check`, `npm --prefix web run build`, `go build ./cmd/... ./internal/...`, `git diff --check`, and gopls diagnostics passed for this checkpoint
6. Slice 3 is closed enough to move the active implementation path to Slice 4 demo-data seeding

### 5.4 Slice 4: demo-data and master-data baseline for user testing

Goal:

1. make the default demo entity usable for realistic bounded user testing by seeding the minimum accounting and master-data baseline needed by the promoted runtime

Scope:

1. define a standard chart of accounts baseline for `North Harbor Works`
2. seed that chart of accounts through the shared backend-owned setup path
3. seed the minimum additional master data needed for realistic testing of promoted admin, list, report, and workflow surfaces
4. keep the demo baseline coherent with the new accounting reports and the user-testing posture rather than as disconnected sample data
5. add focused verification and documentation for how the demo baseline is created, refreshed, or relied on during testing

Guardrail:

1. seed only the minimum realistic baseline needed for supported operator flows and reports; do not turn the demo entity into an uncontrolled catch-all data dump

Current checkpoint recorded on 2026-04-11:

1. `cmd/bootstrap-admin` now seeds the minimum North Harbor Works demo baseline by default through `internal/setup`
2. the seed is idempotent and currently covers a standard chart of accounts, GST 18% sales and purchase tax codes, one open `FY2026-27` accounting period, sample customer and vendor parties with primary contacts, starter inventory items, and starter inventory locations
3. operators can pass `-seed-demo-baseline=false` when they need only the friendly admin login records
4. focused setup integration tests, focused bootstrap/setup package tests, and gopls diagnostics passed for this checkpoint
5. Slice 4 is closed enough to move the active implementation path to production-readiness test expansion and deferred workflow validation

### 5.5 Slice 5: navigation and information-architecture cleanup

Goal:

1. reduce noise and crowding in the promoted desktop runtime by making contextual tabs act as grouped directory pages that link to dedicated destinations

Scope:

1. apply the contextual-tab directory pattern in Admin, starting with categories such as `Master Data` and `Lists`
2. move concrete create flows such as new ledger account, tax code, party, item, or location into dedicated pages reached from those grouped pages
3. move concrete list or review destinations such as chart of accounts, party lists, and similar record lists into dedicated pages reached from grouped list pages
4. apply the same pattern in Accounting and other promoted areas where it materially reduces crowding and improves operator clarity
5. preserve or improve direct workflow continuity while reducing page-level noise
6. add focused route, navigation, and browser verification for the restructured destination model

Guardrail:

1. contextual tabs should represent task groups or sub-areas, while actual work happens on the dedicated destination pages behind those grouped links

### 5.6 Slice 6: live workflow validation and evidence refresh

Goal:

1. execute the bounded workflow-validation backlog on the real seam and record explicit evidence

Scope:

1. draft or parked request lifecycle continuity on the promoted runtime after the Slice 1 corrective pass
2. failed-provider or failed-processing visibility and troubleshooting continuity
3. exact request -> proposal -> approval -> document -> accounting continuity reruns when changed code warrants it
4. updating `docs/workflows/` evidence and checklists to reflect what is now supported, blocked, or deferred

Guardrail:

1. workflow validation should state pass, fail, or blocker explicitly and should not silently rely on stale milestone-closeout language

### 5.7 Slice 7: documentation truth and coverage pass

Goal:

1. make the documentation set reflect the current codebase truth and support stronger operator and maintainer use

Scope:

1. reconcile canonical planning docs with actual implementation status
2. reconcile `docs/workflows/` with current supported operator truth and validation evidence
3. reconcile `docs/user_guides/` with current promoted browser surfaces and supported workflows
4. reconcile `docs/technical_guides/` where technical seams, invariants, or verification guidance have drifted
5. add new focused guides where operators or maintainers currently have missing documentation for implemented capabilities

Guardrail:

1. update workflow truth in `docs/workflows/` first, then update downstream user or technical guides that derive from it

## 6. Required outcomes

Milestone 14 is complete only when:

1. the implementation and documentation mismatches found in the review are resolved or explicitly narrowed with updated docs and tests in the same change
2. production-readiness verification coverage is materially stronger and justified against the current real risk surface
3. baseline accounting reports for trial balance, balance sheet, and income statement are available on the shared reporting seam with sufficient verification and documentation
4. the active workflow-validation backlog for the promoted runtime has explicit evidence or explicit blockers recorded
5. canonical planning docs, workflow docs, user guides, technical guides, and the top-level README no longer materially drift from the current codebase
6. the repository has an explicit bounded user-testing-ready posture for the currently supported runtime, with any remaining exclusions or blockers documented clearly enough that testers are not asked to discover them accidentally
7. the promoted shared shell has the corrected desktop layout contract for sidebar and contextual-tab placement, with verification sufficient to keep that layout from regressing silently
8. adjacent production-shape gaps found during Milestone 14 implementation have either been addressed in bounded form or recorded explicitly rather than left as silent drift
9. the promoted navigation model now uses the intended grouped-directory pattern in the crowded areas selected for Milestone 14, with dedicated pages for concrete reports, lists, master-data creation, or workflow screens
10. `North Harbor Works` has a usable minimum demo-data baseline, including a standard chart of accounts and essential master data for realistic bounded testing
11. the post-Milestone-14 path is explicit: extensive user testing comes next, and Milestone 15 data-exchange implementation remains future work after that testing period rather than the immediate next implementation step

## 7. Verification

For Milestone 14 work:

1. run focused frontend checks and tests for edited Svelte surfaces
2. run focused Go package or integration tests for edited backend seams
3. run the canonical repository verification command before closing implementation slices unless an explicit blocker is recorded
4. use Playwright for real browser validation when the open risk is workflow-critical rendered behavior on `/app`
5. use `cmd/verify-agent -database-url "$DATABASE_URL" -approval-flow` when the browser proof needs exact seeded continuity on the same backend
6. record workflow-validation evidence in `docs/workflows/` in the same change that updates implementation status
7. for accounting-reporting additions, run focused backend verification for report composition and focused browser or route verification for the promoted report surfaces
8. for navigation and information-architecture changes, run focused route, navigation, and real-browser verification on the restructured grouped-directory and dedicated-page flows
9. for demo-data seeding changes, verify both the backend seed path and the visible runtime surfaces that depend on the seeded chart of accounts and minimum master data

## 8. Documentation sync

When Milestone 14 changes:

1. update `new_app_tracker_v2.md` with active status, completed work, and next steps
2. update `new_app_execution_plan_v2.md` when execution order or milestone slices move materially
3. update `docs/workflows/` first when supported workflow truth or validation evidence changes
4. then update `docs/user_guides/` and any affected `docs/technical_guides/`
5. update accounting and reporting documentation when baseline financial statements land or report behavior changes materially
6. update `README.md` when setup, verification workflow, or promoted application behavior changes materially
7. update navigation and admin/accounting usage guides when contextual-tab structure or destination-page behavior changes materially
8. document the demo-data baseline and any bootstrap or seed expectations when `North Harbor Works` becomes a supported user-testing entity

## 9. Current queue position

Milestone 14 is now the active next milestone.

Recommended order:

1. start with the bounded review-gap corrective pass on inbound-request lifecycle and documentation truth
2. include the bounded shared-shell desktop layout correction in that first corrective pass rather than leaving it outside the active milestone
3. then apply the grouped-directory and dedicated-destination navigation model in the most crowded promoted areas, starting with Admin and then Accounting
4. then land baseline accounting reports and close the most obvious adjacent reporting gaps exposed by that work
5. then seed `North Harbor Works` with the minimum realistic chart of accounts and master-data baseline needed for those reports and for bounded user testing
6. then promote the highest-value production-readiness test additions exposed by the corrective work and by focused codebase review
7. then execute the deferred live workflow-validation backlog on the corrected promoted runtime
8. then close with the broader documentation truth and coverage pass needed to make the repository defensible for extensive user testing
9. then move into extensive user testing before promoting Milestone 15 from future candidate to active implementation work
