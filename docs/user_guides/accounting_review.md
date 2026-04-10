# Accounting Review

Date: 2026-04-10
Status: Active
Purpose: explain how to review accounting output and confirm it matches the document workflow chain.

## 1. Open the accounting review surface

Open the accounting review page from the browser navigation or from a linked workflow record:

1. `/app/review/accounting`
2. `/app/review/accounting/{entry_id}`
3. `/app/review/accounting/journal-entries?entry_id={entry_id}`
4. `/app/review/accounting/journal-entries?document_id={document_id}`
5. `/app/review/accounting/control-balances`
6. `/app/review/accounting/tax-summaries`
7. `/app/review/accounting/trial-balance`
8. `/app/review/accounting/balance-sheet`
9. `/app/review/accounting/income-statement`

Use the accounting report directory first, then open the dedicated journal, control-balance, tax-summary, trial-balance, balance-sheet, or income-statement destination needed for the current workflow trace.

## 2. Review the accounting record

Check that the accounting page shows:

1. the correct accounting record identity
2. the current posting or journal state
3. the linked source document
4. any related review details needed to trace the posting
5. financial-statement totals when the task is report review rather than exact journal-entry review

## 3. Confirm continuity

The important checks are:

1. the accounting record traces back to the expected source document
2. the browser page and API read agree on the same accounting facts
3. any downstream summary or balance view matches the record you opened
4. trial-balance, balance-sheet, and income-statement totals come from the shared reporting seam and match the selected effective-date filter

## 4. Troubleshooting

If the accounting record looks incomplete:

1. reopen the source document
2. confirm the posting or review action actually happened
3. verify the org session and record identity are correct
