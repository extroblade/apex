-- Detail-page URLs from the iRacing web catalog (per card). The content sync
-- uses them to fill missing descriptions from each car/track's own page.

ALTER TABLE cars   ADD COLUMN detail_url VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE tracks ADD COLUMN detail_url VARCHAR(255) NOT NULL DEFAULT '';
