# Import/Scrobbling

* Need to use the same logic for those two points
* Need to look at Spotify, Last.fm, and Listenbrainz format and see what's there (usually). Need to have my own internal representation of that and map cleanly

Should probably accept ListenBrainz API on the way in for compatability.

## Flow

Roughly, flow is like this:

1. API ingestion (will define ingestion endpoints in OpenAPI)
2. Persist raw listens
   1. Not sure if I want to expose them yet, without any MBID info
3. Enqueue MB matching
   1. Will re-write the ListenBrainz MBID-Mapper in go. Can rely on AI a lot here to rebuild that functionality natively in Go. Will write a ton of tests to verify behavior.
   2. We should segregate this functionality in its own package and gradually improve matching behavior over time.
   3. We should expose tons of info about the potential match. In the API, we can look up match correctness and verify/update info manually
      1. Each match attempt will be stored in the DB for reference
4. Store highest confidence match results on Listen record directly.
5. Update async analytics (ex: yearly rollups)

## Resources

1. Listen
2. Track
3. Album
4. Artist
5. Rollups (Daily, Weekly, Monthly, etc.). Would be composed of Tracks, Albums, Artists.
6. Analytics (More complex items, like "fav winter season tracks"). Would be composed of Tracks, Albums, Artists.

Do we want o keep actual music info like 'Artist' as core resources? Then we'll need background jobs that, when MBIDs are found, update those resources :thinking:.

Could be OK so long as we have different auth levels. GET, LIST, etc. would be open to any 'READ' user. The others (like POST artist) would be only for admin users.

## Endpoints

```txt
GET /listens/{listen_id} - how to input a listen. needs minimal amount of info.
GET /listens - list listens. we'll support some pretty complex filters here. need to start w/ time-based filter to support 'most recently played'. default will be starting w/ most recent.
POST /listens - this is the endpoint that will support listenbrainz API
PATCH /listens/{listen_id} - how a user would be able to modify the listen manually
