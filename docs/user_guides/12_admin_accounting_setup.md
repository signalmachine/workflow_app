# Admin Accounting Setup

Date: 2026-04-12
Status: Active
Purpose: explain how an admin maintains foundational accounting master data without changing posted accounting truth.

## 1. Open accounting setup

Use an admin session and open one of these browser paths:

1. `/app/admin`
2. `/app/admin/master-data`
3. `/app/admin/lists`
4. `/app/admin/accounting`

Use `/app/admin/accounting` for ledger accounts, tax codes, and accounting periods. Use `/app/review/accounting` when the task is posted accounting review instead of setup.

## 2. Maintain ledger accounts

Use the ledger-account section to create bounded chart-of-accounts rows and to mark accounts active or inactive.

Example:

An admin is preparing the North Harbor Works demo baseline and needs a new expense account for small tools. Open `/app/admin/accounting`, create the ledger account with the right account type and control-account relationship, then confirm the account appears in the ledger-account list before using it in downstream workflow tests.

Do not use this page to edit posted journal entries. Posted accounting truth is reviewed through `/app/review/accounting`.

## 3. Maintain tax codes

Use the tax-code section when a workflow needs a reusable GST or TDS code.

Example:

A submitted vendor purchase request needs an 18 percent GST purchase tax code. Open `/app/admin/accounting`, confirm the GST purchase code already exists from bootstrap data, and only create a new code if the existing control-account mapping does not cover the required workflow.

## 4. Maintain accounting periods

Use the period section to create open accounting periods and close a period when it should no longer accept normal posting.

Example:

Before testing FY2026-27 invoice workflows, confirm the FY2026-27 period is open. If the period has been closed for a closeout review, do not reopen it by changing posted records; create or select the correct open period for the next workflow test.

## 5. Confirm continuity

After setup changes, confirm:

1. the record appears on `/app/admin/accounting`
2. inactive records are visibly governed as inactive
3. downstream review pages still use shared backend truth
4. posted accounting review remains separate from setup maintenance

## 6. Troubleshooting

If a create or status action fails:

1. confirm the signed-in actor has admin access
2. confirm the required fields are filled
3. confirm the control account or period relationship exists
4. reload `/app/admin/accounting` and verify whether the record was already created
