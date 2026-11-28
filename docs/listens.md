# Listen API And Storage

`GET /listens` returns all stored listens, newest creation timestamp first, with
UUID descending as a deterministic tie-breaker. `GET /listens/{listen_id}`
returns one listen, or HTTP 404 when its UUID does not exist.

Responses contain `id`, `createdAt`, `updatedAt`, and `recording`. Recordings
contain `title`, `artistCredits`, `albums`, `mbid`, and `duration`. Duration is an
integer number of nanoseconds, matching Go's `time.Duration`. Creation/update
timestamps describe storage, not playback start/end times.

`internal/listens/api.go` explicitly maps between generated public types and
domain types. API `artistCredit.artist` maps to domain `ArtistCredit.Artists`.
Arrays are returned as `[]`, including when internal collections are nil.

The submission endpoint currently accepts single submissions with one listen.
It persists only uniquely matched recordings. Unsupported, invalid, unmatched,
or ambiguous submissions return HTTP 400 without saving a partial listen.

## Existing Database Upgrade

Fresh databases include the recording title column. The bootstrap is not a
migration system and does not add columns to existing tables.

If an existing `listens` table lacks `title`, stop the application, back up its
database, and apply this statement once using your SQLite client:

```sql
ALTER TABLE listens ADD COLUMN title TEXT NOT NULL DEFAULT '';
```

This preserves existing rows but cannot recover titles that were never stored.
Do not rerun the statement if the column already exists. This change does not
modify an existing database file automatically.
