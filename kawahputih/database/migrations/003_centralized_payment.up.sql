-- -----------------------------------------------------------------------------
-- 003 — SISTEM PEMBAYARAN TERPUSAT (payment gateway)
--
-- Menambahkan kolom transaksi gateway pada payments serta tabel konfigurasi
-- gateway, kanal metode pembayaran, dan log webhook (idempotency).
--
-- Catatan: pada runtime tabel juga dibuat/diselaraskan oleh GORM AutoMigrate,
-- sehingga migration ini idempotent dan aman dijalankan berulang.
-- -----------------------------------------------------------------------------

-- 1. Kolom tambahan pada payments ---------------------------------------------
ALTER TABLE payments
    ADD COLUMN provider VARCHAR(30) NOT NULL DEFAULT 'midtrans' AFTER booking_id,
    ADD COLUMN settlement_status VARCHAR(20) NOT NULL DEFAULT 'UNSETTLED' AFTER status,
    ADD COLUMN proof_url VARCHAR(500) NULL AFTER paid_at,
    ADD COLUMN proof_uploaded_at TIMESTAMP NULL AFTER proof_url;

CREATE INDEX idx_payments_settlement_status ON payments (settlement_status);

-- 2. Konfigurasi payment gateway ---------------------------------------------
-- Kredensial rahasia (server key / access token) TIDAK disimpan di tabel ini.
CREATE TABLE IF NOT EXISTS payment_settings (
    id                          CHAR(36)     NOT NULL,
    provider                    VARCHAR(30)  NOT NULL DEFAULT 'midtrans',
    environment                 VARCHAR(20)  NOT NULL DEFAULT 'sandbox',
    enabled                     TINYINT(1)   NOT NULL DEFAULT 1,
    merchant_name               VARCHAR(150) NULL,
    settlement_bank             VARCHAR(50)  NULL,
    settlement_account_masked   VARCHAR(50)  NULL,
    settlement_status           VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    last_synced_at              TIMESTAMP    NULL,
    created_at                  TIMESTAMP    NULL,
    updated_at                  TIMESTAMP    NULL,
    deleted_at                  TIMESTAMP    NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_payment_settings_provider (provider),
    KEY idx_payment_settings_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Kanal metode pembayaran --------------------------------------------------
-- enabled di sini hanya mengatur tampilan di aplikasi; ketersediaan final tetap
-- ditentukan akun merchant pada provider.
CREATE TABLE IF NOT EXISTS payment_methods (
    id             CHAR(36)     NOT NULL,
    provider       VARCHAR(30)  NOT NULL DEFAULT 'midtrans',
    method_code    VARCHAR(20)  NOT NULL,
    method_name    VARCHAR(100) NOT NULL,
    enabled        TINYINT(1)   NOT NULL DEFAULT 1,
    display_order  INT          NOT NULL DEFAULT 0,
    created_at     TIMESTAMP    NULL,
    updated_at     TIMESTAMP    NULL,
    deleted_at     TIMESTAMP    NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_payment_methods_code (method_code),
    KEY idx_payment_methods_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. Log webhook provider (idempotency + audit) -------------------------------
CREATE TABLE IF NOT EXISTS payment_webhook_logs (
    id                 CHAR(36)     NOT NULL,
    provider           VARCHAR(30)  NOT NULL DEFAULT 'midtrans',
    transaction_id     VARCHAR(100) NULL,
    notification_id    VARCHAR(150) NOT NULL,
    event_type         VARCHAR(50)  NULL,
    processing_status  VARCHAR(20)  NOT NULL DEFAULT 'RECEIVED',
    received_at        TIMESTAMP    NOT NULL,
    processed_at       TIMESTAMP    NULL,
    created_at         TIMESTAMP    NULL,
    updated_at         TIMESTAMP    NULL,
    deleted_at         TIMESTAMP    NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_payment_webhook_logs_notification (notification_id),
    KEY idx_payment_webhook_logs_transaction_id (transaction_id),
    KEY idx_payment_webhook_logs_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
