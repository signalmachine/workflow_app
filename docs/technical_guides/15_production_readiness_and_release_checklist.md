# Production Readiness And Release Checklist

Date: 2026-04-01
Status: Active technical guide
Purpose: define the durable production-readiness checklist for `workflow_app`, including database parity checks, release gates, and repo-specific follow-up work that should be reviewed before production rollout or materially production-like staging validation.

## 1. Role of this guide

Use this guide when the question is no longer "does the local code build and pass the canonical suite?" and has become "is this slice ready to run safely in a production-like environment?"

This guide complements, but does not replace, the canonical day-to-day verification rules in [`07_testing_and_verification.md`](./07_testing_and_verification.md).

Use this guide for:

1. production rollout preparation
2. staging or production-parity validation planning
3. release signoff reviews
4. cloud-environment readiness checks
5. identifying parity gaps between local development and managed production infrastructure

## 2. Core principle

`workflow_app` should continue to use:

1. a local disposable PostgreSQL database for `TEST_DATABASE_URL` and the canonical DB-backed suite
2. a separate application database for `DATABASE_URL`
3. a cloud-hosted PostgreSQL deployment for production when that improves durability and operations

That split is correct. Production readiness depends on parity of engine behavior, migration safety, operational controls, and workflow correctness, not on forcing the everyday test suite to run against the production hosting model.

## 3. Release gate order

Before treating a slice as production-ready, review readiness in this order:

1. build and canonical DB-backed test verification
2. migration safety and schema parity
3. environment and secret readiness
4. auth, approval, and control-boundary validation
5. production-parity smoke verification
6. observability, backup, and rollback posture
7. explicit signoff on known residual risks

Do not collapse these into one informal "it worked locally" decision.

## 4. Database readiness checklist

Use this checklist whenever persistence behavior, schema, migrations, queue execution, approvals, or workflow-critical reporting changed.

### 4.1 Baseline rules

1. production and test both use PostgreSQL
2. the canonical full suite still runs against local disposable `TEST_DATABASE_URL`
3. the main production `DATABASE_URL` is never used as the target for the canonical destructive test suite
4. migrations are applied through the repository migration runner, not through ad hoc SQL drift

### 4.2 Production-parity checks

Confirm or document each of the following:

1. PostgreSQL major version parity between local development, test, staging, and production
2. required extension parity, if any are introduced later
3. SSL expectations for managed or cloud-hosted connections
4. connection pooling expectations and safe connection-count limits
5. migration-runner permissions in staging and production
6. timezone, collation, and text behavior that could affect ordering, filtering, or search
7. backup and restore expectations for production data
8. rollback or forward-fix posture when a migration cannot be reversed safely
9. cloud-latency or cross-network behavior that could affect startup, request timeouts, or queue processing

### 4.3 Repo-specific current state and missing checks

From the current repository inspection on 2026-04-01:

1. the app and migration entry points use PostgreSQL through `pgx` with `database/sql`
2. `.env.example` expects `sslmode=require` for both `DATABASE_URL` and `TEST_DATABASE_URL`
3. local developer usage currently prefers a local disposable `TEST_DATABASE_URL`
4. the checked-in local `.env` shape uses a cloud host for `DATABASE_URL` and `localhost` for `TEST_DATABASE_URL`

That means the local-versus-cloud split is already aligned with the intended posture.

The production-parity checks that are still missing or not yet captured as durable repository policy are:

1. a documented requirement to match PostgreSQL major versions across local, staging, and production
2. a documented checklist for extension parity, even if the current schema does not yet require extra extensions
3. a documented staging or production-parity smoke command set for migrations plus app startup against a managed database
4. a documented check for SSL enforcement and certificate expectations on managed database connections
5. a documented review of connection-pool sizing and idle or open-connection limits for `cmd/app`
6. a documented check for timezone, collation, and text-search assumptions in production
7. a documented backup, restore, and rollback expectation for production rollout
8. a documented latency and timeout review for cloud-hosted database behavior

Until those checks are either documented as intentionally unnecessary or validated explicitly, production readiness should be treated as incomplete.

## 5. Environment and secret checklist

Before rollout or staging validation:

1. confirm `DATABASE_URL` points to the intended environment
2. confirm `TEST_DATABASE_URL` still points to a disposable non-production database
3. confirm OpenAI credentials are present only when the live provider seam is intentionally enabled
4. confirm no local-only defaults or developer passwords are being treated as production values
5. confirm secrets are supplied through the intended environment mechanism rather than committed files
6. confirm listen address, proxy, and network-routing expectations are correct for deployment

## 6. Migration readiness checklist

Before applying a migration to a production-like environment:

1. confirm the migration is represented in the embedded migration runner
2. confirm the migration path has been tested on a fresh database
3. confirm the migration path has been tested on an already-populated database shape when that matters
4. confirm the service and reporting layers agree with the migrated schema
5. confirm reversibility or forward-fix expectations are documented
6. confirm rollout order does not leave the app in a half-compatible state longer than intended

If a migration is operationally risky, add an explicit rollback or break-glass note before rollout.

## 7. Workflow and control-boundary checklist

Before production signoff:

1. confirm request persistence and lifecycle transitions still hold end to end
2. confirm queue-oriented processing still works with the deployed environment characteristics
3. confirm approval creation, approval decision, and conflict behavior still hold
4. confirm browser and API review surfaces still reflect the same underlying truth
5. confirm auth behavior and role checks still match expected operator boundaries
6. confirm operator-visible continuity for exact `REQ-...` references and linked review pages

## 8. Production-parity smoke validation

The repository should prefer a smaller production-parity smoke slice over moving the full canonical suite onto a remote managed database.

A good smoke slice should include:

1. run migrations against the staging or production-like database
2. start `cmd/app` against that environment
3. confirm the app can connect, ping, and serve requests successfully
4. exercise one bounded workflow-critical path end to end
5. confirm the relevant review or reporting path reflects the resulting state

When the live provider seam is part of the rollout target, include the bounded `cmd/verify-agent` path only when the environment is intentionally configured for it.

## 9. Observability, backup, and rollback checklist

Before rollout:

1. confirm application logs are accessible for startup, auth, workflow, and migration failures
2. confirm the managed database has an understood backup posture
3. confirm restore expectations are known, even if full disaster-recovery automation is not yet built
4. confirm rollback means either safe reversal or explicit forward-fix, not an unstated hope
5. confirm operators know how to distinguish app defects from connectivity, credential, or migration issues

## 10. Signoff template

Before calling a slice production-ready, capture:

1. the exact code revision
2. the migration state applied
3. the environments checked
4. the commands run
5. the workflow or smoke checks completed
6. any open risks or deferred parity checks
7. the person or session making the readiness call

Do not mark production readiness complete without recording what was actually verified.
