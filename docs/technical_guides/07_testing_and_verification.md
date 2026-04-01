# Testing And Verification

Date: 2026-04-01
Status: Active technical guide
Purpose: explain the verification strategy for `workflow_app`, especially for workflow-critical slices that cannot be validated by unit tests alone.

## 1. Testing priority

The repository is persistence-heavy and workflow-driven, so testing should focus on real invariants:

1. database constraints
2. service-layer ownership boundaries
3. transaction safety
4. authorization
5. workflow continuity
6. browser and API contract consistency

Coverage matters, but correctness matters more than raw line counts.

## 2. Best test shape by change type

Use these defaults:

1. pure logic changes: unit tests
2. database or transaction changes: integration tests
3. cross-module workflow changes: integration tests plus focused flow verification
4. browser or operator-flow changes: HTTP integration tests and, when needed, live browser review

This repository is not primarily a pure-library codebase. It is a persistence-heavy business system. That means testing should not over-focus on isolated units while under-testing the service and database paths that actually own correctness.

## 3. Canonical verification commands

For code and persistence changes, the standard commands are:

```bash
go build ./cmd/... ./internal/...
set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...
```

For live AI-provider verification when credentials are available:

```bash
set -a; source .env; set +a; go run ./cmd/verify-agent
```

These are not interchangeable. The build verifies compilation, the test command verifies the DB-backed suite, and the verify-agent command exercises the live provider seam.

`cmd/verify-agent` needs both a database URL and OpenAI credentials:

1. `TEST_DATABASE_URL` or `DATABASE_URL`
2. `OPENAI_API_KEY`
3. `OPENAI_MODEL`

For focused reruns, use repository-shaped commands rather than ad hoc shortcuts:

```bash
set -a; source .env; set +a; GOCACHE=/tmp/go-build go test ./path/to/package -count=1
go test -race ./path/to/package
go test -shuffle=on ./path/to/package
go test -count=1 ./path/to/package
git diff --check
```

Use the `.env`-loaded form for DB-backed packages. Do not treat a bare `go test ./path/to/package` as the normal path for DB-backed verification in this repository.

Performance, benchmark, fuzz, or full-repo `-race` runs are opt-in only. Use them when the active change or a concrete defect justifies the extra cost; they are not current repository-wide default checks.

## 4. Why `-p 1` matters

The shared DB-backed suite uses a disposable test database and advisory-lock coordination. Running the suite with package-level parallelism can cause lock contention.

That is why the repository uses a serialized canonical test command for full verification.

## 5. What to test first

When changing a workflow-critical slice, test in this order:

1. the owner package
2. the orchestration path in `internal/app`
3. the relevant reporting read model
4. the browser or API integration path

This order catches the cheapest failures early while still validating the real workflow path.

## 6. Example test boundaries

Good repository tests usually check:

1. request state transitions
2. queue claim behavior
3. approval conflict handling
4. exact `REQ-...` continuity
5. attachment validation
6. auth rejection
7. provider-failure handling

That is more useful than testing private helper functions in isolation.

## 7. Workflow-critical validation

Some app behaviors need more than package tests.

Examples:

1. login then submit then queue then process
2. save draft then amend then queue
3. processed proposal then approval request then approval decision
4. browser detail pages that must preserve exact request continuity

For those cases, use the workflow-reference docs in `docs/workflows/` and validate the real seam instead of assuming the service layer is enough.

For each workflow, assert boundary by boundary:

1. request persistence and lifecycle transitions
2. AI run, step, artifact, recommendation, and delegation persistence where expected
3. approval creation and decision behavior where expected
4. downstream review visibility through `/api/review/...` and `/app/...`
5. exact continuity across linked review pages and upstream provenance

Do not treat broad exploratory manual testing without a checklist as the default approach.

## 8. Failure discipline

If a verification command fails:

1. investigate the cause
2. fix the issue or document the blocker
3. rerun the relevant verification

Do not treat a failed verification run as merely informational.

If a failure is caused by using a non-standard command path, rerun verification with the documented repository command before treating it as a product defect.

If a DB-backed verification command fails because the sandbox cannot reach the configured test database, rerun the documented `.env`-loaded repository command with the required approval path before treating the failure as a product defect.

If DB-backed verification appears hung, check for stale or overlapping sessions holding the disposable advisory lock before treating the symptom as a product defect. If that materially affects validation, document the blocker and cleanup in the canonical planning docs.

Prefer a local disposable PostgreSQL instance for `TEST_DATABASE_URL`. The DB-backed suite is serialized, reset-heavy, and materially faster and more reliable against a local test database than against a remote shared database.

The shared disposable advisory lock should be treated as setup coordination only. Hold it around destructive setup work such as migration and reset, not for the full lifetime of each DB-backed test, or interrupted runs can leave stale holder sessions that poison later suite attempts.

If the suite times out and the diagnostics show an idle session still running `SELECT pg_try_advisory_lock($1)`, terminate the stale backend on the disposable test database, rerun migrations if needed, and then rerun the canonical suite on the local test DB before treating the failure as an application defect.

If migrations or persistence behavior change, verify against the configured development and test databases unless an explicit blocker is documented.

While the application remains pre-production, it is acceptable to drop and recreate the configured test database to recover from schema drift, failed migration experiments, or other disposable development-state issues. This reset rule applies only to the configured test database, not to the application or development database.

## 9. What to keep in mind during review

The hardest-to-see failures in this repository are usually:

1. state drift between services and reporting
2. auth or role-check mismatches
3. queue claim races
4. browser and API contract divergence
5. workflow regressions hidden behind a still-compiling build

That is why integration tests and live seam checks matter here.

## 10. Collaborating on testing work

Codex is strongest when the testing target is a real business invariant rather than just line coverage.

The most useful user inputs are:

1. the business rule that must hold
2. the failure or regression you are worried about
3. important edge cases you already know
4. whether enforcement belongs in the database, service layer, workflow layer, or UI or API contract
5. whether the case is normal-path, error-path, authorization-path, concurrency-path, or migration-path

Useful requests include:

1. add regression coverage for this bug
2. write integration tests for this new service behavior
3. review these tests for gaps and flakiness
4. add authorization coverage for this endpoint or service
5. verify whether the DB constraints and tests match the business rule

The strongest collaboration pattern is:

1. the user explains the business rule or risk
2. Codex identifies the correct test boundary
3. Codex writes or updates the tests
4. Codex runs the appropriate verification
5. the user reviews whether the business meaning is captured correctly
