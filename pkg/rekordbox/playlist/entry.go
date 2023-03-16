package playlist

import (
	`github.com/ambientsound/rex/pkg/marshal`
)

type Entry struct {
	EntryIndex uint32
	TrackID    uint32
	PlaylistID uint32
}

func (entry *Entry) MarshalBinary() ([]byte, error) {
	return marshal.Pack(entry)
}

func (entry *Entry) SetIndexShift(shift uint16) {
}
