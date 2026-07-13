-- Phase 4: persisted race history, synced from iRacing per user. This is the
-- data the comparators (Phase 5) aggregate over.

CREATE TABLE IF NOT EXISTS races (
    user_id            BIGINT UNSIGNED NOT NULL,
    subsession_id      BIGINT UNSIGNED NOT NULL,
    series_id          INT             NOT NULL DEFAULT 0,
    series_name        VARCHAR(255)    NOT NULL DEFAULT '',
    category_id        INT             NOT NULL DEFAULT 0,
    car_id             INT             NOT NULL DEFAULT 0,
    track_id           INT             NOT NULL DEFAULT 0,
    track_name         VARCHAR(255)    NOT NULL DEFAULT '',
    start_position     INT             NOT NULL DEFAULT 0,
    finish_position    INT             NOT NULL DEFAULT 0,
    incidents          INT             NOT NULL DEFAULT 0,
    old_irating        INT             NOT NULL DEFAULT 0,
    new_irating        INT             NOT NULL DEFAULT 0,
    laps_complete      INT             NOT NULL DEFAULT 0,
    session_start_time TIMESTAMP       NULL DEFAULT NULL,
    PRIMARY KEY (user_id, subsession_id),
    KEY idx_races_user_category (user_id, category_id),
    KEY idx_races_user_car (user_id, car_id),
    KEY idx_races_user_track (user_id, track_id),
    CONSTRAINT fk_races_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
