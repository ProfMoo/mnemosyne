-- name: InsertListen :one
INSERT INTO listens (
    id,
    created_at,
    updated_at,
    duration,
    title,
    artist_credits,
    albums,
    mbid
) VALUES (
    sqlc.arg(id),
    sqlc.arg(created_at),
    sqlc.arg(updated_at),
    sqlc.arg(duration),
    sqlc.arg(title),
    sqlc.arg(artist_credits),
    sqlc.arg(albums),
    sqlc.arg(mbid)
)
RETURNING
    id,
    created_at,
    updated_at,
    duration,
    title,
    artist_credits,
    albums,
    mbid;
