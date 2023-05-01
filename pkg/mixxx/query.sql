-- name: ListTracks :many
SELECT library.*, tl.location AS path, tl.filesize AS filesize
FROM library
JOIN track_locations tl on library.location = tl.id
ORDER BY library.id;

-- name: ListPlaylists :many
SELECT * FROM Playlists
ORDER BY position;

-- name: ListPlaylistTracks :many
SELECT tracklist.*, loc.location AS path FROM PlaylistTracks tracklist
JOIN library ON library.id = tracklist.track_id
JOIN track_locations loc ON library.location = loc.id
WHERE tracklist.playlist_id = ?
ORDER BY tracklist.position;

-- name: ListCrates :many
SELECT * FROM crates;

-- name: ListCrateTracks :many
SELECT tracklist.*, loc.location AS path FROM crate_tracks tracklist
JOIN library ON library.id = tracklist.track_id
JOIN track_locations loc ON library.location = loc.id
WHERE tracklist.crate_id = ?;
