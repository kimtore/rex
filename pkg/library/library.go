package library

// Internal music library, used to collect tracks before exporting them to the PDB format.
// Not relevant to the Rekordbox format.

import (
	`time`
)

type TrackInput struct {
	Title  string
	Artist string
	Album  string
}

type Track struct {
	Path        string
	Title       string
	SampleRate  float64
	FileSize    int
	Bitrate     int
	TrackNumber int
	DiscNumber  int
	Tempo       float64
	ReleaseDate *time.Time
	AddedDate   time.Time
	SampleDepth int
	Duration    time.Duration
	Isrc        string

	// Foreign keys
	Artist *Artist
	Album  *Album
	// ArtworkId        uint32
	// KeyId            uint32
	// OriginalArtistId uint32
	// LabelId          uint32
	// RemixerId        uint32
	// ComposerId       uint32
	// GenreId          uint32
	// ColorId          uint8

	// Unused
	// PlayCount       uint16
	// Rating          uint8
	// Composer          string
	// Message         string
	// KuvoPublic      string
	// AutoloadHotcues string
	// MixName         string
	// AnalyzePath     string
	// Comment         string
}

func (t *Track) GetName() string {
	return t.Path
}

type Album struct {
	Artist *Artist
	Title  string
}

func (a *Album) GetName() string {
	return a.Title
}

type Artist struct {
	Name string
}

func (a *Artist) GetName() string {
	return a.Name
}

type Playlist struct {
	ID     ID
	Name   string
	Tracks []*Track
}

func (p *Playlist) GetName() string {
	return p.Name
}

type Library struct {
	tracks    *Collection[*Track]
	artists   *Collection[*Artist]
	albums    *Collection[*Album]
	playlists *Collection[*Playlist]
}

func New() *Library {
	return &Library{
		tracks:    NewCollection[*Track](),
		artists:   NewCollection[*Artist](),
		albums:    NewCollection[*Album](),
		playlists: NewCollection[*Playlist](),
	}
}

func (library *Library) Albums() *Collection[*Album] {
	return library.albums
}

func (library *Library) Artists() *Collection[*Artist] {
	return library.artists
}

func (library *Library) Tracks() *Collection[*Track] {
	return library.tracks
}

func (library *Library) Artist(name string) *Artist {
	artist := library.artists.GetByName(name)
	if artist != nil {
		return artist
	}
	artist = &Artist{
		Name: name,
	}
	library.artists.Insert(artist)
	return artist
}

func (library *Library) Album(title string) *Album {
	album := library.albums.GetByName(title)
	if album != nil {
		return album
	}
	album = &Album{
		Title: title,
	}
	library.albums.Insert(album)
	return album
}

func (library *Library) InsertTrack(track *Track) {
	library.tracks.Insert(track)
}
