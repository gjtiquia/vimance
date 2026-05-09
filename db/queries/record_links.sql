-- name: AddRecordLink :exec
INSERT INTO record_links (parent_id, child_id, created_at, created_by)
VALUES (?, ?, ?, ?);

-- name: RemoveRecordLink :exec
DELETE FROM record_links WHERE parent_id = ? AND child_id = ?;

-- name: RemoveAllRecordLinks :exec
DELETE FROM record_links WHERE child_id = ?;

-- name: GetRecordParents :many
SELECT r.* FROM active_records r
INNER JOIN record_links rl ON r.id = rl.parent_id
WHERE rl.child_id = ?
ORDER BY r.date DESC, r.id DESC;

-- name: GetRecordChildren :many
SELECT r.* FROM active_records r
INNER JOIN record_links rl ON r.id = rl.child_id
WHERE rl.parent_id = ?
ORDER BY r.date DESC, r.id DESC;

-- name: SearchParentCandidates :many
SELECT DISTINCT r.id, r.date, r.amount_cents, r.currency_id, r.notes
FROM active_records r
WHERE r.date BETWEEN ? AND ?
  AND r.currency_id = ?
  AND r.id != ?
ORDER BY r.date DESC, r.id DESC;

-- name: GetRecordTagsByIDs :many
SELECT rt.record_id, t.name FROM tags t
INNER JOIN records_tags rt ON t.id = rt.tag_id
WHERE rt.record_id IN (sqlc.slice('record_ids'))
ORDER BY rt.record_id, t.name;

-- name: GetRecordTagIDsByIDs :many
SELECT record_id, tag_id FROM records_tags
WHERE record_id IN (sqlc.slice('record_ids'))
ORDER BY record_id, tag_id;
