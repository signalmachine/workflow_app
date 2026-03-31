# Testing And Verification

Date: 2026-03-31
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

## 8. Failure discipline

If a verification command fails:

1. investigate the cause
2. fix the issue or document the blocker
3. rerun the relevant verification

Do not treat a failed verification run as merely informational.

## 9. What to keep in mind during review

The hardest-to-see failures in this repository are usually:

1. state drift between services and reporting
2. auth or role-check mismatches
3. queue claim races
4. browser and API contract divergence
5. workflow regressions hidden behind a still-compiling build

That is why integration tests and live seam checks matter here.
