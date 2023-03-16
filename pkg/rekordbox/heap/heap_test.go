package heap_test

import (
	`io`
	`strings`
	`testing`

	`github.com/ambientsound/rex/pkg/rekordbox/heap`
	`github.com/stretchr/testify/assert`
)

func TestPageHeap_MarshalBinary(t *testing.T) {
	t.Run("write to both ends of heap", func(t *testing.T) {
		const heapSize = 32
		page := heap.New(heapSize)

		page.WriteTop([]byte("foo"))
		page.WriteTop([]byte("bar"))
		page.WriteTop([]byte("baz"))

		page.WriteBottom([]byte("foo"))
		page.WriteBottom([]byte("bar"))
		page.WriteBottom([]byte("baz"))

		data, err := page.MarshalBinary()
		assert.NoError(t, err)

		// 00000000  66 6f 6f 62 61 72 62 61  7a 00 00 00 00 00 00 00  |foobarbaz.......|
		// 00000010  00 00 00 00 00 00 00 62  61 7a 62 61 72 66 6f 6f  |.......bazbarfoo|
		assert.Len(t, data, heapSize)
		assert.Equal(t, []byte{
			0x66, 0x6f, 0x6f, 0x62, 0x61, 0x72, 0x62, 0x61,
			0x7a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x62,
			0x61, 0x7a, 0x62, 0x61, 0x72, 0x66, 0x6f, 0x6f,
		}, data)
	})

	t.Run("test top writer", func(t *testing.T) {
		page := heap.New(32)

		input := strings.NewReader("hello, world!")
		w := page.TopWriter()
		_, err := io.Copy(w, input)

		assert.NoError(t, err)

		data, err := page.MarshalBinary()
		assert.NoError(t, err)

		assert.Equal(t, []byte{
			0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		}, data)
	})

	t.Run("test bottom writer", func(t *testing.T) {
		page := heap.New(32)

		input := strings.NewReader("hello, world!")
		w := page.BottomWriter()
		_, err := io.Copy(w, input)

		assert.NoError(t, err)

		data, err := page.MarshalBinary()
		assert.NoError(t, err)

		assert.Equal(t, []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21,
		}, data)
	})
}
