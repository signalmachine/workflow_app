# Implementation Plan Migration Map

Date: 2026-03-18
Status: Active migration map
Purpose: classify the current `implementation_plan/` documents against the active thin-v1 plan and define their legacy-reference role.

## 1. Classification rules

Each current document should be treated as one of:

1. keep in reduced form for thin v1
2. collapse into a smaller replacement doc
3. move mostly to v2 backlog
4. keep only as historical reference

## 2. Accepted mapping

| Current document | Proposed action | Reason |
| --- | --- | --- |
| `implementation_plan/README.md` | Collapse | Replaced by the smaller reading order in [plan_docs/README.md](/home/vinod/PROJECTS/service_day/plan_docs/README.md) now that the thin-v1 reset is canonical |
| `implementation_plan/service_day_initial_plan_2026_03_11.md` | Collapse | It contains too much product breadth for v1; keep only the thin-v1 identity, AI-agent-first stance, database-first stance, and work-order-centered execution posture |
| `implementation_plan/service_day_execution_plan_v1.md` | Collapse | Replaced by the narrower milestone flow in [service_day_refactor_execution_plan_v1.md](/home/vinod/PROJECTS/service_day/plan_docs/service_day_refactor_execution_plan_v1.md) |
| `implementation_plan/service_day_module_boundaries_v1.md` | Collapse | Reduced thin-v1 ownership rules now live in `plan_docs/service_day_schema_and_module_boundaries_v1.md` |
| `implementation_plan/service_day_navigation_launch_experience_v1.md` | Move to v2 backlog | Useful later, but not a thin-v1 foundation concern |
| `implementation_plan/service_day_data_exchange_plan_v1.md` | Move to v2 backlog | Exchange planning is not a v1 center once the scope is reduced |
| `implementation_plan/service_day_schema_foundation_v1.md` | Collapse | Reduced thin-v1 schema guidance now lives in `plan_docs/service_day_schema_and_module_boundaries_v1.md` |
| `implementation_plan/service_day_crm_mvp_scope_v1.md` | Move mostly to v2 backlog | Thin v1 should keep only minimal party/contact support, not CRM as a primary product center |
| `implementation_plan/service_day_ai_architecture_v1.md` | Collapse | Reduced thin-v1 AI guidance now lives in `plan_docs/service_day_ai_architecture_v1.md` |
| `implementation_plan/implementation_decisions_v1.md` | Collapse | Reduced thin-v1 defaults now live in `plan_docs/service_day_implementation_defaults_v1.md` |
| `implementation_plan/implementation_tracker.md` | Collapse | Active thin-v1 status now lives in `plan_docs/service_day_refactor_tracker_v1.md` |
| `implementation_plan/service_day_user_workflow_docs_plan_v1.md` | Collapse | Keep only the rule that user-visible review and reporting workflows need docs when shipped |
| remediation notes under `implementation_plan/` | Historical reference | Keep them for repository history, but do not let them drive thin-v1 scope |

## 3. Thin-v1 canonical document set

The accepted canonical set now consists of:

1. one README and reading-order doc
2. one principles and product-boundary doc
3. one scope and deferral doc
4. one schema and module-boundary doc
5. one implementation-defaults doc
6. one AI and approval-boundary doc
7. one execution-plan doc
8. one tracker doc

## 4. Important cuts

The following current planning topics should stop driving v1:

1. CRM-first product framing
2. launch/navigation planning
3. portal planning
4. spreadsheet exchange planning
5. broad later business-mode expansions
6. payroll and deeper tax sequencing
7. advanced projects framing

## 5. Important keepers

These ideas from the current plan should survive the refactor:

1. database-first and SQL-first enforcement
2. AI-agent-first architecture with strict write boundaries
3. audit-required business mutations
4. work orders as the strongest operational capability
5. balanced accounting and explicit posting boundaries
6. shared task engine with clear ownership semantics
7. tenant safety and idempotent mutation paths

## 6. Migration recommendation

Do not try to edit every current planning file incrementally.

Recommended approach:

1. use `plan_docs/` as the canonical source for new planning and implementation work
2. use `implementation_plan/` only for historical context, existing implementation detail, and focused remediation notes
3. keep legacy files clearly labeled so they do not silently override the thin-v1 plan

## 7. Legacy reference stance

After the thin-v1 reset:

1. `plan_docs/` is the first planning source for new work
2. `implementation_plan/` remains available for historical decisions, existing implementation context, and slice-specific reference
3. `implementation_plan/` should not be treated as the default source for current v1 priorities
