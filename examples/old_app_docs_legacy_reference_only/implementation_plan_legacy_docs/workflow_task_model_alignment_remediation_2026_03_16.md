# Workflow task model alignment remediation note

Date: 2026-03-16
Status: Active
Purpose: document the concrete implementation work needed to align the current workflow/task/activity code and schema with the updated canonical planning set.

## 1. Canonical references

This remediation note implements the current canonical direction in:
1. `implementation_plan/service_day_initial_plan_2026_03_11.md`
2. `implementation_plan/service_day_execution_plan_v1.md`
3. `implementation_plan/service_day_module_boundaries_v1.md`
4. `implementation_plan/service_day_schema_foundation_v1.md`
5. `implementation_plan/service_day_crm_mvp_scope_v1.md`
6. `implementation_plan/implementation_decisions_v1.md`

This note exists because the current implementation still reflects the earlier simpler model:
1. tasks are user-assigned only
2. tasks have one primary context and no secondary related links
3. shared team records do not exist
4. CRM activities still carry some task-like due/owner/completion semantics

## 2. Problem summary

The updated canonical plan now requires:
1. one shared workflow task engine
2. one primary actionable owner per task, which may be either one person or one team queue
3. one primary business context per task
4. optional secondary related links for visibility and analytics, such as project, work order, account, site, or serviced asset
5. clear separation between `workflow`, `task`, and `activity`

The current implementation does not fully satisfy that direction.

Current implementation drift:
1. `tasks.assignee_user_id` hard-codes person-only assignment
2. no canonical `teams` or `team_members` tables exist
3. no `task_related_links` or equivalent secondary-link shape exists
4. CRM `activities` currently act partly like lightweight reminder/task records

## 3. Target outcome

After this remediation:
1. workflow tasks can be owned by either one worker/person or one team queue
2. team queues have stable IDs and membership records
3. tasks retain exactly one primary context while supporting optional secondary related links
4. activities remain factual history records rather than the main owner of accountable follow-up
5. notifications, AI task acceptance, reporting, and timelines continue to work on top of the stronger model

## 4. Recommended schema changes

### 4.1 Shared team foundation

Add:
1. `teams`
   - `id`
   - `org_id`
   - `code`
   - `display_name`
   - `active`
   - audit timestamps
2. `team_members`
   - `id`
   - `org_id`
   - `team_id`
   - `worker_id`
   - optional `role_code`
   - `active`
   - audit timestamps

Rules:
1. teams are tenant-scoped shared operational records
2. membership should point to `workers`, not directly to `users`
3. if a worker also has a login user, downstream workflow/notification logic may resolve the active user through worker-to-user linkage rather than treating `team_members` as auth rows

### 4.2 Task ownership model

Replace the current single-user assignment shape with a primary-owner model.

Recommended `tasks` changes:
1. add `owner_type`
   - constrained to values such as `worker` or `team`
2. add `owner_id`
3. add `claimed_by_worker_id` nullable for later queue-claim flows
4. add `completed_by_worker_id` nullable
5. preserve one primary business context through existing `context_type` and `context_id`

Migration direction:
1. backfill existing rows with `owner_type = 'worker'` or `owner_type = 'user-backed-worker'` only if a worker mapping already exists cleanly
2. if worker mapping does not yet exist for current task assignees, use a staged migration:
   - first allow `owner_type = 'user'` temporarily as an internal compatibility state
   - then add a follow-up migration that converts task ownership to worker/team once worker coverage is complete
3. remove `assignee_user_id` only after API and service code no longer depend on it

Recommended compatibility rule:
1. do not attempt a risky one-shot rewrite if worker coverage is incomplete
2. use an explicit staged compatibility state if needed, but keep the canonical long-term target as worker/team ownership

### 4.3 Secondary related links

Add a separate secondary-link table rather than overloading the primary context columns.

Recommended shape:
1. `task_related_links`
   - `id`
   - `org_id`
   - `task_id`
   - `linked_record_type`
   - `linked_record_id`
   - optional `relationship_role`
   - uniqueness on `(org_id, task_id, linked_record_type, linked_record_id)`

Rules:
1. keep `tasks.context_type/context_id` as the sole primary context
2. use `task_related_links` only for secondary visibility, reporting, search, timeline, and cross-context navigation
3. keep foreign-key linkage tenant-safe through `linked_records` or equivalent owned registry

### 4.4 Activity cleanup

Do not collapse activities into tasks.

