# Running The Application

Date: 2026-03-30
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

## 2. First-run setup

From the repository root:

```bash
go run ./cmd/migrate
go run ./cmd/bootstrap-admin -password 'choose-a-strong-password'
go run ./cmd/app
```

The runnable commands auto-load `.env` from the repository root when it exists. You do not need to `source .env` first for these commands unless you want shell-level overrides.

Migration rule:

1. run `go run ./cmd/migrate` on first setup for a database
2. run it again after pulling code that adds new migrations
3. run it again when you switch to a fresh or reset main database
4. do not treat it as a command you must run before every normal app restart when the database is already up to date

## 3. Default bootstrap admin values

If you run `go run ./cmd/bootstrap-admin -password 'choose-a-strong-password'` without extra flags, the command ensures this login exists in the main database:

1. org name `North Harbor Works`
2. org slug `north-harbor`
3. admin email `admin@northharbor.local`
4. admin display name `North Harbor Admin`

Field meaning:

1. org slug is the short login-facing identifier, for example `north-harbor`
2. org name is the human-readable display name, for example `North Harbor Works`
3. org ID is a separate internal database UUID and is not what you type into the sign-in form

What `bootstrap` means here:

1. create the minimum first-run records needed so the app is usable
2. ensure an organization record exists
3. ensure an admin user record exists
4. ensure that user has an active admin membership in that organization
5. hash and store the password so browser sign-in can work immediately

This does not create a separate application instance. It creates one friendly organization and one friendly admin login inside the main application database.

You can override the defaults:

```bash
go run ./cmd/bootstrap-admin \
  -org-name "Harborline Services" \
  -org-slug harborline \
  -email admin@harborline.local \
  -display-name "Harborline Admin" \
  -password 'choose-a-strong-password'
```

The bootstrap command is idempotent:

1. it creates the org if missing
2. it creates the user if missing
3. it creates or reactivates the org membership if missing
4. it keeps the membership role as `admin`
5. it updates the password to the value you passed

## 4. Browser sign-in

Open:

```text
http://127.0.0.1:8080/app
```

If you used the default bootstrap values, sign in with:

1. Org slug: `north-harbor`
2. User email: `admin@northharbor.local`
3. Password: the password you passed to `cmd/bootstrap-admin`
4. Device label: `browser`

If you sign in with those defaults, you are signing into the organization `North Harbor Works`. The application is multi-tenant at the org level, so the organization is the active tenant context for that browser session.

The sign-in page expects a real org slug and a real email address. A short value like `admin` in the email field will not work unless there is an actual user with that email in the database.

## 5. Recommended verification commands

Use these commands after setup:

```bash
go build ./cmd/... ./internal/...
set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...
```

The first checks that the runnable workspace builds cleanly. The second verifies the code against the configured test database rather than the main application database.

## 6. Troubleshooting

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
