-- name: GetTag :one
SELECT * FROM tags WHERE id = ?;

-- name: GetTagByName :one
SELECT * FROM tags WHERE name = ?;

-- name: ListTags :many
SELECT * FROM tags ORDER BY name;

-- name: ListActiveTags :many
SELECT * FROM active_tags ORDER BY name;

-- name: ListPinnedTags :many
SELECT t.* FROM tags t
JOIN pinned_tags pt ON t.id = pt.tag_id
WHERE t.deleted_at IS NULL
ORDER BY pt.position;

-- name: CreateTag :one
INSERT INTO tags (name, description, notes, created_at, created_by, updated_at, updated_by)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateTag :one
UPDATE tags
SET name = ?, description = ?, notes = ?, updated_at = ?, updated_by = ?
WHERE id = ?
RETURNING *;

-- name: SoftDeleteTag :exec
UPDATE tags
SET deleted_at = ?, deleted_by = ?
WHERE id = ?;

-- name: RestoreTag :one
UPDATE tags
SET deleted_at = NULL, deleted_by = NULL
WHERE id = ?
RETURNING *;

-- name: HardDeleteTag :exec
DELETE FROM tags WHERE id = ?;

-- name: PinTag :exec
INSERT INTO pinned_tags (tag_id, position, created_at, created_by, updated_at, updated_by)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UnpinTag :exec
DELETE FROM pinned_tags WHERE tag_id = ?;

-- name: GetMaxPinnedPosition :one
SELECT COALESCE(MAX(position), 0) FROM pinned_tags;
