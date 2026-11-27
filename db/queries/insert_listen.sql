-- name: InsertListen :one
INSERT INTO listens (
    id,
    duration,
    artist_credits,
    albums,
    mbid
) VALUES (
    sqlc.arg(id),
    sqlc.arg(duration),
    sqlc.arg(artist_credits),
    sqlc.arg(albums),
    sqlc.arg(mbid)
)
RETURNING
    id,
    created_at,
    updated_at,
    duration,
    artist_credits,
    albums,
    mbid;
