# workflow_app Implementation Defaults V2

Date: 2026-04-12
Status: Active canonical defaults
Purpose: record the defaults that active implementation should preserve unless the active v2 planning surface is intentionally updated.

## 1. General defaults

1. root planning docs stay thin and directive
2. use archived planning docs only when the current task needs historical context
3. active v2 implementation may refactor or rebuild weak areas when that is the stronger engineering path
4. if code or behavior conflicts with these defaults, either fix the code or explicitly update the canonical root docs

## 2. Product and workflow defaults

1. keep the system centered on workflows, documents, approvals, ledgers, execution context, and reports
2. keep AI bounded by tool policy, approval boundaries, and normal domain services
3. draft inbound requests must not be processed by AI
4. request identity remains distinct from downstream document identity
5. meaningful workflow and control state must remain durably reconstructible from database records

## 3. Architecture defaults

1. Go owns business logic, workflow rules, write-path invariants, and durable state
2. Svelte is the promoted browser interaction layer on the shared Go backend
3. do not split backend truth into browser-specific versus later-client-specific ownership paths
4. `internal/app` should stay transport and orchestration focused rather than accumulating durable business logic
5. use `docs/technical_guides/` as the forward durable technical reference for the completed Svelte web runtime; treat `docs/svelte_web_guides/` as implementation-era migration guidance that is pending archive, not canonical current-state documentation

## 4. Interface defaults

1. the web layer should support workflow review, approval, inspection, detail continuity, and bounded maintenance rather than becoming a broad manual-entry product by default
2. desktop operator use is the current primary web target
3. avoid Tailwind CSS by default; prefer repo-owned styling on the promoted Svelte stack
4. major landing surfaces should direct operators into workflows cleanly rather than trying to inline every downstream activity

## 5. Implementation hygiene defaults

1. use verification appropriate to the actual risk of the change
2. workflow-critical changes need real `/app` plus `/api/...` continuity validation, not only package tests
3. when workflow-critical validation exercises AI-provider behavior and `.env` provides OpenAI credentials, use the real OpenAI-backed verification path rather than treating mock-only or offline checks as sufficient closeout evidence
4. when a durable workflow or validation checklist exists in `docs/workflows/`, use it and update it when workflow truth changes
5. when implementation reveals drift or weak architecture, fix it or record an explicit active plan rather than leaving silent debt
6. when Playwright is available locally and the open risk is actual browser-rendered behavior on `/app`, prefer Playwright-driven verification over adding more indirect HTTP-only or component-only evidence first
7. when the served app and a verification seed command both participate in one browser closeout, make them target the same backend explicitly; do not rely on implicit `TEST_DATABASE_URL` versus `DATABASE_URL` precedence
8. when browser continuity proof depends on seeded records in a dedicated verification org, the seed command should emit the exact org slug, actor credentials, and continuity ids needed by Playwright rather than expecting reviewers to infer them
9. when the served Go runtime embeds `internal/app/web_dist`, rebuild the frontend artifact and restart the app before treating a browser failure as a product defect
10. prefer stable browser assertions based on route contracts, headings, bounded actions, and exact drill-down ids over brittle copy-only markers that are likely to drift during normal UX refinement
11. when a doc claims supported browser workflow behavior, the promoted Svelte runtime is the source implementation to verify against, not archived template-era behavior or backend-only capability
