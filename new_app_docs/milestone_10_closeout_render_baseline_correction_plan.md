# workflow_app Milestone 10 Closeout Render-Baseline Correction Plan

Date: 2026-04-03
Status: Implemented historical corrective slice for the earlier Go-template browser rebuild; superseded as forward stack guidance by `../docs/svelte_web_guides/svelte_web_ui_migration_plan.md`
Purpose: record the grouped corrective slice promoted out of Milestone 10 closeout review so the retired monolithic Go-template fallback path and its closeout decisions remain available as migration history without serving as the active stack plan.

## 1. Why this corrective slice exists

Milestone 10 claims the rebuilt modular template bundle is the only active browser baseline.

Implementation review found one remaining contradiction:

1. `internal/app/web_templates_bundle.go` still allowed `renderWebPage` to fall back to the old monolithic `webAppHTML` template when page data failed to map to a bundle template
2. that fallback meant the legacy browser baseline was not fully retired in code, even though promoted routes already render from the modular bundle
3. Milestone 10 closeout should enforce the rebuilt bundle as the only valid render path before the browser-review evidence is treated as complete

## 2. Scope

In scope:

1. remove the active fallback from `renderWebPage`
2. fail fast if a web handler tries to render page data that has no modular template mapping
3. add focused test coverage that locks this closeout rule in place
4. remove the dead monolithic inline template payload once the modular bundle is the only render path

Out of scope:

1. browser visual redesign
2. route taxonomy changes
3. broader workflow-validation evidence capture

## 3. Required outcomes

This corrective slice is complete only when:

1. the modular web template bundle is the only active render baseline
2. unmapped page data fails explicitly instead of silently rendering through the retired monolithic template
3. focused `internal/app` coverage proves the failure path is enforced
4. the retired inline browser payload no longer remains in `internal/app/web.go` as dead fallback-era ballast

## 4. Verification

Before closing this corrective slice:

1. run focused `internal/app` tests covering web rendering
2. run `go build ./cmd/... ./internal/...`
3. run `gopls` diagnostics on edited Go files
4. run the canonical `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` verification

## 5. Documentation sync

When this slice lands:

1. update `new_app_tracker.md` with the corrective-slice status
2. update `milestone_10_web_rebuild_plan.md` to note that render-path enforcement is now part of closeout
3. keep `docs/workflows/` focused on browser-review evidence rather than this implementation detail unless closeout expectations materially change
