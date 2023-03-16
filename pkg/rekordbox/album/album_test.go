package album_test

import (
	`testing`

	`github.com/ambientsound/rex/pkg/rekordbox/album`
	`github.com/stretchr/testify/assert`
)

func TestArtist_MarshalBinary(t *testing.T) {
	expected := []byte{
		0x80, 0x00, 0x20, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x0a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x16, 0x15, 0x46,
		0x4a, 0x41, 0x41, 0x4b, 0x20, 0x30, 0x30, 0x36,
	}

	alb := album.Album{
		Name:       "FJAAK 006",
		IndexShift: 288,
		Id:         10,
		ArtistId:   0,
	}

	data, err := alb.MarshalBinary()

	assert.NoError(t, err)
	assert.Equal(t, expected, data)
}
