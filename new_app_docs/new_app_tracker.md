# workflow_app Tracker

Date: 2026-04-03
Status: Thin-v1 is complete through Milestone 9, the active implementation phase is now an ambitious best-practice v2 phase across the implemented Milestone 10 and Milestone 11 browser history, the implemented Milestone 12 admin-maintenance program, and the newly planned Milestone 13 Svelte web migration, the rebuilt modular web bundle now covers the operator-entry surfaces, the promoted review workbench family, the promoted detail-route family, the grouped landing pages for operations, review, and inventory, the searchable route catalog plus utility surfaces, and the new role-aware operator home, `/app/settings` stays explicitly user-scoped while `/app/admin` becomes a grouped privileged maintenance hub with bounded accounting, party, access, and inventory maintenance seams on the shared backend, the approved next browser architecture is now the Svelte-based web replacement on the same Go backend and auth foundation, active implementation posture now explicitly allows stronger capability, stronger architecture, and proactive refactor or rebuild work when justified by established best practices, and Milestone 10 browser-review plus workflow-continuity evidence still needs to catch up on the separate `docs/workflows/` track
Purpose: track the `workflow_app` plan and guard against scope drift during bootstrap and implementation.

## 1. Current status

## 1.0 Browser-plan migration rule

The browser-plan split is now explicit:

1. `docs/svelte_web_guides/svelte_web_ui_migration_plan.md` is the active forward plan for browser implementation work
2. the top-level `new_app_docs/` browser milestone plans for Milestone 10, Milestone 11, and the ERP-style density correction remain useful historical records for route scope, operator-flow intent, and visual lessons, but they are no longer active stack instructions
3. any unfinished or follow-on browser capability should now be planned and implemented through the Svelte migration surface unless the canonical planning set is intentionally changed again

## 1.1 Active posture update

The repository is no longer operating under a thin-scope mindset.

Active implementation rule:

1. thin-v1 is complete and should be treated as historical foundation status, not as the limiter on active implementation choices
2. the current phase is ambitious and best-practice-driven: contributors may implement stronger architecture, stronger product depth, and broader capability when justified
3. `go wild` means freedom to apply established best practices in business software and software engineering without artificial narrowness
4. `go wild` does not mean experimental or novelty-driven implementation that weakens clarity, maintainability, auditability, or workflow control
5. contributors should continuously review the codebase for refactor or rebuild opportunities and either execute that work when appropriate or promote it explicitly into the canonical planning docs
6. large monolithic files and other `God` files should be treated as active refactor or rebuild targets rather than as fixed background debt

