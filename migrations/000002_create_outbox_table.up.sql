CREATE TABLE IF NOT EXISTS Outbox (
                                      ID             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                                      AggregateType  VARCHAR(64)     NOT NULL,
    AggregateID    VARCHAR(64)     NOT NULL,
    EventType      VARCHAR(128)    NOT NULL,
    Payload        JSON            NOT NULL,
    Headers        JSON            NULL,
    Status         ENUM('pending','published','failed') NOT NULL DEFAULT 'pending',
    Attempts       INT UNSIGNED    NOT NULL DEFAULT 0,
    AvailableAt    DATETIME(6)     NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (ID),
    KEY idx_outbox_pending_window (Status, AvailableAt, ID)
    ) ENGINE=InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci;