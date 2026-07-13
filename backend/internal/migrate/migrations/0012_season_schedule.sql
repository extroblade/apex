-- Season schedule: which track each series runs each week (1..13). Populated
-- by a deterministic seeder on startup; a real iRacing sync can overwrite it
-- later. Additive migration — no existing data touched.

CREATE TABLE IF NOT EXISTS season_schedule (
    series_id INT NOT NULL,
    week      INT NOT NULL,
    track_id  INT NOT NULL,
    PRIMARY KEY (series_id, week)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
