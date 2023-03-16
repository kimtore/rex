package album

import (
	`bytes`

	`github.com/ambientsound/rex/pkg/marshal`
	`github.com/ambientsound/rex/pkg/rekordbox/dstring`
)

/**
 * A row that holds an album name and ID, together with the artist ID.
 */
type Album struct {
	Unnamed1   uint16
	IndexShift uint16
	Unnamed2   uint32
	ArtistId   uint32
	Id         uint32
	Unnamed3   uint32
	Unnamed4   uint8
	OfsName    uint8
	Name       string
}

func (album *Album) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}
	name := album.Name // FIXME: if this is a long UTF-16 string, then OfsName is observed to increment by two, and two bytes of zero filled between OfsName and the string.
	album.Name = ""
	album.Unnamed1 = 0x80 // always this value?
	album.Unnamed2 = 0    // always 0
	album.OfsName = 22    // static position
	album.Unnamed3 = 0    // always 0
	album.Unnamed4 = 0x03 // always 0x03
	err := marshal.PackInto(buf, &album)
	if err != nil {
		return nil, err
	}
	nameEncoder := dstring.New(name)
	err = marshal.Into(buf, nameEncoder)
	return buf.Bytes(), err
}

func (album *Album) SetIndexShift(shift uint16) {
	album.IndexShift = shift
}
