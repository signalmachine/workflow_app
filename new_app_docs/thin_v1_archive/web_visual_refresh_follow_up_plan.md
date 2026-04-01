# workflow_app Web Visual Refresh Follow-Up Plan

Date: 2026-03-30
Status: Implemented in code on 2026-03-30; bounded browser-review evidence still pending
Purpose: capture the concrete follow-up work discovered during code review of the landed web visual-refresh implementation so those issues are fixed as one bounded implementation step before the broader browser-surface restructuring continues.

## 1. Why this plan exists

The bounded visual-refresh implementation improved the promoted web layer materially, but the implementation review found two concrete issues that should not be left as silent residuals:

1. the canonical docs and validation checklist now describe `/app/login` as an exact scoped review surface, but the current implementation still renders the sign-in UI only through unauthenticated `GET /app` while `/app/login` remains POST-only
2. the shared template now applies a global `table` minimum width, but only some refreshed pages were wrapped in horizontal-overflow containers, so several non-targeted pages likely now regress on narrow-width layouts because they inherit the new shared rule without the matching containment

These are both bounded browser-layer correctness or continuity issues caused by the same shared-template pass, so they should be fixed together as one cleanup slice rather than deferred into the later larger browser-surface restructuring.

## 2. Planning decision

Current decision:

1. promote one bounded follow-up slice immediately after the visual-refresh implementation review
2. keep this slice limited to route or template correctness and narrow-width continuity caused by the refresh
3. do not mix this slice with the later dashboard or intake or feed or chat restructuring from `operator_communication_and_intake_surfaces_plan.md`
4. do not reopen visual-direction exploration, copywriting churn, or broader workflow redesign under the cover of fixing these review findings

## 3. Exact issues to fix

### 3.1 Login-route and plan alignment

The implementation and canonical docs must agree on the login surface.

Required outcome:

1. either `GET /app/login` must render the intended sign-in surface through the shared unauthenticated shell, or the canonical docs must be narrowed back to the true implemented route if that is the deliberate product decision
2. the implementation should prefer the first option unless implementation review finds a concrete reason to preserve login-only-through-`/app`
3. automated web coverage should assert the chosen GET behavior explicitly so this route drift does not recur

Decision rule:

1. if `GET /app/login` is added, it should render the same sign-in experience and preserve the existing POST login action
2. if the route is intentionally not added, update every affected canonical doc in the same change so the scoped page list and validation checklist no longer claim that page exists

### 3.2 Shared-template table containment

The shared `table` minimum-width rule must not create new narrow-width overflow regressions on inherited pages.

Required outcome:

1. every page that inherits the shared `table` minimum width and can overflow on narrow screens should either be wrapped in the matching overflow container or be restyled so the minimum width is no longer globally unsafe
2. this review should cover at least approvals, proposal detail, document detail, accounting detail, control-account detail, tax-summary detail, inventory detail and related inventory detail pages, work-order detail, and audit detail because they all render tables through the same shared template
3. the fix should stay cheap and shared-first: prefer one coherent template rule or wrapper strategy over page-by-page visual redesign

## 4. Implementation order

Implement in this order:

1. decide and land the login-route alignment fix
2. audit the shared template for all remaining bare tables affected by the global minimum-width rule
3. apply the cheapest coherent containment fix across those pages
4. add or update focused tests that prove the login route behavior and guard at least one representative previously-unwrapped detail page against silent template regression
5. rerun bounded browser review on the exact refreshed pages plus the representative detail pages affected by the shared table rule

## 5. Guardrails

Do not:

1. fold this corrective slice into the later dedicated intake-page, operations-feed, or chat work
2. reopen the visual-refresh slice as an open-ended polish bucket
3. introduce a separate frontend stack, design-system buildout, or larger component rewrite
4. use this slice as a pretext for backend contract changes that are unrelated to the route or narrow-width issues above

## 6. Validation expectations

Required verification for this slice:

1. `gopls` diagnostics on edited Go files
2. focused web tests covering the chosen login-route behavior and the shared-template containment fix
3. `go build ./cmd/... ./internal/...`
4. `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`
5. bounded browser review on desktop and narrow-width layouts for `/app`, the sign-in surface, `/app/inbound-requests/{request_reference_or_id}`, `/app/review/inbound-requests`, `/app/review/proposals`, and at least one representative formerly-unwrapped detail page from the shared template

## 7. Sequencing impact

This corrective slice should land before `operator_communication_and_intake_surfaces_plan.md`.

Reason:

1. the review findings are direct regressions or drift introduced by the landed refresh itself
2. leaving them open would weaken the signal from the next browser-led implementation and validation work
3. the corrective scope is bounded enough that it should not materially delay the later browser-surface restructuring

## 8. Implementation result

Implementation landed on 2026-03-30:

1. `GET /app/login` now renders the same shared sign-in surface used for unauthenticated browser entry while preserving the existing `POST /app/login` action
2. authenticated requests to `GET /app/login` now redirect back to `/app` rather than showing a redundant sign-in screen
3. the shared table minimum-width rule now applies only to tables inside `.table-wrap`, so previously unwrapped review and detail pages no longer inherit unsafe narrow-width overflow by default
4. focused `internal/app` HTTP tests now cover the new login GET behavior and the shared wrapped-table CSS rule

Remaining closeout:

1. bounded manual browser review is still required on desktop and a narrow-width viewport for the pages listed in section 6
