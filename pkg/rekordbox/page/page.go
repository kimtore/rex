package page

import (
	`bytes`
	`encoding`
	`io`

	`github.com/ambientsound/rex/pkg/marshal`
	`github.com/ambientsound/rex/pkg/rekordbox/heap`
)

const MaxRowsInGroup = 16
const TypicalPageSize = 4096
const HeaderSize = 32
const DataHeaderSize = 8 + HeaderSize
const IndexHeaderSize = 28 + HeaderSize

//go:generate stringer -type=Type
type Type uint32

const (
	// Candidates for unknowns:
	// categories, sort, hotcue bank lists
	// the COLUMNS table could be "field which is shown next to track name on CDJ"
	Type_Tracks           Type = 0
	Type_Genres           Type = 1
	Type_Artists          Type = 2
	Type_Albums           Type = 3
	Type_Labels           Type = 4
	Type_Keys             Type = 5
	Type_Colors           Type = 6
	Type_PlaylistTree     Type = 7
	Type_PlaylistEntries  Type = 8
	Type_Unknown9         Type = 9
	Type_Unknown10        Type = 10
	Type_HistoryPlaylists Type = 11
	Type_HistoryEntries   Type = 12
	Type_Artwork          Type = 13
	Type_Unknown14        Type = 14
	Type_Unknown15        Type = 15
	Type_Columns          Type = 16
	Type_Unknown17        Type = 17
	Type_Unknown18        Type = 18
	Type_History          Type = 19
)

/**
 * A table page, consisting of a short header describing the
 * content of the page and linking to the next page, followed by a
 * heap in which row data is found. At the end of the page there is
 * an index which locates all rows present in the heap via their
 * offsets past the end of the page header.
 */
type Data struct {
	Header
	DataHeader
	RowSets []*RowSet
	heap    *heap.Heap
}

type DataHeader struct {
	// Small values, usually 1.
	// When NumRowsLarge is 0x1fff then this field is, too.
	// Equal to the number of rows in the COLORS and COLUMNS and UNKNOWN17-18 tables.
	Unknown5 uint16
	// The value 0x1fff is observed, and when it is, there are certainly deleted rows in the page.
	// Otherwise, a small number. I believe this might be a bitmask, 10 bits?
	// First deletable row?
	NumRowsLarge uint16
	Unknown6     uint16 // Always zero?
	Unknown7     uint16 // Always zero?
}

type RowRefs struct {
	Bitmask uint16
	Padding uint16
}

// Common header for index and data pages.
// 32 bytes (20h)
type Header struct {
	// 16 bytes
	Magic     uint32
	PageIndex uint32
	Type      Type
	NextPage  uint32

	// 16 bytes
	// Transaction increments every time a transaction is made, it follows the value in the global header.
	Transaction  uint32 // Updated when the page is written. For index pages, this value seems to be 1 until the index is changed.
	Unknown2     uint32 // always 00 00 00 00, but not indices?
	NumRowsSmall uint8  // 0x20, 0x09 (doesn't correspond to num playlist rows, which is 2)
	// U3 can maybe indicate the number of active rows in the table?
	// It increases by 0x20 for each active row.
	// This does not hold true for the COLORS table, where this value is always zero.
	// Also not for the COLUMNS table, where we have 27 entries and this value is 0x60.
	Unknown3 uint8 // changed from 0x60 to 0x40 when an entry was deleted.
	// Zero for tracks, albums, artists, genres, lables, artwork, history.
	// 0x01 for color pages.
	// 0x02 for Unknown17-18.
	// 0x03 for columns.
	// Higher values for other tables.
	Unknown4            uint8
	PageFlags           uint8  // 0x64 for index tables. 0x24 and 0x34 for data page.
	FreeSize            uint16 // total size 4050 (plus 40 bytes header, 6 missing?). Always zero on index tables.
	NextHeapWriteOffset uint16 // total size 4050. Always zero on index tables.
}

type Row interface {
	encoding.BinaryMarshaler
	// encoding.BinaryUnmarshaler
	SetIndexShift(shift uint16)
}

func NewPage(pageType Type) *Data {
	return &Data{
		Header: Header{
			Type: pageType,
		},
		heap: heap.New(TypicalPageSize - DataHeaderSize),
	}
}

