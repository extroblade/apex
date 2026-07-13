-- Free/paid content model. is_free marks default content (green in the
-- planner); sku_group ties track configs that come with one purchase; and
-- track_requirements models combined layouts that need several purchases
-- (e.g. Nürburgring Combined = GP + Nordschleife).

ALTER TABLE cars
    ADD COLUMN is_free TINYINT(1)  NOT NULL DEFAULT 0,
    ADD COLUMN price   DECIMAL(6,2) NOT NULL DEFAULT 0;

ALTER TABLE tracks
    ADD COLUMN is_free   TINYINT(1)   NOT NULL DEFAULT 0,
    ADD COLUMN price     DECIMAL(6,2) NOT NULL DEFAULT 0,
    ADD COLUMN sku_group INT          NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS track_requirements (
    track_id          INT NOT NULL,
    requires_track_id INT NOT NULL,
    PRIMARY KEY (track_id, requires_track_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
