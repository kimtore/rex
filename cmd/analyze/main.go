package main

import (
	`bytes`
	`encoding/binary`
	`flag`
	`fmt`
	`io`
	`os`
	`strconv`

	`github.com/ambientsound/rex/pkg/rekordbox/color`
	`github.com/ambientsound/rex/pkg/rekordbox/column`
	`github.com/ambientsound/rex/pkg/rekordbox/dbengine`
	`github.com/ambientsound/rex/pkg/rekordbox/page`
	`github.com/ambientsound/rex/pkg/rekordbox/track`
	`github.com/ambientsound/rex/pkg/rekordbox/unknown17`
	`github.com/ambientsound/rex/pkg/rekordbox/unknown18`
)

/*
This program analyzes a PDB file block by block. That is, each block is parsed separately and not put together into a larger structure.
This means that all kinds of files, also corrupt (mis-generated) can be inspected.
*/

var (
	printIndex = flag.Bool("index", false, "print contents of index structure")
	printRows  = flag.Bool("rows", false, "print individual rows")
	dumb       = flag.Bool("dumb", false, "don't attempt to parse tables")
)

func main() {
	err := run()
	if err != nil {
		fmt.Printf("fatal error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()

	f, err := os.Open(flag.Args()[0])
	if err != nil {
		return err
	}
	defer f.Close()

	if *dumb {
		return run_ordered(f)
	}
	return run_parser(f)
}

func run_parser(f io.ReadWriteSeeker) error {

	db, err := dbengine.Open(f)
	if err != nil {
		return err
	}

	fmt.Printf("PIONEER DJ DeviceSQL file\n")
	fmt.Printf("tables=%d, tx=%d\n", db.Globals.NumTables, db.Globals.Sequence)

	types := db.TableTypes()
	for _, ty := range types {
		table, err := db.GetTable(ty)
		if err != nil {
			return err
		}
		tableName := ty.String()[5:]
		fmt.Printf("Table: %q\n", tableName)
		fmt.Printf(
			"  Meta: u1=%04x u2=%04x numentries=%02d\n",
			table.Index.IndexHeader.Unknown1,
			table.Index.IndexHeader.Unknown2,
			table.Index.IndexHeader.NumEntries,
		)

		totalRows := 0
		totalActive := 0
		totalRowSets := 0

		if *printIndex && table.Index.NumEntries > 0 {
			fmt.Printf("  Indexes:")
			for i := range table.Index.IndexEntries {
				fmt.Printf(" %04x", table.Index.IndexEntries[i])
			}
			fmt.Printf("\n")
		}

		for _, pg := range table.Pages {
			lr := strconv.FormatUint(uint64(pg.NumRowsLarge), 2)
			fmt.Printf("  %s Page: idx=%02x rows=%3d deleted=%3d used=%4d free=%4d large=%010s tx=%04x flags=%02x u3=%04x u4=%04x u5=%04x\n",
				tableName,
				pg.Header.PageIndex,
				pg.Header.NumRowsSmall,
				int(pg.Header.NumRowsSmall)-pg.ActiveRows(),
				pg.Header.NextHeapWriteOffset,
				pg.Header.FreeSize,
				lr,
				pg.Header.Transaction,
				pg.Header.PageFlags,
				pg.Header.Unknown3,
				pg.Header.Unknown4,
				pg.DataHeader.Unknown5,
			)

			totalRows += int(pg.Header.NumRowsSmall)
			totalActive += pg.ActiveRows()
			totalRowSets += len(pg.RowSets)

			if !*printRows {
				continue
			}

			for _, rs := range pg.RowSets {
				bm := strconv.FormatUint(uint64(rs.ActiveRows), 2)
				pd := strconv.FormatUint(uint64(rs.LastWrittenRows), 2)
				// an := strconv.FormatUint(uint64(rs.ActiveRows&rs.LastWrittenRows), 2)
				// xo := strconv.FormatUint(uint64(rs.ActiveRows^rs.LastWrittenRows), 2)
				fmt.Printf("    RowSet: bitmask=%016s lastwrite=%016s\n", bm, pd)
			}

			for rowNum, rowref := range pg.HeapPositions() {
				if table.Type == page.Type_Tracks {

					tr := &track.Track{}
					err = pg.UnmarshalRow(tr, rowref.HeapPosition)
					if err != nil {
						fmt.Printf("      Track: io.EOF\n")
					} else {
						fmt.Printf("      Track: heap=%04x id=%04x shift=%02x exists=%-5v path=%q\n", rowref.HeapPosition, tr.Id, tr.IndexShift, rowref.Exists, tr.FilePath)
					}
				} else if table.Type == page.Type_Unknown18 {
					row := &unknown18.Unknown18{}
					err = pg.UnmarshalRow(row, rowref.HeapPosition)
					fmt.Printf("      %#v\n", row)
				} else if table.Type == page.Type_Columns {
					row := &column.Column{}
					err = pg.UnmarshalRow(row, rowref.HeapPosition)
					fmt.Printf("      %04x %#v\n", rowref.HeapPosition, row)
				} else if table.Type == page.Type_Colors {
					row := &color.Color{}
					err = pg.UnmarshalRow(row, rowref.HeapPosition)
					fmt.Printf("      %04x %#v\n", rowref.HeapPosition, row)
				} else if table.Type == page.Type_Unknown17 {
					row := &unknown17.Unknown17{}
					err = pg.UnmarshalRow(row, rowref.HeapPosition)
					fmt.Printf("      %#v\n", row)
				} else {
					fmt.Printf("      Row: index=%03d heap=%04x exists=%v\n", rowNum, rowref.HeapPosition, rowref.Exists)
				}
			}
		}

		fmt.Printf("  Table summary: records=%d deleted=%d total=%d rowsets=%d\n", totalActive, totalRows-totalActive, totalRows, totalRowSets)
	}

	return nil
}

func run_ordered(f io.ReadWriteSeeker) error {
	flag.Parse()

	const blocksize = 4096
	buf := make([]byte, blocksize)

	db, err := dbengine.Open(f)
	if err != nil {
		return err
	}

	fmt.Printf("%05x: numtables=%d, next_unused_page=%05x, sequence=%d\n", 0, db.Globals.NumTables, db.Globals.NextUnusedPage*blocksize, db.Globals.Sequence)

	for _, ptr := range db.Globals.Pointers {
		fmt.Printf("%-20s first=%02x last=%02x empty_candidate=%02x\n", ptr.Type.String()[5:], ptr.FirstPage, ptr.LastPage, ptr.EmptyCandidate)
	}

	blanks := 0
	i := 1
	_, err = f.Seek(blocksize, io.SeekStart)
	if err != nil {
		return err
	}

	for ; ; i++ {
		_, err = io.ReadAtLeast(f, buf, len(buf))
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		header := &page.Header{}
		idxheader := &page.IndexHeader{}
		r := bytes.NewReader(buf)

		err = binary.Read(r, binary.LittleEndian, header)
		if err != nil {
			return err
		}
		err = binary.Read(r, binary.LittleEndian, idxheader)
		if err != nil {
			return err
		}

		// isIndex := header.PageFlags & 0x64

		if header.Type == 0 && header.PageIndex == 0 {
			fmt.Printf("%05x: NO DATA\n", i*blocksize)
			blanks++
			continue
		}

		fmt.Printf("%05x: idx=%02x next=%05x seq=%d type=%-16s",
			i*blocksize,
			header.PageIndex,
			header.NextPage*blocksize,
			header.Transaction,
			header.Type.String()[5:],
		)

		switch header.PageFlags {
		case 0x64:
			fmt.Printf(" <INDEX>")
			if !*printIndex {
				fmt.Printf("\n")
				continue
			}
			if idxheader.NumEntries == 0 {
				fmt.Printf(" <EMPTY>\n")
				continue
			}
			fmt.Printf(
				" entries=%d u1=%04x u2=%04x break=%04x\n",
				idxheader.NumEntries,
				idxheader.Unknown1,
				idxheader.Unknown2,
				idxheader.FirstEmptyEntry,
			)
			for idxnum := 0; idxnum < int(idxheader.NextOffset); idxnum++ {
				var unknown uint32
				err = binary.Read(r, binary.LittleEndian, &unknown)
				if err != nil {
					return err
				}
				// if unknown == 0x1ffffff8 {
				// 	break
				// }
				s := strconv.FormatUint(uint64(unknown), 2)
				fmt.Printf("> pos=%02d dec=%04d hex=%08x bin=%032s\n", idxnum, unknown, unknown, s)
			}
		case 0x37:
			fmt.Printf(" <XXX>")
			fallthrough
		case 0x34:
			fmt.Printf(" <REF>")
			fallthrough
		case 0x24:
			fmt.Printf(" <DATA>")
			fmt.Printf(" rows=%d\n", header.NumRowsSmall)
		default:
			panic("cannot handle it")
		}
	}

	return nil
}
