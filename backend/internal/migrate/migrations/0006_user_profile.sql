-- Profile fields: a display nickname and an optional avatar stored as a
-- data URL (data:image/...;base64,...). MEDIUMTEXT holds a small image inline;
-- the app caps the size on upload.

ALTER TABLE users
    ADD COLUMN nickname        VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN avatar_data_url MEDIUMTEXT  NULL;
