-- +goose Up
-- +goose StatementBegin

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

-- Pinned tabs table
CREATE TABLE pinned_tabs (
    tab_id INTEGER PRIMARY KEY,
    position INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id),
    updated_at INTEGER NOT NULL,
    updated_by INTEGER NOT NULL REFERENCES users(id)
);

-- Views for active records
CREATE VIEW active_users AS
SELECT * FROM users WHERE deleted_at IS NULL;

CREATE VIEW active_tags AS
SELECT * FROM tags WHERE deleted_at IS NULL;

CREATE VIEW active_records AS
SELECT * FROM records WHERE deleted_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP VIEW IF EXISTS active_records;
DROP VIEW IF EXISTS active_tags;
DROP VIEW IF EXISTS active_users;
DROP TABLE IF EXISTS pinned_tabs;
DROP TABLE IF EXISTS records_tags;
DROP TABLE IF EXISTS records;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS currencies;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd
