# REPL Interface Plan

Date: 2026-04-12
Status: Active planning document — revised with slash-command syntax and full operational surface
Purpose: define the architecture, scope, command surface, implementation slices, and
verification plan for `cmd/repl`, the interactive developer and operator REPL for
`workflow_app`.

---

## 1. Why a REPL

### 1.1 The core problem

The web UI at `/app` is the operator surface, not the developer surface. During active
development of AI hardening, prompt iteration, and workflow continuity work, contributors
face a slow and opaque feedback cycle:

1. start the app and log in through the browser
2. navigate to the submit form and type a request
3. navigate to Operations and click "Process next queued"
4. navigate through request detail → proposal detail → approval → document → accounting
   (up to 6 sequential page hops per iteration)
5. repeat from step 2 for the next prompt or scenario change

Each iteration takes several minutes of clicking even when the backend is correct. The
web UI also does not expose internal AI records that developers need: raw coordinator
output, tool execution traces, run steps, delegation records, provider token usage, and
draft-level DB state.

`cmd/verify-agent` is the right scripted tool for CI-style continuity proof but it does
not support interactive exploration. You cannot submit an arbitrary request, inspect the
result, branch into a different scenario, or interactively drive the approval chain.

### 1.2 What the REPL adds

A REPL reduces the AI hardening iteration cycle from several minutes of browser clicks
to under 30 seconds, with full internal visibility:

1. submit arbitrary request text (with or without attachments) and get `REQ-...` back
   immediately
2. process the next queued request and see coordinator output, tool loop trace, token
   usage, and recommendation in one terminal panel
3. inspect run steps, tool calls, artifacts, delegation chains, and recommendation
   payloads by reference without navigating across 6 pages
4. drive the full approval and document lifecycle interactively
5. query all review reads: requests, proposals, approvals, documents, accounting,
   inventory, work orders, parties, workforce, audit
6. manage master data, tool policies, and org setup without the web admin UI
7. run bounded smoke test scenarios as executable script files for repeatable proof

### 1.3 Dual audience

The REPL serves two audiences through the same binary:

1. **Developer**: fast iteration on AI prompts, coordinator behavior, tool policy, and
   workflow correctness without browser overhead
2. **Operator power user**: a keyboard-first interface to all operational surfaces —
   identical backend seam as the web UI but without the navigation friction

Both audiences share the same command set. The operator may use a narrower subset of
commands; the developer uses the full set including AI internals.

### 1.4 Scope boundary

The REPL is a client of the existing domain service layer, not a second business backend:

1. MUST authenticate through `identityaccess` using the same shared auth model
2. MUST call domain services directly — not reimplementing business logic
3. MUST NOT write to domain-owned tables except via the normal service API
4. MUST be built as `cmd/repl`, following the existing `cmd/` command shape
5. MUST respect workflow state, tool policy, and approval boundaries — it is an
   interactive client, not a bypass path

---

## 2. Command syntax: slash-command format

All commands use a `/` prefix. This serves three important purposes:

1. **disambiguation**: distinguishes REPL commands from free-text input — if the REPL
   ever gains a "send as request" mode, unrecognized non-slash text can optionally be
   forwarded as a new inbound request
