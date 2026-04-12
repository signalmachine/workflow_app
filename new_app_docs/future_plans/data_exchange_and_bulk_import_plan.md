# workflow_app Future Data Exchange and Bulk Import Plan

Date: 2026-04-10
Status: Future implementation candidate
Purpose: define a future v2 implementation candidate for structured data exchange after the Milestone 14 runtime has gone through extensive user testing, starting with bulk master-data creation through CSV upload and export of reports and lists in CSV or Excel-compatible formats.

## 1. Why this plan exists

Once Milestone 14 closes the current readiness, workflow, reporting, navigation, and demo-data gaps, the immediate next operating step is extensive user testing on that corrected runtime. Data-exchange implementation should remain future work until that testing period has produced and triaged its findings.

After that user-testing period, the next strong production-shape need is expected to be data exchange.

Operators will need:

1. a practical way to create or extend master data in bulk without re-entering each record manually
2. a practical way to take lists and reports out of the system for downstream analysis, review, or external use
3. import and export flows that stay governed, auditable, and aligned with the shared backend truth model

The first step in that direction should be CSV-first rather than Excel-first.

Reason:

1. CSV import is materially simpler and lower-risk than native Excel workbook import
2. CSV is easier to validate, diff, test, document, and troubleshoot on the backend
3. CSV can still serve most bulk master-data upload needs when templates and validation feedback are strong
4. export can support both CSV and Excel-compatible delivery, but the import side should start with the simpler bounded contract

## 2. Plan objectives

This future plan should achieve the following:

1. add bulk master-data creation through CSV upload on the shared backend seam
2. add export for promoted reports and lists in CSV and, where justified, Excel-compatible form
3. keep import validation, error reporting, and persistence backend-owned and auditable
4. document the templates, field rules, and operator workflow for import and export

## 3. Scope

In scope:

1. CSV upload for bulk creation of selected master data
2. downloadable templates for those CSV imports
3. import validation feedback that is precise enough for operators to correct and retry files
4. export for promoted reports and lists in CSV
5. export for selected reports and lists in Excel-compatible form when the export path is justified and bounded
6. admin or reporting UI surfaces for initiating import and export
7. tests and docs for the import and export contracts

Out of scope for the first data-exchange pass:

1. broad Excel workbook import with multiple sheets, formulas, or loose spreadsheet parsing
2. unrestricted import into every module at once
3. browser-owned import mapping logic that bypasses shared backend validation
4. edit-heavy spreadsheet-in-the-browser product behavior

## 4. First feature set

The first data-exchange feature set should include:

1. bulk master-data creation through CSV upload
2. export of reports and lists in CSV
3. export of selected reports and lists in Excel-compatible form

Recommended import starting set:

1. chart of accounts
2. parties
3. inventory items
4. inventory locations

Recommended export starting set:

1. chart of accounts
2. parties list
3. inventory item and location lists
4. trial balance
5. balance sheet
6. income statement

## 5. Delivery slices

### 5.1 Slice 1: CSV import foundation

Goal:

1. define and land the shared backend contract for CSV-driven bulk master-data creation

Scope:

1. upload endpoint and processing flow
2. CSV template definitions
3. row-level validation and error reporting
4. auditable result summaries

### 5.2 Slice 2: bulk master-data import surfaces

Goal:

1. expose bounded browser surfaces for import templates, upload, validation feedback, and result review

Scope:

1. admin-side import entry points
2. per-entity import pages or grouped import pages
3. clear operator guidance for correction and retry

### 5.3 Slice 3: report and list export

Goal:

1. let operators download key reports and lists for offline review and external use

Scope:

1. CSV export for promoted lists and reports
2. Excel-compatible export for the most important reports where justified
3. focused verification for exported shape and field continuity

## 6. Architecture rules

1. import parsing, validation, and persistence stay on the shared Go backend
2. browser surfaces should initiate import and show validation results, not own import logic
3. export should derive from the same backend-owned reporting and list seams used by the application UI
4. imported master data should follow the same domain validation and audit rules as manual creation
5. CSV-first is the default import posture; native Excel import should be deferred unless a later bounded need justifies the extra complexity

## 7. Verification

Future data-exchange work should include:

1. focused backend tests for parsing, validation, and persistence
2. focused browser or route tests for upload and export entry points
3. sample-file validation tests for both success and failure cases
4. documentation review to ensure templates and error feedback match the real contract

## 8. Documentation sync

When this data-exchange plan begins:

1. confirm the post-Milestone-14 user-testing findings have been reviewed and any required corrective work has been planned or completed
2. update the active tracker and execution plan to promote it from future candidate to active implementation work
3. add operator documentation for upload templates and export behavior
4. add technical documentation for backend import and export seams
5. keep sample templates and docs aligned with the real validated schema
