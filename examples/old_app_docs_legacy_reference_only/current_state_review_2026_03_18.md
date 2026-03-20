# Current State Review

Date: 2026-03-19
Status: Review note
Purpose: summarize the current repository shape and why the plan reset is justified.

## 1. Review inputs

This review considered:

1. `AGENTS.md`
2. root `README.md`
3. `docs/implementation_objectives/implementation_principles.md`
4. the main documents in `implementation_plan/`
5. the current module layout under `internal/`
6. current migration and package names

## 2. What the codebase currently implements

The implemented backend is strongest in:

1. identity and auth foundations
2. audit and idempotency foundations
3. CRM records and CRM-adjacent workflow
4. AI recommendation persistence and approval traces
5. a first shared documents kernel plus a small accounting shell with balanced journal validation, explicit post action, and transactional journal-to-document linkage

Current top-level implementation modules:

1. `identityaccess`
2. `crm`
3. `documents`
4. `workflow`
5. `ai`
6. `notifications`
7. `attachments`
8. `accounting`

Notably absent as implemented first-class modules today:

1. `work_orders`
2. `inventory_ops`
3. `projects`
4. `billing`
5. `reporting`

## 3. What the current planning set optimizes for

The current `implementation_plan/` set is clear and serious, but it is still too broad for a thin v1 because it tries to keep all of these active at once:

1. CRM MVP
2. project coordination
3. work-order execution
4. billing
5. accounting
6. workforce and later payroll
7. inventory and serviced assets
8. portal/mobile/web launch concerns
9. later rental, trading, marketplace, and other expansion seams

That breadth creates two planning problems:

1. v1 foundations are not isolated sharply enough from later expansion
2. CRM became the earliest concrete center of gravity even though the intended long-term center is `work_order` plus accounting/inventory truth

## 4. Main mismatch against `docs/implementation_objectives/implementation_principles.md`

The doctrine in `docs/implementation_objectives/implementation_principles.md` pushes toward:

1. documents as intent
2. ledgers as truth
3. execution context as process
4. strict posting boundaries
5. reports as derived views
6. AI as a client, not an authority

The current plan partially supports that, but the planning structure still spreads attention across too many module-specific workflows before the core model is reduced to:

1. document engine
2. financial ledger
3. inventory ledger
4. execution context
5. approval and posting rules
6. report layer

## 5. Recommended planning reset

The thin-v1 plan should reset around these priorities:

1. AI-agent-only interaction model for business actions
2. strong document lifecycle with draft, submitted, approved, posted boundaries
3. strong financial and inventory ledger foundations
4. work-order and task execution context as the main operational layer
5. report and query layer for humans to inspect current state
6. minimal party and item foundations instead of broad CRM depth

## 6. Concrete planning cuts needed

These areas should move out of v1 and into v2 unless they are required by a foundation:

1. broad CRM pipeline depth
2. projects as a primary planning concern
3. customer portal
4. web launch/navigation planning
5. spreadsheet exchange planning
6. marketplace, rental, and broader business-mode expansion
7. payroll and deep tax planning
8. advanced customer communication channels

## 7. Refactor conclusion

The repository does not have a code problem first. It has a scope-shape problem first.

The current codebase is capable of supporting a thinner foundation-led plan, but the current document set still frames v1 as a broad operational suite. A new planning set should re-center v1 on:

1. ledger
2. documents
3. execution context
4. AI control boundary
5. reporting
