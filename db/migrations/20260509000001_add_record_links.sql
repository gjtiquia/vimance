-- +goose Up
-- +goose StatementBegin

CREATE TABLE record_links (
    parent_id INTEGER NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    child_id  INTEGER NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id),
    PRIMARY KEY (parent_id, child_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS record_links;

-- +goose StatementEnd
