-- Setups showroom: a repository of car setups users can save privately and,
-- when they choose, share publicly. `data` holds the setup as text (an exported
-- .sto's contents or a values dump); track_id = 0 means a baseline/generic setup
-- not tied to a track. Foreign keys cascade so a user's setups vanish with them.

CREATE TABLE IF NOT EXISTS setups (
    id         INT UNSIGNED    NOT NULL AUTO_INCREMENT,
    user_id    BIGINT UNSIGNED NOT NULL,
    car_id     INT             NOT NULL,
    track_id   INT             NOT NULL DEFAULT 0,
    name       VARCHAR(120)    NOT NULL,
    notes      TEXT            NOT NULL,
    data       MEDIUMTEXT      NOT NULL,
    is_public  TINYINT(1)      NOT NULL DEFAULT 0,
    downloads  INT UNSIGNED    NOT NULL DEFAULT 0,
    created_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_setups_user (user_id),
    KEY idx_setups_car (car_id),
    KEY idx_setups_public (is_public),
    CONSTRAINT fk_setups_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
