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

For one deeper exact request -> proposal -> approval -> document continuity pass on the shared session plus `/api/...` seam:

```bash
set -a; source .env; set +a; go run ./cmd/verify-agent -approval-flow
```

For one deeper exact request -> proposal -> approval -> document -> accounting continuity pass on the same backend the served app is using:

```bash
set -a; source .env; set +a; go run ./cmd/verify-agent -database-url "$DATABASE_URL" -approval-flow
```

For focused live provider integration in the app seam:

```bash
set -a; source .env; set +a; go test -tags integration -count=1 ./internal/app -run TestOpenAIAgentProcessorLiveIntegration -v
```

These are not interchangeable. The build verifies compilation, the test command verifies the DB-backed suite, `cmd/verify-agent` exercises the real provider seam, `cmd/verify-agent -approval-flow` confirms exact request -> proposal -> approval -> document continuity, and `cmd/verify-agent -database-url "$DATABASE_URL" -approval-flow` is the canonical path when the browser proof must continue through one exact posted journal-entry route on the same backend the served app is using.

`cmd/verify-agent` needs both a database URL and OpenAI credentials:

1. `TEST_DATABASE_URL` or `DATABASE_URL`
2. `OPENAI_API_KEY`
3. `OPENAI_MODEL`

Important backend-target rule learned during the 2026-04-10 browser closeout:

1. `cmd/verify-agent` prefers `TEST_DATABASE_URL` over `DATABASE_URL` when no explicit `-database-url` flag is passed
2. `cmd/app` serves against its normal runtime database, so browser proof and verify-agent proof can silently diverge if the seed command is allowed to default to the test database
3. when Playwright or manual browser review must validate seeded continuity on the served app, pass `-database-url "$DATABASE_URL"` explicitly or otherwise prove that both commands are hitting the same backend

Repository rule for workflow-critical AI validation:

1. if the workflow under test depends on real provider behavior and `.env` supplies valid OpenAI credentials, run the live OpenAI-backed verification path as part of closeout
2. do not treat mock-only, fake-provider, or offline verification as sufficient production-shape evidence for those workflow-critical paths
3. document an explicit blocker if live OpenAI-backed verification cannot be run

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

For workflow-critical browser closeout on the served Svelte runtime:

1. prefer Playwright when available
2. if the served Go app embeds `internal/app/web_dist`, rebuild the frontend artifact and restart the app before diagnosing apparent browser regressions
3. if the continuity proof uses a dedicated verification org instead of the default admin org, log in with the exact credentials emitted by the seed command instead of assuming one long-lived local admin actor can see the same records
4. prefer route assertions based on stable headings, bounded primary actions, exact ids, and review-link contracts over brittle text markers that are likely to drift during UI copy cleanup
5. record the exact seed command, org slug, user email, request reference, and continuity ids used for the pass so the proof can be repeated without rediscovery

If the workflow depends on AI-provider behavior, validate it against the actual OpenAI API when `.env` provides the required credentials instead of relying only on mocked or offline provider paths.

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

## 9. Local test DB versus cloud production DB

Using a local disposable PostgreSQL instance for `TEST_DATABASE_URL` and a cloud-hosted PostgreSQL instance for production is the preferred repository posture.

Those environments serve different purposes:

1. the local test DB optimizes for speed, repeatability, destructive reset safety, and isolated verification
2. the production DB optimizes for durability, backups, high availability, monitoring, managed operations, and controlled access

The goal is not to make test and production share the same hosting model. The goal is to keep them aligned on the database engine and on the behaviors that matter for correctness.

For this repository, that means:

1. keep PostgreSQL as the engine in both test and production paths
2. do not substitute SQLite, in-memory fakes, or other engines for the canonical DB-backed suite
3. keep migrations, constraints, transaction behavior, and document or workflow invariants exercised against PostgreSQL in test

The local disposable test DB is still the right default even when production runs in the cloud, because the canonical suite is serialized, reset-heavy, and meant to be run frequently by contributors.

The main risks are configuration drift and environment drift, not the fact that one database is local and the other is cloud-hosted.

Watch for drift in:

1. PostgreSQL major version
2. enabled extensions
3. timezone, collation, and text-search assumptions
4. connection settings, SSL requirements, and pooling behavior
5. migration behavior and DDL permissions
6. latency and concurrency characteristics that only appear in a remotely hosted deployment

The recommended operating model is:

1. run the canonical full DB-backed suite against a local disposable PostgreSQL test database
2. keep the local test DB disposable so migrations, resets, and failure recovery stay cheap
3. verify migrations and a smaller production-parity smoke slice against the configured development or staging environment when persistence behavior, deployment shape, or production-sensitive DB settings change materially
4. do not point the canonical local suite at the main production `DATABASE_URL`

If production readiness depends on a cloud-specific behavior such as SSL enforcement, pooled-connection limits, failover posture, or materially higher network latency, add a focused verification step for that behavior instead of pushing the full day-to-day suite onto a remote shared database.

For the broader release and deployment checklist, including repo-specific production-parity follow-up work, see [`15_production_readiness_and_release_checklist.md`](./15_production_readiness_and_release_checklist.md).

## 10. What to keep in mind during review

The hardest-to-see failures in this repository are usually:

1. state drift between services and reporting
2. auth or role-check mismatches
3. queue claim races
4. browser and API contract divergence
5. workflow regressions hidden behind a still-compiling build

That is why integration tests and live seam checks matter here.

## 11. Collaborating on testing work

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
