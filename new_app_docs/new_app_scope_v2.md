# workflow_app Scope V2

Date: 2026-04-12
Status: Active scope guardrail for post-thin-v1 implementation
Purpose: define the current scope boundaries for active implementation after thin-v1 completion.

## 1. Scope posture

1. thin-v1 is complete and remains the closed foundation baseline
2. active work is now v2: ambitious, best-practice-driven, and allowed to deepen capability where that materially improves correctness, continuity, maintainability, or production readiness
3. stronger scope does not authorize drift away from the workflow-centered product doctrine

## 2. In scope

1. strengthening the shared Go backend, workflow seams, and operator continuity
2. completing the promoted Svelte web layer on the shared backend and auth model
3. bounded refactors or rebuilds when existing code is weak, concentrated, or structurally misaligned
4. improving review, approval, reporting, detail continuity, admin continuity, and workflow support surfaces
5. additive client-neutral backend seams that preserve a clean path to later mobile reuse

## 3. Out of scope by default

1. reviving CRM-first product shape or making support-record breadth the center of gravity
2. splitting the system into separate web-specific and mobile-specific backends
3. building broad manual-entry UI depth that weakens the AI-agent-first and workflow-first operating model
4. reopening completed thin-v1 foundation modeling unless a real correctness defect requires it
5. novelty-driven architecture or experimental autonomy that weakens approvals, auditability, or database truth

## 4. Current scope focus

1. the active implementation focus is post-Milestone-14 user-testing support, bounded corrective slices for real defects, and readiness-preserving follow-through on the promoted Svelte runtime
2. the active workflow follow-through focus is recording user-testing evidence in `docs/workflows/` first and keeping downstream user guides aligned with the current served `/app` route family and supported browser actions
3. completed Milestone 10, Milestone 11, Milestone 12, and Milestone 13 planning detail is now archive material rather than active default context
