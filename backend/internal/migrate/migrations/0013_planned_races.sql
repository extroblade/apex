-- Races the user plans to run: any number of (series, week) picks per week.

CREATE TABLE IF NOT EXISTS planned_races (
    user_id   BIGINT UNSIGNED NOT NULL,
    series_id INT             NOT NULL,
    week      INT             NOT NULL,
    PRIMARY KEY (user_id, series_id, week),
    CONSTRAINT fk_planned_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
