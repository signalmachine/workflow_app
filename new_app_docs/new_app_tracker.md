# workflow_app Tracker

Date: 2026-03-29
Status: Milestone 9 active; slices 1 and 2 complete and post-checkpoint validation still paused after partial completion
Purpose: track the `workflow_app` plan and guard against scope drift during bootstrap and implementation.

## 1. Current status

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
| Post-checkpoint validation Step 2-5 browser and workflow checks | paused | On 2026-03-28 the real browser and API seam was exercised against the configured development database and live provider path: `POST /api/session/login`, `GET /api/session`, `GET /app`, `POST /api/inbound-requests`, `POST /api/agent/process-next-queued-inbound-request`, `GET /api/review/inbound-requests`, `GET /api/review/inbound-request-status-summary`, `GET /api/review/inbound-requests/REQ-000001`, `GET /api/review/processed-proposals`, `GET /api/review/approval-queue`, and `GET /app/inbound-requests/REQ-000001` all returned success responses with the expected persisted request, run, step, artifact, and proposal continuity. The first live blocker from that pass is now cleared: `internal/ai` was hardened so request messages and derived texts are primary evidence, queue-summary context is explicitly secondary, obviously stale `processing`-style briefs are rejected, and the OpenAI provider now gets one bounded repair turn before failing when the first structured brief stays too generic. On 2026-03-28 `go build ./...`, `set -a; source .env; set +a; timeout 600s go test -p 1 ./...`, and `set -a; source .env; set +a; go run ./cmd/verify-agent` all passed after that fix, with the live verification summary returning a request-specific warehouse-pump inspection brief instead of stale queue-status wording. A second bounded shared-backend correctness slice is now also landed inside the same validation phase: processed-proposal review now derives pre-approval document and suggested-queue continuity from recommendation payload when no approval link exists yet, and the shared `/api/...` plus `/app/...` seams now support requesting workflow approval directly from a processed proposal with atomic recommendation-link persistence. `go build ./...`, `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./...`, and targeted web coverage all passed after that change. Failure-path live validation and the remaining draft-amend plus approval-producing live browser workflow passes still remain open, but this validation slice is now intentionally paused until Milestone 9 lands; see `post_checkpoint_validation_and_user_testing_plan.md`. |
| Milestone 9 user-testing readiness hardening | active | The milestone is now active in `milestone_9_user_testing_readiness_hardening_plan.md`. Slice 1 auth hardening is complete: browser-session and bearer-session issuance no longer accept org slug plus email alone, and now require a password-backed credential check on the shared `identityaccess.users` record while preserving one session foundation for the browser and the later mobile client on the same `/api/...` seam. Slice 2 bounded AI capability expansion is now also complete: the OpenAI-backed coordinator keeps the same bounded coordinator-plus-specialist architecture but now exposes request-scoped read-only tools for current inbound-request detail and current processed-proposal continuity, so provider execution can gather more request-relevant review context than queue summary alone without widening into write-capable autonomy or weakening tool-policy enforcement. Slice 3 shared web or API seam decomposition remains pending. The paused live validation slice should resume only after this milestone lands. Anything under `examples/` remains read-only reference material and is not part of the active `workflow_app` codebase. `examples/accounting-agent-app-for-reference-only/` may still be used as a practical comparison point, but it remains a proof-of-concept single-agent reference rather than the architectural target for `workflow_app`. |
| Milestone 6 provider-backed AI execution | done | `internal/ai` now includes optional OpenAI provider configuration loading, the official OpenAI Go SDK, a Responses-API-backed provider adapter, and a coordinator flow that can claim one queued inbound request, assemble request, attachment, and derived-text context, run a hard-capped tool loop, enforce per-capability tool policy, auto-execute the first reporting read tool when policy allows, optionally route the result through one allowlisted specialist capability with a durable child run and delegation record, persist the resulting coordinator and specialist steps with tool-execution metadata, write a provider brief artifact and operator-review recommendation, and mark the request `processed` or `failed` without bypassing existing control boundaries. `internal/app` now exposes a shared backend seam over that path plus submission, attachment transport, operator review, approval decisions, and browser-usable session auth: `POST /api/session/login` now starts an org-scoped browser session from org slug plus user email and sets the session cookies, `GET /api/session` resolves the active browser session, `POST /api/session/logout` revokes that session, `POST /api/agent/process-next-queued-inbound-request` processes the next queued request, `POST /api/inbound-requests` persists the initial request message plus optional inline attachments and queues the request in one backend workflow, `GET /api/attachments/{attachment_id}/content` serves persisted attachment bytes back through the same auth boundary, `GET /api/review/inbound-requests`, `GET /api/review/inbound-request-status-summary`, `GET /api/review/inbound-requests/{request_reference_or_id}`, `GET /api/review/processed-proposals`, `GET /api/review/processed-proposal-status-summary`, and `GET /api/review/approval-queue` surface browser-usable operator review reads backed by `reporting`, and `POST /api/approvals/{approval_id}/decision` routes approval actions through the existing workflow control boundary. `cmd/verify-agent` still provides opt-in live-provider verification on the same seam, `cmd/app` serves the widened runnable API surface, and provider-gated plus API integration coverage now exists; see `ai_provider_execution_plan.md` |
| Milestone 7 usable web application layer | done | the `/app` browser surface now covers the first operator loop plus downstream review continuity on the same shared backend seam. In addition to the already-landed sign-in, request submission, queue processing, approval actions, and document or accounting or inventory or work-order or audit review surfaces, the final closeout sweep is now complete: operators can save new requests as drafts, continue draft editing, add draft attachments, queue a draft from the browser, cancel queued pre-processing requests, return queued or cancelled requests to draft for amendment, hard-delete unprocessed drafts, and use the dashboard plus full inbound-request review as strong browser entry points for parked, failed, cancelled, in-flight, processed, and completed requests. The closeout sweep also fixed the last browser continuity gap by carrying persisted cancellation and failure timestamps plus reasons through exact inbound-request detail and filtered inbound-request review, and browser integration coverage now exercises both parked-request lifecycle management and full request-status visibility. Milestone 8 is now complete, so later backend auth work should follow the additive path documented in `non_browser_auth_evolution_plan.md` rather than reopening the browser milestone. |
| Milestone 8 first shared-API lifecycle hardening slice | done | the first client-neutral backend-hardening pass is now landed on the shared `/api/...` seam: strict JSON decoding now rejects malformed or unknown request bodies across session login, queued-request processing, inbound-request submission and mutation, and approval-decision paths, and inbound-request submission plus draft or queue or cancel or amend responses now return stable lifecycle metadata including timestamps plus cancellation or failure reasons so later non-browser clients do not need browser-specific follow-up assumptions to understand mutation results |
| Milestone 8 review-read contract hardening slice | done | exact-ID review filters across the shared `/api/review/...` seams now fail cleanly with `invalid review filter` instead of leaking database-cast errors when clients send malformed UUID-like lookup values, and API integration coverage now exercises malformed exact-ID filters across approval-queue, document, accounting, inventory, work-order, processed-proposal, and inbound-request review paths |
| Milestone 8 attachment contract hardening slice | done | attachment upload now validates media-type metadata before persistence, malformed attachment IDs now fail cleanly with `invalid attachment` instead of leaking database errors, authenticated attachment downloads now return explicit `Content-Disposition`, `Cache-Control: private, no-store`, and `X-Content-Type-Options: nosniff` headers, and API integration coverage now exercises both the hardened upload and download behavior |
| Milestone 8 approval-action contract hardening slice | done | approval decisions now reject malformed approval IDs as `invalid approval`, keep body validation aligned with the other hardened JSON endpoints, and return current approval plus document state metadata on conflict responses so later non-browser clients can recover without an immediate follow-up read |
| Milestone 8 non-browser auth-evolution planning slice | done | `non_browser_auth_evolution_plan.md` now closes the fifth planned slice by defining an additive bearer-session path for later non-browser clients on top of the existing `identityaccess.sessions` foundation while keeping browser-session cookies as the active v1 auth path and treating UUID actor headers as pre-production automation compatibility rather than the long-term client contract |
| Post-Milestone-8 first additive non-browser auth slice | done | the shared `/api/...` seam now supports additive non-browser bearer sessions on the same `identityaccess.sessions` truth: `POST /api/session/token` issues a device-scoped JSON session plus short-lived bearer access token and refresh token, `POST /api/session/refresh` rotates refresh material and access tokens, `GET /api/session` and `POST /api/session/logout` now accept bearer auth as well as browser cookies, bearer-authenticated shared API writes now resolve through the same actor and authorization rules, and integration coverage plus migration verification are landed |
| Post-Milestone-8 actor-header narrowing slice | done | general shared `/api/...` reads and writes now require browser-session cookies or bearer-session auth, while the UUID actor-header compatibility path remains only on `POST /api/agent/process-next-queued-inbound-request` as a narrow pre-production automation seam with integration coverage proving the broader shared routes reject header-only auth |
| Post-Milestone-8 queued-agent auth retirement slice | done | `POST /api/agent/process-next-queued-inbound-request` now also requires browser-session cookies or bearer-session auth, the last UUID actor-header compatibility seam is retired, integration coverage now exercises queued-request processing through bearer auth, and header-only calls are rejected once session auth is available |
| Minimum thin-v1 party and contact support depth | done | `parties` support records now cover external party identity plus support-depth contacts with tenant-safe service boundaries and integration tests |
| Remaining thin-v1 adopted-document gaps | done | thin v1 adopted document-family ownership is now implemented for work-order, invoice, and payment or receipt document families through module-owned one-to-one payload bridges keyed by `document_id`; see `adopted_document_ownership_remediation_plan.md` |
| Minimum thin-v1 inbound-request and browser-ingress foundation | done | persist-first inbound requests, request messages, PostgreSQL-backed attachments, transcription derivatives, queue claim and status transitions, stable `REQ-...` references, draft editing and hard deletion, queued-request amend-back-to-draft support, AI run causation, and reporting-visible inbound-request and processed-proposal review now exist for thin-v1 browser testing at the service and reporting-read-model level; see `inbound_request_and_attachment_foundation_plan.md` |
| Thin-v1 checkpoint closeout | done | `go build ./...` and `set -a; source .env; set +a; go test -p 1 ./...` both completed cleanly on 2026-03-27, and the resulting review against `new_app_foundation_coverage.md`, this tracker, and the active milestone docs found no material missing foundation slice at the current thin-v1 depth |

