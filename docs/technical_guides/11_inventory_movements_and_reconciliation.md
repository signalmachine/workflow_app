# Inventory Movements And Reconciliation

Date: 2026-03-31
Status: Active technical guide
Purpose: explain how inventory truth is modeled, how movements are recorded, and how execution and accounting handoffs are derived.

## 1. What inventory owns

`internal/inventoryops` owns:

1. item truth
2. location truth
3. inventory movements
4. source-destination semantics
5. inventory document capture
6. execution linkage
7. accounting handoff derivation
8. stock balance derivation

This package is one of the clearest examples of append-only operational truth in the application.

## 2. Items and locations

Inventory begins with two core record types:

1. items
2. locations

Each has a role that describes how it participates in the system.

Examples:

1. resale item
2. service material
3. traceable equipment
4. warehouse location
5. site location
6. installed location

Those role fields matter because different stock and execution flows use the same inventory backbone but with different operational intent.

## 3. Movements

Movements are the durable inventory truth.

```go
movement, err := s.RecordMovement(ctx, inventoryops.RecordMovementInput{
	DocumentID:            document.ID,
	ItemID:                item.ID,
	MovementType:          inventoryops.MovementTypeIssue,
	MovementPurpose:       inventoryops.MovementPurposeServiceConsumption,
	UsageClassification:   inventoryops.UsageBillable,
	SourceLocationID:      warehouse.ID,
	DestinationLocationID: "",
	QuantityMilli:         1000,
	ReferenceNote:         "material issued for work order",
	Actor:                 actor,
})
```

For an `issue` movement, the source is required and the destination must be empty. The example above matches that validation rule.

The important details are:

1. source and destination are explicit
2. movement purpose is explicit
3. usage classification is explicit
4. movement truth is append-only

## 4. Stock balances

Stock is derived from movements.

That means the package can calculate on-hand quantities without treating stock as arbitrary mutable truth. This is important for auditability and reconciliation.

## 5. Document capture

Some inventory flows also capture a document-level view with lines, accounting handoffs, and execution links.

That allows the system to say:

1. what moved
2. why it moved
3. where it moved from and to
4. whether it needs accounting posting
5. whether it is linked to an execution context like a work order

## 6. Reconciliation model

The reporting layer exposes inventory reconciliation as review data.

Operators use that to inspect:

1. movement history
2. pending accounting handoffs
3. execution linkage
4. work-order-related material usage

The reconciliation view is derived from the movement truth, not the source of it.

## 7. What to keep stable

Be careful with:

1. source-destination semantics
2. movement purpose classification
3. item and location role validation
4. stock derivation
5. accounting handoff state
6. execution linkage state
