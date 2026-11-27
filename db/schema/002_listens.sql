CREATE TABLE IF NOT EXISTS listens (
    -- Primary key has b-tree index created automatically.
    id BLOB PRIMARY KEY NOT NULL CHECK (length(id) = 16),
    created_at DATETIME NOT NULL DEFAULT current_timestamp,
    updated_at DATETIME NOT NULL DEFAULT current_timestamp,
    duration INTEGER NOT NULL,
    artist_credits TEXT NOT NULL DEFAULT '[]' CHECK (
        json_valid(artist_credits)
        AND json_type(artist_credits) = 'array'
    ),
    albums TEXT NOT NULL DEFAULT '[]' CHECK (
        json_valid(albums)
        AND json_type(albums) = 'array'
    ),
    mbid BLOB NOT NULL CHECK (length(mbid) = 16)
) WITHOUT ROWID;
