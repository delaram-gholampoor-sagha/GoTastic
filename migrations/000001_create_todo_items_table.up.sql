-- base table used by the app
CREATE TABLE IF NOT EXISTS todos (
                                     id          CHAR(36)       NOT NULL,
    description VARCHAR(255)   NOT NULL,
    due_date    DATETIME(6)    NULL,
    file_id     VARCHAR(255)   NULL,
    created_at  DATETIME(6)    NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at  DATETIME(6)    NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_todos_due_date (due_date)
    ) ENGINE=InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci;
