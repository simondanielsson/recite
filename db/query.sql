-- name: ListRecitals :many
SELECT * FROM recitals
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetRecital :one
SELECT * FROM recitals WHERE id = $1;

-- name: CreateRecital :one
INSERT INTO recitals (url, title, description, status, path, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateRecitalPath :exec
UPDATE recitals SET path = $2 WHERE id = $1
RETURNING *;

-- name: UpdateRecitalStatus :exec
UPDATE recitals SET status = $2 WHERE id = $1
RETURNING *;

-- name: DeleteRecital :exec
DELETE FROM recitals WHERE id = $1
RETURNING *;
