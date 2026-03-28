# AI Agent GST Readiness Assessment

Date: 2026-03-04  
Status: Planning document only (no implementation changes)

## 1) Purpose

This document captures GST readiness of the current AI agent architecture and implementation, based on code review findings as of 2026-03-04.  
It is intended to guide future GST implementation planning.

## 2) Scope

In scope:
- AI agent flow (routing, tool loop, confirmation path)
- Domain/app boundaries and ledger posting pipeline
- Data model and schema readiness for GST complexity
- Test/readiness implications for future GST phases

Out of scope:
- Implementing GST now
- Changing existing non-GST behavior
- Regulatory interpretation details not yet specified

## 3) Current State Summary

The application is currently designed for simple accounting entries and core operations (orders, inventory, vendors, purchase orders).  
GST is explicitly deferred in project documentation and is not yet implemented in runtime code.

## 4) Architecture Strengths (Good Foundation)

1. Clear separation of layers:
- Web adapters -> Application service -> Core domain/ledger
- Good base for introducing a dedicated GST domain module later.

2. Agent safety model already present:
- Read tools can execute autonomously.
- Write actions are proposed and require explicit human confirmation.
- Confirmation endpoints are role-gated (`FINANCE_MANAGER`, `ADMIN`).

3. Transactional integrity in ledger workflows:
- Core posting paths support transaction-aware commits.
- Existing flows (invoice/payment/PO actions) already use atomic update patterns in critical paths.

4. Configurability direction exists:
- Rule engine abstraction exists for account resolution.
- This can be extended for tax-account mapping and GST-specific rule resolution.

## 5) Readiness Gaps for Complex GST

### Gap A: GST is not yet a first-class domain contract
Current proposal schema and core proposal model do not carry GST-specific attributes such as:
- tax rate references
- component-level tax lines (CGST/SGST/IGST/Cess)
- HSN/SAC metadata
- GSTIN and place-of-supply/jurisdiction context
- ITC eligibility flags

Impact:
- AI cannot reliably produce structured GST-ready accounting payloads.
- Ledger cannot natively validate GST dimensions.

### Gap B: Posting logic is simple-accounting oriented
Current sales/purchase posting flows are designed around basic DR/CR outcomes and do not produce GST component splits.

Impact:
- Complex GST scenarios (interstate/intrastate branching, reverse charge, composition dealer behavior, ITC paths) cannot be represented correctly yet.

### Gap C: Agent toolset has no GST capability yet
Current tools support operational queries and PO/vendor workflows, but there are no GST-focused tools (rate applicability, HSN coverage, GST validation, return previews).

Impact:
- Agent cannot gather structured GST context before proposing tax-sensitive entries.

### Gap D: Upload/document intelligence is image-only
Current chat attachment path supports image MIME types only.

Impact:
- Future invoice/document-heavy GST flows may require additional ingestion paths (e.g., PDFs or structured documents).

### Gap E: Test readiness for GST is currently absent
No GST-focused runtime tests were identified in current core test coverage.

Impact:
- High regression risk once GST logic is introduced unless GST test harness is added with the implementation.

## 6) Readiness Verdict

### Is future GST support architecturally feasible?
Yes.

### Is the system GST-ready today for complex transactions?
No.

### Practical conclusion
The architecture is extensible and suitable for phased GST evolution, but meaningful GST support will require explicit domain/schema/posting/tooling/test expansion rather than prompt-only changes.

## 7) Recommended Sequencing (Planning-Level)

1. Keep current MVP stable until GST work starts.
2. Introduce GST capabilities as a cohesive phase set:
- domain model + migrations
- tax computation/rule services
- posting engine extensions
- agent schema/tool updates
- integration/e2e tests
3. Avoid pre-emptive partial refactors now that are not tied to a GST phase.

Rationale:
- Isolated early changes will likely be reworked once final GST data contracts are introduced.

## 8) GST Readiness Checklist for Future Plan Sign-off

Use this checklist before calling the system "GST-ready":

1. Domain model
- Tax entities and component structures are defined and versioned.
- Product/customer/vendor GST metadata is present and validated.

2. Posting engine
- Sales and purchase postings support componentized GST lines.
- ITC handling is explicit and test-covered.
- Edge cases (RCM/composition/zero-rate/exempt) are encoded and tested.

3. Agent contract
- Structured output schema includes GST dimensions required by posting.
- Tool registry includes GST read/validation tools.
- Clarification behavior is deterministic for missing GST-critical fields.

4. Controls and safety
- Write confirmation flow remains mandatory.
- Role-based restrictions are preserved for GST-impacting actions.

5. Tests and observability
- Unit + integration + scenario tests cover GST-critical workflows.
- AI/tool-path audit logs capture GST-relevant decision points.

6. Compliance outputs
- Return-oriented reporting/exports are generated from authoritative posted data.
- Reconciliation checks exist between operational docs and ledger tax balances.

## 9) Planning Risks to Track

1. Under-specification risk:
- GST edge cases and regulatory details are easy to under-define early.

2. Contract drift risk:
- If AI schema and core posting contracts evolve independently, failures will increase.

3. Regression risk:
- Tax logic can silently alter existing non-tax workflows without strong tests.

4. Scope creep risk:
- Mixing GST, TDS/TCS, period locking, and return exports in one phase can delay delivery.

## 10) Final Recommendation

Proceed with GST as a planned, phased architecture upgrade.  
Do not treat it as an agent prompt enhancement only.  
Use this document as the readiness baseline and build the future implementation plan against the checklist in Section 8.