2. **familiarity**: mirrors the interaction model of modern chat tools, IDE terminals,
   and operator consoles (Slack `/commands`, IRC, Claude's interface)
3. **tab completion**: the `/` prefix makes it unambiguous what the completion engine
   should search against

Command names are short, memorable, and action-oriented. Multi-word commands use a
single space separator: `/run tools`, `/approval list`, `/workorder show`.

The `$LAST_*` shorthand tokens (see section 3.11) are expanded by the dispatcher before
argument parsing.

---

## 3. Complete command surface

### 3.1 Quick reference card

This card shows every slash command at a glance, grouped by area.

```
META
  /help [cmd]          Show all commands or help for one command
  /history [--limit N] Show command history for this session
  /context             Show current $LAST_* values and active org/user
  /clear               Clear the terminal screen
  /json                Toggle JSON output mode on/off
  /quiet               Toggle quiet (minimal) output mode on/off
  /script PATH         Execute a .repl script file non-interactively
  /exit  /quit         Exit the REPL

SESSION / AUTH
  /login --org ORG --email EMAIL [--password P]
  /whoami
  /logout
  /switch --org ORG

INBOUND REQUESTS
  /submit "text" [--channel CH] [--label L] [--file PATH]
  /draft "text" [--channel CH] [--label L]
  /queue [REF]
  /cancel [REF] [--reason "..."]
  /amend [REF]
  /delete [REF]
  /request [REF]
  /requests [--status S] [--channel CH] [--limit N]

QUEUE & AI PROCESSING
  /process [--channel CH] [--request REF]
  /queued [--channel CH] [--limit N]
  /qstat

AI RUNS & INTERNALS
  /runs [--limit N]
  /run [RUN_ID]
  /steps [RUN_ID]
  /step STEP_ID
  /tools [RUN_ID]
  /delegation [RUN_ID]
  /artifact ARTIFACT_ID

PROPOSALS / RECOMMENDATIONS
  /proposals [--status S] [--limit N]
  /proposal REC_ID

APPROVALS
  /approvals [--status S] [--limit N]
  /approval APPROVAL_ID
  /reqapproval REC_ID [--reason "..."]
  /approve APPROVAL_ID [--note "..."]
  /reject APPROVAL_ID [--note "..."]

DOCUMENTS
  /documents [--status S] [--limit N]
  /document DOC_ID
  /post DOC_ID [--summary "..."] [--currency CODE]

ACCOUNTING
  /accounts [--class CLASS] [--status S] [--limit N]
  /account ACCT_ID_OR_CODE
  /taxcodes [--status S] [--limit N]
  /taxcode CODE
  /periods [--status S] [--limit N]
  /period PERIOD_ID
  /closeperiod PERIOD_ID [--note "..."]
  /journal [--limit N]
  /entry ENTRY_ID
  /balances [--type TYPE]
  /trialbalance [--as-of DATE]
  /balancesheet [--as-of DATE]
  /income [--period PERIOD_ID]

INVENTORY
  /items [--status S] [--limit N]
  /item ITEM_ID_OR_CODE
  /locations [--status S] [--limit N]
  /location LOC_ID
  /stock [--item ITEM] [--location LOC] [--limit N]
  /movements [--item ITEM] [--location LOC] [--limit N]
  /movement MOVEMENT_ID

WORK ORDERS
  /workorders [--status S] [--limit N]
  /workorder WO_ID
  /wostatus WO_ID --status STATUS [--note "..."]
  /womaterials WO_ID [--limit N]
  /wosync WO_ID

PARTIES & CONTACTS
  /parties [--role ROLE] [--status S] [--limit N]
  /party PARTY_ID
  /contacts [--party PARTY_ID] [--limit N]
  /contact CONTACT_ID

WORKFORCE
  /workers [--limit N]
  /worker WORKER_ID
  /labor [--worker WORKER_ID] [--limit N]

ATTACHMENTS
  /attachments [--request REF] [--limit N]
  /attachment ATTACHMENT_ID

AUDIT
  /audit [--type TYPE] [--entity ENTITY_ID] [--limit N]
  /event EVENT_ID

ADMIN / SETUP
  /org
  /users [--limit N]
  /bootstrap [--org-slug SLUG] [--email EMAIL] [--password P]
  /seed [--minimal]
  /policies [--capability CAP] [--tool TOOL]
  /setpolicy --tool TOOL --capability CAP --policy POLICY [--rationale "..."]
  /createaccount --code CODE --name NAME --class CLASS [--control TYPE]
  /createtaxcode --code CODE --name NAME --rate RATE
  /createperiod --name NAME --start DATE --end DATE
```

---

### 3.2 Meta commands

| Command | Description |
|---|---|
| `/help [cmd]` | List all commands with one-line descriptions, or full help for `cmd` |
| `/history [--limit N]` | Show the last N commands from this session (default 20) |
| `/context` | Print current `$LAST_*` values, active org slug, user, and role |
| `/clear` | Clear the terminal screen |
| `/json` | Toggle JSON output mode — all commands emit raw pretty-printed service responses |
| `/quiet` | Toggle quiet mode — suppress decorative output and headers |
| `/script PATH` | Load and execute a `.repl` file line by line; exit non-zero on first error |
| `/exit` or `/quit` | Exit the REPL (also accepts Ctrl-D) |

---

### 3.3 Session and auth commands

| Command | Description |
|---|---|
| `/login --org ORG --email EMAIL [--password P]` | Authenticate into ORG; prompts securely for password if omitted |
| `/whoami` | Print current org, user display name, role, session ID, and expiry |
| `/logout` | Revoke the current session and clear the active actor |
| `/switch --org ORG [--email EMAIL] [--password P]` | Switch to a different org mid-session |

The REPL can be started without `--org-slug`, `--email`, or `--password` flags and remain
in an unauthenticated state until `/login` is called. Commands that require an actor will
print a clear `not authenticated` error rather than panicking.

---

### 3.4 Inbound request commands

These are the most common commands for both operator and AI-hardening workflows.

| Command | Description |
|---|---|
| `/submit "text" [--channel CH] [--label L] [--file PATH]` | Create and immediately queue a new inbound request; prints `REQ-...` and sets `$LAST_REQ` |
| `/draft "text" [--channel CH] [--label L]` | Create a draft request without queueing; sets `$LAST_REQ` |
| `/queue [REF]` | Queue `$LAST_REQ` or the specified draft reference |
| `/cancel [REF] [--reason "..."]` | Cancel a queued or processing request |
| `/amend [REF]` | Move a queued or cancelled request back to draft for editing |
| `/delete [REF]` | Hard-delete an unprocessed draft (irreversible, prints confirmation) |
| `/request [REF]` | Show full request detail: messages, attachments, lifecycle timestamps, linked runs, proposals |
| `/requests [--status S] [--channel CH] [--limit N]` | List requests with optional status and channel filters (default limit 20) |

`/submit` is the primary entry point for AI hardening. The `--file PATH` flag attaches a
local file and auto-generates a derived text record in the same call.

---

### 3.5 Queue and AI processing commands

| Command | Description |
|---|---|
| `/process [--channel CH] [--request REF]` | Claim and process the next queued request (or a specific request) using the live AI provider |
| `/queued [--channel CH] [--limit N]` | List currently queued requests |
| `/qstat` | Show request count by status (queued, processing, processed, failed, cancelled) |

After a successful `/process`, the output is:

```
✓ Processed REQ-000042
  run:            RUN-abc123             (completed)
  capability:     inbound_request.coordination
  tool loops:     2 iterations, 3 tool calls
  tokens:         input=1240  output=340  total=1580
  priority:       normal
  recommendation: REC-xyz789
  summary:        "Vendor invoice INV-2024-001 from Acme Corp — $1,200 INR..."
  specialist:     — (no delegation)

  $LAST_RUN=RUN-abc123  $LAST_REC=REC-xyz789
  → /tools to inspect tool calls  /proposal to review
```

---

### 3.6 AI run and internals commands

These are the key commands for AI hardening iteration. They expose data that is buried
inside step payloads in the review UI.

| Command | Description |
|---|---|
| `/runs [--limit N]` | List recent AI runs for the current org (default 10) |
| `/run [RUN_ID]` | Show run detail: role, capability, status, request reference, timestamps, metadata |
| `/steps [RUN_ID]` | List all steps for a run with step type, status, and payload summary |
| `/step STEP_ID` | Show full step detail including input and output payloads |
| `/tools [RUN_ID]` | Show the tool execution trace: iteration, tool name, policy enforced, outcome, call ID, argument preview |
| `/delegation [RUN_ID]` | Show delegation record if the coordinator invoked a specialist; prints parent/child run IDs and reason |
| `/artifact ARTIFACT_ID` | Show artifact payload: title, body, rationale bullets, next-action list |

`/tools` is the primary AI hardening diagnostic command. It makes the tool loop
completely transparent:

```
/tools RUN-abc123

  iter  tool                        policy    outcome   preview
  ────  ──────────────────────────  ────────  ────────  ──────────────────────────
  1     list_queued_requests        allow     success   {"count":3,"channel":"repl"}
  1     get_request_detail          allow     success   {"reference":"REQ-000042"...}
  2     (final structured output)   —         —         —

  2 iterations · 2 tool calls · input=1240 output=340
```

All IDs default to `$LAST_RUN` when called without an explicit argument.

---

### 3.7 Proposal and recommendation commands

| Command | Description |
|---|---|
| `/proposals [--status S] [--limit N]` | List recommendations, optionally filtered by status |
| `/proposal [REC_ID]` | Show recommendation detail: type, status, summary, payload, linked artifact, approval, and document |

`/proposal` defaults to `$LAST_REC`. After reviewing the proposal, the typical next step
is `/reqapproval` or inspecting the linked artifact with `/artifact`.

---

### 3.8 Approval commands

| Command | Description |
|---|---|
| `/approvals [--status S] [--limit N]` | List approvals, optionally filtered by status (pending, approved, rejected) |
| `/approval [APPROVAL_ID]` | Show approval detail: linked recommendation, document, decision, decider, note |
| `/reqapproval [REC_ID] [--reason "..."]` | Request approval for a proposal; creates an approval record and sets `$LAST_APPROVAL` |
| `/approve [APPROVAL_ID] [--note "..."]` | Record an approved decision (requires approver role) |
| `/reject [APPROVAL_ID] [--note "..."]` | Record a rejected decision (requires approver role) |

`/approve` and `/reject` are intentionally short action verbs. They wrap `approval
decide` with a cleaner interface and print the resulting document status change.

---

### 3.9 Document commands

| Command | Description |
|---|---|
| `/documents [--status S] [--limit N]` | List documents, optionally filtered by status (draft, submitted, approved, posted) |
| `/document [DOC_ID]` | Show document detail: type, status, linked approval, request reference, recommendation |
| `/post DOC_ID [--summary "..."] [--currency CODE]` | Post a document to accounting (admin/operator; requires journal lines to be pre-configured) |

`/document` defaults to `$LAST_DOC`.

---

### 3.10 Accounting commands

These expose the full accounting surface: ledger structure, tax configuration, period
management, journal entries, control balances, and financial statements.

#### Master data

| Command | Description |
|---|---|
| `/accounts [--class CLASS] [--status S] [--limit N]` | List ledger accounts, optionally filtered by class (asset, liability, equity, revenue, expense) |
| `/account ACCT_ID_OR_CODE` | Show account detail: code, name, class, control type, status, balance |
| `/createaccount --code CODE --name NAME --class CLASS [--control TYPE] [--no-direct-posting]` | Create a new ledger account |
| `/taxcodes [--status S] [--limit N]` | List tax codes |
| `/taxcode CODE` | Show tax code detail: rate, scope, status |
| `/createtaxcode --code CODE --name NAME --rate RATE [--scope SCOPE]` | Create a new tax code |
| `/periods [--status S] [--limit N]` | List accounting periods |
| `/period PERIOD_ID` | Show period detail: name, date range, status |
| `/createperiod --name NAME --start DATE --end DATE` | Create a new accounting period |
| `/closeperiod [PERIOD_ID] [--note "..."]` | Close an open accounting period |

#### Transactions and balances

| Command | Description |
|---|---|
| `/journal [--period PERIOD_ID] [--account ACCT] [--limit N]` | List recent journal entries with source document and request reference |
| `/entry [ENTRY_ID]` | Show journal entry detail: lines, debit/credit amounts, document provenance, linked request and approval |
| `/balances [--type TYPE]` | Show control account balances (receivable, payable, inventory, tax) |

#### Financial statements

| Command | Description |
|---|---|
| `/trialbalance [--as-of DATE]` | Print a trial balance as of a date (default: today) |
| `/balancesheet [--as-of DATE]` | Print a balance sheet with current-earnings treatment |
| `/income [--period PERIOD_ID]` | Print an income statement for a period |

These call `internal/reporting` reads — the same reads the web accounting reports use.

---

### 3.11 Inventory commands

| Command | Description |
|---|---|
| `/items [--status S] [--limit N]` | List inventory items |
| `/item ITEM_ID_OR_CODE` | Show item detail: code, name, unit, status |
| `/locations [--status S] [--limit N]` | List inventory locations |
| `/location LOC_ID` | Show location detail |
| `/stock [--item ITEM] [--location LOC] [--limit N]` | Show stock balances, optionally filtered by item or location |
| `/movements [--item ITEM] [--location LOC] [--limit N]` | List stock movements |
| `/movement MOVEMENT_ID` | Show movement detail: item, location, quantity, direction, source document |

---

### 3.12 Work order commands

| Command | Description |
|---|---|
| `/workorders [--status S] [--limit N]` | List work orders, optionally filtered by status |
| `/workorder [WO_ID]` | Show work order detail: status, linked party and item, labor and material usage |
| `/wostatus WO_ID --status STATUS [--note "..."]` | Advance work order status |
| `/womaterials [WO_ID] [--limit N]` | List material usages for a work order |
| `/wosync [WO_ID]` | Sync inventory usage for a work order (`SyncInventoryUsage`) |

---

### 3.13 Party and contact commands

| Command | Description |
|---|---|
| `/parties [--role ROLE] [--status S] [--limit N]` | List parties; role is `customer`, `vendor`, `supplier`, etc. |
| `/party PARTY_ID` | Show party detail: role, name, status, contact count |
| `/contacts [--party PARTY_ID] [--limit N]` | List contacts for a party or all contacts |
| `/contact CONTACT_ID` | Show contact detail: name, email, phone, party link |

---

### 3.14 Workforce commands

| Command | Description |
|---|---|
| `/workers [--limit N]` | List workers registered in the org |
| `/worker WORKER_ID` | Show worker detail |
| `/labor [--worker WORKER_ID] [--workorder WO_ID] [--limit N]` | List labor entries, filterable by worker or work order |

---

### 3.15 Attachment commands

| Command | Description |
|---|---|
| `/attachments [--request REF] [--limit N]` | List attachments for a specific request or all recent attachments |
| `/attachment ATTACHMENT_ID` | Show attachment detail: file name, media type, size, derived texts |

---

### 3.16 Audit commands

| Command | Description |
|---|---|
| `/audit [--type TYPE] [--entity ENTITY_ID] [--limit N]` | List audit events; filterable by event type and entity ID |
| `/event EVENT_ID` | Show audit event detail: type, entity, actor, payload, timestamp |

---

### 3.17 Admin and setup commands

| Command | Description |
|---|---|
| `/org` | Show current org: slug, name, created at |
| `/users [--limit N]` | List members of the current org with role and session count |
| `/bootstrap [--org-slug SLUG] [--email EMAIL] [--password P]` | Bootstrap a new org with admin user and demo baseline seed |
| `/seed [--minimal]` | Run `internal/setup` seed against the current org (chart of accounts, GST codes, demo parties, inventory) |
| `/policies [--capability CAP] [--tool TOOL]` | List AI tool policies for the current org |
| `/setpolicy --tool TOOL --capability CAP --policy allow\|approval_required\|deny [--rationale "..."]` | Set a tool policy for a capability and tool pair |

---

### 3.18 The `$LAST_*` chaining system

The dispatcher tracks the most recently referenced ID for each entity type. Every command
that creates or identifies a record sets the relevant `$LAST_*` token.

| Token | Set by | Used by |
|---|---|---|
| `$LAST_REQ` | `/submit`, `/draft`, `/request`, `/requests` | `/queue`, `/cancel`, `/amend`, `/delete`, `/process --request`, `/request` |
| `$LAST_RUN` | `/process`, `/run`, `/runs` | `/steps`, `/tools`, `/delegation`, `/run` |
| `$LAST_REC` | `/process`, `/proposal`, `/proposals` | `/proposal`, `/reqapproval`, `/artifact` |
| `$LAST_APPROVAL` | `/reqapproval`, `/approval`, `/approvals` | `/approve`, `/reject`, `/approval` |
| `$LAST_DOC` | `/approve`, `/post`, `/document`, `/documents` | `/document`, `/post`, `/entry` (via doc) |
| `$LAST_ENTRY` | `/entry`, `/journal`, `/post` | `/entry` |
| `$LAST_ARTIFACT` | `/process` | `/artifact` |

Example chaining session:

```
workflow_app> /submit "vendor invoice from Acme Corp for $1200 INR"
✓ REQ-000042 queued  [$LAST_REQ=REQ-000042]

workflow_app> /qstat
  queued=1  processing=0  processed=14  failed=0  cancelled=2

workflow_app> /process
✓ Processed REQ-000042  [$LAST_RUN=RUN-abc123  $LAST_REC=REC-xyz789]

workflow_app> /tools
  (tool trace for RUN-abc123)

workflow_app> /proposal
  (proposal detail for REC-xyz789)

workflow_app> /reqapproval --reason "standard finance review"
✓ APR-def456 created  [$LAST_APPROVAL=APR-def456]

workflow_app> /approve --note "Approved by finance lead"
✓ APR-def456 approved  document=DOC-ghi789 (approved)  [$LAST_DOC=DOC-ghi789]

workflow_app> /document
  (document detail for DOC-ghi789)

workflow_app> /trialbalance
  (current trial balance)
```

The entire AI hardening loop — submit → process → inspect tools → review proposal →
approve — is 6 commands a developer can type in under 20 seconds.

---

## 4. Architecture

### 4.1 File layout

```
cmd/
  repl/
    main.go           ← entry point, flag parsing, DB wiring, REPL loop
    commands.go       ← Command interface, registry, help text, dispatch
    session.go        ← ReplContext struct, actor management, $LAST_* tracking
    renderer.go       ← output formatting (tables, detail blocks, JSON mode, color)
    history.go        ← command history read/write (added in Slice 4)
    completer.go      ← tab completion logic (added in Slice 4)
```

The REPL binary does not share Go source files with `cmd/verify-agent`. Both call the
same domain services independently.

### 4.2 Backend call path

```
slash command input
  → tokenize and expand $LAST_* tokens
  → dispatch to Command.Run(ctx, rc, args)
  → call domain services directly (intake, ai, accounting, etc.)
  → pass result to renderer
  → renderer formats output → print to stdout
  → loop
```

The REPL calls service methods directly using a valid `identityaccess.Actor`. It does NOT
go through the JSON HTTP layer. This avoids HTTP server startup overhead and keeps each
command fast.

An optional `--via-api` flag is reserved for a future slice that routes commands through
the live served `/api/...` seam using session cookies instead, useful for testing HTTP
middleware and response-mapping correctness.

### 4.3 Command interface

```go
type Command interface {
    // Slash-prefixed name: "submit", "tools", "approve" etc.
    Name() string
    // Optional aliases: "quit" → "exit", "q" → "quit"
    Aliases() []string
    // Short one-line usage shown in /help listing
    Usage() string
    // Full help text shown when /help cmd is called
    Help() string
    // Execute the command; rc carries actor, db, $LAST_* state
    Run(ctx context.Context, rc *ReplContext, args []string) error
}
```

### 4.4 ReplContext

```go
type ReplContext struct {
    DB       *sql.DB
    Actor    identityaccess.Actor
    Services ReplServices   // pre-wired service instances

    // $LAST_* tracking
    LastReq      string
    LastRun      string
    LastRec      string
    LastApproval string
    LastDoc      string
    LastEntry    string
    LastArtifact string

    // Output control
    JSONMode  bool
    QuietMode bool
    NoColor   bool
}

type ReplServices struct {
    Auth        *identityaccess.Service
    Intake      *intake.Service
    AI          *ai.Service
    Processor   *app.OpenAIAgentProcessor   // nil until /process is called; lazy-init
    Attachments *attachments.Service
    Documents   *documents.NewService
    Accounting  *accounting.Service
    Reporting   *reporting.Service
    Parties     *parties.Service
    Inventory   *inventoryops.Service
    WorkOrders  *workorders.Service
    Workforce   *workforce.Service
}
```

Services are constructed once at startup and shared across all commands. `Processor` (the
OpenAI agent) is lazily initialized on the first `/process` call so that missing
`OPENAI_API_KEY` does not prevent the REPL from starting.

### 4.5 Output rendering contract

| Scenario | Renderer output |
|---|---|
| Success: single entity | Detail block: `key: value` pairs |
| Success: list | Aligned table with header row |
| Success: mutation | One-line summary: `✓ REQ-000042 queued` |
| Success: action summary | Multi-line block as shown in section 3.5 |
| Error | `✗ error: message` in red if color enabled |
| `--json` / `JSON mode` | Raw pretty-printed Go struct serialized to JSON |

Color is gated on `isatty(stdout)`. Color output is disabled automatically in pipe mode,
script mode, and when `--no-color` is set.

---

## 5. Implementation slices

### Slice 1 — Foundation and core AI hardening loop

**Goal**: a working binary that covers the primary developer feedback loop.

Files created:

```
cmd/repl/main.go
cmd/repl/commands.go
cmd/repl/session.go
cmd/repl/renderer.go
```

Commands included:

```
/help /exit /quit /context /json /quiet
/login /whoami /logout
/submit /draft /queue /cancel /request /requests
/qstat /queued /process
/runs /run /steps /step /tools /delegation /artifact
/proposals /proposal
```

Entry point verification:

```bash
go build ./cmd/repl
go run ./cmd/repl \
  --database-url "$DATABASE_URL" \
  --org-slug north-harbor-works \
  --email admin@nhw.test
# then inside:
/submit "vendor invoice Acme Corp $1200 INR"
/process
/tools
/proposal
```

### Slice 2 — Full workflow lifecycle

**Goal**: complete the chain from request through approval, document, and journal entry.

Commands added:

```
/amend /delete
/approvals /approval /reqapproval /approve /reject
/documents /document
/journal /entry
/balances
```

Verification: run the full interactive chain and compare against the output of
`cmd/verify-agent -approval-flow`.

### Slice 3 — All operational surfaces

**Goal**: expose all remaining domain services so the REPL covers the complete
operational footprint of the web UI.

Commands added:

```
/accounts /account /createaccount
/taxcodes /taxcode /createtaxcode
/periods  /period  /createperiod /closeperiod
/trialbalance /balancesheet /income
/items /item /locations /location /stock /movements /movement
/workorders /workorder /wostatus /womaterials /wosync
/parties /party /contacts /contact
/workers /worker /labor
/attachments /attachment
/audit /event
/org /users /bootstrap /seed /policies /setpolicy
```

### Slice 4 — Developer ergonomics

**Goal**: make the REPL excellent for long development sessions.

Additions:

- `github.com/chzyer/readline` (or equivalent) for line editing, up-arrow history, and
  tab completion on command names
- `/history` command backed by a persistent `~/.workflow_repl_history` file
- tab completion for `/` commands and known `$LAST_*` tokens
- `--quiet` global flag and `/quiet` toggle for minimal output (useful in pipe mode)
- `/clear` screen command
- `/script PATH` non-interactive script execution mode
- `--json` global flag and `/json` toggle

### Slice 5 — Script scenarios and regression tooling

**Goal**: make the REPL useful as an automated smoke test runner.

Additions:

- `scripts/repl/` directory with canonical `.repl` scenario files
- `smoke_ai_loop.repl` — submit → process → inspect tools → confirm proposal
- `smoke_approval_chain.repl` — full request → approval → document → journal proof
- `smoke_accounting_baseline.repl` — seed → trial balance → income statement
- exit-code contract for CI-gate use: exit 0 on pass, non-zero on first `/` command error

Example script:

```
# smoke_ai_loop.repl
/submit "vendor invoice from Acme for $1200 INR" --label "smoke-test"
/qstat
/process
/tools
/proposal
/reqapproval --reason "smoke test approval"
/approve --note "smoke approved"
/document
/trialbalance
```

---

## 6. Configuration and startup

### 6.1 Startup flags

```
--database-url URL     PostgreSQL connection string
                       (default: TEST_DATABASE_URL, then DATABASE_URL)
--org-slug SLUG        Org slug to authenticate into on startup
--email EMAIL          User email for initial authentication
--password PASSWORD    Password (prompted securely from terminal if omitted)
--channel CH           Default inbound request channel (default: repl)
--json                 Default all output to JSON from startup
--no-color             Disable color and box-drawing output
--history-file PATH    History file path (default: ~/.workflow_repl_history)
--timeout DURATION     Per-command context timeout (default: 90s)
--quiet                Suppress decorative output from startup
```

OpenAI credentials source (for `/process`):

```
OPENAI_API_KEY    required when /process is called
OPENAI_MODEL      required when /process is called
```

The REPL starts and is usable for all non-AI commands even when `OPENAI_API_KEY` is not
set. A clear warning is printed only when `/process` is first called without credentials.

### 6.2 Prompt format

```
north-harbor-works (operator)> _
```

The prompt shows the active org slug and role so the developer always knows which context
and authorization boundary they are operating inside.

In unauthenticated state:

```
workflow_app (not logged in)> _
```

### 6.3 Non-interactive and pipe mode

When stdin is not a TTY, the REPL automatically disables prompts and readline, reads
commands line by line, and exits after the last line (or on first error when using
`/script`).

```bash
# pipe a single command
echo "/qstat" | go run ./cmd/repl --database-url "$DATABASE_URL" \
  --org-slug nhw --email op@nhw.test --password p

# run a script file
go run ./cmd/repl \
  --database-url "$DATABASE_URL" \
  --org-slug nhw \
  --email op@nhw.test \
  /script scripts/repl/smoke_ai_loop.repl
```

---

## 7. Service wiring

The REPL calls existing service constructors. No new service interfaces are required.

```go
db := sql.Open("pgx", databaseURL)

services := ReplServices{
    Auth:        identityaccess.NewService(db),
    Intake:      intake.NewService(db),
    AI:          ai.NewService(db),
    Attachments: attachments.NewService(db),
    Documents:   documents.NewService(db),
    Accounting:  accounting.NewService(db, documents.NewService(db)),
    Reporting:   reporting.NewService(db),
    Parties:     parties.NewService(db),
    Inventory:   inventoryops.NewService(db),
    WorkOrders:  workorders.NewService(db),
    Workforce:   workforce.NewService(db),
    // Processor is lazily initialized on first /process call:
    // app.NewOpenAIAgentProcessorFromEnv(db)
}
```

All list and detail commands use `internal/reporting` reads where available. This ensures
the REPL displays the same composite data the web UI reports surface, keeping the
developer view honest with the operator experience.

---

## 8. Relationship to `cmd/verify-agent`

`cmd/verify-agent` remains the canonical scripted tool for CI-style continuity proof and
is not replaced by the REPL.

| Aspect | `cmd/verify-agent` | `cmd/repl` |
|---|---|---|
| Execution | Linear, scripted, exits | Interactive, stateful loop |
| Auth | Creates throwaway verification org | Uses real or named org |
| Output | Flat key=value pairs | Rendered tables, detail blocks, JSON |
| Use case | CI gate, closeout verification | Development, iteration, operator use |
| AI test | Yes (fixed scenario) | Yes (arbitrary scenarios) |
| Script support | No | Yes (Slice 4+) |

Some internal helpers (identity creation, `performJSON`) may be worth extracting into a
shared `internal/devtools` or `internal/testsupport` package if duplication grows across
both commands. That refactor is not required for initial slices.

---

## 9. Safety constraints

The REPL is a power tool. These constraints keep it from undermining the workflow control
model:

1. **No raw SQL escape hatch**: all writes go through service methods; there is no `/sql`
   or `/query` command that bypasses the domain service API
2. **Tool policy is enforced**: `/process` calls the real coordinator with the full tool
   policy engine active; developers cannot ask the REPL to skip policy checks
3. **Draft requests are not processed**: the `queued` state requirement is preserved;
   calling `/process` on a draft returns the same error the service returns
4. **Role-correct authorization**: `/approve` and `/reject` call the same authorization
   checks as the web UI; an operator-role session cannot approve
5. **Org-scoped**: all commands are scoped to the authenticated org; no cross-org reads
   are possible through the REPL context
6. **Mutation commands confirm before executing**: `/delete` and `/closeperiod` print a
   confirmation prompt before executing unless `--force` is passed
7. **No accounting writes outside the service layer**: `/seed` and `/createaccount` call
   `internal/setup` and `internal/accounting` respectively, which already enforce posting
   boundaries

---

## 10. Dependencies

| Package | Purpose | Status |
|---|---|---|
| `bufio` (stdlib) | Initial stdin reading | Already available |
| `golang.org/x/term` | TTY detection (`isatty`) | Likely transitive; add in Slice 1 |
| `github.com/chzyer/readline` | Line editing, history, completion | Add in Slice 4 |

The initial `bufio.Scanner` approach is zero new production dependencies and is the
correct starting point. `readline` is added in Slice 4 when ergonomics become the focus.

---

## 11. Verification plan

### 11.1 Build gate (every slice)

```bash
go build ./cmd/... ./internal/...
```

Must pass cleanly after every slice before moving on.

### 11.2 Slice 1 smoke test — AI hardening loop

```
/whoami                                  → org/user/role printed
/submit "vendor invoice Acme $1200 INR"  → REQ-... printed, $LAST_REQ set
/qstat                                   → queued=1
/process                                 → run summary printed, $LAST_RUN, $LAST_REC set
/tools                                   → tool trace table printed
/steps                                   → step list printed
/proposal                                → proposal detail printed
/requests --limit 5                      → table of 5 requests printed
```

### 11.3 Slice 2 smoke test — full chain

```
/submit "vendor invoice Acme $1200 INR"
/process
/reqapproval --reason "finance review"
/approve --note "approved by finance"
/document                                → status=approved printed, $LAST_DOC set
/journal --limit 3                       → journal entries printed
```

Compare output against `go run ./cmd/verify-agent -database-url "$DATABASE_URL" -approval-flow`.
Both must show identical request reference, recommendation, approval, and document IDs.

### 11.4 Slice 3 smoke test — operational surfaces

```
/accounts --class asset                  → asset account list
/taxcodes                                → tax code list
/periods                                 → period list
/items                                   → inventory item list
/stock                                   → stock balance list
/parties --role customer                 → customer party list
/workers                                 → worker list
/audit --limit 5                         → audit event list
/trialbalance                            → trial balance
/org                                     → org detail
/policies                                → tool policy list
```

### 11.5 Slice 4 script mode test

```bash
set -a; source .env; set +a
go run ./cmd/repl \
  --database-url "$DATABASE_URL" \
  --org-slug north-harbor-works \
  --email admin@nhw.test \
  --password yourpassword \
  /script scripts/repl/smoke_ai_loop.repl
echo "exit code: $?"
```

Exit code must be 0.

### 11.6 Canonical regression suite (after every slice)

```bash
go build ./cmd/... ./internal/...
set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...
```

---

## 12. Future extensions

1. **`--via-api` mode**: route commands through the live served `/api/...` seam using
   session cookies instead of direct service calls — useful for HTTP middleware and
   response-mapping validation
2. **in-browser terminal widget**: a WebSocket-backed terminal inside the web UI giving
   operator power users the same interactive inspection surface without leaving the browser
3. **named scenarios**: `/scenario save NAME` / `/scenario run NAME` — save a sequence of
   REPL commands as a named scenario for repeatable testing
4. **diff mode**: `/process` records structured output; a subsequent `/diff` command
   compares the new output against a saved baseline to surface prompt regressions
5. **watch mode**: `/watch REQ` polls and re-prints request status every few seconds until
   a terminal state is reached, useful for observing async AI processing
6. **auto-submit-and-process shortcut**: `/ask "text"` — combines `/submit` + `/process`
   in one command for rapid prompt iteration

---

## 13. Implementation summary

| Slice | Focus | Priority | Key commands added |
|---|---|---|---|
| 1 | Foundation + AI hardening loop | **Immediate** | `/help` `/login` `/whoami` `/submit` `/draft` `/queue` `/cancel` `/request` `/requests` `/qstat` `/queued` `/process` `/runs` `/run` `/steps` `/step` `/tools` `/delegation` `/artifact` `/proposals` `/proposal` |
| 2 | Full workflow lifecycle | High | `/amend` `/delete` `/approvals` `/approval` `/reqapproval` `/approve` `/reject` `/documents` `/document` `/journal` `/entry` `/balances` |
| 3 | All operational surfaces | High | `/accounts` `/account` `/createaccount` `/taxcodes` `/taxcode` `/createtaxcode` `/periods` `/period` `/createperiod` `/closeperiod` `/trialbalance` `/balancesheet` `/income` `/items` `/item` `/locations` `/location` `/stock` `/movements` `/movement` `/workorders` `/workorder` `/wostatus` `/womaterials` `/wosync` `/parties` `/party` `/contacts` `/contact` `/workers` `/worker` `/labor` `/attachments` `/attachment` `/audit` `/event` `/org` `/users` `/bootstrap` `/seed` `/policies` `/setpolicy` |
| 4 | Developer ergonomics | Medium | `/history` `/clear` `/script` `/json` `/quiet` + readline + persistent history + tab completion |
| 5 | Script scenarios + CI | Medium | canonical `.repl` smoke scripts, `/scenario` save/run, exit-code contract |

---

## 14. Cross-references

- `cmd/verify-agent/main.go` — existing scripted verification; closest implementation
  precedent for service wiring and actor creation
- `docs/technical_guides/03_inbound_request_lifecycle.md` — request state model and
  queue claim behavior
- `docs/technical_guides/04_ai_agent_architecture.md` — coordinator, provider, run, and
  tool policy model
- `docs/technical_guides/05_web_and_api_seams.md` — API seam shape and handler assembly
- `docs/technical_guides/06_identity_session_auth.md` — actor, session, and auth model
- `docs/technical_guides/07_testing_and_verification.md` — canonical verification commands
- `new_app_docs/new_app_tracker_v2.md` — active milestone status and implementation order
- `new_app_docs/new_app_scope_v2.md` — scope boundaries and active focus
