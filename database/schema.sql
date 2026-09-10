-- ---------------------------------------------------------------------------
-- Todo Management API - full database schema (MySQL 8.0)
--
-- This file is a readable, consolidated view of the schema. The authoritative,
-- versioned source is the migrations/ directory (run with golang-migrate).
--
--   mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS todo_db"
--   mysql -u root -p todo_db < database/schema.sql
-- ---------------------------------------------------------------------------

CREATE DATABASE IF NOT EXISTS todo_db
    CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE todo_db;

-- Users -------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name       VARCHAR(255)    NOT NULL,
    email      VARCHAR(255)    NOT NULL,
    password   VARCHAR(255)    NOT NULL, -- bcrypt hash, never plaintext
    created_at DATETIME(3)     NULL,
    updated_at DATETIME(3)     NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_users_email (email)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- Todos -----------------------------------------------------------------
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
    deleted_at  DATETIME(3)                                  NULL, -- soft delete
    PRIMARY KEY (id),
    KEY idx_todos_deleted_at (deleted_at),
    KEY idx_todos_user_status (user_id, status),
    KEY idx_todos_user_priority (user_id, priority),
    KEY idx_todos_user_due_date (user_id, due_date),
    KEY idx_todos_user_created_at (user_id, created_at),
    CONSTRAINT fk_todos_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- Categories / tags ---------------------------------------------------
CREATE TABLE IF NOT EXISTS categories (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id    BIGINT UNSIGNED NOT NULL,
    name       VARCHAR(100)    NOT NULL,
    created_at DATETIME(3)     NULL,
    updated_at DATETIME(3)     NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_categories_user_name (user_id, name),
    CONSTRAINT fk_categories_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS todo_categories (
    todo_id     BIGINT UNSIGNED NOT NULL,
    category_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (todo_id, category_id),
    KEY idx_todo_categories_category (category_id),
    CONSTRAINT fk_todo_categories_todo FOREIGN KEY (todo_id) REFERENCES todos (id) ON DELETE CASCADE,
    CONSTRAINT fk_todo_categories_category FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- Migration bookkeeping table used by golang-migrate ------------------
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT  NOT NULL,
    dirty   BOOLEAN NOT NULL,
    PRIMARY KEY (version)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
