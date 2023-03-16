package pdb_test

import (
	`os`
	`testing`

	`github.com/ambientsound/rex/pkg/rekordbox/page`
	`github.com/ambientsound/rex/pkg/rekordbox/pdb`
	`github.com/ambientsound/rex/pkg/rekordbox/track`
	`github.com/stretchr/testify/assert`
)

func TestLibrary_MarshalBinary(t *testing.T) {
	const pageSize = 512

	lib := pdb.New(pageSize)

	track := track.Track{
		Header: track.Header{
			IndexShift:  0x100,
			Bitmask:     0xC0700,
			SampleRate:  44100,
			FileSize:    42399665,
			ArtworkId:   15,
			KeyId:       8,
			LabelId:     2,
			TrackNumber: 3,
			Tempo:       13300,
			GenreId:     1,
			AlbumId:     15,
			ArtistId:    2,
			Id:          16,
			Year:        2016,
			SampleDepth: 16,
			Duration:    362,
		},
		AnalyzeDate:     "2022-07-27",
		DateAdded:       "2022-07-27",
		Isrc:            "GBJX38209003",
		Filename:        "Dax J - Wir Leben Fur Die Nacht.flac",
		Title:           "Wir Leben Für Die Nacht",
		AutoloadHotcues: "ON",
		FilePath:        "/meteor/techno/Dax J - Wir Leben Fur Die Nacht.flac",
		AnalyzePath:     "/PIONEER/USBANLZ/P03A/0000339E/ANLZ0000.DAT",
	}

	pg := page.NewPage(pageSize)

	trackRow, err := track.MarshalBinary()
	assert.NoError(t, err)

	added, err := pg.AddRows([][]byte{trackRow})
	assert.Equal(t, 1, added)
	assert.NoError(t, err)

	lib.Pages[page.Type_Tracks] = []*page.Data{
		pg,
	}

	data, err := lib.MarshalBinary()
	assert.NoError(t, err)

	assert.Equal(t, []byte{}, data)

	os.WriteFile("/tmp/rex.pdb", data, 0644)
}
