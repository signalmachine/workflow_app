# Documentation alignment remediation 2026-03-18

Date: 2026-03-18
Status: Active remediation note
Purpose: track any remaining repository-documentation cleanup needed after the thin-v1 canonicalization pass so the work does not become orphaned.

## 1. Problem statement

The repository has adopted `plan_docs/` as the active canonical planning set, but parts of the older planning material may still:

1. present themselves as active or canonical
2. point readers to outdated reading orders
3. retain broader roadmap assumptions that should remain historical only

## 2. Required follow-up

1. relabel any remaining legacy `implementation_plan/` files that still read as active canon
2. keep `README.md`, `AGENTS.md`, and `plan_docs/` cross-references aligned whenever the active reading order changes
3. avoid introducing new active implementation rules only in `implementation_plan/`
4. promote any still-needed thin-v1 rule from legacy docs into `plan_docs/` before relying on it as current canon

## 3. Acceptance criteria

1. a new contributor can determine the active planning source without ambiguity
2. legacy files no longer present themselves as the default canonical authority
3. any remaining legacy-reference exceptions are explicit and justified

## 4. Linkage rule

This note must remain linked from `plan_docs/service_day_refactor_tracker_v1.md` until its cleanup items are closed or intentionally retired.
