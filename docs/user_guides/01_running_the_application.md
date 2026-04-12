# Running The Application

Date: 2026-04-12
Status: Active
Purpose: explain the recommended local run path for the main application database, browser sign-in, and the separate test-database workflow.

## 1. Use the right database for the right job

`workflow_app` uses two different database settings locally:

1. `DATABASE_URL` for the main application run path
2. `TEST_DATABASE_URL` for the automated test suite

Use `DATABASE_URL` when you are:

1. running migrations for the app
2. starting `cmd/app`
3. creating or rotating the main browser admin login
4. using the browser UI manually

Use `TEST_DATABASE_URL` when you are:

1. running the canonical Go test suite
2. running focused DB-backed test commands
3. investigating test-only failures

Do not point the canonical test command at the main application database.

Example:

Use `DATABASE_URL` for a manual North Harbor Works browser session where you want seeded demo records to persist across restarts. Use `TEST_DATABASE_URL` for `go test` runs that reset or mutate test data.

## 2. First-run setup

From the repository root:

```bash
go run ./cmd/migrate
go run ./cmd/bootstrap-admin -password 'choose-a-strong-password'
go run ./cmd/app
```

The runnable commands auto-load `.env` from the repository root when it exists. You do not need to `source .env` first for these commands unless you want shell-level overrides.

If you need the browser-login defaults or a friendlier local admin login, see [`02_browser_sign_in_and_admin_bootstrap.md`](./02_browser_sign_in_and_admin_bootstrap.md).

The default bootstrap step also seeds the minimum North Harbor Works demo baseline for bounded testing. Pass `-seed-demo-baseline=false` if you need only the admin login records.

Example:

For a normal local user-testing run, migrate the database, bootstrap North Harbor Works with a real password, start `cmd/app`, then sign in at `/app/login`. You should see the seeded chart of accounts, parties, inventory items, and locations without manually creating them first.

For the durable technical setup contract, see [`Demo Entity: North Harbor Works`](../technical_guides/16_demo_entity_north_harbor_works.md).

Migration rule:

1. run `go run ./cmd/migrate` on first setup for a database
2. run it again after pulling code that adds new migrations
3. run it again when you switch to a fresh or reset main database
4. do not treat it as a command you must run before every normal app restart when the database is already up to date

## 3. Browser sign-in

Open:

```text
http://127.0.0.1:8080/app/login
```

You can also start from `/app` if you prefer the main app entry point. The unauthenticated browser flow should route you to the sign-in surface.

The sign-in page expects a real org slug and a real email address. A short value like `admin` in the email field will not work unless there is an actual user with that email in the database.

## 4. Recommended verification commands

Use these commands after setup:

```bash
go build ./cmd/... ./internal/...
set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...
```

The first checks that the runnable workspace builds cleanly. The second verifies the code against the configured test database rather than the main application database.

## 5. Troubleshooting

If `go run ./cmd/app` prints `DATABASE_URL is required`:

1. confirm `.env` contains `DATABASE_URL=...`
2. make sure you are running from the repository root so the command can auto-load `.env`
3. if you intentionally launch from somewhere else, export `DATABASE_URL` explicitly before running the command

If sign-in fails:

1. rerun `go run ./cmd/bootstrap-admin -password 'new-password'`
2. confirm you are using the matching org slug and full email address
3. confirm you are signing into the app started against `DATABASE_URL`, not a different database

If the app starts but later fails because expected tables or columns are missing:

1. run `go run ./cmd/migrate`
2. restart the app
3. if the database was newly created or reset, rerun `go run ./cmd/bootstrap-admin -password 'choose-a-strong-password'` before signing in
