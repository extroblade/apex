-- Initial schema. Applied automatically on first MySQL container start
-- via docker-entrypoint-initdb.d (see backend/docker-compose.yml).
-- To re-run after changing it during dev, reset the volume:
--   docker compose -f backend/docker-compose.yml down -v

CREATE TABLE IF NOT EXISTS users (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    email         VARCHAR(255)    NOT NULL,
    password_hash VARCHAR(255)    NOT NULL,
    created_at    TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Opaque login sessions. We store only the SHA-256 hash of the token, never
-- the token itself, so a database leak can't be replayed as a valid session.
CREATE TABLE IF NOT EXISTS sessions (
    token_hash CHAR(64)        NOT NULL,
    user_id    BIGINT UNSIGNED NOT NULL,
    expires_at TIMESTAMP       NOT NULL,
    created_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (token_hash),
    KEY idx_sessions_user (user_id),
    CONSTRAINT fk_sessions_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
