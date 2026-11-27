-- name: GetListen :one
SELECT
    id,
    created_at,
    updated_at,
    duration,
    artist_credits,
    albums,
    mbid
FROM listens
WHERE id = sqlc.arg(id);
