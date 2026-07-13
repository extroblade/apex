-- Tracks season-schedule PDF imports so the daily checker only re-parses new
-- files (keyed by content hash).

CREATE TABLE IF NOT EXISTS schedule_imports (
    sha256        CHAR(64)     NOT NULL,
    source        VARCHAR(512) NOT NULL,
    series_found  INT          NOT NULL DEFAULT 0,
    series_matched INT         NOT NULL DEFAULT 0,
    weeks_updated INT          NOT NULL DEFAULT 0,
    imported_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (sha256)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
