-- License class required to run a series (Rookie / D / C / B / A / Pro).
-- Additive column with a default, so existing data is preserved.

ALTER TABLE series
    ADD COLUMN license_needed VARCHAR(16) NOT NULL DEFAULT '';
