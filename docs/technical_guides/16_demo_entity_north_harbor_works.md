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

## 4. Seeded Chart Of Accounts

The current chart of accounts baseline is:

| Code | Name | Class | Control type | Direct posting |
| --- | --- | --- | --- | --- |
| `1000` | Cash and Bank | asset | none | yes |
| `1100` | Accounts Receivable | asset | receivable | no |
| `1200` | Inventory Asset | asset | none | yes |
| `1300` | GST Input Receivable | asset | gst_input | no |
| `2000` | Accounts Payable | liability | payable | no |
| `2100` | GST Output Payable | liability | gst_output | no |
| `3000` | Owner Equity | equity | none | yes |
| `4000` | Service Revenue | revenue | none | yes |
| `4100` | Parts and Materials Revenue | revenue | none | yes |
| `5000` | Cost of Goods Sold | expense | none | yes |
| `5100` | Subcontractor Expense | expense | none | yes |
| `5200` | Inventory Adjustments | expense | none | yes |
| `6000` | Operating Expense | expense | none | yes |

## 5. Seeded Tax Codes And Period

The current tax-code baseline is:

| Code | Name | Type | Rate | Receivable account | Payable account |
| --- | --- | --- | --- | --- | --- |
| `GST18-SALES` | GST 18% Sales | gst | 18% | | `2100` GST Output Payable |
| `GST18-PURCH` | GST 18% Purchases | gst | 18% | `1300` GST Input Receivable | |

The current accounting-period baseline is:

| Period code | Start | End | Status |
| --- | --- | --- | --- |
| `FY2026-27` | `2026-04-01` | `2027-03-31` | open |

## 6. Seeded Parties And Contacts

The current party and contact baseline is:

| Party code | Display name | Kind | Primary contact | Role | Email | Phone |
| --- | --- | --- | --- | --- | --- | --- |
| `CUST-ACME` | Acme Facilities | customer | Asha Rao | Facilities Manager | `asha.rao@acme.example` | `+91-80-5550-0101` |
| `CUST-METRO` | Metro Property Group | customer | Karan Mehta | Operations Lead | `karan.mehta@metro.example` | `+91-80-5550-0102` |
| `VEND-HARBOR` | Harbor Industrial Supply | vendor | Neha Iyer | Account Manager | `neha.iyer@harbor-supply.example` | `+91-80-5550-0201` |
| `VEND-POWER` | Powerline Electricals | vendor | Vikram Singh | Sales Desk | `sales@powerline.example` | `+91-80-5550-0202` |
| `VEND-SUBCO` | Reliable Field Services | vendor | Maya D'Souza | Dispatch Coordinator | `dispatch@reliable-field.example` | `+91-80-5550-0203` |

## 7. Seeded Inventory Master Data

The current inventory item baseline is:

| SKU | Name | Item role | Tracking mode |
| --- | --- | --- | --- |
| `SVC-MAT-FILTER` | Replacement filter kit | service_material | none |
| `SVC-MAT-SEAL` | Industrial sealant pack | service_material | none |
| `RES-PUMP-100` | Pump assembly | resale | none |
| `EQ-METER-200` | Field meter | traceable_equipment | serial |
| `EXP-CLEANUP` | Shop cleanup consumables | direct_expense_consumable | none |

The current inventory location baseline is:

| Code | Name | Location role |
| --- | --- | --- |
| `MAIN-WH` | Main warehouse | warehouse |
| `FIELD-VAN-1` | Field van 1 | van |
| `ADJ-BIN` | Inventory adjustment bin | adjustment |
| `JOB-SITE` | Active job site | site |
| `INSTALLED` | Installed equipment base | installed |

## 8. Idempotency

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

## 9. Login-Only Bootstrap

Use this when a database should get only the friendly org and admin login without demo master data:

```bash
go run ./cmd/bootstrap-admin -password 'choose-a-strong-password' -seed-demo-baseline=false
```

That path is useful for targeted testing where demo master data would make fixture state less obvious.

## 10. Ownership

The seed belongs to the backend setup layer, currently `internal/setup`, and is invoked by `cmd/bootstrap-admin`.

Keep this boundary intact:

1. do not seed the demo org from migrations
2. do not hand-maintain the demo baseline with ad hoc SQL snippets in docs
3. keep business data owned by the relevant database schemas and service conventions
4. keep the command explicit so production and staging environments are not silently populated at app startup

If the demo baseline changes, update this guide, the bootstrap user guide, and the active `new_app_docs` tracker in the same change.
