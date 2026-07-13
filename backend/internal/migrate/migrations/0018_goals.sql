-- Goal tracker: personal, numeric goals the user works toward (e.g. "reach a B
-- road license", "win 5 races"). target/current are decimals so both counts and
-- ratings fit; unit is a free-text label ("wins", "iRating", "license"). done is
-- derived on save but stored so a goal can be marked complete manually too.

CREATE TABLE IF NOT EXISTS goals (
    id         INT UNSIGNED    NOT NULL AUTO_INCREMENT,
    user_id    BIGINT UNSIGNED NOT NULL,
    title      VARCHAR(120)    NOT NULL,
    notes      TEXT            NOT NULL,
    unit       VARCHAR(40)     NOT NULL DEFAULT '',
    target     DECIMAL(12,2)   NOT NULL DEFAULT 0,
    current    DECIMAL(12,2)   NOT NULL DEFAULT 0,
    done       TINYINT(1)      NOT NULL DEFAULT 0,
    due_date   DATE            NULL,
    created_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_goals_user (user_id),
    CONSTRAINT fk_goals_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
