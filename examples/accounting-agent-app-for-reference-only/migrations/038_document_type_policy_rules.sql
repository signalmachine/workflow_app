-- Migration 038: table-driven document type policy rules.

INSERT INTO document_types (
    code,
    name,
    affects_inventory,
    affects_gl,
    affects_ar,
    affects_ap,
    numbering_strategy,
    resets_every_fy
)
VALUES
    ('JE', 'Journal Entry', false, true, false, false, 'global', false),
    ('SI', 'Sales Invoice', true, true, true, false, 'global', false),
    ('PI', 'Purchase Invoice', true, true, false, true, 'global', false),
    ('SO', 'Sales Order', false, false, true, false, 'global', false),
    ('PO', 'Purchase Order', false, false, false, false, 'global', false),
    ('GR', 'Goods Receipt', true, true, false, true, 'global', false),
    ('GI', 'Goods Issue', true, true, false, false, 'global', false),
    ('RC', 'Receipt', false, true, true, false, 'global', false),
    ('PV', 'Payment Voucher', false, true, false, true, 'global', false)
ON CONFLICT (code) DO UPDATE
SET
    numbering_strategy = EXCLUDED.numbering_strategy,
    resets_every_fy = EXCLUDED.resets_every_fy;

CREATE TABLE IF NOT EXISTS document_type_policies (
    id SERIAL PRIMARY KEY,
    intent_code VARCHAR(40) NOT NULL,
    allowed_document_type VARCHAR(10) NOT NULL REFERENCES document_types(code),
    source VARCHAR(30) NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_document_type_policies_intent_doc_source
ON document_type_policies (intent_code, allowed_document_type, COALESCE(source, ''));

WITH seed(intent_code, allowed_document_type, source, is_active) AS (
    VALUES
        ('manual_adjustment', 'JE', NULL, true),
        ('sales_invoice', 'SI', NULL, true),
        ('purchase_invoice', 'PI', NULL, true),
        ('goods_receipt', 'GR', NULL, true),
        ('goods_issue', 'GI', NULL, true),
        ('customer_receipt', 'RC', NULL, true),
        ('vendor_payment', 'PV', NULL, true)
)
INSERT INTO document_type_policies (intent_code, allowed_document_type, source, is_active)
SELECT s.intent_code, s.allowed_document_type, s.source, s.is_active
FROM seed s
WHERE NOT EXISTS (
    SELECT 1
    FROM document_type_policies p
    WHERE p.intent_code = s.intent_code
      AND p.allowed_document_type = s.allowed_document_type
      AND COALESCE(p.source, '') = COALESCE(s.source, '')
);

UPDATE document_type_policies
SET is_active = true
WHERE (intent_code, allowed_document_type) IN (
    ('manual_adjustment', 'JE'),
    ('sales_invoice', 'SI'),
    ('purchase_invoice', 'PI'),
    ('goods_receipt', 'GR'),
    ('goods_issue', 'GI'),
    ('customer_receipt', 'RC'),
    ('vendor_payment', 'PV')
);
