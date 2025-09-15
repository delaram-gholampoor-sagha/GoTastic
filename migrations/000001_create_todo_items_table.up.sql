
CREATE TABLE IF NOT EXISTS TodoItem (
                       ID BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
                       UUID VARCHAR(36) NOT NULL UNIQUE,
                       Description VARCHAR(255) NOT NULL,
                       DueDate DATETIME,
                       FileID VARCHAR(255),
                       CreatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       UpdatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                       INDEX idx_uuid (UUID),
                       INDEX idx_due_date (DueDate),
                       INDEX idx_file_id (FileID)
)ENGINE=InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci;