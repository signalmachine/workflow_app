# workflow_app Milestone 13 Contextual Navigation Refactor Plan

Date: 2026-04-07
Status: First implementation pass landed in code on 2026-04-07
Purpose: define and now track the forward information architecture for refactoring the Svelte shell from the earlier route-family sidebar into a major-area sidebar with contextual section tabs.

## 1. Why this slice exists

The current Svelte shell is structurally sound, but its navigation model is still too route-family oriented.

Current problems:

1. the left sidebar mixes durable application areas with individual task pages such as request submission, operations feed, and agent chat
2. the current shell does not give operators one stable long-running area context for deeper work
3. the current route taxonomy overexposes implementation-era route families such as `review` and `operations` as if they were the final user mental model
4. the current shell does not yet express one contextual second-level navigation layer that can adapt to the active area without creating a second competing global chrome system

This slice exists to correct those problems without weakening the shared backend truth model or workflow ownership boundaries.

## 2. Target navigation doctrine

The forward browser shell should use two layers:

1. a **major-area sidebar** for durable application context
2. a **contextual section-tab row** for view modes inside the selected area

Rules:

1. the sidebar owns major-area selection
2. the contextual tab row owns view selection inside the active area
3. the top bar itself remains brand plus session or user controls and must not become a second global navigation strip
4. contextual tabs may vary by area
5. contextual tabs must not redefine canonical workflow ownership

## 3. Major-area sidebar

The forward sidebar model should be:

1. `Agent`
2. `Accounting`
3. `Inventory`
4. `Operations`
5. `Settings`
6. `Admin`

Notes:

1. `Settings` and `Admin` remain secondary and utility-oriented, but they stay visible because they are stable application areas rather than ephemeral actions
2. `AR` and `AP` remain under `Accounting` until they become first-class route families with coherent owned surfaces
3. the sidebar should support a desktop collapsible state and the existing mobile drawer state

## 4. Contextual tab model

The default tab vocabulary should be:

1. `Overview`
2. `Workflows`
3. `Actions`
4. `Lists`
5. `Reports`
6. `Search`

Area-specific override:

1. `Agent` should lead with `Messages` and `Requests`
2. the rest of the `Agent` tabs may continue with `Workflows`, `Actions`, `Lists`, and `Reports`

Rules:

1. tabs describe operator mode inside the area rather than implementation-era route groups
2. tabs may be hidden per area when no meaningful surface exists yet
3. tab labels should remain mostly fixed across areas to support habit and recognition
4. the agent area is the main allowed exception because it is an observability and control surface, not a normal domain workspace

## 5. Workflow ownership guardrails

This refactor must preserve these rules:

1. accounting workflows remain owned by accounting even if the same activity is visible through an agent lens
2. inventory workflows remain owned by inventory even if surfaced through agent activity or shared reports
3. the agent area shows agent activity, requests, messages, and agent-observed state; it does not become a duplicate source of accounting or inventory truth
4. contextual tabs may pivot the presentation of data, but the shared Go backend remains the canonical business-truth layer

## 6. Concrete proposed mapping from current surfaces

The current Svelte routes should be remapped as follows for the first implementation pass.

### 6.1 Agent

1. `Messages` -> `/app/operations-feed`
2. `Requests` -> `/app/agent-chat` plus request-submission entry affordances
3. `Workflows` -> agent-related workflow status and audit summaries; initially may reuse a grouped landing page backed by existing navigation snapshot seams
4. `Actions` -> coordinator or operator actions that are explicitly agent-related
5. `Lists` -> agent request and proposal views
6. `Reports` -> agent audit and status summaries

### 6.2 Accounting

1. `Overview` -> new accounting landing route
2. `Workflows` -> accounting-related workflow state, approvals, and downstream document movement
3. `Actions` -> bounded accounting actions and handoff entry points
4. `Lists` -> current review accounting, document, proposal, and approval list surfaces as grouped accounting views
5. `Reports` -> control balances, tax summaries, and accounting review reads
6. `Search` -> filtered accounting route discovery or scoped search

