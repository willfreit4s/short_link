-- name: CreateShortLink :one
INSERT INTO links (id, original_url, created_at)
VALUES ($1, $2, NOW())
RETURNING id, original_url, created_at;

-- name: GetShortLink :one
SELECT id, original_url, created_at
FROM links
WHERE id = $1;