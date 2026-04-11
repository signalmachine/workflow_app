# Demo Entity: North Harbor Works

Date: 2026-04-11
Status: Active
Purpose: document the durable setup contract for the North Harbor Works demo entity so a fresh database can be bootstrapped into a useful bounded-testing state without ad hoc SQL.

## 1. Role

`North Harbor Works` is the default demo organization for local browser use, bounded user testing, and report review.

It is not a production tenant template and it is not created by migrations. It is seeded explicitly by the bootstrap command after migrations have created the database schema.

The default identity values are:

1. org name `North Harbor Works`
2. org slug `north-harbor`
3. admin email `admin@northharbor.local`
4. admin display name `North Harbor Admin`

## 2. Create On A Fresh Database

From the repository root:

```bash
go run ./cmd/migrate
go run ./cmd/bootstrap-admin -password 'choose-a-strong-password'
go run ./cmd/app
```

Then sign in at `/app/login` with:

1. org slug `north-harbor`
2. email `admin@northharbor.local`
3. the password passed to `cmd/bootstrap-admin`
4. device label `browser`

The runnable commands auto-load `.env` from the repository root when it exists. Pass `-database-url` only when intentionally targeting a different database.

## 3. What Gets Seeded

By default, `cmd/bootstrap-admin` calls the backend-owned setup seed in `internal/setup`.

The current minimum baseline includes:

1. a standard chart of accounts for cash, receivables, inventory, GST control accounts, payables, equity, service revenue, parts and materials revenue, cost of goods sold, subcontractor expense, inventory adjustments, and operating expense
2. GST 18% sales and purchase tax codes wired to the seeded GST control accounts
3. one open accounting period, `FY2026-27`, from `2026-04-01` through `2027-03-31`
4. sample customer and vendor parties with primary contacts
5. starter inventory items for service materials, resale stock, traceable equipment, and direct-expense consumables
6. starter inventory locations for the main warehouse, field van, adjustment bin, active job site, and installed-equipment base

The seed is intentionally small. It exists to make admin, list, report, and review surfaces usable after sign-in without turning the demo org into a broad fake ERP dataset.

## 4. Idempotency

The seed is idempotent.

Rerunning:

```bash
go run ./cmd/bootstrap-admin -password 'new-password'
```

will update or ensure the friendly admin login, rotate the password to the new value, and restore any missing baseline records without duplicating records that already exist.

Use this after:

1. creating a fresh main database
2. resetting a local database
3. rotating the local demo admin password
4. restoring a partially seeded local environment

## 5. Login-Only Bootstrap

Use this when a database should get only the friendly org and admin login without demo master data:

```bash
go run ./cmd/bootstrap-admin -password 'choose-a-strong-password' -seed-demo-baseline=false
```

That path is useful for targeted testing where demo master data would make fixture state less obvious.

## 6. Ownership

The seed belongs to the backend setup layer, currently `internal/setup`, and is invoked by `cmd/bootstrap-admin`.

Keep this boundary intact:

1. do not seed the demo org from migrations
2. do not hand-maintain the demo baseline with ad hoc SQL snippets in docs
3. keep business data owned by the relevant database schemas and service conventions
4. keep the command explicit so production and staging environments are not silently populated at app startup

If the demo baseline changes, update this guide, the bootstrap user guide, and the active `new_app_docs` tracker in the same change.