| Item | Status | Notes |
| --- | --- | --- |
| New thin-v1 reset accepted | done | The reset plan was accepted, implemented through Milestone 0 through Milestone 8, and closed through explicit checkpoint review |
| New canonical module map | done | `workflow_app` should start without a primary CRM module |
| New thin-v1 scope boundary | done | Thin-v1 scope has been restated for `workflow_app` |
| New execution path | done | Milestone order exists for `workflow_app` |
| Thin-v1 quality rule | done | `workflow_app` is explicitly foundation-heavy and rigor-heavy even though breadth stays narrow |
| Current-codebase v1 gap review | done | Main missing foundation gaps versus the new thin-v1 plan have been documented |
| V2 breadth parking lot | done | Deferred v2 capabilities now have their own plan folder so they do not leak into v1 |
| Foundational document and posting bridge clarified | done | Minimum document families and cross-module posting path are explicit for v1 |
| Multi-agent stance clarified | done | Multi-agent architecture remains in v1 at bounded foundation depth; advanced autonomy is deferred to v2 |
| Implementation defaults captured | done | Locked defaults now exist as a canonical active doc for implementation decisions |
| Foundation coverage checklist captured | done | V1 completion now has an explicit foundation-complete checklist |
| Milestone 0 bootstrap | done | Go module, migration runner, env template, and control-boundary migrations are implemented and verified against primary and test databases |
| Milestone 1 document and approval kernel | done | Shared document identity, approvals, approval queue, decisions, sessions, role-aware service boundaries, and the AI run, tool-policy, artifact, recommendation, delegation trace, inbound-request, and attachment foundations are implemented with queue-oriented request processing seams and reporting-visible causation |
| Milestone 2 accounting foundation | done | Ledger accounts, append-only journal entries and lines, document-linked centralized posting, reversal entries, GST/TDS tax foundation records, accounting periods, effective-date posting control, journal review queries, and control-account balance views are implemented and covered by integration tests |
| Milestone 3 inventory foundation | done | The inventory foundation now includes `inventory_ops` items, locations, movement numbering, append-only movements, derived stock balances, inventory-owned document payload and line records, explicit execution and accounting handoffs, and costed inventory-accounting handoffs consumed through centralized journal posting covered by integration tests |
| Milestone 4 execution foundation | done | `work_orders` now includes first-class work-order records, append-only execution status history, transactional consumption of pending inventory execution links into work-order material-usage truth, workflow-owned work-order tasks with one accountable worker, workforce-owned labor capture with cost snapshots, and centralized accounting consumption of both labor and work-order-linked inventory handoffs covered by integration tests |
| Milestone 5 review and reporting surfaces | done | `reporting` now exposes approval queue, document, accounting journal review, control-account balance review, GST/TDS tax summaries, inventory stock, inventory movement review, inventory reconciliation review, work-order, audit lookup, inbound-request, and processed-proposal review surfaces covered by integration tests; stable inbound-request references now exist for operator tracking and submission acknowledgments, inbound-request list filtering now supports exact `REQ-...` reference lookup, request detail and processed-proposal reads resolve by stable `REQ-...` reference inside the authorized reporting read path instead of depending on raw UUID-only lookup, inbound-request review now also surfaces persisted cancellation and failure reasons with their timestamps for operator troubleshooting plus submitter, session, metadata, attachment provenance, AI step and delegation detail, AI artifact detail, and recommendation payload context, and queue-oriented reporting summaries now provide status-count rollups for inbound requests and processed proposals, so remaining v1 work has moved from reporting polish to provider-backed AI execution and the web layer |
| Post-checkpoint validation Step 1 live-provider verification | done | On 2026-03-28 the OpenAI Responses loop was hardened to use provider-safe tool names plus stateless continuation compatible with `store: false`; `go build ./...`, `set -a; source .env; set +a; go test -p 1 ./...`, and `set -a; source .env; set +a; go run ./cmd/verify-agent` all passed after that fix |
| Workflow validation track handoff | done | Active workflow testing, live review, and readiness evidence now move to `docs/workflows/workflow_validation_track.md` so `new_app_docs/` stays focused on implementation planning. Future workflow-review findings should add bounded fix plans back into `new_app_docs/` before implementation begins. |
| Post-checkpoint validation Step 2-5 browser and workflow checks | deferred_to_workflow_track | The active validation order, deferred workflow backlog, and issue-handling rule now live in `docs/workflows/workflow_validation_track.md`; use that track plus `docs/workflows/end_to_end_validation_checklist.md` for future live workflow review rather than treating this tracker row as the active testing plan. |
| Web visual refresh slice | implemented_with_follow_up_required | The shared web template now uses a low-glare slate-and-blue enterprise visual system with refreshed sans-serif typography, navigation, cards, forms, tables, and targeted page-hierarchy updates on `/app`, the current sign-in surface, `/app/inbound-requests/{request_reference_or_id}`, `/app/review/inbound-requests`, and `/app/review/proposals`, backed by focused `internal/app` HTTP test coverage plus repository build and test verification. Implementation review then found two bounded follow-up issues: canonical docs currently claim `/app/login` as a GET review surface even though the current implementation still renders sign-in only through unauthenticated `GET /app`, and the new shared table minimum-width rule likely causes narrow-width overflow on several non-targeted pages that still lack matching containment. These issues are now planned in `thin_v1_archive/web_visual_refresh_follow_up_plan.md`, and bounded browser-review evidence is still pending until that corrective slice lands. |
| Web visual refresh follow-up corrective slice | implemented_with_browser_review_pending | `thin_v1_archive/web_visual_refresh_follow_up_plan.md` is now implemented in code: `GET /app/login` renders the shared sign-in surface, authenticated `GET /app/login` requests redirect to `/app`, the shared table minimum-width rule now applies only inside `.table-wrap`, and focused `internal/app` HTTP tests cover both fixes. Bounded manual browser-review evidence on desktop and narrow-width layouts is still pending before the later browser-surface restructuring proceeds. |
| Web UI streamlining plan | superseded_as_problem_statement | `thin_v1_archive/web_ui_streamlining_plan.md` still captures the density and hierarchy problems visible in the current browser layer, but it is no longer the preferred implementation path. The repository now promotes a full web-layer rebuild through `milestone_10_web_rebuild_plan.md` instead of another bounded cleanup pass on the current monolithic template structure. |
| Milestone 10 web rebuild planning | implemented_in_code_with_closeout_pending | `milestone_10_web_rebuild_plan.md` is now the active implementation record for a landed-in-code browser rebuild across all three promoted slices. The modular embedded bundle under `internal/app/web_templates` now covers operator-entry routes, review workbench routes, and the promoted detail route family, so the remaining Milestone 10 work is bounded browser-review and workflow-continuity closeout on the separate `docs/workflows/` track rather than more open-ended browser implementation. |
| Milestone 10 Slice 1 architecture and operator-entry implementation | implemented_with_browser_review_pending | The modular embedded bundle under `internal/app/web_templates` remains the active shell and shared-primitive baseline for `/app`, `/app/login`, `/app/submit-inbound-request`, `/app/operations-feed`, and `/app/agent-chat`. Phase-labeled production filenames are now explicitly disallowed in the canonical docs and contributor guidance, and the shared bundle code now also uses responsibility-based naming rather than milestone-labeled hooks. `go build ./cmd/... ./internal/...`, focused `go test ./internal/app -run '^TestHandleWeb' -count=1`, and the canonical `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` verification all passed after the review-workbench migration. Bounded manual browser review on the rebuilt operator-entry routes is still pending on the separate workflow-validation track. |
| Milestone 10 Slice 2 review workbench family rebuild | implemented_with_browser_review_pending | The promoted review list family now renders from the rebuilt modular bundle instead of the legacy monolithic template path: `/app/review/inbound-requests`, `/app/review/approvals`, `/app/review/proposals`, `/app/review/documents`, `/app/review/accounting`, `/app/review/inventory`, `/app/review/work-orders`, and `/app/review/audit` now share one calmer review-workbench structure with shared filters, summary treatment, contained tables, and continuity links. The implementation also restored operator-visible continuity required by the DB-backed browser suite across approval, document, accounting, inventory, work-order, and audit pivots. `go build ./cmd/... ./internal/...`, focused `go test ./internal/app -run '^TestHandleWeb' -count=1`, focused DB-backed reruns for the affected browser-reporting coverage, and the canonical `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` command all passed. Bounded desktop and narrow-width browser-review evidence for the rebuilt review routes is still pending before Slice 2 can be treated as fully closed. |
| Milestone 11 operator shell and navigation direction | implemented_in_code_with_validation_follow_up_pending | `milestone_11_operator_shell_and_navigation_plan.md` remains the active browser-direction record, but all three slices are now implemented in code. Slice 1 landed the lighter grouped top-bar shell, Slice 2 landed bundle landing pages at `/app/operations`, `/app/review`, and `/app/inventory`, and Slice 3 now lands the searchable route catalog at `/app/routes`, explicit utility surfaces at `/app/settings` plus access-scoped `/app/admin`, and a role-aware `/app` home that recommends first workflow routes from current workload instead of behaving as a permanently generic dashboard. Focused `go test ./internal/app -run '^TestHandleWeb' -count=1`, `go build ./cmd/... ./internal/...`, `gopls` diagnostics, and the canonical `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` verification all passed after the slice. Milestone 10 browser-review plus workflow-continuity closeout still remains open on the separate `docs/workflows/` track rather than being silently treated as complete. |
| Milestone 11 route-catalog operator-intent corrective slice | done | The bounded follow-up on `/app/routes` is now implemented in code: catalog matching no longer depends on one raw substring, tokenized multi-term queries such as `pending approvals` now resolve against route title, summary, and keywords with simple relevance ordering, and focused `internal/app` web coverage now locks that behavior in place. `go build ./cmd/... ./internal/...`, `gopls` diagnostics on the edited Go files, and the canonical `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` verification all passed for this corrective slice. |
| Milestone 12 admin maintenance and master-data planning | slices_1_through_6_implemented_in_code | `milestone_12_admin_maintenance_and_master_data_plan.md` is now active in implementation: Slice 1 landed in code through an explicit user-scoped `Settings` surface plus a grouped privileged `Admin` maintenance hub, Slice 2 landed in code through `/app/admin/accounting` plus bounded admin-only `/api/admin/accounting/...` seams for ledger-account, tax-code, and accounting-period maintenance, Slice 3 landed in code through `/app/admin/parties` plus bounded admin-only `/api/admin/parties` list, create, filtered list, exact-detail, and exact-detail contact-create maintenance seams, Slice 4 landed in code through `/app/admin/access` plus bounded admin-only `/api/admin/access/users` list and provision flows and `/api/admin/access/users/{membership_id}/role` updates on the shared `identityaccess` seam, Slice 5 landed in code through `/app/admin/inventory` plus bounded admin-only `/api/admin/inventory/items` and `/api/admin/inventory/locations` list and create flows on the shared `inventory_ops` seam, and Slice 6 has now landed in code through bounded active or inactive status controls for ledger accounts, tax codes, parties, inventory items, and inventory locations on those same shared admin seams. |
| Milestone 13 Svelte web migration planning | planned | `milestone_13_svelte_web_migration_plan.md` is now the canonical next-browser implementation milestone, with Slice 1 foundation and shell work, Slice 2 workflow-surface migration, and Slice 3 detail or admin parity plus cutover defined as the next implementation planning surface. |
| ERP-style web-density corrective slice | implemented_in_code | `web_ui_erp_style_density_correction_plan.md` is now implemented in the modular `internal/app` web bundle: the promoted shell uses `Workflow App` left branding with a thinner blue-gray top bar, session controls move to a compact right-side utility path, home plus landing plus utility pages now use plain hyperlink-first route directories instead of summary-card or result-card mosaics, login is simplified to a thinner operator-entry surface, and dedicated workflow pages now default to single-column stacked composition. A later in-place styling refinement has now strengthened that same modular bundle without changing the stack: shared CSS now uses a more deliberate blue-gray plus parchment visual system with stronger shell contrast, clearer route-directory rows, richer route-catalog and utility summaries, more intentional table treatment, and calmer but more distinctive operator-page hierarchy. Focused `internal/app` HTTP coverage plus canonical repo verification for the refinement are complete; bounded browser-review evidence on desktop and narrow width still belongs on the separate `docs/workflows/` track. |
| Milestone 10 Slice 3 detail continuity and closeout rebuild | implemented_with_browser_review_pending | The promoted detail route family now renders from the rebuilt modular bundle instead of the legacy monolithic template path: `/app/inbound-requests/{request_reference_or_id}`, `/app/review/approvals/{approval_id}`, `/app/review/proposals/{recommendation_id}`, `/app/review/documents/{document_id}`, `/app/review/accounting/{entry_id}`, `/app/review/accounting/control-accounts/{account_id}`, `/app/review/accounting/tax-summaries/{tax_code}`, `/app/review/inventory/{movement_id}`, `/app/review/inventory/items/{item_id}`, `/app/review/inventory/locations/{location_id}`, `/app/review/work-orders/{work_order_id}`, and `/app/review/audit/{event_id}` now share the rebuilt shell, detail primitives, and responsive styling. Focused detail-route `internal/app` HTTP coverage passed after the migration, `go build ./cmd/... ./internal/...` passed, the canonical `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` command completed cleanly again, and `gopls` reported no diagnostics on the edited Go rendering entrypoint. Bounded browser review plus final workflow-continuity evidence remain the open Milestone 10 closeout work on the separate workflow-validation track. |
| Milestone 10 closeout render-baseline corrective slice | done | The grouped corrective slice in `milestone_10_closeout_render_baseline_correction_plan.md` is now fully closed in code: `internal/app.renderWebPage` no longer falls back to the retired monolithic `webAppHTML` template, unmapped page-data states now fail explicitly as `500` instead of silently reviving the legacy browser baseline, the old inline browser payload has now also been removed from `internal/app/web.go`, focused `internal/app` render-path coverage passed, `go build ./cmd/... ./internal/...` passed again, `gopls` reported no diagnostics on the edited web rendering files, and the canonical `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` verification completed cleanly. |
| `internal/app` transport-boundary review | third_cleanup_slice_done | The bounded cleanup review now has three implemented slices in code. The first routes repeated workflow-navigation summary composition used by the role-aware home, settings, operations landing, and review landing through one shared `reporting.WorkflowNavigationSnapshot` contract. The second routes operations landing, durable operations feed, and inventory landing composition through dedicated shared `reporting` snapshot contracts instead of leaving those web handlers to orchestrate repeated multi-read reporting fetches directly. The third now routes the remaining dashboard recent-request or proposal composition and agent-chat continuity reads through dedicated shared `reporting.DashboardSnapshot` and `reporting.AgentChatSnapshot` contracts, so `internal/app` stays closer to route, auth, response, and rendering ownership while preserving the same browser behavior. |
| Dedicated inbound-request page and dashboard-only home | done | `thin_v1_archive/operator_communication_and_intake_surfaces_plan.md` is implemented in code: `/app` is dashboard-only, `/app/submit-inbound-request` is the dedicated intake page, and submit or draft-save flows from that page return clear result messaging with exact `REQ-...` continuity plus next-step links back into detail, dashboard, and review. |
| Operations feed surface | done | `GET /app/operations-feed` now provides a durable one-way coordinator or system communication page assembled from current request, proposal, and approval truth with exact workflow continuity links and focused `internal/app` web-test coverage. |
| Agent chat surface | done | `GET /app/agent-chat` now provides a separate coordinator communication surface for guidance-oriented requests on the same persisted inbound-request foundation via a dedicated `agent_chat` channel, with exact request or proposal continuity and focused `internal/app` web-test coverage. |
| `internal/app` test-suite performance pass | done | The bounded performance pass is now complete. The shared DB-backed test harness no longer reruns schema migrations on every `dbtest.Open` call inside one test process, because each test already performs a full data reset. That removes repeated no-op migration work from the large `internal/app` integration surface and other DB-backed packages without weakening isolation. Closeout also fixed one stale dashboard-shell assertion in `internal/app/web_test.go` so the required repo suite matches the current shared template again. |
| Milestone 9 user-testing readiness hardening | done | Milestone 9 is now complete in `thin_v1_archive/milestone_9_user_testing_readiness_hardening_plan.md`, and the explicit implementation-versus-plan review was also completed cleanly on 2026-03-29. Slice 1 auth hardening is complete: browser-session and bearer-session issuance now require a password-backed credential check on the shared `identityaccess.users` record while preserving one session foundation for the browser and the later mobile client on the same `/api/...` seam. Slice 2 bounded AI capability expansion is complete: the OpenAI-backed coordinator keeps the same bounded coordinator-plus-specialist architecture but now exposes request-scoped read-only tools for current inbound-request detail and current processed-proposal continuity, so provider execution can gather more request-relevant review context than queue summary alone without widening into write-capable autonomy or weakening tool-policy enforcement. Slice 3 shared web or API seam decomposition is complete: `internal/app` is split by seam into dedicated API session, inbound, review, and approval handler files plus dedicated web session or inbound and review-surface files, reducing the largest maintenance-risk concentration without changing the shared HTTP contracts or the thin-v1 web stack. Closeout verification passed on 2026-03-29: `go build ./cmd/... ./internal/...`, `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`, and `set -a; source .env; set +a; go run ./cmd/verify-agent` all completed cleanly. The required post-closeout review then found no material drift across the auth, AI, or seam-decomposition slices, and a fresh `set -a; source .env; set +a; go run ./cmd/verify-agent` rerun also completed cleanly on 2026-03-29 with `REQ-000001` processed through a completed coordinator run and a request-specific urgent warehouse-pump recommendation summary, so the repository should now continue the paused post-checkpoint live validation slice rather than reopen Milestone 9 hardening. Anything under `examples/` remains read-only reference material and is not part of the active `workflow_app` codebase. The historical accounting-agent proof-of-concept remains an external comparison point at https://github.com/signalmachine/accounting-agent-app rather than an in-repo reference tree. |
| Milestone 6 provider-backed AI execution | done | `internal/ai` now includes optional OpenAI provider configuration loading, the official OpenAI Go SDK, a Responses-API-backed provider adapter, and a coordinator flow that can claim one queued inbound request, assemble request, attachment, and derived-text context, run a hard-capped tool loop, enforce per-capability tool policy, auto-execute the first reporting read tool when policy allows, optionally route the result through one allowlisted specialist capability with a durable child run and delegation record, persist the resulting coordinator and specialist steps with tool-execution metadata, write a provider brief artifact and operator-review recommendation, and mark the request `processed` or `failed` without bypassing existing control boundaries. `internal/app` now exposes a shared backend seam over that path plus submission, attachment transport, operator review, approval decisions, and browser-usable session auth: `POST /api/session/login` now starts an org-scoped browser session from org slug plus user email and sets the session cookies, `GET /api/session` resolves the active browser session, `POST /api/session/logout` revokes that session, `POST /api/agent/process-next-queued-inbound-request` processes the next queued request, `POST /api/inbound-requests` persists the initial request message plus optional inline attachments and queues the request in one backend workflow, `GET /api/attachments/{attachment_id}/content` serves persisted attachment bytes back through the same auth boundary, `GET /api/review/inbound-requests`, `GET /api/review/inbound-request-status-summary`, `GET /api/review/inbound-requests/{request_reference_or_id}`, `GET /api/review/processed-proposals`, `GET /api/review/processed-proposal-status-summary`, and `GET /api/review/approval-queue` surface browser-usable operator review reads backed by `reporting`, and `POST /api/approvals/{approval_id}/decision` routes approval actions through the existing workflow control boundary. `cmd/verify-agent` still provides opt-in live-provider verification on the same seam, `cmd/app` serves the widened runnable API surface, and provider-gated plus API integration coverage now exists; see `thin_v1_archive/ai_provider_execution_plan.md` |
| Milestone 7 usable web application layer | done | the `/app` browser surface now covers the first operator loop plus downstream review continuity on the same shared backend seam. In addition to the already-landed sign-in, request submission, queue processing, approval actions, and document or accounting or inventory or work-order or audit review surfaces, the final closeout sweep is now complete: operators can save new requests as drafts, continue draft editing, add draft attachments, queue a draft from the browser, cancel queued pre-processing requests, return queued or cancelled requests to draft for amendment, hard-delete unprocessed drafts, and use the dashboard plus full inbound-request review as strong browser entry points for parked, failed, cancelled, in-flight, processed, and completed requests. The closeout sweep also fixed the last browser continuity gap by carrying persisted cancellation and failure timestamps plus reasons through exact inbound-request detail and filtered inbound-request review, and browser integration coverage now exercises both parked-request lifecycle management and full request-status visibility. Milestone 8 is now complete, so later backend auth work should follow the additive path documented in `thin_v1_archive/non_browser_auth_evolution_plan.md` rather than reopening the browser milestone. |
| Milestone 8 first shared-API lifecycle hardening slice | done | the first client-neutral backend-hardening pass is now landed on the shared `/api/...` seam: strict JSON decoding now rejects malformed or unknown request bodies across session login, queued-request processing, inbound-request submission and mutation, and approval-decision paths, and inbound-request submission plus draft or queue or cancel or amend responses now return stable lifecycle metadata including timestamps plus cancellation or failure reasons so later non-browser clients do not need browser-specific follow-up assumptions to understand mutation results |
| Milestone 8 review-read contract hardening slice | done | exact-ID review filters across the shared `/api/review/...` seams now fail cleanly with `invalid review filter` instead of leaking database-cast errors when clients send malformed UUID-like lookup values, and API integration coverage now exercises malformed exact-ID filters across approval-queue, document, accounting, inventory, work-order, processed-proposal, and inbound-request review paths |
| Milestone 8 attachment contract hardening slice | done | attachment upload now validates media-type metadata before persistence, malformed attachment IDs now fail cleanly with `invalid attachment` instead of leaking database errors, authenticated attachment downloads now return explicit `Content-Disposition`, `Cache-Control: private, no-store`, and `X-Content-Type-Options: nosniff` headers, and API integration coverage now exercises both the hardened upload and download behavior |
| Milestone 8 approval-action contract hardening slice | done | approval decisions now reject malformed approval IDs as `invalid approval`, keep body validation aligned with the other hardened JSON endpoints, and return current approval plus document state metadata on conflict responses so later non-browser clients can recover without an immediate follow-up read |
| Milestone 8 non-browser auth-evolution planning slice | done | `thin_v1_archive/non_browser_auth_evolution_plan.md` now closes the fifth planned slice by defining an additive bearer-session path for later non-browser clients on top of the existing `identityaccess.sessions` foundation while keeping browser-session cookies as the active v1 auth path and treating UUID actor headers as pre-production automation compatibility rather than the long-term client contract |
| Post-Milestone-8 first additive non-browser auth slice | done | the shared `/api/...` seam now supports additive non-browser bearer sessions on the same `identityaccess.sessions` truth: `POST /api/session/token` issues a device-scoped JSON session plus short-lived bearer access token and refresh token, `POST /api/session/refresh` rotates refresh material and access tokens, `GET /api/session` and `POST /api/session/logout` now accept bearer auth as well as browser cookies, bearer-authenticated shared API writes now resolve through the same actor and authorization rules, and integration coverage plus migration verification are landed |
| Post-Milestone-8 actor-header narrowing slice | done | general shared `/api/...` reads and writes now require browser-session cookies or bearer-session auth, while the UUID actor-header compatibility path remains only on `POST /api/agent/process-next-queued-inbound-request` as a narrow pre-production automation seam with integration coverage proving the broader shared routes reject header-only auth |
| Post-Milestone-8 queued-agent auth retirement slice | done | `POST /api/agent/process-next-queued-inbound-request` now also requires browser-session cookies or bearer-session auth, the last UUID actor-header compatibility seam is retired, integration coverage now exercises queued-request processing through bearer auth, and header-only calls are rejected once session auth is available |
| Minimum thin-v1 party and contact support depth | done | `parties` support records now cover external party identity plus support-depth contacts with tenant-safe service boundaries and integration tests |
| Remaining thin-v1 adopted-document gaps | done | thin v1 adopted document-family ownership is now implemented for work-order, invoice, and payment or receipt document families through module-owned one-to-one payload bridges keyed by `document_id`; see `thin_v1_archive/adopted_document_ownership_remediation_plan.md` |
| Minimum thin-v1 inbound-request and browser-ingress foundation | done | persist-first inbound requests, request messages, PostgreSQL-backed attachments, transcription derivatives, queue claim and status transitions, stable `REQ-...` references, draft editing and hard deletion, queued-request amend-back-to-draft support, AI run causation, and reporting-visible inbound-request and processed-proposal review now exist for thin-v1 browser testing at the service and reporting-read-model level; see `thin_v1_archive/inbound_request_and_attachment_foundation_plan.md` |
| Thin-v1 checkpoint closeout | done | `go build ./...` and `set -a; source .env; set +a; go test -p 1 ./...` both completed cleanly on 2026-03-27, and the resulting review against `new_app_foundation_coverage.md`, this tracker, and the active milestone docs found no material missing foundation slice at the current thin-v1 depth |
| Thin-v1 marked complete | done | The canonical planning set now treats thin-v1 as complete foundation work rather than the active implementation phase |
| V2 implementation phase start | in_progress | Starting at Milestone 10, active implementation work is v2 work aimed at broader application enhancement and production readiness on top of the completed shared foundation; the first live code slice is the Milestone 10 Slice 1 web architecture and operator-entry rebuild |

