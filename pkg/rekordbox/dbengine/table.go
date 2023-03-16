package dbengine

import (
	`fmt`
	`io`

	`github.com/ambientsound/rex/pkg/marshal`
	`github.com/ambientsound/rex/pkg/rekordbox/page`
)

type Table struct {
	Type  page.Type
	Index page.Index
	Pages []page.Data
}

func (db *DbEngine) GetTable(pageType page.Type) (*Table, error) {
	pageNum, err := db.tableIndex(pageType)
	if err != nil {
		return nil, err
	}

	err = db.seekToPage(pageNum)
	if err != nil {
		return nil, err
	}

	idx, err := db.readIndex()
	if err != nil {
		return nil, err
	}

	table := &Table{
		Type:  pageType,
		Pages: make([]page.Data, 0),
		Index: *idx,
	}

	nextPage := idx.IndexHeader.NextPage

	for {
		var data *page.Data
		err = db.seekToPage(nextPage)
		if err != nil {
			return nil, err
		}
		data, err = db.readData()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		table.Pages = append(table.Pages, *data)
		nextPage = data.NextPage
	}

	return table, nil
}

func (db *DbEngine) TableTypes() []page.Type {
	types := make([]page.Type, db.Globals.NumTables)
	var i uint32
	for i = 0; i < db.Globals.NumTables; i++ {
		types[i] = db.Globals.Pointers[i].Type
	}
	return types
}

func (db *DbEngine) tableIndex(pageType page.Type) (uint32, error) {
	for _, ptr := range db.Globals.Pointers {
		if ptr.Type == pageType {
			return ptr.FirstPage, nil
		}
	}
	return 0, fmt.Errorf("table '%s' not found", pageType)
}

func (db *DbEngine) readIndex() (*page.Index, error) {
	index := &page.Index{}
	err := marshal.UnpackFrom(db.backend, index)
	if err != nil {
		return nil, err
	}

	if index.PageFlags != 0x64 {
		return nil, fmt.Errorf("index page flags not 0x64")
	}

	return index, nil
}

func (db *DbEngine) readData() (*page.Data, error) {
	data := make([]byte, 4096)
	_, err := io.ReadFull(db.backend, data)
	if err != nil {
		return nil, err
	}

	p := &page.Data{}
	err = p.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}

	switch p.PageFlags {
	case 0x10, 0x24, 0x34, 0x37:
		return p, nil
	case 0x0:
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("invalid page flags 0x%02x", p.PageFlags)
	}
}