## 2. Immediate next steps

The thin-v1 checkpoint closeout is complete, and the next active step is now explicit.

Planned next step:

1. start Milestone 9 user-testing readiness hardening as the next active bounded milestone
2. treat the post-checkpoint validation slice as paused after partial completion rather than silently incomplete
3. land the Milestone 9 slices in the planned order: auth hardening, bounded AI capability expansion, then shared web or API seam decomposition
4. resume the remaining canonical live workflows for draft-amend continuity, approval-producing flow continuity, and failed-provider or failed-processing visibility only after Milestone 9 is complete
5. end the resumed validation slice with one explicit readiness result for supervised AI-backed user testing or with an explicit blocker list if readiness is not yet achieved

Planned next-session review and testing order:

1. start with Milestone 9 planning and implementation rather than continuing the paused live workflows
2. implement the remaining Milestone 9 slice in the planned order recorded in `milestone_9_user_testing_readiness_hardening_plan.md`
3. finish the remaining slice with focused review, tests, and tracker updates before moving to resumed live validation
4. when Milestone 9 closes, return to `post_checkpoint_validation_and_user_testing_plan.md` and resume the deferred live workflows in the documented order
5. update `docs/workflows/application_workflow_catalog.md` and `docs/workflows/end_to_end_validation_checklist.md` later if the durable workflow reference or reusable workflow checklist drifts when the resumed live workflow validation completes

