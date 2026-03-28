# REPL + Stateless CLI Upgrade Plan (2026-03-05)

## Summary
This plan upgrades the REPL and stateless CLI for power-user and testing workflows, while keeping default behavior backward compatible for existing scripts.

## Current Gaps
1. REPL confirms AI-proposed write tools but does not execute them.
2. CLI loads default company before command dispatch, which blocks `validate`/`commit` in some multi-company/pipeline cases.
3. CLI `propose` only reads one argument token instead of full trailing text.
4. No adapter-level tests exist for REPL/CLI behavior.
5. Minor help/docs mismatch (`/quit` supported in code but not shown in REPL help).

## Implementation Plan
1. REPL write-tool execution
- In REPL proposed-action flow, replace placeholder message with `ExecuteWriteTool(...)`.
- Preserve explicit confirmation prompt (`y/n`) before any write execution.
- Print success payload and actionable error output.

2. REPL help and command consistency
- Update REPL help output to include `/quit` (already implemented alias).
- Keep existing command names and routing unchanged.

3. CLI company-loading refactor
- Move company resolution into command-specific branches.
- `propose` and `balances`: require company context.
- `validate` and `commit`: read stdin proposal first and validate/commit directly without unconditional default-company load.

4. CLI argument and output hardening
- Parse `propose` event text from all remaining args (`args[1:]` joined with spaces).
- Keep current default stdout/stderr semantics.
- Add opt-in machine-readable output flags for power users without changing defaults:
  - `--json` on `balances` prints JSON to stdout (instead of table format).
  - `--quiet` on `validate` prints no success text; exit code only.
  - `--quiet` on `commit` prints no success text; exit code only.
  - `propose` remains JSON-by-default (no behavior change).

5. Docs and usage contract
- Document exact CLI behavior for stdout, stderr, and exit codes by command:
  - Success: exit code `0`; failures: non-zero.
  - Validation/domain errors print to stderr; machine-readable output (when requested) remains on stdout.
- Align README command examples with implemented behavior.

## Test Plan
1. CLI tests
- `propose` accepts multi-token text.
- `validate`/`commit` work without default-company precondition when proposal is valid.
- Invalid JSON and domain failures return non-zero exit with stable stderr messaging.

2. REPL tests
- Proposed write tool confirm path calls `ExecuteWriteTool`.
- Cancel path never executes write tool.
- Write-tool execution errors are surfaced and REPL loop continues.
- Write-tool success output format:
  - If `ExecuteWriteTool` returns valid JSON string, print pretty JSON.
  - If it returns non-JSON text, print as plain text.

3. Consistency checks
- REPL `/help` output matches supported slash commands.
- README command reference matches runtime behavior.

## Assumptions
1. No DB schema changes are needed.
2. Backward compatibility of default CLI outputs is mandatory.
3. REPL remains a local power-user/testing interface with existing service-layer safeguards.
