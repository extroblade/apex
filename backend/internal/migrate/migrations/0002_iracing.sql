-- Linked iRacing accounts (OAuth). We store the OAuth *refresh token*,
-- encrypted at rest with AES-GCM (see internal/secretbox). One iRacing account
-- per app user. Access tokens are short-lived (~10 min) and kept only in memory.
--
-- NOTE: schema changes only apply to a fresh DB volume (migrations run via
-- docker-entrypoint-initdb.d on first init). After editing, reset with:
--   docker compose -f backend/docker-compose.yml down -v

CREATE TABLE IF NOT EXISTS iracing_accounts (
    user_id           BIGINT UNSIGNED NOT NULL,
    cust_id           BIGINT UNSIGNED NOT NULL,
    display_name      VARCHAR(255)    NOT NULL DEFAULT '',
    refresh_token_enc TEXT            NOT NULL,
    linked_at         TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
                                      ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id),
    KEY idx_iracing_cust (cust_id),
    CONSTRAINT fk_iracing_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
