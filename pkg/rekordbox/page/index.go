package page

import (
	`bytes`
	`encoding/binary`
	`io`

	`github.com/ambientsound/rex/pkg/marshal`
	`github.com/ambientsound/rex/pkg/rekordbox/heap`
	`github.com/lunixbochs/struc`
)

// Index type pages.
type Index struct {
	Header
	IndexHeader
}

// The first page entry for any table is an index table.
// It is not clear how to use this functionality.
// Instead, we use commonly found values that are assumed to work.
// 28 bytes?
type IndexHeader struct {
	// unknown1-2 changes from (ff1f ff1f ec03 0000)
	//                    to   (0100 0000 ec03 0100)
	//                    to   (ff1f ff1f ec03 0200)
	// upon deleting, renaming entries in this table.
	Unknown1   uint16 // Usually 0x1fff, sometimes 0x0001 (keys, history), sometimes corresponding to NumEntries
	Unknown2   uint16 // Usually 0x1fff, sometimes 0x0000 (keys, history)
	Unknown3   uint16 // Always 0x03ec
	NextOffset uint16 // Byte offset of where to insert the next entry. Relative to the start of index entries, usually zero for empty pages.
	PageIndex  uint32 // |
	NextPage   uint32 // |- reflects the values from PageIndex and NextPage, except empty tables are 0x1ffffff8

	// Some kind of flags
	Unknown5        uint32 // Always 0x03ffffff
	Unknown6        uint32 // Always 0x00000000
	NumEntries      uint16 `struc:"sizeof=IndexEntries"` // Number of index entries. The actual number of pages with flag 0x34 might be (one) higher than this value.
	FirstEmptyEntry uint16 // Points to the first offset of index entries where the index value equals 0x1ffffff8. If none, then this value is 0x1fff. Garbage collector?

	IndexEntries []uint32

	// Next part is the indexes. This seems to be implemented as a heap,
	// Filling up the rest of the page.
	// Most of the heap is filled with 0x1fffffff8.
	// The heap ends with 18 bytes which are observed to be always zero (???)

	// Changing/deleting single entry in the table results in this change on the first line of indices:
	// -00001030: ffff ff03 0000 0000 0000 ff1f f8ff ff1f  ................
	// ------ CHANGE 1
	// +00001030: ffff ff03 0000 0000 0100 ff1f 8b01 0000  ................
	// ------ CHANGE 2
	// +00001030: ffff ff03 0000 0000 0200 ff1f 8801 0000  ................
	// +00001040: 1000 0000 f8ff ff1f f8ff ff1f f8ff ff1f  ................
	//
	// Renaming an entry in the table results in this change on the first line of indices.
	// -00001020: ffff ff03 0000 0000 0000 ff1f f8ff ff1f  ................
	// +00001020: ffff ff03 0000 0000 0100 ff1f 8b01 0000  ................

	// Deleting an entry can also not trigger the index.
	// Simply the bit is flipped in the row table on the page.
	//
	// Theories on data
	//
	// - bit mask, seems to be at most 11 bits long = 2048.
	//   for TRACKS and most other tables, there seem to be eight bits in use. The three least significant bits are always zero.
	//   for KEYS, the three LSB have been observed to be 011.
	//
	// - the largest auto-incrementing ID for a field?
	// - the values seem to be mostly increasing, but also decreasing
	// - always greater than the highest ID of any track on that page
	//
	// bin(0x10)       // '0b000010000'
	// bin(0x1b0)      // '0b110110000'
	// bin(0x1a8)      // '0b110101000'
	// bin(0x180)      // '0b110000000'
	// bin(0x1ffffff8) // '0b11111111111111111111111111000'
	//
	// bin(0x1ab)      // '0b110101011'
	// Observation
	//
	// The KEYS table can have both camelot and standard notation.
	// id and id2 are increasing monotonically.
	// on the keys table the index is
}

func (page *Index) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	page.Header.PageFlags = 0x64
	page.Header.FreeSize = 0
	page.Header.NextHeapWriteOffset = 0

	// I believe this combination of values means "no index entries for this table"
	page.IndexHeader.PageIndex = page.Header.PageIndex
	// page.IndexHeader.NextPage = page.Header.NextPage
	page.IndexHeader.Unknown1 = 0x1fff
	page.IndexHeader.Unknown2 = 0x1fff
	page.IndexHeader.Unknown3 = 0x03ec
	page.IndexHeader.Unknown5 = 0x03ffffff
	page.IndexHeader.FirstEmptyEntry = 0x1fff

	// Empty bitmask

	hp := heap.New(TypicalPageSize - IndexHeaderSize)
	err := marshal.PackInto(hp.BottomWriter(), [20]byte{})
	if err != nil {
		return nil, err
	}

	// Instead of zero-filling, we fill with a known constant.
	for err == nil {
		err = marshal.PackInto(hp.TopWriter(), uint32(0x1ffffff8))
	}
	if err != io.ErrShortWrite {
		return nil, err
	}

	err = struc.PackWithOptions(buf, page, &struc.Options{
		Order: binary.LittleEndian,
	})
	if err != nil {
		return nil, err
	}

	data, err := hp.MarshalBinary()
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(data)
	return buf.Bytes(), err
}

func NewIndex(pageType Type) *Index {
	return &Index{
		Header: Header{
			Type: pageType,
		},
	}
}
