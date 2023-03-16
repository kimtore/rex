package page

import (
	`sort`

	`github.com/ambientsound/rex/pkg/marshal`
)

const rowsetLength = 36
const rowsInRowSet = 16

type RowSet struct {
	// Heap positions for row data.
	Positions []uint16 `struc:"[16]uint16"`
	// Bitmask of rows that exists in the table. The bits are set to 0 for deleted rows.
	ActiveRows uint16
	// Bitmask of rows that were affected by the previous operation. When a new row is written, this value is cleared, and the bit representing the row position is set.
	// If a record in this table is deleted, that bit is set with a bitwise OR operation.
	LastWrittenRows uint16
}

type RowReference struct {
	Exists       bool
	HeapPosition uint16
}

func (r *RowSet) RowExists(index int) bool {
	bit := uint16(1 << index)
	return r.ActiveRows&bit > 0
}

func (r *RowSet) MarshalBinary() ([]byte, error) {
	rev := &RowSet{
		Positions:       make([]uint16, len(r.Positions)),
		ActiveRows:      r.ActiveRows,
		LastWrittenRows: r.LastWrittenRows,
	}
	copy(rev.Positions, r.Positions)
	sort.SliceStable(rev.Positions, func(i, j int) bool {
		return j < i
	})
	return marshal.Pack(rev)
}

func (r *RowSet) UnmarshalBinary(data []byte) error {
	err := marshal.Unpack(r, data)
	if err != nil {
		return err
	}
	sort.SliceStable(r.Positions, func(i, j int) bool {
		return j < i
	})
	return nil
}
