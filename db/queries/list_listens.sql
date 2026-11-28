-- name: ListListens :many
SELECT
    id,
    created_at,
    updated_at,
    duration,
    title,
    artist_credits,
    albums,
    mbid
FROM listens
ORDER BY created_at DESC, id DESC
LIMIT
    sqlc.arg(page_limit)
    OFFSET sqlc.arg(page_offset);
