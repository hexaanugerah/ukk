-- Rollback 003 — sistem pembayaran terpusat.

DROP TABLE IF EXISTS payment_webhook_logs;
DROP TABLE IF EXISTS payment_methods;
DROP TABLE IF EXISTS payment_settings;

DROP INDEX idx_payments_settlement_status ON payments;

ALTER TABLE payments
    DROP COLUMN proof_uploaded_at,
    DROP COLUMN proof_url,
    DROP COLUMN settlement_status,
    DROP COLUMN provider;
