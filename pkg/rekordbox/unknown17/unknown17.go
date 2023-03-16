package unknown17

import (
	`github.com/ambientsound/rex/pkg/marshal`
)

// A bunch of rows that are found in the pristine rekordbox export.
// The rows in this table seems to be static.
var InitialDataset = []*Unknown17{
	{Unknown1: 0x1, Unknown2: 0x1, Unknown3: 0x163, Unknown4: 0x0},
	{Unknown1: 0x5, Unknown2: 0x6, Unknown3: 0x105, Unknown4: 0x0},
	{Unknown1: 0x6, Unknown2: 0x7, Unknown3: 0x163, Unknown4: 0x0},
	{Unknown1: 0x7, Unknown2: 0x8, Unknown3: 0x163, Unknown4: 0x0},
	{Unknown1: 0x8, Unknown2: 0x9, Unknown3: 0x163, Unknown4: 0x0},
	{Unknown1: 0x9, Unknown2: 0xa, Unknown3: 0x163, Unknown4: 0x0},
	{Unknown1: 0xa, Unknown2: 0xb, Unknown3: 0x163, Unknown4: 0x0},
	{Unknown1: 0xd, Unknown2: 0xf, Unknown3: 0x163, Unknown4: 0x0},
	{Unknown1: 0xe, Unknown2: 0x13, Unknown3: 0x104, Unknown4: 0x0},
	{Unknown1: 0xf, Unknown2: 0x14, Unknown3: 0x106, Unknown4: 0x0},
	{Unknown1: 0x10, Unknown2: 0x15, Unknown3: 0x163, Unknown4: 0x0},
	{Unknown1: 0x12, Unknown2: 0x17, Unknown3: 0x163, Unknown4: 0x0},
	{Unknown1: 0x2, Unknown2: 0x2, Unknown3: 0x2, Unknown4: 0x1},
	{Unknown1: 0x3, Unknown2: 0x3, Unknown3: 0x3, Unknown4: 0x2},
	{Unknown1: 0x4, Unknown2: 0x4, Unknown3: 0x1, Unknown4: 0x3},
	{Unknown1: 0xb, Unknown2: 0xc, Unknown3: 0x63, Unknown4: 0x4},
	{Unknown1: 0x11, Unknown2: 0x5, Unknown3: 0x63, Unknown4: 0x5},
	{Unknown1: 0x13, Unknown2: 0x16, Unknown3: 0x63, Unknown4: 0x6},
	{Unknown1: 0x14, Unknown2: 0x12, Unknown3: 0x63, Unknown4: 0x7},
	{Unknown1: 0x1b, Unknown2: 0x1a, Unknown3: 0x263, Unknown4: 0x8},
	{Unknown1: 0x18, Unknown2: 0x11, Unknown3: 0x63, Unknown4: 0x9},
	{Unknown1: 0x16, Unknown2: 0x1b, Unknown3: 0x63, Unknown4: 0xa},
}

type Unknown17 struct {
	Unknown1 uint16
	Unknown2 uint16
	Unknown3 uint16
	Unknown4 uint16
}

func (uk *Unknown17) MarshalBinary() ([]byte, error) {
	return marshal.Pack(uk)
}

func (uk *Unknown17) UnmarshalBinary(data []byte) error {
	return marshal.Unpack(uk, data)
}

func (uk *Unknown17) SetIndexShift(shift uint16) {
}
