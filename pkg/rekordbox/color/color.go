package color

import (
	`bytes`

	`github.com/ambientsound/rex/pkg/marshal`
	`github.com/ambientsound/rex/pkg/rekordbox/dstring`
)

// A bunch of rows that are found in the pristine rekordbox export.
// The rows in this table seems to be _mostly_ static.
var InitialDataset = []*Color{
	{Header: Header{Unknown1: 0x0, Unknown2: 0x1, ID: 0x1, Unknown3: 0x0}, Name: "Pink"},
	{Header: Header{Unknown1: 0x0, Unknown2: 0x2, ID: 0x2, Unknown3: 0x0}, Name: "Red"},
	{Header: Header{Unknown1: 0x0, Unknown2: 0x3, ID: 0x3, Unknown3: 0x0}, Name: "Orange"},
	{Header: Header{Unknown1: 0x0, Unknown2: 0x4, ID: 0x4, Unknown3: 0x0}, Name: "Yellow"},
	{Header: Header{Unknown1: 0x0, Unknown2: 0x5, ID: 0x5, Unknown3: 0x0}, Name: "Green"},
	{Header: Header{Unknown1: 0x0, Unknown2: 0x6, ID: 0x6, Unknown3: 0x0}, Name: "Aqua"},
	{Header: Header{Unknown1: 0x0, Unknown2: 0x7, ID: 0x7, Unknown3: 0x0}, Name: "Blue"},
	{Header: Header{Unknown1: 0x0, Unknown2: 0x8, ID: 0x8, Unknown3: 0x0}, Name: "Purple"},
}

type Header struct {
	Unknown1 uint32
	Unknown2 uint8
	ID       uint16
	Unknown3 uint8
}

type Color struct {
	Header
	Name string
}

func (c *Color) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := marshal.PackInto(buf, &c.Header)
	if err != nil {
		return nil, err
	}
	dstr := dstring.New(c.Name)
	err = marshal.Into(buf, dstr)
	return buf.Bytes(), err
}

func (c *Color) UnmarshalBinary(data []byte) error {
	err := marshal.Unpack(&c.Header, data)
	if err != nil {
		return err
	}
	c.Name, err = dstring.UnmarshalBinary(data[8:])
	return err
}

func (c *Color) SetIndexShift(shift uint16) {
}