Follow-on rule:

1. keep Milestone 8 closed as a bounded backend-hardening milestone unless a later review promotes one explicit new client-neutral slice
2. treat the additive bearer-session and queued-agent auth-retirement slices as complete unless follow-up verification exposes a concrete defect
3. if backend auth work continues later, do not reintroduce UUID actor-header compatibility; promote only new client-neutral auth work that materially improves correctness or reuse
4. keep widening or correcting the shared backend only in client-neutral slices that strengthen correctness, continuity, or reuse rather than creating a browser-specific versus mobile-specific split
5. if later work exposes a browser-layer regression or a newly discovered residual Milestone 7 blocker, document it explicitly and fix it narrowly rather than reopening broad browser-surface expansion
6. keep the codebase centered on the approved first-class modules while allowing support-depth records such as `parties` and `contacts` where the canonical module-boundary doc explicitly permits them
7. add attachments only where they support approval evidence, document support flows, or persisted inbound request intake
8. use `new_app_v1_gap_review_from_current_codebase.md` as historical context only, not as the live list of remaining missing foundation areas
9. use `new_app_implementation_defaults.md` as the default-rules reference during implementation
10. use `new_app_foundation_coverage.md` as the v1 completion checklist and foundation coverage control
11. keep the thin-v1 web stack unchanged unless the canonical planning set explicitly changes it

