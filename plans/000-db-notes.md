- sqlite for simplicity
- `sqlc` for generating go from .sql
- `goose` for migrations

PRAGMA foreign_keys = ON; // enable foreign key constraints, not on by default as sqlite favors backwards compatibility and it didnt hv this before

users
- id
- username // force unique
- created_at // utc int
- created_by // normally reference self, but could be admin/system
- updated_at // utc int
- updated_by // normally reference self, but could be admin/system
- deleted_at // utc int, nullable, for soft deletes, for recovery
- deleted_by // foreign key, nullable, for soft deletes, for recovery

records
// create a view active_records that filters out deleted records
- id
- date (TEXT format 2026-01-02), constraint ISO 8601 (YYYY-MM-DD) which rejects 2026-13-99
- amount_cents (INT, store amount in cents to avoid floating point issues)
- currency // foreign key
- notes // required, but can be empty string
- created_at // utc int
- created_by // foreign key
- updated_at // utc
- updated_by // foreign key
- deleted_at // utc int, nullable, for soft deletes, for recovery
- deleted_by // foreign key, nullable, for soft deletes, for recovery

tags
// create a view active_tags that filters out deleted records
- id
- name // force unique, perhaps enforce lowercase and no space to avoid duplicates
- description // required, but can be empty string, will be shown in filter view
- notes // required, but can be empty string, will not be shown in filter view
- created_at // utc int
- created_by // foreign key
- updated_at // utc int
- updated_by // foreign key
- deleted_at // utc int, nullable, for soft deletes, for recovery
- deleted_by // foreign key, nullable, for soft deletes, for recovery

records_tags
- // composite primary key from record_id and tag_id
- record_id // foreign key // ON DELETE CASCADE, will be cleaned on hard delete
- tag_id // foreign key // ON DELETE CASCADE, will be cleaned on hard delete
- created_at // utc int
- created_by // foreign key
- updated_at // utc int
- updated_by // foreign key

pinned_tags
- tag_id // primary key referencing tags, ON DELETE CASCADE, will be cleaned on hard delete
- position // for ordering pinned tags
- created_at // utc int
- created_by // foreign key
- updated_at // utc int
- updated_by // foreign key

currencies
- id
- code (e.g. USD, CAD, EUR) // force unique TEXT
- created_at // utc int
- created_by // foreign key
- updated_at // utc int
- updated_by // foreign key


