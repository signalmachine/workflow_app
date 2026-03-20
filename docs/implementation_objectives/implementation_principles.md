Everything is a ledger + documents + execution context

Status: Reference-only implementation principles guidance

Use this as a reference note for designing or refactoring business systems. It provides implementation-principles guidance, but the implementation principles for `workflow_app` are not limited to this document alone.

---

## Core Idea

Everything in the system is understood through three things:

* **Documents** — why something happened
* **Ledger** — what value changed
* **Execution Context** — how it actually happened

If a feature does not map cleanly to these three, the model is incomplete.

---

## 1. Documents (Intent Layer)

A document represents a **business decision or event**.

Examples:

* invoice
* payment
* purchase order
* delivery
* timesheet

Rules:

* A document captures **intent**, not outcome.
* Documents are created first, validated, then optionally **posted**.
* Documents may be edited only until they are finalized (e.g., `draft → confirmed → posted`).
* Documents never directly store financial or inventory truth.
* Documents must be traceable and human-readable.

Think:

> A document answers: *“What did the business decide or agree?”*

---

## 2. Ledger (Truth Layer)

The ledger represents **what actually changed in value**.

There are typically multiple ledgers:

* financial ledger (money)
* inventory ledger (quantity)

Rules:

* The ledger is **append-only**. No updates, no deletes.
* Every change must be recorded as a **balanced movement**.
* Balance is enforced by the database (not by application logic).
* All derived values (balances, stock, profit) come from ledger aggregation.
* If a value cannot be derived from the ledger, it is not trustworthy.

Key invariant:

> Nothing is created or destroyed — only moved or reclassified.

Think:

> The ledger answers: *“What changed, in measurable terms?”*

---

## 3. Execution Context (Process Layer)

Execution context represents **how work happens in reality**.

Examples:

* projects (services)
* timesheets
* deliveries
* GRNs (goods received)
* workflows

Rules:

* Execution context explains real-world processes but does not itself define financial truth.
* It may trigger documents or be referenced by them.
* It is allowed to be messy and domain-specific — unlike ledger.
* It must always be linkable to documents for traceability.

Think:

> Execution context answers: *“What actually happened operationally?”*

---

## 4. Strict Separation of Responsibilities

Never mix responsibilities across layers.

* Documents must not behave like ledgers.
* Ledgers must not contain business workflow logic.
* Execution context must not directly manipulate financial truth.

If you see:

* “balance” stored in a document → wrong
* “status” inside ledger → wrong
* “financial logic” inside execution tables → wrong

---

## 5. Posting: The Bridge

Posting is the controlled transformation:

> Document → Ledger entries

Rules:

* Only **posted documents** affect the ledger.
* Posting must be:

  * deterministic
  * idempotent
  * transactional
* A document must never be posted more than once.
* Posting logic must be explicit and centralized.

---

## 6. Derivation Over Storage

Never store what can be derived.

* account balance = sum of ledger entries
* stock = sum of inventory movements
* invoice outstanding = invoice total − allocations

You may cache results (snapshots), but:

> Ledger remains the only source of truth.

---

## 7. Double-Entry Thinking (Generalized)

Apply this beyond accounting.

Every meaningful change should have:

* a source
* a destination

Examples:

* money: cash → revenue
* stock: warehouse → customer
* allocation: payment → invoice

Rule:

> Every movement must balance globally.

---

## 8. Items Are Unified

There is no fundamental difference between:

* product
* service

Both are “items” with different behavior.

Rules:

* behavior is controlled via attributes (e.g., `has_inventory`)
* documents treat them uniformly
* downstream effects differ (inventory vs non-inventory)

---

## 9. Parties Are Unified

A party is any external entity.

Rules:

* do not split into separate customer/vendor tables
* a party can play multiple roles
* financial accounts are linked, not embedded

---

## 10. State Is a Projection, Not Storage

Avoid storing mutable state as truth.

Bad:

* `balance` column
* `stock` column

Good:

* compute from ledger
* cache separately if needed

---

## 11. Immutability as Default

Critical tables must be append-only:

* financial ledger
* inventory ledger

Corrections are done via:

* reversal entries
* compensating transactions

Never by mutation.

---

## 12. Constraints Over Code

Business invariants must be enforced at the database level.

Examples:

* ledger must balance
* no duplicate posting
* foreign key integrity
* valid state transitions

Application code is not trusted to maintain correctness.

---

## 13. Idempotency Everywhere

Operations must be safe to retry.

* posting must not duplicate
* payments must not double-apply
* document creation must handle retries

Use:

* unique constraints
* idempotency keys

---

## 14. Execution Does Not Equal Accounting

Do not assume:

* delivery = revenue
* work done = invoice
* invoice = payment

Each step is independent and must be modeled separately.

---

## 15. Time Is First-Class

Everything must be time-aware.

* ledger entries have timestamps
* documents have dates
* reports are time-bounded

This enables:

* auditing
* reconstruction
* historical reporting

---

## 16. Reports Are Views, Not Data

All reports must derive from:

* ledger
* documents
* execution context

Never maintain separate “report tables” as truth.

---

## 17. AI Is a Client, Not Authority

AI interacts only through:

* document creation
* controlled operations

AI must not:

* write ledger entries directly
* bypass constraints
* assume correctness

The database remains the final authority.

---

## 18. Error Handling Philosophy

If something is invalid:

> Fail loudly and early.

* reject invalid data
* never silently adjust
* never auto-correct without trace

---

## 19. System Flow (Always Think in This Direction)

```text
User / AI
   ↓
Documents (intent)
   ↓
Posting / Execution
   ↓
Ledgers (truth)
   ↓
Snapshots / Reports (derived)
```

---

## 20. Final Rule (The One That Matters Most)

When designing anything, always ask:

1. What is the **document** here?
2. What **ledger entries** should this produce?
3. What is the **execution context** (if any)?

If you can’t answer all three clearly, the design is incomplete.

---

This is not just a pattern. It’s a constraint-driven way of thinking that keeps systems correct even as complexity grows.
