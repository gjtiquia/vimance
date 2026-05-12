-- name: CreateTarget :one
INSERT INTO targets (name, saved_query_id, target_cents, created_at, created_by, updated_at, updated_by)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetTarget :one
SELECT * FROM targets WHERE id = ?;

-- name: ListTargets :many
SELECT * FROM targets ORDER BY created_at ASC;

-- name: UpdateTarget :one
UPDATE targets
SET name = ?, saved_query_id = ?, target_cents = ?, updated_at = ?, updated_by = ?
WHERE id = ?
RETURNING *;

-- name: DeleteTarget :exec
DELETE FROM targets WHERE id = ?;