## 2. Immediate next steps

Thin-v1 is complete, and the next active implementation phase is now explicit.

Planned next step:

1. treat Milestone 9 as closed after the successful 2026-03-29 closeout verification
2. treat the required same-day Milestone 9 implementation review against the milestone plan and related canonical planning docs as complete with no material drift recorded
3. treat the post-review `set -a; source .env; set +a; go run ./cmd/verify-agent` rerun as complete, with the live provider seam reconfirmed at the paused-validation start point
4. treat the bounded repo-verification pass for the Phase 1 foundational workflows as complete on 2026-03-29 through end-to-end `internal/app` integration coverage on the shared `/api/...` plus `/app/...` seams
5. treat the bounded shared-backend correctness-hardening slice as landed and verified on 2026-03-30
6. the bounded test-harness hardening slice around disposable test-database advisory-lock behavior and stale-session diagnostics is now complete
7. `go build ./cmd/... ./internal/...` plus the canonical `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` verification are now complete for that slice
8. the bounded corrective slice in `thin_v1_archive/web_visual_refresh_follow_up_plan.md` is now complete
9. `thin_v1_archive/operator_communication_and_intake_surfaces_plan.md` is now fully implemented in code through the dashboard-only home, dedicated intake page, durable operations feed, and dedicated agent-chat surface
10. `go build ./cmd/... ./internal/...` plus the canonical `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` verification also completed cleanly for that browser-surface slice on 2026-03-30
11. the bounded performance pass on the `internal/app` test suite is now complete through one shared-harness speedup: DB-backed tests now reuse one migrated schema per test process while preserving per-test resets and advisory-lock discipline
12. workflow validation remains on the separate `docs/workflows/` track and is now the next active validation track unless that validation feeds a new bounded fix slice back into `new_app_docs/`
13. the next active browser implementation milestone is now Milestone 13, and it should be treated as the canonical Svelte migration program rather than as another extension of the old Go-template browser stack
14. if the team chooses the next promoted browser implementation work rather than resuming ad hoc workflow validation first, the correct planning shape is now Milestone 13 Slice 1 foundation and shell work rather than additional old-stack browser cleanup

