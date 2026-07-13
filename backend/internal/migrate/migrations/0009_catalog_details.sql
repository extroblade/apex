-- Catalog detail fields for the info view + filters. Additive only (new columns
-- with defaults), so it applies to an existing database without touching data.

ALTER TABLE cars
    ADD COLUMN category    VARCHAR(64)  NOT NULL DEFAULT '',
    ADD COLUMN description VARCHAR(600) NOT NULL DEFAULT '';

ALTER TABLE tracks
    ADD COLUMN description VARCHAR(600) NOT NULL DEFAULT '';

ALTER TABLE series
    ADD COLUMN description VARCHAR(600) NOT NULL DEFAULT '';
