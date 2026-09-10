CREATE TABLE IF NOT EXISTS todos (
    id          BIGINT UNSIGNED                              NOT NULL AUTO_INCREMENT,
    user_id     BIGINT UNSIGNED                              NOT NULL,
    title       VARCHAR(255)                                 NOT NULL,
    description TEXT                                         NULL,
    status      ENUM ('pending', 'in_progress', 'completed') NOT NULL DEFAULT 'pending',
    priority    ENUM ('low', 'medium', 'high')               NOT NULL DEFAULT 'medium',
    due_date    DATE                                         NULL,
    created_at  DATETIME(3)                                  NULL,
    updated_at  DATETIME(3)                                  NULL,
    deleted_at  DATETIME(3)                                  NULL,
    PRIMARY KEY (id),
    KEY idx_todos_deleted_at (deleted_at),
    KEY idx_todos_user_status (user_id, status),
    KEY idx_todos_user_priority (user_id, priority),
    KEY idx_todos_user_due_date (user_id, due_date),
    KEY idx_todos_user_created_at (user_id, created_at),
    CONSTRAINT fk_todos_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
