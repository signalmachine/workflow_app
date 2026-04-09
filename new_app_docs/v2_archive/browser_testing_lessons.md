# workflow_app Browser Testing Lessons

Date: 2026-04-04
Status: Archived reference log for browser-serving lessons still relevant to future validation
Purpose: preserve concrete browser-testing failures, root causes, and durable prevention rules discovered during the Svelte cutover and related browser-serving work.

## 1. How to use this document

Use this document as archived reference when future browser-serving or cutover work needs past failure lessons.

Keep entries:

1. concrete
2. short
3. tied to one real failure mode
4. written as prevention rules, not just postmortem notes

If future work uncovers another durable lesson of the same kind, either update this archived log deliberately or promote a more permanent workflow-testing home under `docs/workflows/`.

## 2. Update rule

Add or update an entry when:

1. the real browser behavior differs from what existing tests implied
2. a route renders HTML but fails to hydrate or load assets correctly
3. a static-asset, routing-base, or frontend-build integration bug escapes normal verification
4. a new browser issue reveals that the current tests are checking the wrong thing

Do not add vague reminders. Add one specific lesson with one specific prevention rule.

## 3. Current lessons

### 3.1 Real asset requests matter more than shell-only checks

Failure:

1. `/app` returned a valid-looking Svelte shell, but the browser still showed a blank page because JavaScript module assets were not loading correctly.

What the tests missed:

1. existing tests checked that `/app` returned HTML containing Svelte bootstrap markup
2. they did not request one real asset URL from that shell and verify that the asset itself was served correctly

Prevention rule:

1. whenever the Go server embeds and serves a browser bundle, include at least one test that requests a real built asset such as `/app/_app/...js` and verifies that it returns asset content rather than the SPA HTML fallback

### 3.2 Root-path browser behavior must be tested, not inferred

Failure:

1. loading `/app` without a trailing slash caused relative asset URLs to resolve incorrectly from the browser's point of view

What the tests missed:

1. the root-path test asserted one specific HTML import shape instead of asserting the browser-relevant outcome

Prevention rule:

1. tests for SPA root routes should assert browser-relevant path behavior, especially that root requests produce asset URLs that remain valid when loaded exactly at `/app`

### 3.3 Go embed rules can silently exclude frontend assets

Failure:

1. the Svelte `_app` asset directory was not embedded into the Go binary because Go embed excludes underscore-prefixed paths unless `all:` is used

What the tests missed:

1. no test proved that one embedded `_app` asset could actually be opened and served from the running Go handler

Prevention rule:

1. when embedding frontend build output with Go, include one test that proves the generated asset tree used by the browser is actually present in the embedded filesystem, especially for paths that may be excluded by default embed rules

### 3.4 HTML fallback behavior must not hide asset-serving defects

Failure:

1. missing or misclassified JavaScript asset requests fell through to the SPA HTML fallback, which turned a file-serving bug into a browser MIME-type error

What the tests missed:

1. earlier tests did not distinguish between correct SPA fallback for route navigation and incorrect fallback for static asset requests

Prevention rule:

1. tests must cover both sides of SPA serving behavior:
2. non-asset application routes should fall back to the SPA shell
3. real static assets should never fall back to the SPA shell
4. missing asset-like requests should return `404` so routing bugs or broken bundles do not get disguised as HTML fallback

### 3.5 Browser smoke checks still matter after focused handler tests

Failure:

1. the server-side tests were too narrow to prove that the browser would hydrate the app successfully

What the tests missed:

1. no bounded browser-smoke step checked the actual rendered `/app` experience after the cutover

Prevention rule:

1. after changes to frontend embedding, SPA fallback, routing base paths, or built-asset serving, run one bounded browser smoke check on the real `/app` route in addition to handler-level tests

## 4. Current minimum checklist for future browser-serving changes

When changing Svelte static serving, Go embed behavior, SPA fallback, or app base-path handling, verify all of the following:

1. request `GET /app` and confirm it returns the SPA shell
2. request one real built JavaScript asset under `/app/_app/...` and confirm it returns asset content, not HTML fallback
3. confirm the asset response has a JavaScript-appropriate content type
4. confirm one non-asset app route such as `/app/review` still falls back to the SPA shell
5. run one real browser smoke check against `/app` and confirm the page hydrates instead of showing a blank screen

## 5. Seed incidents

### 2026-04-04: Milestone 13 Svelte cutover white-screen bug

Observed behavior:

1. the browser showed a blank white screen on `/app`
2. the console reported module-load failures caused first by wrong root-path asset resolution and then by HTML fallback being served for JavaScript chunk requests

Confirmed causes:

1. `/app` root serving needed the SPA shell with `/app`-rooted asset imports
2. `//go:embed web_dist` did not include the Svelte `_app` directory, so built chunk files were missing from the embedded filesystem

Required enduring tests:

1. root-path shell test for `/app`
2. direct embedded-asset serving test for one built `/app/_app/...js` file
3. bounded real-browser smoke check after changes to the embedded Svelte serving path
