-- Racing Planner: cached iRacing catalog (global), per-user owned content and
-- favorite series, and the user's manual plan rows.

-- Global catalog, synced from the iRacing API.
CREATE TABLE IF NOT EXISTS cars (
    car_id   INT          NOT NULL,
    car_name VARCHAR(255) NOT NULL,
    PRIMARY KEY (car_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS tracks (
    track_id    INT          NOT NULL,
    track_name  VARCHAR(255) NOT NULL,
    config_name VARCHAR(255) NOT NULL DEFAULT '',
    category    VARCHAR(64)  NOT NULL DEFAULT '',
    PRIMARY KEY (track_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS series (
    series_id   INT          NOT NULL,
    series_name VARCHAR(255) NOT NULL,
    category    VARCHAR(64)  NOT NULL DEFAULT '',
    category_id INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (series_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Per-user ownership / favorites (join tables).
CREATE TABLE IF NOT EXISTS owned_cars (
    user_id BIGINT UNSIGNED NOT NULL,
    car_id  INT             NOT NULL,
    PRIMARY KEY (user_id, car_id),
    CONSTRAINT fk_owned_cars_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS owned_tracks (
    user_id  BIGINT UNSIGNED NOT NULL,
    track_id INT             NOT NULL,
    PRIMARY KEY (user_id, track_id),
    CONSTRAINT fk_owned_tracks_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS favorite_series (
    user_id   BIGINT UNSIGNED NOT NULL,
    series_id INT             NOT NULL,
    PRIMARY KEY (user_id, series_id),
    CONSTRAINT fk_fav_series_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- The user's manual plan rows: series + track + car (+ note).
CREATE TABLE IF NOT EXISTS plan_entries (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id    BIGINT UNSIGNED NOT NULL,
    series_id  INT             NOT NULL DEFAULT 0,
    track_id   INT             NOT NULL DEFAULT 0,
    car_id     INT             NOT NULL DEFAULT 0,
    note       VARCHAR(500)    NOT NULL DEFAULT '',
    created_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_plan_user (user_id),
    CONSTRAINT fk_plan_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
