package page_test

import (
	`bytes`
	`encoding/binary`
	`fmt`
	`io`
	`os`
	`testing`

	`github.com/ambientsound/rex/pkg/rekordbox/dbengine`
	`github.com/ambientsound/rex/pkg/rekordbox/page`
	`github.com/ambientsound/rex/pkg/rekordbox/pdb`
	`github.com/ambientsound/rex/pkg/rekordbox/track`
	`github.com/lunixbochs/struc`
	`github.com/stretchr/testify/assert`
)

// Test that we can try to squeeze many records into a page,
// detect when it's full, and that it is correctly written.
func TestPage_MarshalBinary(t *testing.T) {
	const numRows = 64
	const pageSize = 256
	const expectedAdds = 5

	pg := page.NewPage(page.Type_Tracks)

	rows := make([][]byte, numRows)
	for i := 0; i < numRows; i++ {
		rows[i] = []byte("Xo")
	}

	totalWritten := 0
	for {
		rowsAdded, err := pg.AddRows(rows)
		rows = rows[rowsAdded:]
		totalWritten += rowsAdded
		if err == io.ErrShortWrite {
			break
		}
	}

	assert.Equal(t, expectedAdds, totalWritten)

	data, err := pg.MarshalBinary()
	assert.Len(t, data, pageSize)
	assert.NoError(t, err)
	assert.Equal(t, []byte{
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xc5, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x14, 0x14, 0x0, 0x24, 0x8, 0x0, 0xd0, 0x0, 0x0, 0x0, 0x14, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x98, 0x0, 0x90, 0x0, 0x88, 0x0, 0x80, 0x0, 0xf, 0x0, 0x0, 0x0, 0x78, 0x0, 0x70, 0x0, 0x68, 0x0, 0x60, 0x0, 0x58, 0x0, 0x50, 0x0, 0x48, 0x0, 0x40, 0x0, 0x38, 0x0, 0x30, 0x0, 0x28, 0x0, 0x20, 0x0, 0x18, 0x0, 0x10, 0x0, 0x8, 0x0, 0x0, 0x0, 0xff, 0xff, 0x0, 0x0,
	}, data)
	// 00000000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// 00000010  c5 0a 00 00 00 00 00 00  14 14 00 24 08 00 d0 00  |...........$....|
	// 00000020  00 00 14 00 00 00 00 00  58 6f 00 00 00 00 00 00  |........Xo......|
	// 00000030  58 6f 00 00 00 00 00 00  58 6f 00 00 00 00 00 00  |Xo......Xo......|
	// 00000040  58 6f 00 00 00 00 00 00  58 6f 00 00 00 00 00 00  |Xo......Xo......|
	// 00000050  58 6f 00 00 00 00 00 00  58 6f 00 00 00 00 00 00  |Xo......Xo......|
	// 00000060  58 6f 00 00 00 00 00 00  58 6f 00 00 00 00 00 00  |Xo......Xo......|
	// 00000070  58 6f 00 00 00 00 00 00  58 6f 00 00 00 00 00 00  |Xo......Xo......|
	// 00000080  58 6f 00 00 00 00 00 00  58 6f 00 00 00 00 00 00  |Xo......Xo......|
	// 00000090  58 6f 00 00 00 00 00 00  58 6f 00 00 00 00 00 00  |Xo......Xo......|
	// 000000a0  58 6f 00 00 00 00 00 00  58 6f 00 00 00 00 00 00  |Xo......Xo......|
	// 000000b0  58 6f 00 00 00 00 00 00  58 6f 00 00 00 00 00 00  |Xo......Xo......|
	// 000000c0  58 6f 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |Xo..............|
	// 000000d0  98 00 90 00 88 00 80 00  0f 00 00 00 78 00 70 00  |............x.p.|
	// 000000e0  68 00 60 00 58 00 50 00  48 00 40 00 38 00 30 00  |h.`.X.P.H.@.8.0.|
	// 000000f0  28 00 20 00 18 00 10 00  08 00 00 00 ff ff 00 00  |(. .............|

	// assert.Equal(t, []byte{}, data)
}

func TestPage_Insert(t *testing.T) {
	var err error
	f, err := os.Create("/tmp/dbengine.pdb")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	db := dbengine.New(f)
	assert.NoError(t, db.CreateTable(page.Type_Tracks))

	tr := &track.Track{}
	pg := page.NewPage(page.Type_Tracks)
	for err == nil {
		tr.Id++
		tr.Title = fmt.Sprintf("Track %d", tr.Id)
		err = pg.Insert(tr)
	}

	assert.NoError(t, db.InsertPage(pg))
}

