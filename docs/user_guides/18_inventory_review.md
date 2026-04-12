# Inventory Review

Date: 2026-04-12
Status: Active
Purpose: explain how to review inventory output and confirm it matches the document workflow chain.

## 1. Open the inventory review surface

Open the inventory review page from the browser navigation or from a linked workflow record:

1. `/app/review/inventory`
2. `/app/review/inventory/{movement_id}`
3. `/app/review/inventory?movement_id={movement_id}`
4. `/app/review/inventory?item_id={item_id}`
5. `/app/review/inventory?location_id={location_id}`
6. `/app/review/inventory/items/{item_id}`
7. `/app/review/inventory/locations/{location_id}`

Use this surface when you need to inspect inventory truth tied to a workflow record.

## 2. Review the inventory record

Check that the inventory page shows:

1. the correct inventory record identity
2. the current stock or movement state
3. the linked source document or execution record
4. any related review details needed to trace the change

Example:

If a work-order workflow consumes `Pump seal kit` from `Main warehouse`, open `/app/review/inventory?item_id={item_id}` or the exact movement route and confirm the movement, source document, and location context match the work-order chain.

## 3. Confirm continuity

The important checks are:

1. the inventory record traces back to the expected source document or movement
2. the browser page and API read agree on the same inventory facts
3. the linked workflow records still point to the same inventory event

## 4. Troubleshooting

If the inventory record looks incomplete:

1. reopen the source workflow record
2. confirm the inventory action actually happened
3. verify the org session and record identity are correct
