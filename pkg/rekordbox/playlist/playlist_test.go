package playlist_test

import (
	`testing`

	`github.com/ambientsound/rex/pkg/rekordbox/playlist`
	`github.com/stretchr/testify/assert`
)

func TestPlaylist_MarshalBinary(t *testing.T) {
	p := &playlist.Playlist{
		PlaylistHeader: playlist.PlaylistHeader{
			ParentId:    0,
			Unknown1:    0,
			SortOrder:   0,
			Id:          1,
			RawIsFolder: 0,
		},
		Name: "Rekordbåks",
	}

	data, err := p.MarshalBinary()

	assert.NoError(t, err)
	assert.Equal(t, []byte{}, data)
}
