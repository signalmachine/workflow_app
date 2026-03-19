# service_day Data exchange plan v1

Date: 2026-03-16
Status: Legacy reference
Purpose: preserve the later data-exchange planning from the broader roadmap as historical reference.

## 1. Intent

`service_day` should be able to support later Excel and CSV import/export workflows where they materially help real operators move data into or out of the system.

This is not a requirement to build spreadsheet-driven product workflows as the primary operating model.

The intended direction is:
1. support later controlled CSV and spreadsheet import for selected high-value migration, bulk-edit, and operational data-entry workflows
2. support later controlled export for reporting, finance, migration, and interoperability needs
3. prepare early schema, API, audit, and attachment foundations so data exchange can land later without bypassing normal product rules

## 2. Product stance

Rules:
1. the product should reduce dependence on spreadsheets for day-to-day system-of-record work
2. spreadsheet import and export should still be available where bulk entry, migration, offline review, or external interoperability are practical
3. imports must not become a backdoor around domain validation, permissions, approval rules, or audit boundaries
4. exports must respect tenant boundaries, user permissions, and sensitive-data controls
5. CSV should be treated as the minimum interoperable exchange format; Excel support is valuable later but should not be the only path

## 3. Expected later use cases

Likely later import use cases:
1. CRM account, contact, and lead migration
2. item, price, or estimate-line seeding where a structured import is faster than manual entry
3. accounting or billing journal-support uploads where policy allows preparation but not direct posting
4. narrow channel-order or settlement import for later external commerce or channel workflows

Likely later export use cases:
1. CRM list and pipeline extracts
2. estimate, invoice, receipt, and accounting exports
3. reporting and reconciliation extracts
4. migration and backup-friendly operational extracts

## 4. Core rules

### 4.1 Import rules

Imports should:
1. enter through explicit import jobs or equivalent bounded workflows
2. validate through the same domain rules used by normal product writes
3. preserve row-level error visibility so operators can correct bad input safely
4. remain idempotent or replay-safe where duplicate execution would be harmful
5. support staged review for sensitive or high-volume imports when policy requires it

### 4.2 Export rules

Exports should:
1. use explicit query/read contracts rather than direct table dumps as the normal product path
2. preserve permission-aware field visibility
3. support stable column naming where downstream reuse is expected
4. remain derivable from canonical domain truth or approved read models rather than from ad hoc client-only transformations

### 4.3 File and evidence rules

Imported source files and generated export files may need bounded persistence.

Rules:
1. uploaded import files should use normal attachment or file-transport contracts rather than hidden local-only side paths
2. import runs should retain enough metadata for audit, support, and replay diagnosis
3. the system should be able to answer who imported what, when, into which org, with which result status

## 5. Early foundational implications

This later capability does require some early technical discipline.

Prepare now:
1. keep importable business flows available through explicit domain services rather than direct table mutation
2. keep write paths idempotent where bulk retry or batch replay could otherwise create duplicates
3. keep attachment/file transport contracts generic enough to carry later import files and export artifacts
4. keep audit boundaries explicit so import-created business changes remain explainable
5. keep reporting/read-model seams explicit so exports do not depend on unstable frontend-only field assembly
6. keep background-job or batch-processing seams possible for later long-running imports and exports

Do not require now:
1. a full batch-processing subsystem before current milestones need it
2. a universal spreadsheet abstraction for every module before real import/export workflows are chosen
3. Excel-specific parsing choices to be locked before concrete implementation work begins

## 6. Preferred architecture direction

Recommended model:
1. a later `data_exchange` or equivalent bounded capability may orchestrate import/export jobs
2. domain modules still own validation, write semantics, and business invariants
3. reporting or other read-model layers may support export-friendly projections where transactional reads would otherwise be too raw or expensive
4. attachment/file-transport infrastructure should remain reusable for import sources and generated exports

Avoid:
1. direct spreadsheet-to-table writes
2. module-specific hidden import code that bypasses service contracts
3. one-off exports assembled entirely in frontend code
4. treating Excel formulas or workbook structure as the canonical business model

## 7. Schema and persistence posture

The current schema does not need to ship full import/export tables immediately, but future support should remain easy to add.

Likely later records:
1. `data_exchange_jobs`
2. `data_exchange_job_items` or equivalent row/result detail
3. references to source attachments or generated export artifacts

Design guidance:
1. keep those future records tenant-safe
2. keep them metadata-oriented rather than duplicating the imported business truth
3. separate import job tracking from the business records created by that import

## 8. Acceptance direction for later implementation

The later spreadsheet/data-exchange capability should be considered technically solid when:
1. high-value imports run through explicit job boundaries
2. row-level validation failures are visible and correctable
3. duplicate import execution cannot silently corrupt business state
4. exports reflect authorized, stable business or reporting views
5. audit, permissions, and tenant safety remain intact throughout
