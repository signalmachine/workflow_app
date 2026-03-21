# Testing Guide

Date: 2026-03-20
Status: Active working guide
Purpose: explain how testing should be approached in `workflow_app`, what Codex can handle effectively, and what input from the user makes testing faster and more accurate.

## 1. Testing stance

`workflow_app` should follow industry-standard testing best practices unless there is a concrete reason to deviate.

In this repository, testing should prioritize:

1. business correctness before superficial coverage
2. database and schema invariants where the database is the real safety boundary
3. service-layer behavior at module boundaries
4. transaction safety, auditability, and idempotency
5. realistic integration coverage for the document -> approval -> posting -> inventory/execution chain

This repository is not primarily a pure-library codebase. It is a persistence-heavy business system. That means testing should not over-focus on isolated units while under-testing the service and database paths that actually own correctness.

## 2. What Codex can do well in testing

Codex is effective at:

1. designing tests for service-layer business rules
2. writing Go unit tests and integration tests
3. extending existing package-level integration test patterns
4. building realistic test fixtures for orgs, users, sessions, documents, approvals, inventory, accounting, and work orders
5. testing lifecycle transitions such as draft -> submitted -> approved -> posted or reversed
6. testing authorization boundaries and role-sensitive behavior
7. testing database-enforced invariants such as uniqueness, append-only constraints, balance checks, and status restrictions
8. testing idempotency, duplicate-write prevention, and retry-safe behavior
9. testing migration-backed persistence behavior
10. debugging failing tests and tightening weak or flaky setup
11. identifying missing edge cases and regression risks
12. aligning tests with module ownership and repository architecture rules

In practice, Codex is strongest when the testing target is a real business invariant rather than just line coverage.

## 3. What Codex expects from the user

Codex does not require deep Go testing expertise from the user, but testing becomes more effective when the user provides clear domain guidance.

The most useful inputs are:

1. the business rule that must hold
2. the failure or regression you are worried about
3. any edge cases you already know are important
4. whether the rule is expected to be enforced in the database, service layer, workflow layer, or UI/API contract
5. whether a case is normal-path, error-path, authorization-path, concurrency-path, or migration-path
6. any real-world examples or expected outcomes that clarify the intended behavior

Good guidance from the user usually sounds like:

1. "an approved inventory issue linked to a work order must create pending execution linkage but must not write accounting rows directly"
2. "this operation should be idempotent on retry"
3. "operators should be allowed here but approvers should not"
4. "a posted journal entry must never be mutated in place"

What helps most is domain clarity, not Go test syntax.

## 4. What the user does not need to know deeply

The user does not need deep expertise in:

1. Go test package structure
2. `testing.T` usage details
3. fixture helper design
4. table-driven test syntax
5. transaction setup mechanics
6. migration-backed test database wiring
7. low-level assertion style choices

Codex can usually translate business requirements into the appropriate test structure.

## 5. What makes testing effective in this repository

Testing is most effective when it follows these rules:

1. test the real ownership boundary
2. test the smallest meaningful public behavior, not private implementation trivia
3. prefer integration tests when correctness depends on SQL constraints, transactions, migrations, audit writes, or cross-module persistence
4. prefer unit tests when logic is local, deterministic, and does not depend on database truth
5. cover both success paths and failure paths
6. cover authorization and invalid-state transitions where applicable
7. make assertions specific enough to catch regressions in status, counts, links, and side effects
8. keep fixtures realistic but not bloated
9. avoid fragile tests that assert irrelevant ordering or incidental formatting
10. test behaviors that matter to future refactors, not just what was easy to assert

## 6. Recommended testing hierarchy

Use this rough priority order unless there is a special reason not to:

1. migration and schema safety for new persistence behavior
2. service integration tests for module-owned write paths
3. authorization and lifecycle tests around shared workflow and document boundaries
4. narrowly targeted unit tests for isolated logic that would otherwise be awkward to cover through integration
5. repo-wide build and test verification after meaningful changes

For this repository, service integration tests are often the highest-value test type.

## 7. When Codex should prefer integration tests

Prefer integration tests when a change involves:

1. SQL constraints
2. triggers or append-only enforcement
3. migrations
4. cross-table invariants
5. document lifecycle state
6. approval state
7. posting rules
8. stock derivation
9. work-order, inventory, accounting, or workflow linkage
10. audit writes that must succeed or fail transactionally with business actions

These are core repository concerns, so testing them only with mocks is usually not enough.

## 8. When unit tests are enough

Unit tests are often enough when:

1. logic is pure or nearly pure
2. there is no meaningful database invariant involved
3. the behavior is computational or formatting-oriented rather than persistence-oriented
4. the test is guarding a dense branch structure that would be expensive to reach repeatedly through full integration setup

Even then, unit tests should still reflect real business expectations rather than arbitrary implementation details.

## 9. How to ask Codex for testing help effectively

Useful requests include:

1. "add regression coverage for this bug"
2. "write integration tests for this new service behavior"
3. "what edge cases are missing in this module"
4. "review these tests for gaps and flakiness"
5. "add authorization coverage for this endpoint or service"
6. "verify whether the DB constraints and tests match the business rule"
7. "convert this weak test into one that checks the real invariant"

If the user is unsure what test is needed, it is enough to describe the intended behavior and the risk.

## 10. Practical collaboration model

The strongest collaboration pattern is:

1. the user explains the business rule or risk
2. Codex identifies the correct test boundary
3. Codex writes or updates the tests
4. Codex runs the appropriate verification
5. the user reviews whether the business meaning is captured correctly

This division of labor works well because domain intent usually matters more than framework syntax.

## 11. Known limits

Codex is effective at backend and persistence-heavy testing, but some limits still matter:

1. if external dependencies are unavailable, verification may require mocks, fixtures, or explicit approval to access configured systems
2. UI-heavy exploratory testing is weaker than backend/service testing unless the interface and expected behaviors are tightly specified
3. if the intended behavior is ambiguous, Codex may produce technically valid tests for the wrong rule
4. weak seams in the codebase may need to be improved before tests can be clean and maintainable

These are usually solvable, but they are real constraints.

## 12. Repository-specific verification expectations

For implementation work in this repository, the normal verification path is:

1. run `gopls` diagnostics on edited Go files
2. run `go build ./...`
3. run targeted package tests when developing a slice
4. run `set -a; source .env; set +a; go test -p 1 ./...` when code or persistence behavior changed because the shared test database uses an advisory lock and package-parallel full-suite runs can contend on it
5. if migrations changed, verify that `go run ./cmd/migrate` applies cleanly against the configured development database
6. document any blocker explicitly if full verification cannot run

## 13. Bottom line

The user does not need to know deep Go testing mechanics to get strong results from Codex.

What Codex needs most is:

1. clear business intent
2. known risks or edge cases
3. correction when domain expectations are wrong or incomplete

With that guidance, Codex can usually handle the testing design, implementation, and verification work effectively.
