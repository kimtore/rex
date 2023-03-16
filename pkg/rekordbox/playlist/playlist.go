package playlist

import (
	`bytes`
	`encoding/binary`

	`github.com/ambientsound/rex/pkg/rekordbox/dstring`
	`github.com/lunixbochs/struc`
)

type PlaylistHeader struct {
	ParentId    uint32
	Unknown1    uint32
	SortOrder   uint32
	Id          uint32
	RawIsFolder uint32
}

/**
 * A row that holds a playlist name, ID, indication of whether it
 * is an ordinary playlist or a folder of other playlists, a link
 * to its parent folder, and its sort order.
 */
type Playlist struct {
	PlaylistHeader
	Name string
}

func (playlist *Playlist) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	err := struc.PackWithOptions(buf, &playlist.PlaylistHeader, &struc.Options{
		Order: binary.LittleEndian,
	})
	if err != nil {
		return nil, err
	}

	s := dstring.New(playlist.Name)
	data, err := s.MarshalBinary()
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (playlist *Playlist) SetIndexShift(shift uint16) {
}
