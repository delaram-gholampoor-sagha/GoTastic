CREATE TABLE IF NOT EXISTS outbox (
                                      id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                                      aggregate_type VARCHAR(64)     NOT NULL,
    aggregate_id   VARCHAR(64)     NOT NULL,
    event_type     VARCHAR(128)    NOT NULL,
    payload        JSON            NOT NULL,
    headers        JSON            NULL,
    status         ENUM('pending','published','failed') NOT NULL DEFAULT 'pending',
    attempts       INT UNSIGNED    NOT NULL DEFAULT 0,
    available_at   DATETIME(6)     NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    locked_until   DATETIME(6)     NULL,
    lock_token     CHAR(36)        NULL,
    created_at     DATETIME(6)     NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    published_at   DATETIME(6)     NULL,
    error          TEXT            NULL,
    PRIMARY KEY (id),
    KEY idx_outbox_pending_window (status, available_at, locked_until, id)
    ) ENGINE=InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci;
