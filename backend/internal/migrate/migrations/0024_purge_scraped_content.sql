-- IP hygiene: purge content that originated from scraping iracing.com.
--
-- image_path only ever held scraped/rehosted iRacing artwork (webcatalog card
-- images) — clear it everywhere; catalog art is generated client-side now and
-- the API no longer serves the field.
--
-- Scraped descriptions were written only to rows with a detail_url (the
-- fillDescriptions step); clear those. Rows whose description came from the
-- authored seed keep it, and SeedCatalog re-applies authored copy on the next
-- startup for any row this touches (the seed only writes non-empty values), so
-- no original content is lost.

UPDATE cars   SET image_path = '';
UPDATE tracks SET image_path = '';
UPDATE series SET image_path = '';

UPDATE cars   SET description = '' WHERE detail_url <> '';
UPDATE tracks SET description = '' WHERE detail_url <> '';
