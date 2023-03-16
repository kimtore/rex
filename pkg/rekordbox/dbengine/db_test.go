package dbengine_test

import (
	`os`
	`testing`

	`github.com/ambientsound/rex/pkg/rekordbox/dbengine`
	`github.com/ambientsound/rex/pkg/rekordbox/page`
	`github.com/stretchr/testify/assert`
)

func TestDbEngine_WriteHeader(t *testing.T) {
	f, err := os.Create("/tmp/dbengine.pdb")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	db := dbengine.New(f)
	assert.NoError(t, db.WriteHeader())
	assert.NoError(t, db.CreateTable(page.Type_Tracks))
	assert.NoError(t, db.CreateTable(page.Type_Genres))

	p := page.NewPage(page.Type_Tracks, page.TypicalPageSize)
	assert.NoError(t, db.InsertPage(p))
	assert.NoError(t, db.InsertPage(p))
	assert.NoError(t, db.InsertPage(p))

	assert.NoError(t, db.CreateTable(page.Type_Artists))
}