Current next-step order:

1. treat the completed Milestone 10 rebuild, the full Milestone 11 slice set, the route-catalog corrective slice, the ERP-style density-correction slice, and the bounded `internal/app` transport-boundary cleanup slice as landed implementation state rather than pending implementation work
2. keep Milestone 10 closeout on the separate `docs/workflows/` track, and run it as one larger browser-review plus workflow-continuity sweep across the full promoted route family instead of many tiny disconnected validation passes
3. use the current promoted browser baseline for that sweep: `/app/login`, `/app`, `/app/routes`, `/app/settings`, `/app/admin` for an admin actor, `/app/admin/accounting` for an admin actor, `/app/admin/parties` for an admin actor, `/app/admin/access` for an admin actor, `/app/admin/inventory` for an admin actor, `/app/operations`, `/app/review`, `/app/inventory`, `/app/submit-inbound-request`, `/app/operations-feed`, `/app/agent-chat`, the promoted review routes, and the promoted detail-route family
4. keep Milestone 12 treated as landed implementation state except for later policy or operational controls that may still be promoted on the shared backend
5. keep the separate Milestone 10 closeout sweep on the workflow track, but treat it as parity and validation input to Milestone 13 rather than as a blocker on beginning the Svelte foundation work
6. if that workflow sweep exposes real product defects or missing support seams, group tightly related findings into one bounded corrective slice or absorb them into the relevant Milestone 13 slice instead of reopening scattered old-stack browser follow-up work
7. direct any new browser implementation into the Milestone 13 Svelte plan unless the canonical planning set is intentionally changed again
8. do not silently widen the landed Go-template browser-navigation and density-correction work; new browser architecture work should now happen only on the Svelte path
9. keep using `new_app_implementation_defaults.md` as the default-rules reference, `new_app_foundation_coverage.md` as the completed thin-v1 coverage control, and `docs/workflows/` as the active workflow-validation surface

