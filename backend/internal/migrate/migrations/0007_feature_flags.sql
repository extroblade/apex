-- Backend feature toggles. Seed iRacing OAuth OFF for everyone (registration is
-- paused by iRacing); the planner works without it via the seeded catalog.

CREATE TABLE IF NOT EXISTS feature_flags (
    flag_key    VARCHAR(64)  NOT NULL,
    enabled     TINYINT(1)   NOT NULL DEFAULT 0,
    description VARCHAR(255)  NOT NULL DEFAULT '',
    updated_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (flag_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO feature_flags (flag_key, enabled, description) VALUES
    ('iracing_oauth', 0, 'iRacing OAuth linking, driver lookup, live stats, sync, comparators')
ON DUPLICATE KEY UPDATE description = VALUES(description);
