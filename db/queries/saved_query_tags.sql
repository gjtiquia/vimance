-- name: AddSavedQueryTag :exec
INSERT INTO saved_query_tags (query_id, tag_id) VALUES (?, ?);

-- name: RemoveSavedQueryTag :exec
DELETE FROM saved_query_tags WHERE query_id = ? AND tag_id = ?;

-- name: RemoveAllSavedQueryTags :exec
DELETE FROM saved_query_tags WHERE query_id = ?;

-- name: ListSavedQueryTags :many
SELECT t.* FROM tags t
INNER JOIN saved_query_tags sqt ON t.id = sqt.tag_id
WHERE sqt.query_id = ?
ORDER BY t.name;
