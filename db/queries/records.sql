-- name: GetRecord :one
SELECT * FROM records WHERE id = ?;

-- name: ListRecords :many
SELECT * FROM records ORDER BY date DESC, id DESC;

-- name: ListActiveRecords :many
SELECT * FROM active_records ORDER BY date DESC, id DESC;

-- name: ListActiveRecordsByDateRange :many
SELECT * FROM active_records
WHERE date BETWEEN ? AND ?
ORDER BY date DESC, id DESC;

-- name: CreateRecord :one
INSERT INTO records (date, amount_cents, currency_id, notes, created_at, created_by, updated_at, updated_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateRecord :one
UPDATE records
SET date = ?, amount_cents = ?, currency_id = ?, notes = ?, updated_at = ?, updated_by = ?
WHERE id = ?
RETURNING *;

-- name: SoftDeleteRecord :exec
UPDATE records
SET deleted_at = ?, deleted_by = ?
WHERE id = ?;

-- name: RestoreRecord :one
UPDATE records
SET deleted_at = NULL, deleted_by = NULL
WHERE id = ?
RETURNING *;

-- name: HardDeleteRecord :exec
DELETE FROM records WHERE id = ?;
