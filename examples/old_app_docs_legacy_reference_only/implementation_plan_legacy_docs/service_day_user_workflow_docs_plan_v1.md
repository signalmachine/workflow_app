# service_day User workflow documentation plan v1

Date: 2026-03-14
Status: Legacy reference
Purpose: preserve the pre-thin-v1 user-guide planning detail as historical reference.

## 1. Intent

This document exists to keep user workflow documentation visible as implementation quality work while separating it from code-correctness remediation.

The plan should:
1. track workflow-guide expectations for shipped features
2. keep user-guide sequencing aligned with real implementation state
3. prevent placeholder documentation from being mistaken for delivered user guidance
4. make deferrals explicit when shipped workflows still lack user-facing docs

## 2. Scope

This plan covers:
1. internal-user workflow guides under `docs/user_guides/`
2. setup and first-run user guidance for shipped functionality
3. role-oriented workflow guidance where the current implementation is real enough to document accurately

This plan does not cover:
1. product strategy already owned by the other canonical planning documents
2. technical implementation remediation already tracked in `docs/archive/review_remediation_2026_03_14.md`
3. speculative guides for unimplemented modules

## 3. Documentation stance

Rules:
1. only document workflows that are actually implemented and verified
2. do not describe planned behavior as if it already exists
3. keep one guide per workflow or role when that makes user operations clearer
4. treat missing user guides for shipped workflows as a planning and tracker gap, not as an excuse to overstate completion

## 4. Current state

Current repo state:
1. `docs/user_guides/README.md` exists as a placeholder and folder contract
2. no shipped internal-user workflow guides exist yet for the current CRM and auth surface
3. the current implementation now includes bootstrap/login, device-session refresh plus session management, notification-device registration, notification inbox/read plus a first push-dispatch baseline for task assignment, CRM relationship management, estimates, AI proposal acceptance, attachment upload/download, and an explicit v1 additive-compatibility policy, so the first workflow guides are no longer blocked by missing core behavior alone
4. Milestone B planning already expects the first workflow guides for the actually shipped CRM surface

## 5. Initial guide set

The first documentation wave should cover the workflows that are already closest to real operator use.

### 5.1 Wave 1: first-run and auth

Create when the current auth/bootstrap flow remains stable enough to document:
1. `admin_bootstrap_and_login.md`

Minimum content:
1. first bootstrap
2. login
3. device-session login, refresh, list, and revoke behavior for the current API-driven workflow
4. notification-device registration, inbox listing, read acknowledgement, push-dispatch configuration, and current notification limitations
5. bearer-token use expectations for the current API-driven workflow if no UI exists yet
6. current limitations, especially where the system is still API-first and online-first

### 5.2 Wave 2: CRM relationship workspace baseline

Create when the corresponding CRM behavior is technically solid:
1. `crm_accounts_and_contacts.md`
2. `lead_to_opportunity.md`
3. `tasks_and_follow_up.md`

Minimum content:
1. creating accounts and contacts
2. creating, qualifying, disqualifying, and converting leads
3. logging notes, activities, and communications
4. creating and completing follow-up tasks
5. current task-notification behavior for cross-user assignment, including the current queued-push dispatch baseline and the remaining limitation that richer retry/backoff handling is still later work
6. current limitations or deferrals for relationship linking or opportunity editing until those gaps are closed

### 5.3 Wave 3: estimates and AI assistance

Create when the currently shipped estimate and AI recommendation flows are stable enough to document without immediate churn:
1. `estimates.md`
2. `ai_assisted_follow_up.md`

Minimum content:
1. estimate creation and commercial-state progression
2. AI estimate draft recommendations
3. AI task proposal review, acceptance, and discard behavior
4. any approval or audit-visible behavior users need to understand

### 5.4 Later wave: launch home and global search

Create when the later web client has a stable signed-in launch experience:
1. `home_and_navigation.md`

Minimum content:
1. pinned home tiles
2. searching for workflows and records
3. pinning or unpinning activities
4. the difference between direct workflow launch, record search, and role or permission limits

## 6. Sequencing rule

The workflow documentation plan should follow this order:
1. bootstrap and login guidance first
2. CRM relationship-workspace guides next
3. estimate and AI-assist guides next once the current shipped behavior and limitations can be documented plainly
4. launch-home and global-search guidance later when the web client actually ships with that experience

Interpretation rule:
1. do not write polished end-user guides for opportunity management, estimates, or AI action flows until the currently shipped behavior and remaining limitations are narrow enough to describe without guessing

## 7. Deferral policy

When a shipped workflow still lacks user documentation:
1. record the gap in `plan_docs/service_day_refactor_tracker_v1.md` or the relevant active thin-v1 documentation note
2. name the specific missing guide or guide set
3. state whether the guide is blocked by unstable implementation or simply not yet written

## 8. Acceptance rule

User workflow documentation for a shipped slice is complete only when:
1. the guide exists under `docs/user_guides/`
2. the guide matches the current implementation
3. any important limitations or deferred capabilities are stated plainly
4. the tracker records the guide as delivered or explicitly deferred

## 9. Immediate next documentation actions

Current next steps:
1. keep workflow-guide work out of the current technical remediation sequence in `docs/archive/review_remediation_2026_03_14.md`
2. treat the first real workflow guides as the next documentation wave unless a higher-priority implementation bug interrupts it
3. keep auth/mobile guidance explicit about the current online-first contract, versioned API headers, device-session controls, notification-device registration, current inbox/read notification behavior, optional push-dispatch configuration, and the fact that broader notification event coverage plus richer push retry/backoff handling are still later work
4. keep polished opportunity-management, estimate, and AI workflow guides scoped carefully to the currently implemented behavior and limitations instead of waiting for hypothetical perfect stability
5. create the first real guides for:
   - `admin_bootstrap_and_login.md`
   - `crm_accounts_and_contacts.md`
   - `lead_to_opportunity.md`
   - `estimates.md`
   - `tasks_and_follow_up.md`
