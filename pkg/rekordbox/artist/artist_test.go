package artist_test

import (
	`testing`

	`github.com/ambientsound/rex/pkg/rekordbox/artist`
	`github.com/stretchr/testify/assert`
)

func TestArtist_MarshalBinary(t *testing.T) {
	expected := []byte{
		0x60, 0x00, 0x40, 0x00, 0x76, 0x00, 0x00, 0x00, 0x03, 0x0a, 0x47, 0x54,
		0x6f, 0x74, 0x61, 0x6c, 0x6c, 0x79, 0x20, 0x45, 0x6e, 0x6f, 0x72, 0x6d,
		0x6f, 0x75, 0x73, 0x20, 0x45, 0x78, 0x74, 0x69, 0x6e, 0x63, 0x74, 0x20,
		0x44, 0x69, 0x6e, 0x6f, 0x73, 0x61, 0x75, 0x72, 0x73,
	}

	art := artist.Artist{
		Name:       "Totally Enormous Extinct Dinosaurs",
		IndexShift: 0x40,
		Id:         118,
	}

	data, err := art.MarshalBinary()

	assert.NoError(t, err)
	assert.Equal(t, expected, data)
}
