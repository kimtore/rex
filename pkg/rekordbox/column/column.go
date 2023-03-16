package column

import (
	`bytes`

	`github.com/ambientsound/rex/pkg/marshal`
	`github.com/ambientsound/rex/pkg/rekordbox/dstring`
)

// A bunch of rows that are found in the pristine rekordbox export.
var InitialDataset = []*Column{
	{Header: Header{ID: 0x1, Unknown1: 0x80}, Name: "\ufffaGENRE\ufffb"},
	{Header: Header{ID: 0x2, Unknown1: 0x81}, Name: "\ufffaARTIST\ufffb"},
	{Header: Header{ID: 0x3, Unknown1: 0x82}, Name: "\ufffaALBUM\ufffb"},
	{Header: Header{ID: 0x4, Unknown1: 0x83}, Name: "\ufffaTRACK\ufffb"},
	{Header: Header{ID: 0x5, Unknown1: 0x85}, Name: "\ufffaBPM\ufffb"},
	{Header: Header{ID: 0x6, Unknown1: 0x86}, Name: "\ufffaRATING\ufffb"},
	{Header: Header{ID: 0x7, Unknown1: 0x87}, Name: "\ufffaYEAR\ufffb"},
	{Header: Header{ID: 0x8, Unknown1: 0x88}, Name: "\ufffaREMIXER\ufffb"},
	{Header: Header{ID: 0x9, Unknown1: 0x89}, Name: "\ufffaLABEL\ufffb"},
	{Header: Header{ID: 0xa, Unknown1: 0x8a}, Name: "\ufffaORIGINAL ARTIST\ufffb"},
	{Header: Header{ID: 0xb, Unknown1: 0x8b}, Name: "\ufffaKEY\ufffb"},
	{Header: Header{ID: 0xc, Unknown1: 0x8d}, Name: "\ufffaCUE\ufffb"},
	{Header: Header{ID: 0xd, Unknown1: 0x8e}, Name: "\ufffaCOLOR\ufffb"},
	{Header: Header{ID: 0xe, Unknown1: 0x92}, Name: "\ufffaTIME\ufffb"},
	{Header: Header{ID: 0xf, Unknown1: 0x93}, Name: "\ufffaBITRATE\ufffb"},
	{Header: Header{ID: 0x10, Unknown1: 0x94}, Name: "\ufffaFILE NAME\ufffb"},
	{Header: Header{ID: 0x11, Unknown1: 0x84}, Name: "\ufffaPLAYLIST\ufffb"},
	{Header: Header{ID: 0x12, Unknown1: 0x98}, Name: "\ufffaHOT CUE BANK\ufffb"},
	{Header: Header{ID: 0x13, Unknown1: 0x95}, Name: "\ufffaHISTORY\ufffb"},
	{Header: Header{ID: 0x14, Unknown1: 0x91}, Name: "\ufffaSEARCH\ufffb"},
	{Header: Header{ID: 0x15, Unknown1: 0x96}, Name: "\ufffaCOMMENTS\ufffb"},
	{Header: Header{ID: 0x16, Unknown1: 0x8c}, Name: "\ufffaDATE ADDED\ufffb"},
	{Header: Header{ID: 0x17, Unknown1: 0x97}, Name: "\ufffaDJ PLAY COUNT\ufffb"},
	{Header: Header{ID: 0x18, Unknown1: 0x90}, Name: "\ufffaFOLDER\ufffb"},
	{Header: Header{ID: 0x19, Unknown1: 0xa1}, Name: "\ufffaDEFAULT\ufffb"},
	{Header: Header{ID: 0x1a, Unknown1: 0xa2}, Name: "\ufffaALPHABET\ufffb"},
	{Header: Header{ID: 0x1b, Unknown1: 0xaa}, Name: "\ufffaMATCHING\ufffb"},
}

type Header struct {
	ID       uint16
	Unknown1 uint16
}

type Column struct {
	Header
	// All entries in this table have their names wrapped in 0xfffa and 0xfffb.
	Name string
}

func (c *Column) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := marshal.PackInto(buf, &c.Header)
	if err != nil {
		return nil, err
	}
	dstr := dstring.New(c.Name)
	err = marshal.Into(buf, dstr)
	return buf.Bytes(), err
}

func (c *Column) UnmarshalBinary(data []byte) error {
	err := marshal.Unpack(&c.Header, data)
	if err != nil {
		return err
	}
	c.Name, err = dstring.UnmarshalBinary(data[4:])
	return err
}

func (c *Column) SetIndexShift(shift uint16) {
}
