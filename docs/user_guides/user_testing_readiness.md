# User Testing Readiness

Date: 2026-04-11
Status: Active
Purpose: describe the supported Milestone 14 runtime, setup path, workflow scope, and validation limits for bounded user testing.

## 1. Supported Runtime

Use the promoted Svelte browser runtime served by the Go app:

1. start from `/app/login` or `/app`
2. use the shared `/api/...` backend seam through the browser session
3. test against the `DATABASE_URL` application database, not `TEST_DATABASE_URL`
4. use the North Harbor Works demo baseline unless the test session needs a deliberately empty org

The retired Go-template browser path is not the active runtime.

## 2. Setup Path

From the repository root:

```bash
go run ./cmd/migrate
go run ./cmd/bootstrap-admin -password 'choose-a-strong-password'
go run ./cmd/app
```

Sign in with the default demo identity unless the session owner supplied a different bootstrap command:

1. org slug `north-harbor`
2. email `admin@northharbor.local`
3. the password passed to `cmd/bootstrap-admin`
4. device label `browser`

The default bootstrap command seeds the minimum North Harbor Works baseline: chart of accounts, GST tax codes, an open FY2026-27 accounting period, sample customer and vendor parties with contacts, starter inventory items, and starter locations.

## 3. Supported Test Focus

The current bounded user-testing pass should focus on:

1. browser sign-in and session continuity
2. inbound-request draft, edit, queue, cancel, amend, and delete lifecycle
3. operations processing through the process-next action
4. processed proposal review and exact drill-down
5. approval request, approval decision, document review, and upstream request continuity
6. failed-processing visibility from list and exact request detail
7. accounting review, including journal entries, control balances, tax summaries, trial balance, balance sheet, and income statement
8. inventory stock, movement, reconciliation, and work-order review continuity
9. Admin grouped directory pages, accounting setup, party setup, access maintenance, and inventory setup

Use `docs/workflows/end_to_end_validation_checklist.md` as the checklist when a session is validating workflow continuity rather than only browsing the app.

## 4. Current Limits

The handoff is for bounded user testing, not production release.

Known limits:

1. deeper live workflow-validation backlog execution is intentionally deferred into the user-testing period
2. Milestone 15 data exchange, CSV import, and CSV or Excel-compatible export are future work after user-testing findings are triaged
3. mobile-specific UX depth remains out of scope for the served web runtime
4. production-parity release checks in `docs/technical_guides/15_production_readiness_and_release_checklist.md` are still required before production rollout
5. OpenAI-backed processing requires valid `OPENAI_API_KEY` and `OPENAI_MODEL` in the environment

## 5. Evidence Baseline

Current pre-handoff evidence includes:

1. Milestone 13 real-browser closeout evidence recorded in `docs/workflows/workflow_validation_track.md`
2. focused frontend checks for request-detail lifecycle controls, grouped navigation, and accounting report destinations
3. focused backend and API integration coverage for inbound-request lifecycle mutations, provider failure visibility, approval and proposal org boundaries, accounting-report org boundaries, inventory and work-order org boundaries, and Admin exact-record org boundaries
4. canonical Go verification from the landed Milestone 14 production-readiness checkpoints

Record new user-testing pass, fail, blocker, or deferral notes in `docs/workflows/workflow_validation_track.md` before promoting follow-up implementation work.
