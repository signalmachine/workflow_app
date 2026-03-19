# Estimate line shape alignment remediation note

Date: 2026-03-16
Status: Active
Purpose: document the follow-up needed to align the current estimate-line implementation with the broader canonical commercial line-shape expectations.

## 1. Canonical references

This remediation note aligns the shipped estimate implementation with:
1. `implementation_plan/service_day_initial_plan_2026_03_11.md`
2. `implementation_plan/service_day_execution_plan_v1.md`
3. `implementation_plan/service_day_crm_mvp_scope_v1.md`
4. `implementation_plan/service_day_schema_foundation_v1.md`
5. `plan_docs/service_day_refactor_tracker_v1.md`

## 2. Problem summary

The current implementation only accepts estimate line types `service` and `milestone`.

That is narrower than the current canonical plan, which expects the estimate foundation to remain usable for:
1. service lines
2. milestone lines
3. scoped-work lines
4. stocked-item lines where the business also sells inventory directly

This is not a request to turn CRM estimates into a full item/inventory workflow immediately.
It is a request to avoid baking an unnecessarily narrow line-shape contract into the commercial baseline.

## 3. Current implementation drift

Current code behavior:
1. CRM estimate normalization rejects any `line_type` other than `service` or `milestone`
2. AI estimate-draft acceptance applies the same narrowed line-type rule
3. tracker wording currently mirrors the shipped implementation rather than recording this narrower contract as a deliberate temporary limit

## 4. Target outcome

After this remediation:
1. the estimate model remains service-business-first but no longer hard-codes a two-type ceiling
2. estimate lines can represent scoped work and stocked-item quoting without redesigning the aggregate later
3. AI estimate acceptance follows the same broadened and validated line-shape contract
4. the commercial baseline stays compatible with later delivery, billing, and limited direct-item sales flows

## 5. Recommended implementation direction

### 5.1 Minimum contract expansion

Expand the accepted estimate line types to include at least:
1. `service`
2. `milestone`
3. `scope`
4. `item`

Naming rule:
1. use one canonical vocabulary in code, API, tests, and docs
2. avoid adding multiple overlapping synonyms such as `stocked_item`, `product_line`, and `work_scope` unless a later canonical decision explicitly needs that complexity

Recommended direction:
1. `scope` should cover scoped-work quoting that is not best expressed as a milestone or generic service line
2. `item` should cover stocked-item or material-style quoted lines without forcing immediate inventory reservation or costing behavior into CRM

### 5.2 Scope boundary

This remediation should not:
1. introduce inventory allocation, stock deduction, or item-master lookup requirements into CRM estimate creation
2. force delivery or billing conversion work to land in the same session
3. broaden pricing into a deep CPQ or retail catalog system

The goal is contract breadth and future-safe line semantics, not downstream automation depth.

### 5.3 Validation and persistence

Update:
1. CRM estimate input normalization
2. AI estimate-draft acceptance normalization
3. any line-type validation tests and fixtures
4. public documentation and tracker wording that currently imply only `service` or `milestone`

Keep:
1. current quantity, unit-price, and total behavior unless a separate commercial-model decision changes it
2. revision-safe estimate lineage and current audit/idempotency behavior

## 6. Required tests

1. CRM service tests proving `scope` and `item` line types are accepted
2. AI acceptance tests proving estimate-draft acceptance supports the same expanded line-type set
3. integration coverage proving persisted estimate lines round-trip correctly through create, revise, and list flows
4. regression coverage proving invalid unknown line types still fail with stable invalid-input behavior

## 7. Acceptance criteria

This remediation is complete when:
1. the estimate-line contract matches the canonical planning set instead of the narrower current implementation
2. live CRM estimate creation and revision accept the expanded line-type set intentionally
3. AI estimate acceptance stays aligned with the same line-type rules
4. tracker and README wording no longer understate or misstate the intended commercial line-shape baseline

## 8. Recommended session timing

Recommended sequencing:
1. land RP-09 workflow task-model alignment first because it is the broader structural drift already prioritized in the tracker
2. implement this estimate-line remediation after RP-09 or alongside a closely related commercial-model hardening session
3. avoid coupling it to inventory or billing milestones unless a session is already touching those seams

## 9. Non-goals

This remediation should not expand into:
1. inventory master data implementation
2. automatic quote-to-invoice conversion
3. automatic work-order material planning
4. advanced pricing matrices or CPQ rules
