-- Migration 036: enforce global document numbering uniqueness per (company_id, type_code)

-- Normalize go-live document types to global, no FY reset.
UPDATE document_types
SET numbering_strategy = 'global',
    resets_every_fy = false
WHERE code IN ('JE', 'SI', 'PI', 'SO', 'PO', 'GR', 'GI', 'RC', 'PV');

-- Guard: fail early with a clear message if strict global uniqueness would be violated.
DO $$
DECLARE
    duplicate_count BIGINT;
BEGIN
    SELECT COUNT(*) INTO duplicate_count
    FROM (
        SELECT company_id, type_code, document_number
        FROM documents
        WHERE document_number IS NOT NULL
        GROUP BY company_id, type_code, document_number
        HAVING COUNT(*) > 1
    ) d;

    IF duplicate_count > 0 THEN
        RAISE EXCEPTION 'cannot apply global document uniqueness: found % duplicate (company_id, type_code, document_number) groups', duplicate_count;
    END IF;
END $$;

-- Consolidate historical sequence rows across FY/branch into global rows.
CREATE TEMP TABLE tmp_document_sequences_global AS
SELECT company_id, type_code, MAX(last_number) AS last_number
FROM document_sequences
GROUP BY company_id, type_code;

TRUNCATE TABLE document_sequences;

INSERT INTO document_sequences (company_id, type_code, financial_year, branch_id, last_number)
SELECT company_id, type_code, NULL, NULL, last_number
FROM tmp_document_sequences_global;

DROP INDEX IF EXISTS document_sequences_unique_idx;

CREATE UNIQUE INDEX document_sequences_unique_idx
ON document_sequences (company_id, type_code);

DROP INDEX IF EXISTS documents_unique_number_idx;

CREATE UNIQUE INDEX documents_unique_number_idx
ON documents (company_id, type_code, document_number)
WHERE document_number IS NOT NULL;
