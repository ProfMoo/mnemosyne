# Submit Listens Examples

Example request bodies for `POST https://api.listenbrainz.org/1/submit-listens`.
Each subdirectory contains 10 standalone JSON payloads:

- `single/`: exactly one completed listen, with `listened_at`.
- `import/`: historical batches of one or more listens, each with `listened_at`.
- `playing_now/`: exactly one current listen, without `listened_at`.

Examples range from minimal metadata to albums, durations, services, players,
Unicode names, multiple artists, and MusicBrainz identifiers. MusicBrainz UUIDs
are illustrative, not verified matches. Example service URLs are placeholders.
Timestamps are fixed Unix epoch seconds. No credentials are included; submitting
requests requires authentication.
