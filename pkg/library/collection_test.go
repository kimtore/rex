package library_test

import (
	`testing`

	`github.com/ambientsound/rex/pkg/library`
	`github.com/stretchr/testify/assert`
)

func TestCollection_All(t *testing.T) {
	c := library.NewCollection[*library.Artist]()
	artist := &library.Artist{
		Name: "foobar",
	}
	id := c.Insert(artist)

	assert.Len(t, c.All(), 1)
	assert.Equal(t, library.ID(1), id)
	assert.Equal(t, "foobar", c.GetByName("foobar").Name)
	assert.Equal(t, "foobar", c.GetByID(id).Name)
	assert.Equal(t, id, c.ID(artist))
}
