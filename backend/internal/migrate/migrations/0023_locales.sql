-- Backend-driven i18n: the list of languages and their translation bundles live
-- in the DB, served at /api/locales and /api/locales/{code}. `en` is also
-- bundled in the frontend as the instant fallback + the Translation type source,
-- so the frontend never needs to fetch it. Built-ins (en, ru) are re-seeded from
-- embedded JSON on startup; a new language is just a row here (or a future
-- Cockpit editor) and appears in the app with no frontend deploy.

CREATE TABLE IF NOT EXISTS locales (
    code       VARCHAR(16)  NOT NULL PRIMARY KEY,
    name       VARCHAR(64)  NOT NULL,
    bundle     LONGTEXT     NOT NULL,
    sort_order INT          NOT NULL DEFAULT 0,
    enabled    TINYINT(1)   NOT NULL DEFAULT 1
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
