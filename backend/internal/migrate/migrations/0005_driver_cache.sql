-- Public driver browser: cache of iRacing API payloads, keyed by cust_id + kind
-- ("member" | "career" | "recent"). Served for `ttl` before being re-fetched,
-- so all public traffic through the shared service account stays within limits.

CREATE TABLE IF NOT EXISTS driver_cache (
    cust_id      BIGINT UNSIGNED NOT NULL,
    kind         VARCHAR(32)     NOT NULL,
    payload_json JSON            NOT NULL,
    fetched_at   TIMESTAMP       NOT NULL,
    PRIMARY KEY (cust_id, kind)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
