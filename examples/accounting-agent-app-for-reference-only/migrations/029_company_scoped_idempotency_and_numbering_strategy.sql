-- Migration 029: company-scoped idempotency and numbering strategy normalization

-- Replace global idempotency uniqueness with company-scoped uniqueness.
DO $$ BEGIN
    ALTER TABLE journal_entries DROP CONSTRAINT IF EXISTS journal_entries_idempotency_key_key;
EXCEPTION WHEN undefined_table THEN NULL;
END $$;

DROP INDEX IF EXISTS journal_entries_idempotency_key_key;
DROP INDEX IF EXISTS journal_entries_company_idempotency_key_idx;

CREATE UNIQUE INDEX IF NOT EXISTS journal_entries_company_idempotency_key_idx
    ON journal_entries (company_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

-- Normalize legacy numbering strategy values and enforce one vocabulary.
UPDATE document_types
SET numbering_strategy = 'global'
WHERE numbering_strategy = 'sequential';

ALTER TABLE document_types DROP CONSTRAINT IF EXISTS chk_document_types_numbering_strategy;
ALTER TABLE document_types
    ADD CONSTRAINT chk_document_types_numbering_strategy
    CHECK (numbering_strategy IN ('global', 'per_fy', 'per_branch'));