Follow-on rule:

1. keep Milestone 8 closed as a bounded backend-hardening milestone unless a later review promotes one explicit new client-neutral slice
2. treat the additive bearer-session and queued-agent auth-retirement slices as complete unless follow-up verification exposes a concrete defect
3. if backend auth work continues later, do not reintroduce UUID actor-header compatibility; promote only new client-neutral auth work that materially improves correctness or reuse
4. keep widening or correcting the shared backend only in client-neutral slices that strengthen correctness, continuity, or reuse rather than creating a browser-specific versus mobile-specific split
5. if later work exposes a browser-layer regression or a newly discovered residual Milestone 7 blocker, document it explicitly and fix it narrowly rather than reopening broad browser-surface expansion
6. keep the codebase centered on the approved first-class modules while allowing support-depth records such as `parties` and `contacts` where the canonical module-boundary doc explicitly permits them
7. add attachments only where they support approval evidence, document support flows, or persisted inbound request intake
8. use `thin_v1_archive/new_app_v1_gap_review_from_current_codebase.md` as historical context only, not as the live list of remaining missing foundation areas
9. use `new_app_implementation_defaults.md` as the default-rules reference during implementation
10. use `new_app_foundation_coverage.md` as the v1 completion checklist and foundation coverage control
11. treat the earlier Go-template browser layer as an implemented interim phase and use the approved Svelte migration plan as the active web-direction reference for forward browser work

