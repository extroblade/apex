-- Account lifecycle: email verification + password reset.
--
-- email_verified on users defaults to 0; new accounts are unverified until they
-- click the link in the welcome email. The flag is read by the frontend to
-- show a "verify your email" banner, and the API may (later) gate sensitive
-- flows on it. Existing accounts created before this migration are left
-- unverified — they can verify via the resend endpoint.

ALTER TABLE users
    ADD COLUMN email_verified TINYINT(1) NOT NULL DEFAULT 0;

-- One row per outstanding token. `kind` separates "verify" (email verification)
-- from "reset" (password reset) so the same table serves both flows. We store
-- only the SHA-256 hash of the token (never the token itself), so a DB leak
-- can't be replayed. The composite PK (token_hash, kind) lets a user have one
-- of each kind outstanding at a time. expires_at is checked on consume.
CREATE TABLE IF NOT EXISTS email_tokens (
    token_hash CHAR(64)        NOT NULL,
    user_id    BIGINT UNSIGNED NOT NULL,
    kind        VARCHAR(16)     NOT NULL,
    expires_at  TIMESTAMP       NOT NULL,
    created_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (token_hash, kind),
    KEY idx_email_tokens_user (user_id, kind),
    CONSTRAINT fk_email_tokens_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
