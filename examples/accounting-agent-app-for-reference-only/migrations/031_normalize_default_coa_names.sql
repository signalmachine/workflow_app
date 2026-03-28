-- Migration 031: Normalize default Company 1000 chart-of-accounts names.
-- This is idempotent and safe to re-run.
-- Scope is intentionally limited to the default seeded company (1000).

UPDATE accounts a
SET name = 'Bank Account'
FROM companies c
WHERE a.company_id = c.id
  AND c.company_code = '1000'
  AND a.code = '1100'
  AND a.name <> 'Bank Account';

UPDATE accounts a
SET name = 'Accounts Receivable'
FROM companies c
WHERE a.company_id = c.id
  AND c.company_code = '1000'
  AND a.code = '1200'
  AND a.name <> 'Accounts Receivable';

UPDATE accounts a
SET name = 'Sales Revenue'
FROM companies c
WHERE a.company_id = c.id
  AND c.company_code = '1000'
  AND a.code = '4000'
  AND a.name <> 'Sales Revenue';

UPDATE accounts a
SET name = 'Cost of Goods Sold'
FROM companies c
WHERE a.company_id = c.id
  AND c.company_code = '1000'
  AND a.code = '5000'
  AND a.name <> 'Cost of Goods Sold';
