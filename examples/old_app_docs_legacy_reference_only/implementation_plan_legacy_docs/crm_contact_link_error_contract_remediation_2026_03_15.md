# CRM contact-link error-contract remediation note

Date: 2026-03-15
Status: Done
Purpose: record the resolved cross-tenant contact-link regression and the explicit store-level contract now enforced.

## Resolved issue

The broad CRM integration suite had a failing test:

1. `TestStoreIntegrationRejectsCrossTenantContactLinking`

Previously observed behavior:

1. the unsafe cross-tenant contact link is still rejected
2. the rejection now appears as a raw schema/foreign-key failure from `account_contacts_org_account_fk`
3. the existing test still expects the older application-mapped `account not found` style error

Tenant safety still held, but the contract at the store or service boundary had drifted.

## Resolution

1. `internal/crm/store.go` now derives one canonical account-link set for `CreateContact` from either `AccountLinks` or `PrimaryAccountID`
2. the store now pre-validates that canonical link set through `ensureLinkedRecordExists` before any `account_contacts` insert runs
3. direct store callers and service-normalized callers now both get the same application-level `account not found` behavior for cross-tenant account references

## Why this matters

1. a lower-level database error may now escape farther than intended
2. the store and tests no longer agree on whether cross-tenant account references should be classified as `not found` or left as raw constraint failures
3. if this drift reaches HTTP unchanged, clients can see inconsistent error semantics across similar invalid-reference paths

## Chosen contract

1. cross-tenant account references in contact-link creation map to application-level `ErrNotFound`
2. schema-backed foreign keys remain the second enforcement layer, but the store should not leak a raw SQL error for this caller-visible path
3. direct store usage and service-mediated usage must agree on that rule

## Acceptance criteria

1. `TestStoreIntegrationRejectsCrossTenantContactLinking` passes again
2. the chosen error contract is explicit in code rather than accidental
3. similar cross-tenant contact-link paths do not leak inconsistent raw SQL errors
4. tracker notes that currently rely on targeted CRM integration suites can be narrowed once the broad CRM suite is healthy again

## Verification

1. `go test ./internal/crm`
2. `go test ./...`
3. `/bin/bash -lc 'set -a; source .env; set +a; go test -tags integration -count=1 ./internal/crm -run TestStoreIntegrationRejectsCrossTenantContactLinking'`
