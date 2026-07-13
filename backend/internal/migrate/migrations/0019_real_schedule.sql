-- Real season data. series_cars is the precise series → eligible-car mapping
-- (replaces the by-category approximation for "can run"); race_date lets the
-- current week be derived from the actual calendar instead of ISO-week math.

CREATE TABLE IF NOT EXISTS series_cars (
    series_id INT NOT NULL,
    car_id    INT NOT NULL,
    PRIMARY KEY (series_id, car_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE season_schedule ADD COLUMN race_date DATE NULL;
