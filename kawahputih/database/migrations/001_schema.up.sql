-- =============================================================================
-- SISTEM BOOKING TIKET KAWAH PUTIH ONLINE
-- Target schema per ERD & Database Schema:
--   ROLES, USERS, BOOKINGS, PAYMENTS, TICKETS,
--   BOOKING_DETAILS, TICKET_VERIFICATIONS, REVIEWS
--
-- Auth-support tables (refresh_tokens, auth_tokens) are retained so
-- login / refresh / logout / password-reset keep working; they are not part
-- of the ERD but are required by the authentication feature. users.id is
-- BIGINT, so their user_id is BIGINT as well.
--
-- Every table is created only if it does not exist, so this file is safe to
-- apply on top of a fresh MySQL container (docker-entrypoint-initdb.d) or
-- manually.
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. ROLES
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS roles (
    id         BIGINT       NOT NULL AUTO_INCREMENT,
    name       VARCHAR(30)  NOT NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_roles_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- 2. USERS
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id         BIGINT       NOT NULL AUTO_INCREMENT,
    role_id    BIGINT       NOT NULL,
    name       VARCHAR(100) NOT NULL,
    email      VARCHAR(100) NOT NULL,
    password   VARCHAR(255) NOT NULL,
    phone      VARCHAR(20)  NULL,
    status     VARCHAR(20)  NOT NULL DEFAULT 'AKTIF',
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_email (email),
    KEY idx_users_role_id (role_id),
    KEY idx_users_status (status),
    CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- AUTH SUPPORT: refresh tokens (login session) + one-shot auth tokens
-- (email verification / password reset)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         BIGINT       NOT NULL AUTO_INCREMENT,
    user_id    BIGINT       NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at BIGINT       NOT NULL,
    revoked    BOOLEAN      NOT NULL DEFAULT FALSE,
    user_agent VARCHAR(255) NULL,
    ip_address VARCHAR(45)  NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_refresh_tokens_user_id (user_id),
    KEY idx_refresh_tokens_token_hash (token_hash),
    CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS auth_tokens (
    id         BIGINT       NOT NULL AUTO_INCREMENT,
    user_id    BIGINT       NOT NULL,
    purpose    VARCHAR(30)  NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at BIGINT       NOT NULL,
    used_at    BIGINT       NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_auth_tokens_user_purpose (user_id, purpose),
    KEY idx_auth_tokens_token_hash (token_hash),
    CONSTRAINT fk_auth_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- 5. TICKETS (bookable ticket catalog)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tickets (
    id              BIGINT        NOT NULL AUTO_INCREMENT,
    name            VARCHAR(100)  NOT NULL,
    description     TEXT          NULL,
    price           DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    quota           INT           NOT NULL DEFAULT 0,
    available_quota INT           NOT NULL DEFAULT 0,
    status          VARCHAR(20)   NOT NULL DEFAULT 'AKTIF',
    created_at      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_tickets_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- 3. BOOKINGS
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS bookings (
    id           BIGINT        NOT NULL AUTO_INCREMENT,
    user_id      BIGINT        NOT NULL,
    booking_code VARCHAR(30)   NOT NULL,
    visit_date   DATE          NOT NULL,
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    status       VARCHAR(30)   NOT NULL DEFAULT 'PENDING',
    created_at   TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_bookings_booking_code (booking_code),
    KEY idx_bookings_user_id (user_id),
    KEY idx_bookings_visit_date (visit_date),
    KEY idx_bookings_status (status),
    CONSTRAINT fk_bookings_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- 4. PAYMENTS
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS payments (
    id                 BIGINT        NOT NULL AUTO_INCREMENT,
    booking_id         BIGINT        NOT NULL,
    transaction_number VARCHAR(100)  NOT NULL,
    payment_method     VARCHAR(50)   NOT NULL,
    amount             DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    status             VARCHAR(30)   NOT NULL DEFAULT 'PENDING',
    paid_at            TIMESTAMP     NULL,
    created_at         TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_payments_booking_id (booking_id),
    KEY idx_payments_transaction_number (transaction_number),
    KEY idx_payments_status (status),
    CONSTRAINT fk_payments_booking FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- 6. BOOKING_DETAILS (booking line items)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS booking_details (
    id         BIGINT        NOT NULL AUTO_INCREMENT,
    booking_id BIGINT        NOT NULL,
    ticket_id  BIGINT        NOT NULL,
    quantity   INT           NOT NULL DEFAULT 1,
    price      DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    subtotal   DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    PRIMARY KEY (id),
    KEY idx_booking_details_booking_id (booking_id),
    KEY idx_booking_details_ticket_id (ticket_id),
    CONSTRAINT fk_booking_details_booking FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE,
    CONSTRAINT fk_booking_details_ticket FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- 7. TICKET_VERIFICATIONS (e-ticket / gate verification per booking line)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ticket_verifications (
    id                BIGINT       NOT NULL AUTO_INCREMENT,
    booking_detail_id BIGINT       NOT NULL,
    verified_by       BIGINT       NULL,
    qr_code           VARCHAR(255) NOT NULL,
    verified_at       TIMESTAMP    NULL,
    status            VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    PRIMARY KEY (id),
    UNIQUE KEY uk_ticket_verifications_qr_code (qr_code),
    KEY idx_ticket_verifications_booking_detail_id (booking_detail_id),
    KEY idx_ticket_verifications_verified_by (verified_by),
    KEY idx_ticket_verifications_status (status),
    CONSTRAINT fk_ticket_verifications_detail FOREIGN KEY (booking_detail_id) REFERENCES booking_details(id) ON DELETE CASCADE,
    CONSTRAINT fk_ticket_verifications_user FOREIGN KEY (verified_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- 8. REVIEWS
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS reviews (
    id         BIGINT      NOT NULL AUTO_INCREMENT,
    user_id    BIGINT      NOT NULL,
    booking_id BIGINT      NOT NULL,
    rating     INT         NOT NULL DEFAULT 0,
    comment    TEXT        NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_reviews_user_id (user_id),
    KEY idx_reviews_booking_id (booking_id),
    KEY idx_reviews_status (status),
    CONSTRAINT fk_reviews_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_reviews_booking FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