### 6.3 Inventory

1. `Overview` -> `/app/inventory`
2. `Workflows` -> inventory-related workflow movement and reconciliation status
3. `Actions` -> bounded inventory actions when they exist
4. `Lists` -> stock, movements, and reconciliation review surfaces
5. `Reports` -> inventory reporting views
6. `Search` -> scoped search or route discovery for inventory

### 6.4 Operations

1. `Overview` -> `/app`
2. `Workflows` -> `/app/operations`
3. `Actions` -> `/app/submit-inbound-request`
4. `Lists` -> grouped operational lists and request queues
5. `Reports` -> cross-workflow reporting views that are operational rather than module-owned
6. `Search` -> `/app/routes`

### 6.5 Settings

1. `Overview` -> `/app/settings`

### 6.6 Admin

1. `Overview` -> `/app/admin`
2. `Accounting` -> `/app/admin/accounting`
3. `Parties` -> `/app/admin/parties`
4. `Access` -> `/app/admin/access`
5. `Inventory` -> `/app/admin/inventory`

## 7. Implementation posture

This slice is primarily an information-architecture and shell-state refactor.

Expected technical posture:

1. low architectural risk
2. moderate implementation effort
3. limited backend dependency for the first pass
4. most work concentrated in frontend navigation config, shell composition, active-state logic, and area landing pages

Likely first-pass technical approach:

1. introduce one declarative navigation registry describing sidebar areas, contextual tabs, route matching, and access rules
2. keep current route files where practical and remap them under the new shell model before creating many new pages
3. add new area landing routes only where the current route family does not map cleanly enough
4. keep the first pass light on backend API additions unless a tab needs a new grouped snapshot surface for correctness

## 8. Recommended implementation order

1. refactor shell configuration into a declarative area-plus-tabs navigation model
2. replace the current mixed sidebar entries with the new major-area sidebar
3. add contextual tabs below the top bar and wire active-state logic from the current path
4. add desktop sidebar collapse with preference persistence
5. remap existing Svelte route families under the new area-plus-tab structure without changing canonical backend ownership
6. add new area landing pages where the current surfaces are only placeholders or where grouped domain entry is needed
7. tighten labels, summaries, and page headers so they reflect the new area model rather than legacy route taxonomy
8. update workflow validation checklists after the new shell is implemented

## 9. Verification expectations

Before closing implementation for this slice:

1. run frontend type-check and tests
2. verify desktop expanded and collapsed sidebar behavior
3. verify mobile drawer behavior
4. verify active-state correctness for sidebar areas and contextual tabs
5. verify that agent, accounting, inventory, operations, settings, and admin each preserve continuity into existing routes
6. verify that no route loses access because of the new shell model
7. update `docs/workflows/` validation material when operator-visible browser behavior changes materially

## 10. Current implementation checkpoint

The first pass of this refactor is now landed in code.

Landed result:

1. the Svelte shell no longer treats the sidebar as one flat route-family list with mixed task links and utility destinations
2. the sidebar now exposes the planned major application areas: Agent, Accounting, Inventory, Operations, Settings, and Admin for privileged actors
3. the shell now renders a contextual section-tab row beneath the top bar so each major area owns its own second-level navigation
4. active-state logic now comes from one declarative navigation registry in the frontend rather than being duplicated across shell components
5. the first-pass area mapping now keeps inbound-request detail and proposal continuity under `Agent`, accounting review surfaces under `Accounting`, inventory review surfaces under `Inventory`, operational workflow entry plus route discovery under `Operations`, and privileged maintenance under `Admin`
6. focused frontend verification now covers the navigation-model mapping logic directly, while `npm --prefix web run check`, `npm --prefix web run test`, `npm --prefix web run build`, and `go build ./cmd/... ./internal/...` all passed after the refactor

Follow-up still open:

1. later passes may still add desktop sidebar collapse persistence, new area landing routes where the current mapping remains thin, and broader live browser validation evidence on the workflow track
