package mediascanner

// Use FFMPEG to scan music files for metadata.

import (
	`context`
	`encoding/json`
	`fmt`
	`io`
	`os`
	`os/exec`
	`path/filepath`
	`strconv`
	`strings`
	`time`

	`github.com/ambientsound/rex/pkg/library`
	`github.com/ambientsound/rex/pkg/mixxx`
	`github.com/ambientsound/rex/pkg/rekordbox/album`
	`github.com/ambientsound/rex/pkg/rekordbox/artist`
	`github.com/ambientsound/rex/pkg/rekordbox/track`
)

type Probe struct {
	Format struct {
		Filesize string `json:"size"`
		Duration string `json:"duration"`
		Tags     struct {
			Title       string `json:"title"`
			Artist      string `json:"artist"`
			Album       string `json:"album"`
			Genre       string `json:"genre"`
			TrackNumber string `json:"track"`
			Date        string `json:"date"`
		} `json:"tags"`
	} `json:"format"`
}

func ProbeMetadata(ctx context.Context, src string) (*Probe, error) {
	proc := exec.CommandContext(ctx, "ffprobe", "-show_format", "-print_format", "json", src)
	output, err := proc.Output()
	if err != nil {
		return nil, err
	}
	probe := &Probe{}
	err = json.Unmarshal(output, probe)
	if err != nil {
		return nil, err
	}

	return probe, nil
}

func intOrZero[T int | uint16 | uint32](s string) T {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return T(i)
}

func detectDate(input string) *time.Time {
	if len(input) == 4 {
		tm, _ := time.Parse("2006", input)
		return &tm
	} else if len(input) == 7 {
		tm, _ := time.Parse("2006-01", input)
		return &tm
	} else if len(input) >= 10 {
		tm, _ := time.Parse("2006-01-02", input[:10])
		return &tm
	}
	return nil
}

func parseDuration(input string) time.Duration {
	dur, _ := time.ParseDuration(input + "s")
	return dur
}

func yearOrZero(tm *time.Time) uint16 {
	if tm == nil {
		return 0
	}
	return uint16(tm.Year())
}

func TrackFromMixxx(track mixxx.ListTracksRow) *library.Track {
	trackNumber, _ := strconv.Atoi(track.Tracknumber.String)
	return &library.Track{
		Path:        track.Path.String,
		Title:       track.Title.String,
		SampleRate:  float64(track.Samplerate.Int64),
		FileSize:    int(track.Filesize.Int64),
		Bitrate:     int(track.Bitrate.Int64),
		TrackNumber: trackNumber,
		Tempo:       track.Bpm.Float64,
		FileType:    track.Filetype.String,
		AddedDate:   detectDate(track.DatetimeAdded.String),
		Duration:    time.Duration(track.Duration.Float64),
		Artist:      track.Artist.String,
		Album:       track.Album.String,
		// SampleDepth
		// DiscNumber
		// ReleaseDate
		// Isrc
	}
}

type RenderResult struct {
	Action string
}

func RenderTo(ctx context.Context, t *library.Track, outputDir string) (*RenderResult, error) {
	{
		filename := filepath.Base(t.Path)
		outputPath := filepath.Join(outputDir, filename)

		if t.FileType != "mp3" {
			outputPath += ".mp3"
		}
		t.OutputPath = outputPath
	}

	_, err := os.Stat(t.OutputPath)
	if err == nil {
		return &RenderResult{Action: "skip"}, nil
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	switch t.FileType {
	case "mp3":
		err = CopyFile(t.Path, t.OutputPath)
		return &RenderResult{Action: "copy"}, err
	default:
		err = ConvertToMP3(ctx, t.Path, t.OutputPath)
		return &RenderResult{Action: "encode"}, err
	}
}

func ConvertToMP3(ctx context.Context, src, dst string) error {
	proc := exec.CommandContext(ctx, "ffmpeg",
		"-i", src,
		"-map_metadata", "0",
		"-codec:a", "libmp3lame",
		"-qscale:a", "0",
		"-joint_stereo", "0",
		dst,
	)
	out, err := proc.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\n%s", err, string(out))
	}
	return nil
}

func CopyFile(inputPath, outputPath string) error {
	in, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)

	if err != nil {
		os.Remove(out.Name())
	}

	return err
}

func TrackFromFile(lib *library.Library, path string, probe Probe) *library.Track {
	now := time.Now()
	return &library.Track{
		Path:        path,
		Bitrate:     320,   // FIXME
		Tempo:       128,   // FIXME
		SampleDepth: 16,    // FIXME
		SampleRate:  44100, // FIXME
		DiscNumber:  0,     // FIXME
		Isrc:        "",    // FIXME
		FileSize:    intOrZero[int](probe.Format.Filesize),
		TrackNumber: intOrZero[int](probe.Format.Tags.TrackNumber),
		ReleaseDate: detectDate(probe.Format.Tags.Date),
		AddedDate:   &now,
		Artist:      probe.Format.Tags.Artist,
		Album:       probe.Format.Tags.Album,
		Duration:    parseDuration(probe.Format.Duration),
		Title:       probe.Format.Tags.Title,
	}
}

func PdbTrack(lib *library.Library, t *library.Track, baseDir string) track.Track {
	const isoDateFormat = "2006-01-02"
	baseDir = strings.TrimRight(baseDir, "/")
	filePath := t.OutputPath
	if strings.HasPrefix(filePath, baseDir) {
		filePath = filePath[len(baseDir):]
	}

	return track.Track{
		Header: track.Header{
			FileSize:    uint32(t.FileSize),
			TrackNumber: uint32(t.TrackNumber),
			Year:        yearOrZero(t.ReleaseDate),
			Duration:    uint16(t.Duration.Seconds()),
			Bitrate:     uint32(t.Bitrate),
			Tempo:       uint32(t.Tempo * 100),
			Id:          uint32(lib.Tracks().ID(t)),
			ArtistId:    uint32(lib.Artists().ID(lib.Artist(t.Artist))),
			AlbumId:     uint32(lib.Albums().ID(lib.Album(t.Album))),
			SampleDepth: uint16(t.SampleDepth),
			SampleRate:  uint32(t.SampleRate),
			FileType:    track.FileTypeMP3,
		},
		AnalyzeDate: time.Now().Format(isoDateFormat),
		FilePath:    filePath,
		DateAdded:   t.AddedDate.Format(isoDateFormat),
		Filename:    filepath.Base(t.Path),
		Title:       t.Title,
		// AnalyzePath: "/PIONEER/USBANLZ/P016/0000875E/ANLZ0000.DAT",
	}
}

func PdbArtist(lib *library.Library, a *library.Artist) artist.Artist {
	return artist.Artist{
		Id:   uint32(lib.Artists().ID(a)),
		Name: a.Name,
	}
}

func PdbAlbum(lib *library.Library, a *library.Album) album.Album {
	return album.Album{
		Id:       uint32(lib.Albums().ID(a)),
		ArtistId: 0, // FIXME: multiple artist albums with ID 0, otherwise get ref?
		Name:     a.Title,
	}
}

func FileTypeFromString(t string) (track.FileType, error) {
	switch t {
	case "mp3":
		return track.FileTypeMP3, nil
	case "aac":
		return track.FileTypeM4A, nil
	case "wav":
		return track.FileTypeWAV, nil
	case "flac":
		return track.FileTypeFLAC, nil
	default:
		return track.FileTypeUnknown, fmt.Errorf("unimplemented file format '%s'", t)
	}
}
