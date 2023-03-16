package artist

import (
	`bytes`

	`github.com/ambientsound/rex/pkg/marshal`
	`github.com/ambientsound/rex/pkg/rekordbox/dstring`
)

/**
 * A row that holds an artist name and ID.
 */
type Artist struct {
	Subtype     uint16
	IndexShift  uint16
	Id          uint32
	Unnamed3    uint8
	OfsNameNear uint8
	Name        string
}

func (artist *Artist) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}
	name := artist.Name
	artist.Name = ""
	artist.Subtype = 0x60
	artist.OfsNameNear = 0xA
	artist.Unnamed3 = 0x03 // always observed to be 0x03
	err := marshal.PackInto(buf, &artist)
	if err != nil {
		return nil, err
	}
	nameEncoder := dstring.New(name)
	err = marshal.Into(buf, nameEncoder)
	return buf.Bytes(), err
}

func (artist *Artist) SetIndexShift(shift uint16) {
	artist.IndexShift = shift
}
