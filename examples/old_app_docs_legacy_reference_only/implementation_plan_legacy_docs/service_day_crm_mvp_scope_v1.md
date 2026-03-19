# service_day CRM MVP Scope v1

Date: 2026-03-11
Status: Legacy reference
Purpose: preserve the broader pre-thin-v1 CRM-first acceptance scope for historical context and shipped-slice back-reference.

Legacy note:
1. the active thin-v1 plan no longer treats CRM as the primary v1 module center
2. use this file only when maintaining or interpreting the already-implemented CRM slice

## 1. MVP Goal

The first CRM MVP should allow a service business to:
1. manage accounts and contacts
2. capture leads and convert them into opportunities
3. manage opportunity follow-up and relationship history
4. prepare service-oriented estimates
5. keep a useful customer timeline
6. use bounded AI assistance for summaries, drafts, and next-step suggestions

The CRM MVP is successful when an internal team can run day-to-day pipeline and customer relationship work without needing spreadsheets or disconnected notes, while preserving a clean, low-friction handoff into the later `work_order`-centered operating model.

## 2. Product Position of the MVP

The CRM MVP is:
1. internal-user-first
2. service-business-first, not generic marketing CRM
3. designed to hand work into later `project` and `work_order` flows
4. built with AI assistance, not AI autonomy

Positioning rule:
1. CRM is important, but it is not intended to be the product's strongest long-term differentiator.
2. The CRM layer should make later execution and commercial handoffs cleaner, not compete with the core `work_order` operating model for product identity.

The CRM MVP is not:
1. a marketing automation suite
2. a customer portal
3. a dispatch system
4. a full CPQ engine
5. a full project-delivery module

## 3. Core User Outcomes

The MVP should support these outcomes cleanly:

1. create and maintain customer accounts and contacts
2. capture a new lead with source and ownership
3. qualify or disqualify a lead
4. convert a qualified lead into account/contact/opportunity records
5. keep an explicit, lightweight customer lifecycle visible with clear owner and next step
6. move an opportunity through stages with visible owner and expected close date
7. log calls, meetings, visits, reminders, and notes
8. see all relevant relationship history in one place
9. create a technically solid quote or estimate for service work
10. assign follow-up tasks
11. use AI to summarize notes, draft follow-ups, and propose next actions

## 4. In-Scope Modules and Entities

### 4.1 Accounts

Required capabilities:
1. create and edit accounts
2. support customer, prospect, partner, and branch/location style account records
3. store addresses, phone, email, website, and tags
4. track owner and status
5. track a lightweight customer lifecycle stage
6. keep next-action and last-contact visibility on the relationship record

### 4.2 Contacts

Required capabilities:
1. create and edit contacts
2. link one contact to one or more accounts where needed
3. store role/title, communication details, and preferences
4. mark primary contact relationships

Recommended post-MVP extension:
1. keep deeper contact-edit, merge/deduplication, and richer relationship-role maintenance as later phased work after the explicit many-to-many link baseline is stable

### 4.3 Leads

Required capabilities:
1. capture source, owner, qualification state, and notes
2. convert into account/contact/opportunity
3. record disqualification reason when not converted
4. preserve lead history after conversion

### 4.4 Opportunities

Required capabilities:
1. create and edit opportunities
2. assign owner
3. track stage, amount/value, expected close date, probability, and service category
4. link to account and primary contact
5. support notes, activities, and estimate linkage

### 4.4A Customer Lifecycle

Required capabilities:
1. represent a lightweight lifecycle across prospect, active sales pursuit, customer, and inactive states without introducing a separate heavyweight module
2. keep lifecycle stages configurable through lookup data rather than hard-coded per-market assumptions
3. record lifecycle transition timestamps and enough history to explain how an account reached its current state
4. keep lifecycle progression linked to lead conversion, opportunity creation, and later commercial activity where relevant

Recommended v1 limit:
1. no separate customer-success workspace
2. no marketing-automation scoring or campaign orchestration
3. no region-specific lifecycle taxonomy baked into the core model

### 4.5 Activities and Notes

Required capabilities:
1. log calls, meetings, visits, emails, and reminders
2. capture freeform notes
3. support due date, responsible person visibility, and completion state where the activity itself is a factual reminder record
4. attach notes and activities to lead, account, contact, and opportunity context
5. keep activities distinct from shared workflow tasks so factual history and accountable follow-up are not conflated

### 4.6 Communications Baseline

Required capabilities:
1. manually capture communication records
2. record direction, subject/summary, date, participants, and linked context
3. support future imported-message flows without redesigning the table shape

Recommended v1 limit:
1. manual communication logging first
2. optional lightweight email import later, not required for MVP acceptance

### 4.7 Tasks

Required capabilities:
1. create user tasks manually
2. allow AI-proposed tasks
3. track due date, assignee, status, and linked context
4. allow assignment to one person or one team queue as the primary actionable owner
5. keep one primary context while allowing secondary related links where project, work order, site, or asset visibility becomes relevant
6. support follow-up workflows from leads and opportunities

### 4.8 Estimates

Required capabilities:
1. create service-led estimates
2. support line items for services, milestones, scoped work, or stocked items when the business also sells inventory directly
3. link estimate to opportunity
4. maintain draft and final-ready states
5. capture quote number, issue date, expiry date, currency, tax, discount, subtotal, and total
6. preserve revision-safe progression so later changes do not silently overwrite prior commercial state
7. keep clean conversion seams to later invoice, project, or work-order flows without forcing those flows into MVP scope

