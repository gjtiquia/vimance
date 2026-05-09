-- +goose Up
-- +goose StatementBegin

ALTER TABLE record_links ADD COLUMN created_by INTEGER NOT NULL REFERENCES users(id) DEFAULT 1;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

CREATE TABLE record_links_temp (
    parent_id INTEGER NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    child_id  INTEGER NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL,
    PRIMARY KEY (parent_id, child_id)
);

INSERT INTO record_links_temp (parent_id, child_id, created_at)
SELECT parent_id, child_id, created_at FROM record_links;

DROP TABLE record_links;

ALTER TABLE record_links_temp RENAME TO record_links;

-- +goose StatementEnd
