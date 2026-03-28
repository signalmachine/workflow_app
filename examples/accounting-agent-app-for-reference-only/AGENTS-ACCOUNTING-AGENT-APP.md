# Repository Guidelines

## Project Structure & Module Organization
This repository is a Go monolith organized around clear layers and multiple interfaces.

- `cmd/`: runnable entry points.
- `cmd/server`: web server bootstrap.
- `cmd/app`: REPL and stateless CLI (`propose`, `validate`, `commit`, `balances`).
- `cmd/verify-db`, `cmd/verify-db-health`, `cmd/verify-agent`: operational verification tools.
- `internal/adapters/`: interface adapters (`web`, `repl`, `cli`).
- `internal/app/`: application service contract and orchestration across domains.
- `internal/core/`: domain logic (ledger, orders, inventory, reporting, vendors, purchase orders, users).
- `internal/db/`: PostgreSQL connection wiring.
- `internal/ai/`: OpenAI-backed interpretation and tool orchestration.
- `web/templates/`: `templ` layouts/pages.
- `web/static/`: CSS, JS, fonts, and static assets.
- `migrations/`: ordered SQL migrations (`NNN_description.sql`).
- `docs/`: architecture notes, deployment docs, and user/testing references.

## Architecture Overview
Contributors should preserve separation of concerns:

- Adapters translate input/output (HTTP, terminal, JSON) and delegate business work.
- `internal/app` coordinates use-cases and cross-service calls.
- `internal/core` owns accounting and operational rules.
- `internal/db` and `internal/ai` provide infrastructure integrations.

When adding features, prefer extending `internal/app` APIs instead of routing adapter-specific logic directly into `internal/core`.

## Collaboration Expectations
- It is expected to challenge suggestions when there is a stronger technical alternative.
- Recommendations should prioritize accounting integrity, software quality, operability, and delivery pragmatism.
- When proposing an alternative, include clear reasoning and tradeoffs.

## AI Integration Policy
- AI agent integrations in this repository must use the OpenAI **Responses API** via `github.com/openai/openai-go` (`client.Responses.New` / `client.Responses.NewStreaming`).
- Do **not** use the Chat Completions API (`client.Chat.Completions.New`) for agent flows in this application.
- Keep OpenAI integration code in `internal/ai/` and align changes with existing patterns in `internal/ai/agent.go` and `internal/ai/tools.go`.

## Build, Test, and Development Commands
Use these commands as the default workflow:

- `make generate`: compile `templ` files into Go.
- `make css`: build `web/static/css/app.css` from Tailwind input.
- `make dev`: template generation + run web server.
- `make build`: build server artifact.
- `make test`: run `go test ./internal/core -v`.
- `go build -o app.exe ./cmd/app`: build REPL/CLI binary.
- `go run ./cmd/verify-db`: apply/verify schema migrations.
- `go run ./cmd/verify-db-health`: validate database consistency checks.
- `go run ./cmd/verify-agent`: validate AI integration path.

Useful targeted test examples:

- `go test ./internal/core -v -run TestProposal`
- `go test ./internal/core -v -run TestInventory`
- `go test ./internal/core -v -run TestPurchaseOrder`

## Coding Style & Naming Conventions
- Run `gofmt` on changed Go files before commit.
- Use `snake_case` file names (`order_service.go`, `vendor_model.go`).
- Keep tests in `*_test.go`; integration-heavy suites use `*_integration_test.go`.
- Use `github.com/shopspring/decimal` for currency/amount fields; avoid `float64` in accounting paths.
- Keep SQL explicit and readable; this repo uses hand-written SQL with `pgx`, not ORM-generated queries.

## Testing Guidelines
- Test framework: Go `testing` package.
- Integration tests in `internal/core` depend on `TEST_DATABASE_URL` and may truncate data.
- Tests auto-skip when `TEST_DATABASE_URL` is absent; set it explicitly for CI/local integration validation.
- Never point test variables to a live database.
- After adding/changing migrations, migrate test DB first:
  - `DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db`

