-- Migration 032: Seed GST/TDS-ready chart-of-accounts entries for Company 1000.
-- Scope: master data only (no tax engine logic, no rule mapping changes).
-- Safe to re-run.

INSERT INTO accounts (company_id, code, name, type)
SELECT c.id, a.code, a.name, a.type
FROM companies c
CROSS JOIN (VALUES
    -- Input tax credits (assets)
    ('1350', 'ITC - CGST', 'asset'),
    ('1360', 'ITC - SGST', 'asset'),
    ('1370', 'ITC - IGST', 'asset'),
    ('1380', 'ITC - CESS', 'asset'),

    -- Output tax liabilities
    ('2150', 'GST Payable - CGST', 'liability'),
    ('2160', 'GST Payable - SGST', 'liability'),
    ('2170', 'GST Payable - IGST', 'liability'),
    ('2180', 'GST Payable - CESS', 'liability'),

    -- Direct tax placeholders (future phases)
    ('2190', 'TDS Payable', 'liability'),
    ('2195', 'TCS Payable', 'liability')
) AS a(code, name, type)
WHERE c.company_code = '1000'
ON CONFLICT (company_id, code) DO NOTHING;
