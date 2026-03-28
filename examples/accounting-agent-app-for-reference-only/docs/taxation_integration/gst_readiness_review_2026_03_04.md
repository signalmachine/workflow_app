# GST Readiness Review Note

Date: 2026-03-04
Updated: 2026-03-04 (expanded with new review findings)
Scope: Runtime GST readiness for complex workflows and AI-agent execution capability.

## Summary Verdict

- Architecturally extensible: **Yes**
- Strong GST foundation today for complex workflows: **No**
- Overall: Core accounting platform is solid for phased extension, but runtime GST capabilities are not yet implemented.

## Findings

### 1) GST support is deferred and not implemented in runtime code (Requirement gap)

What was observed:
- GST roadmap is explicitly marked deferred and under-specified.
- Current sales invoicing posts AR vs revenue only; no GST component lines (CGST/SGST/IGST).
- Runtime models and posting paths are not yet GST-first.

Evidence references:
- `docs/taxation_integration/Tax_Regulatory_Future_Plan.md:10`
- `docs/taxation_integration/Tax_Regulatory_Future_Plan.md:12`
- `internal/core/order_service.go:443`
- `internal/core/order_service.go:454`

Impact:
- Current behavior is not GST-compliant for Indian tax-bearing invoicing scenarios.

### 2) AI agent does not yet have GST capability surface (Tooling gap)

What was observed:
- Agent routing/system prompt is general-purpose for accounting/operations.
- No live GST tools are registered in the tool registry (e.g., GST rate applicability, HSN coverage, GSTR previews/exports).

Evidence references:
- `internal/ai/agent.go:273`
- `internal/app/app_service.go:1050`
- `internal/app/app_service.go:1235`

Impact:
- AI cannot reliably gather GST context or execute GST-specific compliance workflows today.

### 3) Existing accounting integrity gaps can amplify GST errors (Control gap)

What was observed:
- `ReceivePO` processes lines with per-line posting calls and transitions PO status after loop completion; failure mid-way can leave partial side effects.
- Vendor invoice variance handling is warning-only in runtime flow.

Evidence references:
- `internal/core/purchase_order_service.go:257`
- `internal/core/purchase_order_service.go:333`
- `internal/core/purchase_order_service.go:380`

Impact:
- Once GST components are introduced, partial posting and unresolved variance handling can create larger reconciliation and compliance defects.

### 4) Compliance prerequisites not active in runtime yet (Period/reporting readiness gap)

What was observed:
- Period locking and downstream return-oriented controls remain future-plan items.
- Runtime service/reporting surface currently provides standard accounting reports only.

Evidence references:
- `docs/taxation_integration/Tax_Regulatory_Future_Plan.md:32`
- `internal/app/service.go:23`
- `internal/ai/agent.go:284`

Impact:
- Robust GST filing-period discipline and return workflows are not yet enforceable in production runtime.

## Resolvability Assessment

All identified issues are **resolvable** with phased implementation. None are structural dead-ends.

Reasoning:
- The current architecture already has clear layering (adapter -> app -> core), transactional posting patterns, and an extensible AI tool registry.
- The GST plan already defines target phases for schema, posting, tools, and reporting.
- Required changes are implementation and control hardening tasks, not fundamental re-architecture.

## Suggested Implementation Sequence (High-level)

1. Stabilize accounting integrity controls used by GST-sensitive flows.
2. Introduce GST domain schema + TaxEngine + tax-aware posting.
3. Add GST read tools for the AI agent before GST write actions.
4. Add period-lock and return/report controls.
5. Add GST integration tests and AI workflow tests before rollout.
