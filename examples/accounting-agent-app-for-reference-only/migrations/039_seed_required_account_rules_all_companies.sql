-- Migration 039: ensure required core accounts + account_rules exist for all companies.

-- Ensure required account codes exist per company.
INSERT INTO accounts (company_id, code, name, type)
SELECT c.id, a.code, a.name, a.type
FROM companies c
CROSS JOIN (
    VALUES
        ('1100', 'Bank Account', 'asset'),
        ('1200', 'Accounts Receivable', 'asset'),
        ('1400', 'Inventory', 'asset'),
        ('2000', 'Accounts Payable', 'liability'),
        ('5000', 'Cost of Goods Sold', 'expense')
) AS a(code, name, type)
LEFT JOIN accounts existing
       ON existing.company_id = c.id
      AND existing.code = a.code
WHERE existing.id IS NULL;

-- Ensure required core rule mappings exist for every company.
WITH required(rule_type, account_code) AS (
    VALUES
        ('AR', '1200'),
        ('AP', '2000'),
        ('INVENTORY', '1400'),
        ('COGS', '5000'),
        ('BANK_DEFAULT', '1100'),
        ('RECEIPT_CREDIT', '2000')
),
missing AS (
    SELECT c.id AS company_id, r.rule_type, r.account_code
    FROM companies c
    CROSS JOIN required r
    LEFT JOIN account_rules ar
           ON ar.company_id = c.id
          AND ar.rule_type = r.rule_type
    WHERE ar.id IS NULL
)
INSERT INTO account_rules (company_id, rule_type, account_code, account_id, priority, effective_from)
SELECT m.company_id, m.rule_type, m.account_code, a.id, 100, DATE '1970-01-01'
FROM missing m
JOIN accounts a
  ON a.company_id = m.company_id
 AND a.code = m.account_code;

-- Keep account_id aligned for pre-existing rules where it was left NULL.
UPDATE account_rules ar
SET account_id = a.id
FROM accounts a
WHERE ar.company_id = a.company_id
  AND ar.account_code = a.code
  AND ar.account_id IS NULL;