Recommended pre-PR local gate:

1. `make test`
2. `go run ./cmd/verify-db-health`
3. `go run ./cmd/verify-agent` (if change touches AI, chat, or tool flow)

For full-repo test runs that share one integration database, prefer serial package execution to avoid cross-package truncation races:

- `go test -p 1 ./...`

## DB Schema Change Workflow (Major Implementations)
For any major implementation that changes database schema, follow this sequence end-to-end:

1. Add migration(s) and update related code/tests in the same change set.
2. Migrate local test database first:
   - `DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db`
3. Run local DB health checks against test DB:
   - `DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db-health`
4. Run full local test suite (prefer serial when sharing one integration DB):
   - `go test -p 1 ./...`
5. Only after local validation passes, migrate cloud/prod DB:
   - `DATABASE_URL=<CLOUD_DATABASE_URL> go run ./cmd/verify-db`
6. Run cloud/prod DB health checks immediately after migration:
   - `DATABASE_URL=<CLOUD_DATABASE_URL> go run ./cmd/verify-db-health`

Completion rule (mandatory):

- A schema-change implementation is **not complete** until steps 5 and 6 have been executed successfully on the target live/dev cloud database (`DATABASE_URL`), and results are recorded in the change notes/PR.

Required safeguards:

- Never run test suites against cloud/live DBs.
- Never skip `verify-db-health` after a schema migration.
- If any step fails, stop rollout and fix before proceeding.

## Migration & Data Change Rules
- Add new migration files; do not edit old applied migrations.
- Use safe/idempotent patterns where possible:
  - `CREATE ... IF NOT EXISTS`
  - guarded `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`
  - `ON CONFLICT DO NOTHING` for seed rows
- Keep migration names descriptive (example: `033_add_vendor_tax_fields.sql`).
- If a migration impacts test fixtures or assumptions, update related integration tests in the same PR.

Recent control-account migrations:
- `033_add_control_account_flags.sql`
- `034_manual_je_control_account_audit.sql`
- `035_manual_je_control_account_enforcement_audit.sql`

Recent document-type governance migrations:
- `037_document_type_policy_expansion.sql`
- `038_document_type_policy_rules.sql`
- `040_backfill_document_types_and_core_rules.sql`
- `041_document_type_policy_violation_audit.sql`

## Control Account Guardrails (Current State)
- Phase 1-4 implementation is present:
  - control-account metadata + backfill
  - manual JE warning/enforcement/audit
  - reconciliation diagnostics report (`/reports/control-account-reconciliation`)
- Low-risk compatibility rule currently in place:
  - enforce-mode blocking is applied for explicit `manual_web` JE context.
  - AI chat JE posting remains non-blocking for control-account enforcement to avoid breaking existing workflows.

## Document Type Governance (Current State)
- Operational payments post with dedicated document types:
  - customer receipt -> `RC`
  - vendor payment -> `PV`
- `DOCUMENT_TYPE_POLICY_MODE=off|warn|enforce` is implemented in shared proposal validation/commit path.
- In `enforce` mode, both manual and AI proposal sources are blocked for document-type violations.
- Governance diagnostics report is available at `/reports/document-type-governance`.

## Commit & Pull Request Guidelines
Git history is currently minimal, so use a clear convention going forward.

Commit style:

- Prefer `<scope>: <imperative summary>`.
- Examples:
  - `core: enforce account rule lookup for AP postings`
  - `web: add PO invoice validation errors`
  - `migrations: add company-scoped sequence constraint`

PR checklist:

1. Explain what changed and why.
2. List affected modules (`internal/core`, `internal/app`, `web/templates`, etc.).
3. Include exact commands run for validation.
4. Call out schema/migration impact and rollback considerations.
5. Add screenshots for UI/template changes.

