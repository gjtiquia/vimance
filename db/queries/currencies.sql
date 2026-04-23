-- name: GetCurrency :one
SELECT * FROM currencies WHERE id = ?;

-- name: GetCurrencyByCode :one
SELECT * FROM currencies WHERE code = ?;

-- name: ListCurrencies :many
SELECT * FROM currencies ORDER BY code;

-- name: CreateCurrency :one
INSERT INTO currencies (code, created_at, updated_at)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateCurrency :one
UPDATE currencies
SET code = ?, updated_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteCurrency :exec
DELETE FROM currencies WHERE id = ?;