Recommended direction:
1. keep CRM `activities` as factual history records
2. preserve `activity_type`, `subject`, `details`, timestamps, and linked-record context
3. treat `due_at` and `completed_at` on activities as transitional legacy fields rather than the future accountable-work model
4. when follow-up becomes accountable, create a linked workflow task instead of deepening activity-state semantics

Recommended migration posture:
1. do not force immediate destructive removal of `activities.due_at` or `activities.completed_at`
2. first stop expanding activity workflow semantics in service/API code
3. only later decide whether to keep those fields for reminder-style activities or narrow them away

## 5. Recommended service and API changes

### 5.1 Workflow API

Evolve task commands from single-user assignee to explicit owner semantics.

Recommended create payload direction:
1. keep `context_type` and `context_id`
2. replace `assignee_user_id` with:
   - `owner_type`
   - `owner_id`
3. optionally accept:
   - `related_links`
   - `priority`
   - `due_at`

Recommended list/filter additions:
1. filter by `owner_type`
2. filter by `owner_id`
3. later filter by queue/claimed state
4. optionally filter by related linked record where reporting or UI needs it

Recommended compatibility rollout:
1. support legacy `assignee_user_id` for one deprecation window
2. translate it internally to the new owner model where possible
3. remove the legacy field only after clients and AI acceptance paths are updated

### 5.2 AI task-acceptance flows

Current AI acceptance creates live tasks through the existing workflow path.

Required follow-up:
1. update AI acceptance input and accepted-artifact persistence to use task owner semantics instead of only `assignee_user_id`
2. preserve audit metadata and accepted-artifact replay behavior during the transition
3. ensure AI can target either a specific person or a team queue where policy allows

### 5.3 Notification behavior

The current notification system is person-assignment-centric.

Required follow-up:
1. keep direct user notifications for person-owned tasks
2. define team-queue notification behavior explicitly before implementation

Recommended first team behavior:
1. team-owned tasks create queue visibility first
2. do not fan out push notifications to every team member by default
3. after a claim or explicit delegation, notify the claiming/delegated person normally

This avoids notification spam and weak accountability.

## 6. Recommended rollout order

1. Add `teams` and `team_members`.
2. Add worker/team ownership support to workflow services and schema while preserving current user-assigned compatibility.
3. Update task APIs, service logic, AI acceptance, and notifications to the new owner model.
4. Add `task_related_links` and update search/timeline/reporting surfaces.
5. Stop adding new task-like semantics to CRM activities.
6. Add reporting support for queue aging, reassignment, overdue rates, and throughput.
7. Remove deprecated task user-assignment-only paths after the compatibility window closes.

## 7. Required tests

### 7.1 Schema and migration tests

1. migration coverage for backfilling legacy task rows
2. tenant-safety coverage for `teams`, `team_members`, and `task_related_links`
3. constraint tests proving one primary owner and one primary context remain enforced

### 7.2 Workflow tests

1. create/list/complete flows for person-owned tasks
2. create/list/claim/delegate flows for team-owned tasks
3. queue visibility tests
4. reassignment and audit-history tests
5. related-link persistence and duplicate-prevention tests

### 7.3 Integration tests

1. AI task acceptance into person-owned and team-owned tasks
2. notification fan-out behavior for person ownership
3. queue-only behavior for team ownership until later fan-out rules are explicitly added
4. CRM timeline projection showing primary context plus any derived secondary visibility

## 8. Acceptance criteria

This remediation is complete when:
1. team ownership is modeled through canonical tables rather than freeform worker text
2. workflow tasks no longer depend only on `assignee_user_id`
3. the API supports the owner model intentionally
4. tasks retain one primary context while supporting optional secondary links
5. activities remain factual history and no longer act as the main accountable-work model
6. AI, notifications, timelines, and reporting still behave coherently on the stronger task model

## 9. Non-goals for this remediation

This remediation should not be allowed to expand into:
1. a broad BPMN-style workflow engine
2. many-assignee tasks as the default model
3. full dispatch optimization
4. a separate facilities-management product
5. a second workflow state system outside the normal domain-service and audit boundaries

## 10. Recommendation for future implementation sessions

Treat this remediation as the next canonical workflow-model correction before:
1. broader work-order execution depth
2. richer mobile task flows
3. team-based notification expansion
4. task/activity efficiency reporting

The current codebase is still early enough that this should be corrected now rather than after larger delivery and mobile surfaces depend on the older person-only task contract.
