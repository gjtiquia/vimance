-- +goose Up
-- +goose StatementBegin

PRAGMA foreign_keys = ON;

-- Users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    deleted_at INTEGER,
    deleted_by INTEGER REFERENCES users(id)
);

-- Currencies table
CREATE TABLE currencies (
    id INTEGER PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Tags table
CREATE TABLE tags (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id),
    updated_at INTEGER NOT NULL,
    updated_by INTEGER NOT NULL REFERENCES users(id),
    deleted_at INTEGER,
    deleted_by INTEGER REFERENCES users(id)
);

-- Records table
CREATE TABLE records (
    id INTEGER PRIMARY KEY,
    date TEXT NOT NULL CHECK (
        date LIKE '____-__-__' AND
        substr(date, 1, 4) BETWEEN '0000' AND '9999' AND
        substr(date, 6, 2) BETWEEN '01' AND '12' AND
        substr(date, 9, 2) BETWEEN '01' AND '31'
    ),
    amount_cents INTEGER NOT NULL,
    currency_id INTEGER NOT NULL REFERENCES currencies(id),
    notes TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id),
    updated_at INTEGER NOT NULL,
    updated_by INTEGER NOT NULL REFERENCES users(id),
    deleted_at INTEGER,
    deleted_by INTEGER REFERENCES users(id)
);

-- Records-Tags junction table
CREATE TABLE records_tags (
    record_id INTEGER NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id),
    updated_at INTEGER NOT NULL,
    updated_by INTEGER NOT NULL REFERENCES users(id),
    PRIMARY KEY (record_id, tag_id)
);

-- Pinned tags table
CREATE TABLE pinned_tags (
    tag_id INTEGER PRIMARY KEY REFERENCES tags(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id),
    updated_at INTEGER NOT NULL,
    updated_by INTEGER NOT NULL REFERENCES users(id),
    UNIQUE(position)
);

-- Record links table (multiple parents per child)
CREATE TABLE record_links (
    parent_id INTEGER NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    child_id  INTEGER NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id),
    PRIMARY KEY (parent_id, child_id),
    CHECK(parent_id != child_id)
);

-- Views for active records
CREATE VIEW active_users AS
SELECT * FROM users WHERE deleted_at IS NULL;

CREATE VIEW active_tags AS
SELECT * FROM tags WHERE deleted_at IS NULL;

CREATE VIEW active_records AS
SELECT * FROM records WHERE deleted_at IS NULL;

-- Saved queries table (hard delete, user convenience config)
CREATE TABLE saved_queries (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    date_from TEXT NOT NULL,
    date_to TEXT NOT NULL,
    currency_id INTEGER REFERENCES currencies(id),
    fuzzy_text TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id),
    updated_at INTEGER NOT NULL,
    updated_by INTEGER NOT NULL REFERENCES users(id),
    CHECK(date_from <= date_to)
);

CREATE TABLE saved_query_tags (
    query_id INTEGER NOT NULL REFERENCES saved_queries(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (query_id, tag_id)
);

CREATE TABLE targets (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    saved_query_id INTEGER NOT NULL REFERENCES saved_queries(id) ON DELETE CASCADE,
    target_cents INTEGER NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    created_by INTEGER NOT NULL REFERENCES users(id),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_by INTEGER NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_targets_saved_query_id ON targets(saved_query_id);

-- Indexes for common query patterns
CREATE INDEX idx_records_date ON records(date);
CREATE INDEX idx_records_currency_id ON records(currency_id);
CREATE INDEX idx_records_deleted_at ON records(deleted_at);
CREATE INDEX idx_records_tags_tag_id ON records_tags(tag_id);
CREATE INDEX idx_record_links_child_id ON record_links(child_id);
CREATE INDEX idx_record_links_parent_id ON record_links(parent_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_records_date;
DROP INDEX IF EXISTS idx_records_currency_id;
DROP INDEX IF EXISTS idx_records_deleted_at;
DROP INDEX IF EXISTS idx_records_tags_tag_id;
DROP INDEX IF EXISTS idx_record_links_child_id;
DROP INDEX IF EXISTS idx_record_links_parent_id;
DROP INDEX IF EXISTS idx_targets_saved_query_id;
DROP TABLE IF EXISTS targets;
DROP VIEW IF EXISTS active_records;
DROP VIEW IF EXISTS active_tags;
DROP VIEW IF EXISTS active_users;
DROP TABLE IF EXISTS saved_query_tags;
DROP TABLE IF EXISTS saved_queries;
DROP TABLE IF EXISTS pinned_tags;
DROP TABLE IF EXISTS record_links;
DROP TABLE IF EXISTS records_tags;
DROP TABLE IF EXISTS records;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS currencies;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd
