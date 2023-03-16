// After these header fields comes the page heap. Rows are allocated within this heap starting at byte 28.
// Since rows can be different sizes, there needs to be a way to locate them. This takes the form of a row index,
// which is built from the end of the page backwards, in groups of up to sixteen row pointers along with a bitmask
// saying which of those rows are still part of the table (they might have been deleted). The number of row index
// entries is determined, as described above, by the value of either num_rows_small or num_rows_large.
//
// The bit mask for the first group of up to sixteen rows, labeled rowpf0 in the diagram (meaning “row presence flags group 0”),
// is found near the end of the page. The last two bytes after each row bitmask (for example pad0 after rowpf0) have an unknown
// purpose and may always be zero, and the rowpf0 bitmask takes up the two bytes that precede them. The low order bit of this
// value will be set if row 0 is really present, the next bit if row 1 is really present, and so on. The two bytes before these flags,
// labeled ofs0, store the offset of the first row in the page. This offset is the number of bytes past the end of the page header
// at which the row itself can be found. So if row 0 begins at the very beginning of the heap, at byte 28 in the page, ofs0 would have the value 0000.
//
// As more rows are added to the page, space is allocated for them in the heap, and additional index entries are added
// at the end of the heap, growing backwards. Once there have been sixteen rows added, all of the bits in rowpf0 are accounted for,
// and when another row is added, before its offset entry ofs16 can be added, another row bit-mask entry rowpf1 needs to be allocated,
// followed by its corresponding pad1. And so the row index grows backwards towards the rows that are being added forwards,
// and once they are too close for a new row to fit, the page is full, and another page gets allocated to the table.

// TODO: alignment
// This is technically dependent on the compile time ABI
// but for x86 (which this is likely compiled on), we should be able
// to assume that
// the alignment of a type is just the size of its largest
// member. That likely matches the assumptions made for the 32-bit
// MCU (Renesas R8A77240D500BG) built into the different CDJ-2000 variants.
// In either way, its better to overshoot the alignment
// than to undershoot it. For CDJ-3000s, this assumption
// is likely also correct since they use a 64-bit ARM CPU (Renesas R8A774C0HA01BG)

// TODO(Swiftb0y): Write with proper alignment.
// Rows don't seem to be directly adjacent to each other
// but instead have gaps in between. They probably adhere to their
// member variable alignment.
// I have seen gaps of 52 to 55 bytes (ending after the last char
// of the previous row and the first byte of the next row).
// I have 0 idea why these gaps are this big or how to accurately
// guess their size.
// Rows also don't have a fixed size. Their sizes seem to fluctuate
// between 0 and 48 bytes in size (though the fluctuations always
// were multiple of 12)
//
// TODO(ambientsound): artist rows have between 6-9 chars of zero padding
// FIXME(ambientsound): best guess is pad with six null bytes, then up to two more up to next 4-byte boundary?
// 15->24
// 16->24
// 17->24
// 18->24
// 19->28
// 20->28
// 22->28
// 23->32
// 24->32
// 25->32
// 26->32
// 27->36
// 28->36
// 29->36
// 37->44
// 52->60
// 56->64
// 74->80

package heap

import (
	`bytes`
	`io`
)

type Heap struct {
	top    *bytes.Buffer
	bottom *bytes.Buffer
	size   int
}

func New(size int) *Heap {
	return &Heap{
		size:   size,
		top:    &bytes.Buffer{},
		bottom: &bytes.Buffer{},
	}
}

func Load(data []byte, bottomBytes int) (*Heap, error) {
	hp := New(len(data))
	err := hp.WriteTop(data[:len(data)-bottomBytes])
	if err == nil && bottomBytes > 0 {
		err = hp.WriteBottom(data[len(data)-bottomBytes:])
	}
	return hp, err
}

type heapWriter struct {
	write func([]byte) error
}

var _ io.Writer = &heapWriter{}

func (h heapWriter) Write(p []byte) (n int, err error) {
	err = h.write(p)
	if err == nil {
		n = len(p)
	}
	return
}

func (heap *Heap) ResetBottom() {
	heap.bottom.Reset()
}

func (heap *Heap) Reset() {
	heap.top.Reset()
	heap.bottom.Reset()
}

func (heap *Heap) Bytes() []byte {
	return append(heap.top.Bytes(), heap.bottom.Bytes()...)
}

func (heap *Heap) TopSize() int {
	return heap.top.Len()
}

func (heap *Heap) Size() int {
	return heap.size
}

func (heap *Heap) Free() int {
	return heap.size - heap.top.Len() - heap.bottom.Len()
}

func (heap *Heap) CursorTop() int {
	return heap.top.Len()
}

func (heap *Heap) CursorBottom() int {
	return heap.size - heap.bottom.Len()
}

// Pad the data on the top buffer with zeroes until its length is a multiple of `align`
func (heap *Heap) AlignTop(align int) error {
	remainder := align - (heap.top.Len() % align)
	padding := make([]byte, remainder)
	_, err := heap.top.Write(padding)
	return err
}

// Write to the start of the buffer. New data appends on the right of the existing data.
func (heap *Heap) WriteTop(data []byte) error {
	if len(data) > heap.Free() {
		return io.ErrShortWrite
	}
	_, err := heap.top.Write(data)
	return err
}

// Write to the end of the buffer. New data appends on the left of the existing data.
func (heap *Heap) WriteBottom(data []byte) error {
	if len(data) > heap.Free() {
		return io.ErrShortWrite
	}
	buf := &bytes.Buffer{}
	_, err := buf.Write(data)
	if err != nil {
		return err
	}
	_, err = buf.Write(heap.bottom.Bytes())
	if err != nil {
		return err
	}
	heap.bottom = buf
	return err
}

func (heap *Heap) TopWriter() io.Writer {
	return &heapWriter{write: heap.WriteTop}
}

func (heap *Heap) BottomWriter() io.Writer {
	return &heapWriter{write: heap.WriteBottom}
}

// Write the whole heap, filling in the blanks with null bytes.
func (heap *Heap) MarshalBinary() ([]byte, error) {
	buf := make([]byte, heap.size)

	copy(buf, heap.top.Bytes())
	copy(buf[heap.CursorBottom():], heap.bottom.Bytes())

	return buf, nil
}