## 2.1 Current implementation order

Recommended sequence from the current repository state:

1. treat the current repository state as completed thin-v1 foundation plus implemented Milestone 10 rebuild and implemented Milestone 11 shell or navigation follow-on work rather than as an implicitly unfinished milestone chain
2. keep the separate Milestone 10 browser-review and workflow-continuity closeout sweep as the next active validation step on `docs/workflows/`
3. begin browser implementation planning and execution through Milestone 13 rather than reopening the old browser-stack milestone docs
4. if the Milestone 10 validation sweep exposes a real product defect or missing support seam, promote one bounded corrective slice or absorb it into the matching Milestone 13 slice and rerun the affected validation
5. keep each newly promoted v2 slice bounded to one coherent concern rather than reopening completed milestone buckets
6. keep richer draft-attachment editing beyond the landed additive upload flow as residual only if later evidence proves it materially necessary

Reason:

1. the adopted document-family ownership mismatch is closed for work-order, invoice, and payment or receipt families
2. inbound request intake, attachment support, queue claim semantics, and reporting-visible AI causation sit on top of the stabilized document-adoption model
3. the reporting foundation is complete enough for thin-v1 review and browser-ready read seams, and the provider-backed coordinator plus browser-session auth make the shared backend usable from a real browser client
4. the landed coordinator slice includes a hard-capped tool loop with policy-enforced read-tool execution plus bounded specialist delegation while keeping the default contributor workflow provider-independent
5. the real `/app` plus `/api/...` seam is operational for login, intake, queued processing, review reads, and request-detail continuity outside direct service calls
6. the additive non-browser auth slice is landed on that same seam through bearer-session issue, refresh rotation, token-authenticated introspection, token-authenticated logout, shared handler reuse for bearer-authenticated writes, and bearer-authenticated queued-request processing
7. the queued-agent UUID actor-header compatibility seam is retired, so later auth work should not preserve or widen that temporary path again
8. the Milestone 10 modular rebuild, the grouped render-baseline corrective slice, the full Milestone 11 slice set, the route-catalog corrective slice, the ERP-style density-correction slice, and the bounded `internal/app` transport-boundary cleanup slice are all implemented in code
9. the next planned browser implementation gap is now the Milestone 13 Svelte migration itself, while later privileged-maintenance follow-on work remains bounded shared-backend product work rather than old-stack browser work
10. workflow validation continues on its own track rather than being silently mixed back into implementation planning