## Security & Configuration Tips
- Copy `.env.example` to `.env` for local setup.
- Do not commit secrets (`OPENAI_API_KEY`, `JWT_SECRET`, real DB URLs).
- Common required vars: `DATABASE_URL`, `TEST_DATABASE_URL`, `JWT_SECRET`.
- `DATABASE_URL` should point to the deployed cloud/live database.
- `TEST_DATABASE_URL` should point to a local test database used only for integration tests.
- Set `COMPANY_CODE` when running REPL/CLI in multi-company environments.
- Use separate local databases for dev and integration tests.
# Project Instructions for AI Code Assistant with gopls-mcp

## Context
You are an AI programming assistant helping users with Go code. You have access to gopls-mcp tools for semantic code analysis.

## CRITICAL PROHIBITIONS (NEVER DO THIS)
1. NEVER use `go_search` for text content (comments, strings, TODOs). Use `Grep` tool.
2. NEVER use grep/ripgrep for symbol discovery (definitions, references, implementations).
3. NEVER fall back from exclusive capabilities (see Tool Selection Guide).

<!-- Marker: AUTO-GEN-START -->
## Tool Selection Guide

### Code relationships (Exclusive Capabilities - NO FALLBACK)
| Task | Tool |
|------|------|
| Find interface implementations | go_implementation |
| Trace call relationships | go_get_call_hierarchy |
| Find symbol references | go_symbol_references |
| Jump to definition | go_definition |
| Analyze dependencies | go_get_dependency_graph |
| Preview renaming | go_dryrun_rename_symbol |

### Code exploration (Enhanced Capabilities - FALLBACK ALLOWED)
| Task | Tool | Fallback after 3 failures |
|------|------|---------------------------|
| List package symbols | go_list_package_symbols | Glob + Read |
| List module packages | go_list_module_packages | find |
| Analyze workspace | go_analyze_workspace | Manual exploration |
| Quick project overview | go_get_started | Read README + go.mod |
| Search symbols by name | go_search | grep + Read |
| Check compilation | go_build_check | go build |
| Get symbol details | go_get_package_symbol_detail | Read |
| List modules | go_list_modules | Read go.mod |
<!-- Marker: AUTO-GEN-END -->

## Integration Workflow
1. **Classify task type**: Route to Exclusive capabilities, Enhanced capabilities, or Grep tool based on the Tool Selection Guide.
2. **Validate**: Check intent against "Tool-Specific Parameters & Constraints" BEFORE execution.
3. **Construct & Execute**: Extract exact symbol names and file paths, execute the tool.
4. **Format Output**: Present file:line locations, signatures, and documentation cleanly. Do not dump raw JSON.

## Tool-Specific Parameters & Constraints

* **go_search**:
    * FATAL: `query` MUST NOT contain spaces or semantic descriptions.
    * Must be symbol names only (single token). Correct: `query="ParseInt"`.
    * Does NOT search comments or documentation.
* **go_implementation**:
    * Only for interfaces and types. STRICTLY PROHIBITED for functions.
* **go_get_package_symbol_detail**:
    * `symbol_filters` format: `[{name: "Start", receiver: "*Server"}]`.
    * `receiver` requires exact string match (`"*Server"` != `"Server"`).
* **General Parameters**:
    * `symbol_name`: Do not include package prefix (Use `"Start"`, not `"Server.Start"`).
    * `context_file`: Obtain strictly from the current file being analyzed.

## Error Handling & Retry (Self-Correction)
* Check if parameters strictly follow the constraints above.
* Try a shorter/simpler symbol name.
* Re-analyze code context before retrying.

## Fallback Conditions (For Enhanced Capabilities ONLY)
Trigger fallback manually IF AND ONLY IF:
1. 3 consecutive tool failures.
2. Timeout exceeds 30 seconds.
3. Empty result returned when code existence is absolutely certain.
*Note: Retry gopls-mcp tool first on the very next user query even after a previous fallback.*
