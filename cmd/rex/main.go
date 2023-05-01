package main

import (
	`context`
	`database/sql`
	`flag`
	`fmt`
	`io`
	`os`
	`path/filepath`

	`github.com/ambientsound/rex/pkg/library`
	`github.com/ambientsound/rex/pkg/mediascanner`
	`github.com/ambientsound/rex/pkg/mixxx`
	`github.com/ambientsound/rex/pkg/rekordbox/color`
	`github.com/ambientsound/rex/pkg/rekordbox/column`
	`github.com/ambientsound/rex/pkg/rekordbox/dbengine`
	`github.com/ambientsound/rex/pkg/rekordbox/page`
	`github.com/ambientsound/rex/pkg/rekordbox/pdb`
	`github.com/ambientsound/rex/pkg/rekordbox/playlist`
	`github.com/ambientsound/rex/pkg/rekordbox/unknown17`
	`github.com/ambientsound/rex/pkg/rekordbox/unknown18`

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	err := run()
	if err != nil {
		fmt.Printf("fatal error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	var err error

	fmt.Printf("REX: unofficial Pioneer DJ export.pdb generator\n")
	fmt.Printf("This software is neither supported nor endorsed by Pioneer.\n")
	fmt.Printf("Please do not rely on it for serious use.\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	lib := library.New()

	// Initialize options
	basedir := flag.String("root", "./", "Root path of USB drive")
	trackDir := flag.String("trackdir", "rex", "Where on the USB drive to put exported files, relative to root path")
	forceOverwrite := flag.Bool("f", false, "Overwrite export file if it exists")
	mixxxdbPath := flag.String("mixxxdb", defaultMixxxDbPath(), "Path to Mixxx database")
	flag.Parse()

	*basedir, err = filepath.Abs(*basedir)
	if err != nil {
		return err
	}

	// Open Mixxx database
	sqliteHandle, err := sql.Open("sqlite3", *mixxxdbPath)
	if err != nil {
		return fmt.Errorf("open Mixxx database: %w", err)
	}
	defer sqliteHandle.Close()
	mixxxdb := mixxx.New(sqliteHandle)
	fmt.Printf("Mixxx database opened: %s\n", *mixxxdbPath)

	// Create output directories
	outputPath := filepath.Join(*basedir, "PIONEER", "rekordbox")
	err = os.MkdirAll(outputPath, 0755)
	if err != nil {
		return err
	}
	*trackDir = filepath.Join(*basedir, *trackDir)
	*trackDir, err = filepath.Abs(*trackDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(*trackDir, 0755)
	if err != nil {
		return err
	}

	// Open output file for writing
	outputFile := filepath.Join(outputPath, "export.pdb")
	outputFile, err = filepath.Abs(outputFile)
	if err != nil {
		return err
	}
	var flags = os.O_CREATE | os.O_RDWR
	if *forceOverwrite {
		flags |= os.O_TRUNC
	}
	out, err := os.OpenFile(outputFile, flags, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	fmt.Printf("PIONEER database created: %s\n", outputFile)

	// Scan Mixxx library for tracks
	srcTracks, err := mixxxdb.ListTracks(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Found %d tracks in Mixxx database\n", len(srcTracks))
	trackCandidates := make(map[string]*library.Track, len(srcTracks))
	for i, track := range srcTracks {
		t := mediascanner.TrackFromMixxx(track)
		trackCandidates[t.Path] = t
		fmt.Printf("\033[2K\r[%6d/%6d] %s", i+1, len(srcTracks), t.Title)
	}
	fmt.Printf("\033[2K\r")
	fmt.Printf("Tracks imported.\n")

	// Create playlists
	mixxPlaylists, err := mixxxdb.ListPlaylists(ctx)
	if err != nil {
		return err
	}
	for _, plist := range mixxPlaylists {
		if plist.Hidden > 0 {
			continue
		}
		tracks, err := mixxxdb.ListPlaylistTracks(ctx, sql.NullInt64{Int64: plist.ID, Valid: true})
		if err != nil {
			return err
		}
		pplist := &library.Playlist{
			ID:     library.ID(plist.ID),
			Name:   "P: " + plist.Name.String,
			Tracks: make([]*library.Track, 0),
		}
		for _, track := range tracks {
			t := lib.Tracks().GetByName(track.Path.String)
			if t == nil {
				var found bool
				t, found = trackCandidates[track.Path.String]
				if !found {
					return fmt.Errorf("database incoherent: %s not found", track.Path.String)
				}
				lib.InsertTrack(t)
				delete(trackCandidates, track.Path.String)
			}
			pplist.Tracks = append(pplist.Tracks, t)
		}
		lib.Playlists().Insert(pplist)
		fmt.Printf("Playlist %q loaded with %d tracks\n", pplist.Name, len(pplist.Tracks))
	}

	// Create playlists from crates
	mixxCrates, err := mixxxdb.ListCrates(ctx)
	if err != nil {
		return err
	}
	for _, crate := range mixxCrates {
		if crate.Show.Int64 == 0 {
			continue
		}
		if crate.Locked.Int64 > 0 {
			continue
		}
		tracks, err := mixxxdb.ListCrateTracks(ctx, crate.ID)
		if err != nil {
			return err
		}
		pplist := &library.Playlist{
			ID:     library.ID(crate.ID),
			Name:   "C: " + crate.Name,
			Tracks: make([]*library.Track, 0),
		}
		for _, track := range tracks {
			t := lib.Tracks().GetByName(track.Path.String)
			if t == nil {
				var found bool
				t, found = trackCandidates[track.Path.String]
				if !found {
					return fmt.Errorf("database incoherent: %s not found", track.Path.String)
				}
				lib.InsertTrack(t)
				delete(trackCandidates, track.Path.String)
			}
			pplist.Tracks = append(pplist.Tracks, t)
		}
		lib.Playlists().Insert(pplist)
		fmt.Printf("Crate %q loaded with %d tracks\n", pplist.Name, len(pplist.Tracks))
	}

	fmt.Printf("Tracks marked for export: %6d used/%6d total\n", len(lib.Tracks().All()), len(srcTracks))
	fmt.Printf("Copying or encoding tracks to %s\n", *trackDir)

	for i, t := range lib.Tracks().All() {
		fmt.Printf("\r[%6d/%6d] ", i+1, len(lib.Tracks().All()))
		result, err := mediascanner.RenderTo(ctx, t, *trackDir)
		if err != nil {
			fmt.Printf("\n")
			return fmt.Errorf("render %q: %w\n", t.OutputPath, err)
		}
		fmt.Printf("\033[2K\r[%6d/%6d] %s %s", i+1, len(lib.Tracks().All()), result.Action, t.OutputPath)
	}

	fmt.Printf("\033[2K\r")
	fmt.Printf("All tracks copied to destination\n")
	fmt.Printf("Writing PDB file...\n")

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

	// Generate playlists
	playlists := lib.Playlists().All()
	for playlistID := range playlists {
		pl := &playlist.Playlist{
			PlaylistHeader: playlist.PlaylistHeader{
				Id: uint32(playlistID),
			},
			Name: playlists[playlistID].GetName(),
		}
		inserts = append(inserts, Insert{
			Type: page.Type_PlaylistTree,
			Row:  pl,
		})
		for trackIndex, t := range playlists[playlistID].Tracks {
			ent := &playlist.Entry{
				EntryIndex: uint32(trackIndex + 1),
				TrackID:    uint32(lib.Tracks().ID(t)),
				PlaylistID: uint32(playlistID),
			}
			inserts = append(inserts, Insert{
				Type: page.Type_PlaylistEntries,
				Row:  ent,
			})
		}
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
		fmt.Printf("Finished successfully.\n")
	}

	return nil
}

func defaultMixxxDbPath() string {
	homedir, _ := os.UserHomeDir()
	return filepath.Join(homedir, ".mixxx", "mixxxdb.sqlite")
}