## 2.1.1 Next-session decision gate

The checkpoint decision is now closed with outcome 1.

1. thin-v1 checkpoint complete
Result:
the full repository verification completed cleanly on 2026-03-27, the foundation checklist still matches the implemented codebase, and no material missing foundation slice remains at the current thin-v1 depth
2. promote one more bounded slice
Result:
the review finds a real remaining gap that is still foundation work or client-neutral shared-backend hardening, and that slice is written into this tracker before implementation begins

The repository should not return to an implicit middle state where it is treated as both complete and not complete at the same time.

## 2.2 Milestone 8 planned slices

Milestone 8 is now explicitly bounded to five planned slices.

Planned slices:

1. shared-API lifecycle contract hardening
Status: done
This slice tightened JSON request validation and made inbound-request mutation responses carry lifecycle metadata directly.
2. review-read contract hardening
Status: done
This slice standardized malformed exact-ID filter handling across the shared review reads so clients now receive `invalid review filter` instead of database-cast failures when they send malformed UUID-like lookup values.
3. attachment contract hardening
Status: done
This slice standardized attachment upload metadata validation, download-header behavior, and malformed attachment-ID handling so later non-browser clients inherit a cleaner bounded attachment seam.
4. approval-action contract hardening
Status: done
This slice now rejects malformed approval IDs as `invalid approval`, keeps approval-decision body validation aligned with the other hardened shared JSON endpoints, and returns current approval plus document state metadata on approval-decision conflicts so later non-browser clients can recover without an immediate follow-up read.
5. non-browser auth-evolution planning
Status: done
This slice is now documented in `thin_v1_archive/non_browser_auth_evolution_plan.md`, which keeps browser-session cookies as the active v1 auth path, treats UUID actor headers as pre-production automation compatibility, and defines the next additive bearer-session path for lightweight non-browser clients on the same backend foundation.

