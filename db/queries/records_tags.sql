-- name: GetRecordTags :many
SELECT t.* FROM tags t
INNER JOIN records_tags rt ON t.id = rt.tag_id
WHERE rt.record_id = ?
ORDER BY t.name;

-- name: GetTagRecords :many
SELECT r.* FROM records r
INNER JOIN records_tags rt ON r.id = rt.record_id
WHERE rt.tag_id = ?
ORDER BY r.date DESC, r.id DESC;

-- name: AddRecordTag :exec
INSERT INTO records_tags (record_id, tag_id, created_at, created_by, updated_at, updated_by)
VALUES (?, ?, ?, ?, ?, ?);

-- name: RemoveRecordTag :exec
DELETE FROM records_tags WHERE record_id = ? AND tag_id = ?;

-- name: RemoveAllRecordTags :exec
DELETE FROM records_tags WHERE record_id = ?;

-- name: RemoveAllTagRecords :exec
DELETE FROM records_tags WHERE tag_id = ?;
