# Admin Inventory Setup

Date: 2026-04-12
Status: Active
Purpose: explain how an admin maintains inventory items and locations that downstream workflows can reference.

## 1. Open inventory setup

Use an admin session and open:

1. `/app/admin`
2. `/app/admin/master-data`
3. `/app/admin/inventory`

Use `/app/admin/inventory` for setup. Use `/app/review/inventory` when the task is stock, movement, or reconciliation review.

## 2. Maintain items

Use the item section to create bounded inventory item records and to mark items active or inactive.

Example:

Before testing a work-order fulfillment request, an admin creates item `Pump seal kit` with the right SKU and active status. Later inventory movement and work-order review should point to that same item rather than a free-text item name.

## 3. Maintain locations

Use the location section to create warehouse, van, adjustment, job-site, or installed-equipment locations.

Example:

A field-service workflow needs movement from `Main warehouse` to `Van 1`. Confirm both locations exist and are active on `/app/admin/inventory` before creating or processing the request that should produce inventory movement.

## 4. Confirm continuity

After setup changes, confirm:

1. item and location records appear on `/app/admin/inventory`
2. inactive records are visibly governed as inactive
3. `/app/inventory` and `/app/review/inventory` continue to show stock and movement review from the shared backend seam
4. downstream work-order review can still trace back to the item and location records

## 5. Troubleshooting

If an item or location is missing from review:

1. check `/app/admin/inventory` first
2. confirm the record belongs to the current org
3. confirm it is active when the downstream workflow requires active records
4. confirm the inventory movement or work-order workflow actually ran
