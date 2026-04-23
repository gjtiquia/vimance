-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;

-- name: ListUsers :many
SELECT * FROM users ORDER BY id;

-- name: ListActiveUsers :many
SELECT * FROM active_users ORDER BY id;

-- name: CreateUser :one
INSERT INTO users (username, created_at, updated_at)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET username = ?, updated_at = ?
WHERE id = ?
RETURNING *;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = ?, deleted_by = ?
WHERE id = ?;

-- name: RestoreUser :one
UPDATE users
SET deleted_at = NULL, deleted_by = NULL
WHERE id = ?
RETURNING *;

-- name: HardDeleteUser :exec
DELETE FROM users WHERE id = ?;
