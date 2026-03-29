# workflow_app Milestone 9 User-Testing Readiness Hardening Plan

Date: 2026-03-29
Status: Complete on 2026-03-29; all three slices landed and closeout verification passed
Purpose: define the bounded readiness-hardening milestone that should land before the paused post-checkpoint live workflow validation resumes.

## 1. Why this milestone exists

The post-checkpoint validation pass produced useful live signal, but it also showed that `workflow_app` should not continue deeper user-testing-oriented workflow validation until a few bounded readiness gaps are addressed first.

Those gaps are not thin-v1 breadth gaps.

They are readiness gaps inside the already chosen architecture:

1. auth maturity is still acceptable for controlled internal validation but weaker than it should be for broader guided user testing
2. the coordinator-plus-specialist AI architecture is structurally stronger than the reference application, but the currently exposed bounded read-tool surface is still too thin
3. the shared browser and API layer is functionally rich but concentrated into very large files that increase implementation and regression risk

Reference-codebase stance for this milestone:

1. anything under `examples/` remains read-only reference material during Milestone 9
2. `examples/accounting-agent-app-for-reference-only/` remains a valid reference codebase during Milestone 9 implementation
3. no folder under `examples/` should be treated as part of the active `workflow_app` codebase
4. the accounting-agent reference app should be used as a practical comparison point for operator usefulness, guardrail quality, and implementation clarity
5. it must not be treated as the target architecture because it was a proof-of-concept single-agent application rather than the coordinator-plus-specialist architecture chosen for `workflow_app`
6. it still deserves explicit attention because it worked and achieved its intended objectives within those narrower design limits
7. it must remain read-only, must not be updated as part of Milestone 9 work, and must not be treated as part of the current `workflow_app` codebase

## 2. Planning decision

Milestone 9 is a bounded readiness-hardening milestone.

It is not:

1. a restart of thin-v1 foundation work
2. a general architecture rewrite
3. a broad UX expansion or product-surface expansion
4. a v2 promotion

## 3. Objective

Improve `workflow_app` enough that the paused live workflow validation can resume against a stronger and safer implementation baseline.

Milestone 9 should:

1. harden the auth path to a level more appropriate for guided user testing
2. strengthen the bounded coordinator or specialist read-tool surface without weakening approval and posting boundaries
3. reduce the biggest implementation-risk concentration in the shared browser and API layer

## 4. Scope

Milestone 9 contains three planned slices.

### 4.1 Slice 1: auth hardening

Goal:

1. move the active sign-in and session-issuance path beyond org-slug-plus-email-only issuance

Status:

1. done
2. browser-session and bearer-session issuance now require org slug, user email, and a password verified against the shared `identityaccess.users` credential record
3. the browser and later mobile or non-browser clients still share one session foundation, one tenant-context model, and one shared `/api/...` seam

Required outcomes:

1. one explicit target auth model is chosen and documented before implementation widens
2. browser-session and bearer-session issuance both sit on a stronger authentication boundary
3. the resulting auth path is acceptable for guided user testing rather than only tightly controlled internal validation
4. existing session truth, role enforcement, and shared backend auth reuse remain intact

### 4.2 Slice 2: bounded AI capability expansion

Goal:

1. keep the existing coordinator-plus-specialist architecture but make it materially more useful through a stronger bounded read-tool surface

Status:

1. done
2. the OpenAI-backed coordinator now exposes request-scoped read-only tools for current inbound-request detail and current processed-proposal continuity in addition to the existing org-level status summary tool
3. tool-policy enforcement, durable tool execution traces, artifacts, recommendations, and bounded delegation remain unchanged while the coordinator gains more request-relevant context than queue summary alone

Required outcomes:

1. the coordinator can gather more request-relevant review context than queue summary alone
2. tool additions stay read-only unless an existing approval-governed workflow already owns the write path
3. tool-policy enforcement, durable run traces, artifacts, recommendations, and delegation records remain intact
4. the milestone does not broaden into autonomy-heavy behavior

### 4.3 Slice 3: shared web or API seam decomposition

Goal:

1. reduce the maintenance and regression risk concentrated in the very large shared browser and API files

Status:

1. done
2. `internal/app` is now decomposed by seam rather than concentrated in two oversized files: API session, inbound, review, and approval handlers now live in dedicated files, and web session or inbound plus review surfaces now do the same
3. the shared HTTP contracts and route registration stayed unchanged while browser rendering, auth, inbound lifecycle, review reads, and approval actions became easier to inspect in isolation

Required outcomes:

1. decomposition follows domain or seam boundaries rather than arbitrary file splitting
2. HTTP contracts stay stable unless an explicit correction is necessary
3. browser rendering, auth, inbound-request lifecycle, review reads, and approval actions become easier to reason about in isolation
4. the milestone does not turn into a redesign of the thin-v1 web stack

## 5. Recommended execution order

1. auth hardening
2. bounded AI capability expansion
3. shared web or API seam decomposition

Reason:

1. auth is the highest readiness risk if guided user testing widens beyond tightly controlled internal use
2. AI usefulness should improve before the remaining live workflows are exercised again, otherwise those workflow passes will still be constrained by known capability thinness
3. file decomposition is valuable, but it is safest after the highest-risk behavior changes are already landed

## 6. Guardrails

During Milestone 9:

1. keep the current coordinator-plus-specialist architecture and improve it rather than replacing it
2. keep writes behind existing document, approval, posting, and workflow control boundaries
3. do not broaden scope into new business workflows, new modules, or v2 surfaces
4. do not reopen the paused validation slice as if it were still the active milestone
5. record any new scope decision in the canonical tracker before widening beyond the slices above
6. continue to use `examples/accounting-agent-app-for-reference-only/` as a reference input when it sharpens implementation decisions, but borrow selectively and only when the result still fits `workflow_app`'s multi-agent architecture
7. do not modify anything under `examples/` and do not let Milestone 9 implementation work spill into that reference tree

## 7. Verification expectations

For every Milestone 9 implementation slice:

1. add or update tests appropriate to the change
2. run `go build ./cmd/... ./internal/...`
3. run `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`
4. rerun narrower focused checks where the slice materially changes live seams, auth flows, or AI execution behavior

Closeout result on 2026-03-29:

1. the shared web or API seam decomposition slice remained implemented without contract drift
2. `go build ./cmd/... ./internal/...` passed
3. `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` passed against the configured PostgreSQL test database
4. `set -a; source .env; set +a; go run ./cmd/verify-agent` also passed again on the configured live-provider path, returning a processed request with completed coordinator and specialist runs plus a request-specific warehouse-pump recommendation summary
5. the tracker and README now record Milestone 9 as complete and set the next session to begin with a full implementation-versus-plan review before the paused live workflow validation resumes

At Milestone 9 closeout:

1. `go build ./cmd/... ./internal/...` passes
2. `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` passes
3. the tracker and README reflect the milestone result and the resumed next step
4. the repository is ready to resume the paused live workflow validation captured in `post_checkpoint_validation_and_user_testing_plan.md`

## 8. Stop rule

Milestone 9 is complete when:

1. the auth path is materially stronger for guided user testing
2. the bounded AI read-tool surface is materially stronger without weakening control boundaries
3. the biggest shared browser or API risk concentration has been decomposed enough to lower iteration risk
4. the repo is ready to resume the paused live workflows rather than continue broad hardening indefinitely

Milestone 9 should not stay open as an unbounded “quality improvements” bucket.
