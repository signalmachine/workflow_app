-- Migration 042: Flexible purchase invoice flow (direct/bypass invoice + settlement + PO close)
-- Idempotent migration. Adds only new tables/columns/indexes.

ALTER TABLE purchase_orders
    ADD COLUMN IF NOT EXISTS closed_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS close_reason TEXT NULL,
    ADD COLUMN IF NOT EXISTS closed_by_user_id INT NULL REFERENCES users(id);

CREATE TABLE IF NOT EXISTS vendor_invoices (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id),
    vendor_id INT NOT NULL REFERENCES vendors(id),
    po_id INT NULL REFERENCES purchase_orders(id),
    source VARCHAR(20) NOT NULL DEFAULT 'direct',
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    invoice_number VARCHAR(100) NOT NULL,
    invoice_date DATE NOT NULL,
    currency VARCHAR(3) NOT NULL,
    exchange_rate NUMERIC(15,6) NOT NULL DEFAULT 1,
    invoice_amount NUMERIC(14,2) NOT NULL,
    amount_paid NUMERIC(14,2) NOT NULL DEFAULT 0,
    last_paid_at TIMESTAMPTZ NULL,
    idempotency_key VARCHAR(120) NOT NULL,
    pi_document_number VARCHAR(30) NULL,
    journal_entry_id INT NULL REFERENCES journal_entries(id),
    created_by_user_id INT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_vendor_invoices_source
        CHECK (source IN ('direct', 'po_strict', 'po_bypass')),
    CONSTRAINT chk_vendor_invoices_status
        CHECK (status IN ('OPEN', 'PARTIALLY_PAID', 'PAID', 'VOID')),
    CONSTRAINT chk_vendor_invoices_invoice_amount_positive
        CHECK (invoice_amount > 0),
    CONSTRAINT chk_vendor_invoices_amount_paid_bounds
        CHECK (amount_paid >= 0 AND amount_paid <= invoice_amount),
    CONSTRAINT uq_vendor_invoices_company_idempotency
        UNIQUE (company_id, idempotency_key)
);

CREATE TABLE IF NOT EXISTS vendor_invoice_lines (
    id SERIAL PRIMARY KEY,
    vendor_invoice_id INT NOT NULL REFERENCES vendor_invoices(id) ON DELETE CASCADE,
    line_number INT NOT NULL,
    description TEXT NOT NULL,
    expense_account_code VARCHAR(20) NOT NULL,
    amount NUMERIC(14,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_vendor_invoice_lines_amount_positive CHECK (amount > 0),
    CONSTRAINT uq_vendor_invoice_lines_number UNIQUE (vendor_invoice_id, line_number)
);

CREATE TABLE IF NOT EXISTS vendor_invoice_payments (
    id SERIAL PRIMARY KEY,
    vendor_invoice_id INT NOT NULL REFERENCES vendor_invoices(id) ON DELETE CASCADE,
    payment_document_number VARCHAR(30) NOT NULL,
    payment_amount NUMERIC(14,2) NOT NULL,
    payment_date DATE NOT NULL,
    journal_entry_id INT NULL REFERENCES journal_entries(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_vendor_invoice_payments_positive CHECK (payment_amount > 0),
    CONSTRAINT uq_vendor_invoice_payment_doc UNIQUE (vendor_invoice_id, payment_document_number)
);

CREATE INDEX IF NOT EXISTS idx_vendor_invoices_company_created_at
    ON vendor_invoices(company_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_vendor_invoices_company_po_id
    ON vendor_invoices(company_id, po_id);

CREATE INDEX IF NOT EXISTS idx_vendor_invoices_company_status
    ON vendor_invoices(company_id, status);

CREATE UNIQUE INDEX IF NOT EXISTS idx_vendor_invoices_company_vendor_invoice_norm
    ON vendor_invoices(company_id, vendor_id, lower(btrim(invoice_number)));

CREATE INDEX IF NOT EXISTS idx_vendor_invoice_payments_invoice_created_at
    ON vendor_invoice_payments(vendor_invoice_id, created_at DESC);