func (page *Data) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	// Unknown unknowns
	// page.NumRowsLarge = 0x1fff
	// page.DataHeader.Unknown5 = 0x1fff // 0x1 // seems to be the row count for this table?
	page.DataHeader.Unknown5 = 1
	page.Header.Unknown4 = 0
	// page.DataHeader.Unknown5 = uint16(page.Header.NumRowsSmall)
	// page.Header.NumRowsSmall = uint8(page.DataHeader.NumRowsLarge) // I wonder what happens when this overflows.

	// In files exported by rekordbox, free + used = 4050.
	// This does not make sense, as the data header size is 40 bytes, and indeed this is where the heap starts.
	// Does that mean that six bytes are padded on top?
	// page.Header.FreeSize = uint16(page.heap.Free())
	// page.Header.NextHeapWriteOffset = uint16(page.heap.Size()) - page.Header.FreeSize
	// page.Header.FreeSize -= 6

	// For normal pages, this value is 0x24.
	// The value 0x34 is also observed mostly on Track tables (seen in PlaylistTree, History, Keys).
	// In a big file, around a third of the entries have the 0x10 bit set.
	//
	// Pages with 0x34 seems to have some connection with indexes.
	// The number of non-0x1ffffff8 index entries seem to correspond with the number of 0x34 tables.
	// Keys=0x19b, PlaylistTree=0x80, History=0x140
	page.Header.PageFlags = 0x34

	err := marshal.PackInto(buf, &page.Header)
	if err != nil {
		return nil, err
	}

	err = marshal.PackInto(buf, &page.DataHeader)
	if err != nil {
		return nil, err
	}

	err = marshal.Into(buf, page.heap)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (page *Data) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	page.RowSets = make([]*RowSet, 0)

	// Unpack header fields
	err := marshal.UnpackFrom(r, &page.Header)
	if err != nil {
		return err
	}

	err = marshal.UnpackFrom(r, &page.DataHeader)
	if err != nil {
		return err
	}

	raw, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// Read row tables from the end of the buffer
	remain := int(page.NumRowsSmall)
	sz := len(raw)
	for remain > 0 {
		sz -= rowsetLength
		remain -= rowsInRowSet
		rowset := &RowSet{}
		err = rowset.UnmarshalBinary(raw[sz : sz+rowsetLength])
		if err != nil {
			return err
		}
		page.RowSets = append(page.RowSets, rowset)
	}

	page.heap, err = heap.Load(raw, len(page.RowSets)*rowsetLength)

	return err
}

// Read one row directly from the heap and unmarshal it into a struct.
// Heap position is relative to the start of the heap.
// These values always come from HeapPositions().
func (page *Data) UnmarshalRow(row encoding.BinaryUnmarshaler, heapPosition uint16) error {
	if int(heapPosition) >= page.heap.TopSize() {
		return io.ErrShortBuffer
	}
	return row.UnmarshalBinary(page.heap.Bytes()[heapPosition:])
}

func (page *Data) ActiveRows() (numRows int) {
	for _, rowref := range page.HeapPositions() {
		if rowref.Exists {
			numRows++
		}
	}
	return
}

func (page *Data) HeapPositions() []RowReference {
	rowsToParse := page.NumRowsSmall
	refs := make([]RowReference, 0)

	for _, rs := range page.RowSets {
		for i, heapPos := range rs.Positions {
			if rowsToParse <= 0 {
				break
			}
			rowsToParse--
			refs = append(refs, RowReference{
				Exists:       rs.RowExists(i),
				HeapPosition: heapPos,
			})
		}
	}

	return refs
}

func (page *Data) Insert(row Row) error {
	const align = 4

	row.SetIndexShift(uint16(page.Header.NumRowsSmall) * 0x20)

	data, err := row.MarshalBinary()
	if err != nil {
		return err
	}

	heapPosition := uint16(page.heap.CursorTop())

	err = page.heap.WriteTop(data)
	if err != nil {
		return err
	}

	err = page.heap.AlignTop(align)
	if err != nil {
		return err
	}

	page.Header.NextHeapWriteOffset = uint16(page.heap.CursorTop())
	page.Header.FreeSize = uint16(page.heap.Free())

	index := page.Header.NumRowsSmall % 16

	if index == 0 {
		page.RowSets = append(page.RowSets, &RowSet{
			Positions:       make([]uint16, 16),
			ActiveRows:      0,
			LastWrittenRows: 0,
		})

		err = page.writeRowsets()
		if err != nil {
			return err
		}
	}

	rowsetNum := len(page.RowSets) - 1

	page.RowSets[rowsetNum].ActiveRows |= 1 << index
	page.RowSets[rowsetNum].LastWrittenRows = 1 << index
	page.RowSets[rowsetNum].Positions[index] = heapPosition

	err = page.writeRowsets()
	if err != nil {
		return err
	}

	page.Header.NumRowsSmall++
	page.Header.Unknown3 += 0x20

	return nil
}

func (page *Data) writeRowsets() error {
	page.heap.ResetBottom()
	for rs := range page.RowSets {
		err := marshal.Into(page.heap.BottomWriter(), page.RowSets[rs])
		if err != nil {
			return err
		}
	}
	return nil
}
