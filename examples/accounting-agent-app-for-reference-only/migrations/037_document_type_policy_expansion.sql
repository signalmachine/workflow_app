-- Migration 037: add operational payment document types used by posting policy.

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
    ('RC', 'Receipt', false, true, true,  false, 'global', false),
    ('PV', 'Payment Voucher', false, true, false, true,  'global', false)
ON CONFLICT (code) DO UPDATE
SET
    name = EXCLUDED.name,
    affects_inventory = EXCLUDED.affects_inventory,
    affects_gl = EXCLUDED.affects_gl,
    affects_ar = EXCLUDED.affects_ar,
    affects_ap = EXCLUDED.affects_ap,
    numbering_strategy = EXCLUDED.numbering_strategy,
    resets_every_fy = EXCLUDED.resets_every_fy;