func TestPage_Analyze(t *testing.T) {

	// test candidates
	const testdump = "/home/kimt/windows/pioneers/PIONEER/rekordbox/export.pdb"
	const meteor = "/run/media/kimt/METEOR/PIONEER/rekordbox/export.pdb"
	const rex = "/tmp/rex.pdb"
	const dbengine = "/tmp/dbengine.pdb"

	const blocksize = 4096
	buf := make([]byte, blocksize)
	f, err := os.Open(dbengine)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = io.ReadAtLeast(f, buf, len(buf))
	if err == io.EOF {
		return
	}
	if err != nil {
		panic(err)
	}
	r := bytes.NewReader(buf)
	ph := &pdb.FileHeader{}
	err = struc.UnpackWithOptions(r, ph, &struc.Options{Order: binary.LittleEndian})
	if err != nil && err != io.EOF {
		panic(err)
	}

	t.Logf("%05x: numtables=%d, next_unused_page=%05x, sequence=%d", 0, ph.NumTables, ph.NextUnusedPage*blocksize, ph.Sequence)

	for _, ptr := range ph.Pointers {
		t.Logf("%-20s first=%02x last=%02x empty_candidate=%02x", ptr.Type.String()[5:], ptr.FirstPage, ptr.LastPage, ptr.EmptyCandidate)
	}

	blanks := 0
	i := 1

	for ; ; i++ {
		_, err = io.ReadAtLeast(f, buf, len(buf))
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		header := &page.Header{}
		idxheader := &page.IndexHeader{}
		r := bytes.NewReader(buf)

		err = binary.Read(r, binary.LittleEndian, header)
		if err != nil {
			panic(err)
		}
		err = binary.Read(r, binary.LittleEndian, idxheader)
		if err != nil {
			panic(err)
		}

		// isIndex := header.PageFlags & 0x64

		if header.Type == 0 && header.PageIndex == 0 {
			t.Logf("%05x: NO DATA", i*blocksize)
			blanks++
			continue
		}

		t.Logf("%05x: idx=%02x next=%05x type=%-20s flags=0x%02x seq=%d",
			i*blocksize,
			header.PageIndex,
			header.NextPage*blocksize,
			header.Type.String()[5:],
			header.PageFlags,
			header.Transaction,
		)

		if header.PageFlags&0x64 == 0x64 {
			t.Logf(
				"index: u1=%04x, u2=%04x, u4=%04x, u7=%04x, u8=%04x",
				idxheader.Unknown1,
				idxheader.Unknown2,
				idxheader.NextOffset,
				idxheader.NumEntries,
				idxheader.FirstEmptyEntry,
			)
			continue
			var unknown [8]uint32
			for i := 0; i < 2; i++ {
				err = binary.Read(r, binary.LittleEndian, &unknown)
				if err != nil {
					panic(err)
				}
				t.Logf(
					"index: %08x %08x %08x %08x %08x %08x %08x %08x",
					unknown[0],
					unknown[1],
					unknown[2],
					unknown[3],
					unknown[4],
					unknown[5],
					unknown[6],
					unknown[7],
				)
			}
		}
	}

	t.Logf("Finished with blanks=%d total=%d", blanks, i)

	// # Table order
	//
	// Every time a page is written to disk, the next page address in that table series is allocated.
	// This address comes from the global unused page variable in Data.NextPage.
	//
	// Let's say we create the Track table. First, an index is created with flags 0x64:
	// 00000: numtables=20, next_unused_page=b8000, version=4422
	// 01000: idx=01 next=02000 type=Type_Tracks          flags=0x64|01100100
	// 02000: idx=02 next=34000 type=Type_Tracks          flags=0x24|00100100
	//
	// Then a bunch of other tables are created. What's at 34000?
	// 34000: idx=34 next=37000 type=Type_Tracks          flags=0x24|00100100
	// ...
	// 37000: idx=37 next=38000 type=Type_Tracks          flags=0x24|00100100
	// ... and so forth ...
	//
	// Finally, at the end of the file we have:
	// b2000: idx=b2 next=b4000 type=Type_Tracks          flags=0x34|00110100
	// b3000: NO DATA
	// b4000: idx=b4 next=b7000 type=Type_Tracks          flags=0x34|00110100
	//
	// The last record is at b4000. But b5000, b6000, b7000 have all been referenced by pages.
	// Thus, the next page entry in the file header is the next = b8000.

}
