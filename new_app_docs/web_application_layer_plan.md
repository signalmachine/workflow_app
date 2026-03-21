# workflow_app Web Application Layer Plan

Date: 2026-03-21
Status: Planned v1 implementation slice
Purpose: define the promoted web-layer work required so `workflow_app` is usable as an application in v1 rather than only through service-layer seams, minimal API contracts, or direct developer tooling.

## 1. Promotion decision

The earlier thin-v1 planning set intentionally kept web work minimal and deferred broader operational surfaces to v2.

That is now changed.

Current decision:

1. v1 should include a usable web application layer
2. the web layer should support the real provider-backed AI path, not only report on its outputs after the fact
3. this does not remove the approval, posting, audit, or database-first control model
4. this does mean web work is no longer limited to minimal browser-testing seams

## 2. V1 objective

Land the web-layer foundations and operator-facing surfaces required so users can actually operate the v1 system through the browser with the AI path, review path, approval path, and downstream inspection path all reachable in one coherent application layer.

## 3. V1 scope

In scope:

1. a real HTTP or API application surface
2. session-auth and active-org handling suitable for browser use
3. attachment upload and download contracts on top of the existing attachment persistence model
4. inbound request submission and tracking through browser-usable paths
5. approval queue and approval decision surfaces
6. inbound-request, processed-proposal, document, accounting, inventory, work-order, and audit review surfaces
7. browser-usable initiation and review of the provider-backed AI flow
8. enough navigation, page structure, and operator workflow continuity that the application is usable in practice rather than only technically exposed

Out of scope:

1. portal-product depth
2. polished consumer-chat UX
3. full mobile-product work
4. broad CRM revival
5. broad manual-entry ERP behavior that displaces the AI-agent-first model

## 4. Required web-layer characteristics

The promoted v1 web layer should preserve the core doctrine:

1. AI remains the main operator interface
2. humans still review through explicit queues and read models
3. financially meaningful actions stay approval-gated and auditable
4. the browser layer should orchestrate domain services rather than become a second truth owner
5. the web layer should make the request -> AI -> recommendation -> approval -> document -> posting or execution chain understandable without raw SQL or developer-only tools
6. the promoted web layer should use backend contracts that a later v2 mobile client can also reuse, even if the web and mobile clients differ in interaction design and presentation

## 5. Minimum usable v1 web outcomes

V1 web work is not complete until the following are true:

1. a user can sign in and operate in one active org context through the browser
2. a user can submit an inbound request with attachments
3. a user can observe queued, processing, processed, failed, cancelled, and completed request states
4. a user can inspect resulting AI artifacts, recommendations, and approval needs
5. an authorized user can act on approval queues through the browser
6. downstream documents and review surfaces are reachable without dropping back to service-layer-only tooling
7. the application is usable enough that real operator testing does not depend on bespoke scripts or direct database access

## 6. Sequencing guidance

Recommended order after the remaining Milestone 5 reporting polish:

1. complete the provider-backed AI execution milestone
2. land the minimal API and attachment transport contracts required for the live path
3. build the usable web application layer on top of those live contracts and existing reporting reads
4. keep broad mobile, chat-product, portal, and multimodal-polish work deferred until after this v1 web layer is credible

Execution rule:

1. treat this milestone as an umbrella for multiple small vertical slices, not as one monolithic web build
2. prefer one usable end-to-end operator loop first, then widen surrounding web coverage incrementally on the same shared backend contracts

## 7. Success criteria

This slice is complete only when:

1. the browser layer is a real application surface rather than only a testing seam
2. it supports the live AI path and approval flow end to end
3. it remains aligned with the database-first and approval-first control model
4. the application is materially usable by operators in v1
