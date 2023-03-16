package unknown18

import (
	`github.com/ambientsound/rex/pkg/marshal`
)

// A bunch of rows that are found in the pristine rekordbox export.
// The rows in this table seems to be _mostly_ static.
var InitialDataset = []*Unknown18{
	{Unknown1: 0x1, Unknown2: 0x6, Unknown3: 0x1, Unknown4: 0x0},
	{Unknown1: 0x15, Unknown2: 0x7, Unknown3: 0x1, Unknown4: 0x0},
	{Unknown1: 0xe, Unknown2: 0x8, Unknown3: 0x1, Unknown4: 0x0},
	{Unknown1: 0x8, Unknown2: 0x9, Unknown3: 0x1, Unknown4: 0x0},
	{Unknown1: 0x9, Unknown2: 0xa, Unknown3: 0x1, Unknown4: 0x0},
	{Unknown1: 0xa, Unknown2: 0xb, Unknown3: 0x1, Unknown4: 0x0},
	{Unknown1: 0xf, Unknown2: 0xd, Unknown3: 0x1, Unknown4: 0x0},
	{Unknown1: 0xd, Unknown2: 0xf, Unknown3: 0x1, Unknown4: 0x0},
	{Unknown1: 0x17, Unknown2: 0x10, Unknown3: 0x1, Unknown4: 0x0},
	{Unknown1: 0x16, Unknown2: 0x11, Unknown3: 0x1, Unknown4: 0x0},
	{Unknown1: 0x19, Unknown2: 0x0, Unknown3: 0x100, Unknown4: 0x0},
	{Unknown1: 0x1a, Unknown2: 0x1, Unknown3: 0x200, Unknown4: 0x0},
	{Unknown1: 0x2, Unknown2: 0x2, Unknown3: 0x302, Unknown4: 0x0},
	{Unknown1: 0x3, Unknown2: 0x3, Unknown3: 0x400, Unknown4: 0x0},
	{Unknown1: 0x5, Unknown2: 0x4, Unknown3: 0x500, Unknown4: 0x0},
	{Unknown1: 0x6, Unknown2: 0x5, Unknown3: 0x600, Unknown4: 0x0},
	{Unknown1: 0xb, Unknown2: 0xc, Unknown3: 0x700, Unknown4: 0x0},
}

type Unknown18 struct {
	Unknown1 uint16
	Unknown2 uint16
	Unknown3 uint16
	Unknown4 uint16
}

func (uk *Unknown18) MarshalBinary() ([]byte, error) {
	return marshal.Pack(uk)
}

func (uk *Unknown18) UnmarshalBinary(data []byte) error {
	return marshal.Unpack(uk, data)
}

func (uk *Unknown18) SetIndexShift(shift uint16) {
}
