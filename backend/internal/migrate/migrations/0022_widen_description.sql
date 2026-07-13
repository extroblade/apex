-- Scraped catalog descriptions (from the iRacing detail pages) can exceed the
-- original 600-char limit, which aborted the content-sync fill with
-- "Error 1406: Data too long for column 'description'". Widen to 2000 chars;
-- the fill step also caps the text it extracts. Additive/safe: MODIFY preserves
-- existing values and keeps the '' default, so catalog upserts that omit
-- description still work.

ALTER TABLE cars   MODIFY COLUMN description VARCHAR(2000) NOT NULL DEFAULT '';
ALTER TABLE tracks MODIFY COLUMN description VARCHAR(2000) NOT NULL DEFAULT '';
ALTER TABLE series MODIFY COLUMN description VARCHAR(2000) NOT NULL DEFAULT '';
