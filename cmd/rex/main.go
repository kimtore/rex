package main

import (
	`context`
	`flag`
	`fmt`
	`io`
	`io/fs`
	`os`
	`path/filepath`

	`github.com/ambientsound/rex/pkg/library`
	`github.com/ambientsound/rex/pkg/mediascanner`
	`github.com/ambientsound/rex/pkg/rekordbox/color`
	`github.com/ambientsound/rex/pkg/rekordbox/column`
	`github.com/ambientsound/rex/pkg/rekordbox/dbengine`
	`github.com/ambientsound/rex/pkg/rekordbox/page`
	`github.com/ambientsound/rex/pkg/rekordbox/pdb`
	`github.com/ambientsound/rex/pkg/rekordbox/playlist`
	`github.com/ambientsound/rex/pkg/rekordbox/unknown17`
	`github.com/ambientsound/rex/pkg/rekordbox/unknown18`
)

func main() {
	lib := library.New()

	// Initialize options
	basedir := flag.String("root", "./", "Root path of USB drive")
	forceOverwrite := flag.Bool("f", false, "Overwrite export file if it exists")
	scandir := flag.String("scan", "./", "Path to music files, scanned recursively. Must only contain MP3 files.")
	flag.Parse()

	// Create output directory
	outputPath := filepath.Join(*basedir, "PIONEER", "rekordbox")
	err := os.MkdirAll(outputPath, 0755)
	if err != nil {
		panic(err)
	}

	// Open output file for writing
	outputFile := filepath.Join(outputPath, "export.pdb")
	var flags = os.O_CREATE | os.O_RDWR
	if *forceOverwrite {
		flags |= os.O_TRUNC
	}
	out, err := os.OpenFile(outputFile, flags, 0644)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Traverse directories and scan for music
	walk := func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		probe, err := mediascanner.ProbeMetadata(context.TODO(), path)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Add %q\n", path)
		t := mediascanner.TrackFromFile(lib, path, *probe)
		lib.InsertTrack(t)
		return nil
	}

	err = filepath.Walk(*scandir, walk)
	if err != nil {
		panic(err)
	}

	// Intermediary type for storing "INSERT statements"
	type Insert struct {
		Type page.Type
		Row  page.Row
	}
	inserts := make([]Insert, 0)

	// Create PDB data types for tracks, artists, albums and playlists.
	tracks := lib.Tracks().All()
	for i := range tracks {
		pdbtrack := mediascanner.PdbTrack(lib, tracks[i], *basedir)
		inserts = append(inserts, Insert{
			Type: page.Type_Tracks,
			Row:  &pdbtrack,
		})
	}

	artists := lib.Artists().All()
	for i := range artists {
		pdbartist := mediascanner.PdbArtist(lib, artists[i])
		inserts = append(inserts, Insert{
			Type: page.Type_Artists,
			Row:  &pdbartist,
		})
	}

	albums := lib.Albums().All()
	for i := range albums {
		pdbalbum := mediascanner.PdbAlbum(lib, albums[i])
		inserts = append(inserts, Insert{
			Type: page.Type_Albums,
			Row:  &pdbalbum,
		})
	}

	pl := &playlist.Playlist{
		PlaylistHeader: playlist.PlaylistHeader{
			Id: 1,
		},
		Name: "REX tracks",
	}
	inserts = append(inserts, Insert{
		Type: page.Type_PlaylistTree,
		Row:  pl,
	})

	for i, t := range tracks {
		ent := &playlist.Entry{
			EntryIndex: uint32(i + 1),
			TrackID:    uint32(lib.Tracks().ID(t)),
			PlaylistID: 1,
		}
		inserts = append(inserts, Insert{
			Type: page.Type_PlaylistEntries,
			Row:  ent,
		})
	}

	for _, uk := range unknown17.InitialDataset {
		inserts = append(inserts, Insert{
			Type: page.Type_Unknown17,
			Row:  uk,
		})
	}

	for _, uk := range unknown18.InitialDataset {
		inserts = append(inserts, Insert{
			Type: page.Type_Unknown18,
			Row:  uk,
		})
	}

	for _, uk := range color.InitialDataset {
		inserts = append(inserts, Insert{
			Type: page.Type_Colors,
			Row:  uk,
		})
	}

	for _, uk := range column.InitialDataset {
		inserts = append(inserts, Insert{
			Type: page.Type_Columns,
			Row:  uk,
		})
	}

	// Initialize the database.
	db := dbengine.New(out)

	// Create all tables found in a typical rekordbox export.
	for _, pageType := range pdb.TableOrder {
		err = db.CreateTable(pageType)
		if err != nil {
			panic(err)
		}
	}

	// Generate data pages with the inserts generated earlier.
	// When a data page is full, it is inserted into the db.
	// This is a quick and dirty way for export ONLY,
	// it will not work to modify existing databases.
	dataPages := make(map[page.Type]*page.Data)
	for _, insert := range inserts {
		if dataPages[insert.Type] == nil {
			dataPages[insert.Type] = page.NewPage(insert.Type)
		}
		err = dataPages[insert.Type].Insert(insert.Row)
		if err == nil {
			continue
		}
		if err == io.ErrShortWrite {
			err = db.InsertPage(dataPages[insert.Type])
			if err != nil {
				panic(err)
			}
			dataPages[insert.Type] = nil
			continue
		}
		panic(err)
	}

	// Insert the remainding pages.
	for _, pg := range dataPages {
		if pg == nil {
			continue
		}
		err = db.InsertPage(pg)
		if err != nil {
			panic(err)
		}
	}

	// Flush buffers and exit program.
	err = out.Close()
	if err == nil {
		fmt.Printf("Wrote %s\n", out.Name())
	}
}