## 2.1 Planned next implementation order

Recommended sequence after checkpoint closeout:

1. treat the current repository state as the thin-v1 foundation checkpoint rather than an implicitly unfinished milestone chain
2. begin with the bounded Milestone 9 readiness-hardening plan in `milestone_9_user_testing_readiness_hardening_plan.md`
3. keep the paused post-checkpoint validation slice intact but deferred while Milestone 9 lands
4. after Milestone 9 closes, resume the remaining canonical live workflows for draft-amend continuity, approval-producing flow continuity, and failed-provider or failed-processing visibility
5. once those workflows finish, decide explicitly whether the next session is:
6. one bounded post-checkpoint shared-backend or correctness slice, or
7. one explicitly promoted v2 work item under `new_app_docs/app_v2_plans/`
8. keep each newly promoted slice bounded to one concern rather than reopening completed milestone buckets
9. keep richer draft-attachment editing beyond the landed additive upload flow as residual only if later evidence proves it materially necessary

Reason:

1. the adopted document-family ownership mismatch is now closed for work-order, invoice, and payment or receipt families
2. inbound request intake, attachment support, queue claim semantics, and reporting-visible AI causation now sit on top of the stabilized document-adoption model
3. the reporting foundation is complete enough for thin-v1 review and browser-ready read seams, and the provider-backed coordinator plus browser-session auth now make the shared backend usable from a real browser client
4. the landed coordinator slice includes a hard-capped tool loop with policy-enforced read-tool execution plus bounded specialist delegation while keeping the default contributor workflow provider-independent
5. the first real `/app` plus `/api/...` seam pass proved that the shared backend is operational for login, intake, queued processing, review reads, and request-detail continuity outside direct service calls
6. that same live pass exposed one concrete blocker in the first provider-backed operator brief, and that blocker is now closed through request-centering instructions, request-centered validation, and one bounded OpenAI repair turn for generic first-pass output
7. the first additive non-browser auth slice is now also landed on that same seam through bearer-session issue, refresh rotation, token-authenticated introspection, token-authenticated logout, shared handler reuse for bearer-authenticated writes, and bearer-authenticated queued-request processing
8. the queued-agent UUID actor-header compatibility seam is now retired, so later auth work should not preserve or widen that temporary path again
9. the browser milestone is complete enough that residual browser work should now be treated as regression fixes or later UX refinement rather than as an active milestone plan
10. the verification gap is now closed, and the foundation checklist review found no material missing v1 structure
11. the next unanswered question is still operational readiness for supervised AI-backed user testing rather than missing thin-v1 foundation breadth
12. the validation pass has already produced enough signal to justify one bounded readiness-hardening milestone before deeper workflow testing resumes
13. the correct next move is therefore Milestone 9, followed by the deferred live workflows, rather than continuing the paused validation slice immediately or treating the repository as silently incomplete

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
This slice is now documented in `non_browser_auth_evolution_plan.md`, which keeps browser-session cookies as the active v1 auth path, treats UUID actor headers as pre-production automation compatibility, and defines the next additive bearer-session path for lightweight non-browser clients on the same backend foundation.

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
