package dbengine

import (
	`bytes`
	`encoding`
	`io`

	`github.com/ambientsound/rex/pkg/marshal`
	`github.com/ambientsound/rex/pkg/rekordbox/page`
	`github.com/ambientsound/rex/pkg/rekordbox/pdb`
)

const emptyTable = 0x03ffffff

type DbEngine struct {
	Globals *pdb.FileHeader
	tables  []*page.Index
	indices map[page.Type]*page.Index
	backend io.ReadWriteSeeker
}

func New(dbFile io.ReadWriteSeeker) *DbEngine {
	return &DbEngine{
		Globals: &pdb.FileHeader{
			LenPage:        page.TypicalPageSize,
			NextUnusedPage: 1,
			Sequence:       2,
			Unknown1:       0x5,
		},
		tables:  make([]*page.Index, 0),
		indices: make(map[page.Type]*page.Index),
		backend: dbFile,
	}
}

func Open(dbFile io.ReadWriteSeeker) (*DbEngine, error) {
	db := New(dbFile)
	err := db.seekToPage(0)
	if err != nil {
		return nil, err
	}
	err = marshal.UnpackFrom(db.backend, db.Globals)
	return db, err
}

func (db *DbEngine) WriteHeader() error {
	_, err := db.backend.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	err = marshal.Into(buf, db.Globals)
	if err != nil {
		return err
	}
	padLength := page.TypicalPageSize - buf.Len()
	_, err = buf.Write(make([]byte, padLength))
	if err != nil {
		return err
	}
	_, err = io.Copy(db.backend, buf)
	return err
}

func (db *DbEngine) CreateTable(pageType page.Type) error {
	p := page.NewIndex(pageType)
	p.Header.PageIndex = db.Globals.NextUnusedPage

	// When the table is completely empty, then NextPage will be set to 0x03ffffff.
	// After the first page is written to the table, then NextPage is set to its usual value.
	p.Header.NextPage = p.Header.PageIndex + 1
	p.IndexHeader.NextPage = emptyTable
	p.Header.Transaction = 1 // Sequence no. always 1 for index tables

	err := db.writeBlock(p.Header.PageIndex, p)
	if err != nil {
		return err
	}

	db.indices[pageType] = p

	// Update database state
	ptr := pdb.TablePointer{
		Type:           pageType,
		EmptyCandidate: p.Header.PageIndex + 1,
		FirstPage:      p.Header.PageIndex,
		LastPage:       p.Header.PageIndex,
	}
	db.Globals.Pointers = append(db.Globals.Pointers, ptr)
	db.Globals.NumTables++
	db.Globals.NextUnusedPage += 2

	return db.WriteHeader()
}

func (db *DbEngine) InsertPage(p *page.Data) error {
	p.PageIndex = db.nextFreePage(p.Type)
	p.NextPage = db.Globals.NextUnusedPage
	p.Transaction = db.Globals.Sequence

	err := db.writeBlock(p.PageIndex, p)
	if err != nil {
		return err
	}

	err = db.updateIndex(p.Type)
	if err != nil {
		return err
	}

	// Update database state
	db.Globals.NextUnusedPage++
	db.Globals.Sequence++
	db.setTableLimits(p.Type, p.PageIndex, p.NextPage)

	return db.WriteHeader()
}

func (db *DbEngine) nextFreePage(pageType page.Type) uint32 {
	for _, ptr := range db.Globals.Pointers {
		if ptr.Type != pageType {
			continue
		}
		return ptr.EmptyCandidate
	}
	panic("table does not exist")
}

func (db *DbEngine) updateIndex(pageType page.Type) error {
	index := db.indices[pageType]
	if index == nil {
		panic("table does not exist")
	}
	index.IndexHeader.NextPage = index.Header.NextPage
	return db.writeBlock(index.Header.PageIndex, index)
}

func (db *DbEngine) setTableLimits(pageType page.Type, lastPage uint32, emptyCandidate uint32) {
	for i, ptr := range db.Globals.Pointers {
		if ptr.Type != pageType {
			continue
		}
		db.Globals.Pointers[i] = pdb.TablePointer{
			Type:           ptr.Type,
			EmptyCandidate: emptyCandidate,
			FirstPage:      ptr.FirstPage,
			LastPage:       lastPage,
		}
		return
	}
	panic("table does not exist")
}

func (db *DbEngine) seekToPage(pageIndex uint32) (err error) {
	_, err = db.backend.Seek(int64(pageIndex)*int64(page.TypicalPageSize), io.SeekStart)
	return
}

func (db *DbEngine) writeBlock(pageIndex uint32, block encoding.BinaryMarshaler) (err error) {
	err = db.seekToPage(pageIndex)
	if err == nil {
		err = marshal.Into(db.backend, block)
	}
	return
}

func (db *DbEngine) parseIndexPage(header page.Header) (p *page.Index, err error) {
	idx := &page.IndexHeader{}
	err = marshal.UnpackFrom(db.backend, idx)
	if err != nil {
		return
	}
	return &page.Index{
		Header:      header,
		IndexHeader: *idx,
	}, nil
}

func (db *DbEngine) parseDataPage(header page.Header) (p *page.Data, err error) {
	switch header.Type {
	case page.Type_Tracks:
	case page.Type_Genres:
	case page.Type_Artists:
	case page.Type_Albums:
	case page.Type_Labels:
	case page.Type_Keys:
	case page.Type_Colors:
	case page.Type_PlaylistTree:
	case page.Type_PlaylistEntries:
	case page.Type_Unknown9:
	case page.Type_Unknown10:
	case page.Type_HistoryPlaylists:
	case page.Type_HistoryEntries:
	case page.Type_Artwork:
	case page.Type_Unknown14:
	case page.Type_Unknown15:
	case page.Type_Columns:
	case page.Type_Unknown17:
	case page.Type_Unknown18:
	case page.Type_History:
	}

	data := &page.DataHeader{}
	err = marshal.UnpackFrom(db.backend, data)
	if err != nil {
		return
	}
	return &page.Data{
		Header:     header,
		DataHeader: *data,
	}, nil
}