Milestone 8 stop rule:

1. Milestone 8 is complete when the five planned slices above are implemented and reviewed
2. additional Milestone 8 slices should be added only after that review shows a real remaining client-neutral hardening gap
3. do not let Milestone 8 remain an open-ended hardening bucket without updating this planned-slices list explicitly

Current result:

1. this stop rule is now satisfied, so Milestone 8 should be treated as complete until a later review explicitly opens a new backend-hardening milestone or promotes a concrete follow-up slice

## 2.3 Milestone 8 preview

Milestone 8 followed Milestone 7 rather than competing with its remaining browser work.

Planned Milestone 8 focus:

1. client-neutral backend hardening for later lightweight mobile reuse on the same backend foundation
2. explicit contract discipline for shared `/api/...` paths already exercised by the web layer
3. standardization of request-status, review-read, approval-action, and attachment semantics where later non-browser clients would otherwise inherit browser-specific assumptions
4. next-step auth-evolution planning for non-browser clients without replacing the active browser-session model prematurely

Milestone 8 guardrails:

1. do not let Milestone 8 erase or defer the still-pending Milestone 7 browser work
2. do not treat Milestone 8 as the mobile-product build milestone
3. keep mobile UX, full mobile auth-product depth, push behavior, offline behavior, and broader multimodal client breadth outside Milestone 8 unless the canonical planning set is later updated explicitly

## 2.4 Remaining Milestone 7 slice analysis

The Milestone 7 closeout sweep is now complete.

Closeout result:

1. the final structured pass against the browser exit criteria found one real late-stage continuity blocker: exact inbound-request detail and filtered inbound-request review were not carrying persisted cancellation and failure reasons forward even though the dashboard already surfaced them
2. that blocker is now fixed, and browser integration coverage now exercises parked-request lifecycle management plus browser visibility for draft, queued, processing, failed, cancelled, processed, and completed request states
3. Milestone 7 is therefore complete, and that browser closeout no longer blocks the next additive backend-auth slice that should follow the now-complete Milestone 8 review

Planned-slice control rule:

1. treat the active slices above as the current explicit implementation plan
2. if later implementation reveals a concrete browser regression, record it explicitly as residual Milestone 7 follow-up rather than silently reopening Milestone 7 as a broad work bucket
3. do not use the possibility of later residual browser work as a reason to defer Milestone 8

## 3. Scope guardrail

The replacement effort fails if it repeats the old pattern of letting support workflows become the center of gravity.

Do not:

1. add a primary `crm` module
2. add opportunity or estimate breadth as the first workflow center
3. build large human operational UI before review and report surfaces
4. broaden scope because a feature feels commercially attractive

## 4. Success test

The replacement effort is on track only if:

1. documents, approvals, ledgers, inventory, execution, and reports are visibly the center of the model
2. the module list stays narrow
3. every new feature can be justified as foundation, not convenience
