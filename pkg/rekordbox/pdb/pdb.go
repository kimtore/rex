package pdb

import (
	`github.com/ambientsound/rex/pkg/marshal`
	`github.com/ambientsound/rex/pkg/rekordbox/page`
)

// The first page of a PIONEER DJ DeviceSQL database file.
type FileHeader struct {
	Magic          uint32 // always zero.
	LenPage        uint32 // typical page length 4096
	NumTables      uint32 `struc:"sizeof=Pointers"` // Unique tables present in file, usually around 20
	NextUnusedPage uint32 // Block index of next unused page.
	Unknown1       uint32 // observed to be 0x5, 0x4, or 0x1, NOT always zero.
	Sequence       uint32 // (next) commit number
	Gap            uint32 // always zero.
	Pointers       []TablePointer
}

type TablePointer struct {
	Type           page.Type
	EmptyCandidate uint32
	FirstPage      uint32
	LastPage       uint32
}

func (t *FileHeader) MarshalBinary() ([]byte, error) {
	return marshal.Pack(t)
}

// Perhaps relevant, perhaps not.
// Useful for trying to end up with a file that looks exactly like Pioneer's.
var TableOrder = []page.Type{
	page.Type_Tracks,
	page.Type_Genres,
	page.Type_Artists,
	page.Type_Albums,
	page.Type_Labels,
	page.Type_Keys,
	page.Type_Colors,
	page.Type_PlaylistTree,
	page.Type_PlaylistEntries,
	page.Type_Unknown9,
	page.Type_Unknown10,
	page.Type_HistoryPlaylists,
	page.Type_HistoryEntries,
	page.Type_Artwork,
	page.Type_Unknown14,
	page.Type_Unknown15,
	page.Type_Columns,
	page.Type_Unknown17,
	page.Type_Unknown18,
	page.Type_History,
}
