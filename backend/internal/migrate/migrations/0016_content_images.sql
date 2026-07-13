-- Artwork for catalog content. image_path holds a /media-relative path to a
-- car/track/series image (sourced from the iRacing web catalog by the content
-- sync). season_assets stores whole-schedule artwork pulled from the season PDF
-- (currently just the header banner — the PDF has no per-series logos).

ALTER TABLE cars   ADD COLUMN image_path VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE tracks ADD COLUMN image_path VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE series ADD COLUMN image_path VARCHAR(255) NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS season_assets (
    name       VARCHAR(64)  NOT NULL,
    path       VARCHAR(255) NOT NULL,
    sha256     CHAR(64)     NOT NULL,
    updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
