-- name: CreateSavedQuery :one
INSERT INTO saved_queries (name, date_from, date_to, currency_id, fuzzy_text, created_at, created_by, updated_at, updated_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetSavedQuery :one
SELECT * FROM saved_queries WHERE id = ?;

-- name: ListSavedQueries :many
SELECT * FROM saved_queries ORDER BY name;

-- name: UpdateSavedQuery :one
UPDATE saved_queries
SET name = ?, date_from = ?, date_to = ?, currency_id = ?, fuzzy_text = ?, updated_at = ?, updated_by = ?
WHERE id = ?
RETURNING *;

-- name: DeleteSavedQuery :exec
DELETE FROM saved_queries WHERE id = ?;