Recommended v1 limit:
1. no deep quote configuration engine
2. no automatic project/work-order creation at MVP acceptance
3. no approval-heavy quote workflow as an MVP requirement
4. no region-specific quoting assumptions baked into the core commercial document shape
5. no advanced retail or trading-company pricing depth as an MVP requirement

Later-phase extension note:
1. quote approvals, richer commercial templates, and broader downstream document conversion can land after MVP on top of the immutable revision-safe baseline rather than replacing it
2. stocked-item quoting in MVP should remain compatible with a later limited inventory-trading extension so direct-sale commercial flows do not need a second estimate or invoice model

### 4.9 Timeline

Required capabilities:
1. show activities, communications, notes, tasks, and estimates in one relationship timeline
2. aggregate by linked customer and opportunity context
3. remain append-oriented, not the owner of record truth

### 4.10 Attachments

Required capabilities:
1. attach files to CRM records
2. support notes and estimate-related attachments
3. retain metadata and access rules

### 4.11 Search

Required capabilities:
1. search accounts, contacts, leads, opportunities, and activities
2. support fast lookup by name, phone, email, and key reference fields
3. support tenant-safe filtering

## 5. AI Scope Inside CRM MVP

AI should be useful from the MVP, but bounded.

### 5.1 In Scope

1. note cleanup
2. call summary generation
3. opportunity health summary
4. suggested follow-up task
5. draft customer reply
6. draft estimate from notes
7. timeline summary across related records

### 5.2 Out of Scope

1. silent customer communication sending
2. silent record mutation without explicit tool path
3. autonomous pipeline progression
4. autonomous commercial approval

### 5.3 Acceptance Rule

Every AI-assisted action must be:
1. reviewable by a human
2. auditable
3. traceable to an explicit run, recommendation, or approval record

## 6. UX Rules for the MVP

1. CRM should feel like the customer-context entry layer of the app, not a disconnected product competing with execution.
2. Account and opportunity pages should show activity and communication context without excessive navigation.
3. Opportunity workflow should be clear enough for a small team without CRM-specialist training.
4. Estimate creation should feel connected to the opportunity, not like a disconnected finance screen.
5. AI assistance should appear as help, not as a second hidden workflow system.

## 7. Suggested MVP Pages or Screens

1. account list
2. account detail
3. contact list/detail
4. lead list/detail
5. opportunity list
6. opportunity detail
7. task list
8. estimate list/detail
9. unified quick search

## 8. Acceptance Workflows

The MVP should be accepted only if these workflows are demonstrably working.

### Workflow 1: Lead to Opportunity

1. user creates lead
2. user qualifies lead
3. lead converts into account/contact/opportunity
4. original lead history remains traceable

### Workflow 2: Relationship Follow-Up

1. user logs meeting or call
2. user creates a follow-up task
3. task appears in the user task view
4. activity appears in the customer timeline

### Workflow 3: Opportunity Management

1. user creates opportunity for an account
2. user updates stage and expected close date
3. user adds notes and activities
4. opportunity history remains visible in one place

### Workflow 4: Estimate Preparation

1. user creates estimate from an opportunity
2. user adds service or milestone lines
3. estimate totals and dates are calculated and saved correctly
4. estimate remains retrievable as a distinct commercial state
5. estimate appears in relevant customer/opportunity context

### Workflow 5: AI-Assisted Follow-Up

1. user records notes or communication history
2. AI produces summary and next-step suggestion
3. user accepts or discards the suggestion
4. accepted output is persisted through normal product flows

## 9. Roles in MVP

Minimum internal roles should cover:
1. admin
2. sales/relationship owner
3. manager

Role expectations:
1. admins manage setup and broad access
2. sales/relationship owners manage their records and tasks
3. managers can review broader pipeline activity

## 10. Non-Goals

These should not be treated as MVP blockers:

1. customer portal
2. work-order execution UI
3. project delivery UI
4. timesheets
5. collections workflows
6. GST/TDS logic
7. dispatch planning
8. marketing campaigns
9. approval-heavy quote workflow

## 11. Data and Architecture Requirements

The CRM MVP must respect the broader platform architecture.

1. All tenant-relevant records must be tenant-safe.
2. CRM records must retain clean seams for later project/work-order conversion.
3. Timeline should be derived or append-oriented.
4. AI writes must go through explicit contracts.
5. Attachments must remain reusable across modules.
6. business-managed progression models such as customer lifecycle and opportunity stages should be configurable through lookup data
7. monetary calculations, revision integrity, and downstream commercial linkage must remain enforced by schema and application logic
8. CRM quality is judged partly by how tightly it integrates with later project and `work_order` flows, not only by isolated CRM usability

## 12. Testing Expectations

Minimum expected test coverage should include:
1. tenant isolation across account/contact/lead/opportunity records
2. lead conversion correctness
3. customer lifecycle stage progression and transition-history correctness
4. activity and task linking correctness
5. estimate lifecycle baseline, including totals and revision-safe state progression
6. search correctness on core entities
7. AI approval/tool-path enforcement for CRM AI features

## 13. Exit Criteria

The CRM MVP is complete when:
1. a service business can run lead, contact, account, and opportunity workflows end to end
2. relationship history is visible and useful
3. customer lifecycle state is explicit enough to support day-to-day relationship ownership and follow-up
4. service-oriented estimates can be created and tracked as sound commercial records
5. AI assistance is genuinely useful but safely bounded
6. the module remains cleanly prepared for later project and work-order execution

## 14. Immediate Follow-On After CRM MVP

The next layer after CRM MVP should be:
1. project creation baseline
2. work-order creation baseline
3. worker assignment and time-entry baseline
4. delivery costing baseline

That sequence preserves continuity from sales into service execution without reworking CRM foundations.
