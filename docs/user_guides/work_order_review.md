# Work-Order Review

Date: 2026-03-31
Status: Active
Purpose: explain how to review a work-order record and confirm it matches the workflow chain that produced it.

## 1. Open the work-order review surface

Open the work-order review page from the browser navigation or from a linked workflow record:

1. `/app/review/work-orders`
2. `/app/review/work-orders/{work_order_id}`
3. `/app/review/work-orders?work_order_id={work_order_id}`
4. `/app/review/work-orders?document_id={document_id}`

Use this surface when you need to inspect a single work-order record.

## 2. Review the work-order record

Check that the work-order page shows:

1. the correct work-order identity
2. the current execution state
3. any linked inventory, labor, or source document context
4. any related downstream review details

## 3. Confirm continuity

The important checks are:

1. the work-order record traces back to the expected source workflow
2. the browser page and API read agree on the same work-order facts
3. the linked workflow records still point to the same work-order

## 4. Troubleshooting

If the work-order record looks incomplete:

1. reopen the source workflow record
2. confirm the work-order action actually happened
3. verify the org session and record identity are correct
