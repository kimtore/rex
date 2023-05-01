package track

import (
	`bytes`
	`encoding`
	`encoding/binary`
	`fmt`
	`io`

	`github.com/ambientsound/rex/pkg/marshal`
	`github.com/ambientsound/rex/pkg/rekordbox/dstring`
)

/**
 * A row that describes a track that can be played, with many
 * details about the music, and links to other tables like artists,
 * albums, keys, etc.
 */
type Track struct {
	Header
	StringOffsets
	UnknownString8  string
	UnknownString6  string
	AnalyzeDate     string
	FilePath        string
	AutoloadHotcues string // ON = true, "" = false
	DateAdded       string
	PhraseAnalyzed  string // 1 = false, 2 = true
	Composer        string
	KuvoPublic      string
	MixName         string
	UnknownString5  string
	UnknownString4  string
	Message         string
	KeyAnalyzed     string // 1 = false, 2 = true
	Isrc            string
	UnknownString7  string
	Filename        string
	AnalyzePath     string
	Comment         string
	ReleaseDate     string
	Title           string
}

type FileType uint16

const (
	FileTypeUnknown FileType = 0
	FileTypeMP3     FileType = 0x1
	FileTypeM4A     FileType = 0x4
	FileTypeFLAC    FileType = 0x5
	FileTypeWAV     FileType = 0xb
)

// All numerical values go here.
type Header struct {
	Unnamed0         uint16 // Always 0x24. Identifies this as a track row?
	IndexShift       uint16 // Starts at zero and increases by 0x20 every row in a RowSet.
	Bitmask          uint32
	SampleRate       uint32
	ComposerId       uint32
	FileSize         uint32
	Checksum         uint32 // There seems to be uniform distribution of values, so I dubbed this field "checksum". 28 bits???
	Unnamed7         uint16
	Unnamed8         uint16
	ArtworkId        uint32
	KeyId            uint32
	OriginalArtistId uint32
	LabelId          uint32
	RemixerId        uint32
	Bitrate          uint32
	TrackNumber      uint32
	Tempo            uint32
	GenreId          uint32
	AlbumId          uint32
	ArtistId         uint32
	Id               uint32
	DiscNumber       uint16
	PlayCount        uint16
	Year             uint16
	SampleDepth      uint16
	Duration         uint16
	Unnamed26        uint16
	ColorId          uint8
	Rating           uint8
	FileType         FileType
	// If this value is zero, Rekordbox crashes, leaving behind a "brokendb" file with 1 byte length.
	// Set it to 0x3.
	Unnamed30 uint16
}

// Order matters
type StringOffsets struct {
	Isrc            uint16
	Composer        uint16
	Num1            uint16
	Num2            uint16
	UnknownString4  uint16
	Message         uint16
	KuvoPublic      uint16
	AutoloadHotcues uint16
	UnknownString5  uint16
	UnknownString6  uint16
	DateAdded       uint16
	ReleaseDate     uint16
	MixName         uint16
	UnknownString7  uint16
	AnalyzePath     uint16
	AnalyzeDate     uint16
	Comment         uint16
	Title           uint16
	UnknownString8  uint16
	Filename        uint16
	FilePath        uint16
}

func (t *Track) WriteTSV(w io.Writer) {
	io.WriteString(w, fmt.Sprintf("%04x\t", t.Header.Unnamed0))
	io.WriteString(w, fmt.Sprintf("%04x\t", t.Header.Bitmask))
	io.WriteString(w, fmt.Sprintf("%04x\t", t.Header.Checksum))
	io.WriteString(w, fmt.Sprintf("%04x\t", t.Header.Unnamed7))
	io.WriteString(w, fmt.Sprintf("%04x\t", t.Header.Unnamed8))
	io.WriteString(w, fmt.Sprintf("%04x\t", t.Header.Unnamed26))
	io.WriteString(w, fmt.Sprintf("%04x\t", t.Header.FileType))
	io.WriteString(w, fmt.Sprintf("%04x\t", t.Header.Unnamed30))

	io.WriteString(w, t.KeyAnalyzed+"\t")
	io.WriteString(w, t.PhraseAnalyzed+"\t")
	io.WriteString(w, t.KuvoPublic+"\t")
	io.WriteString(w, t.AutoloadHotcues+"\t")
	io.WriteString(w, t.UnknownString4+"\t")
	io.WriteString(w, t.UnknownString5+"\t")
	io.WriteString(w, t.UnknownString6+"\t")
	io.WriteString(w, t.UnknownString7+"\t")
	io.WriteString(w, t.UnknownString8+"\t")
	io.WriteString(w, t.DateAdded+"\t")
	io.WriteString(w, t.ReleaseDate+"\t")
	io.WriteString(w, t.MixName+"\t")
	io.WriteString(w, t.AnalyzePath+"\t")
	io.WriteString(w, t.AnalyzeDate+"\t")
	io.WriteString(w, t.FilePath+"\t")
	io.WriteString(w, "\n")
}

// Unmarshal a single track row.
func (t *Track) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	err := marshal.UnpackFrom(buf, &t.Header)
	if err != nil {
		return err
	}
	err = marshal.UnpackFrom(buf, &t.StringOffsets)
	if err != nil {
		return err
	}

	t.FilePath, err = dstring.UnmarshalBinary(data[t.StringOffsets.FilePath:])
	if err != nil {
		return err
	}
	t.Title, err = dstring.UnmarshalBinary(data[t.StringOffsets.Title:])
	if err != nil {
		return err
	}

	load := func(dst *string, offset uint16) {
		if err != nil {
			return
		}
		if int(offset) >= len(data) {
			err = fmt.Errorf("heap pointer is past dataset")
		}
		*dst, err = dstring.UnmarshalBinary(data[offset:])
	}

	load(&t.Composer, t.StringOffsets.Composer)
	load(&t.KeyAnalyzed, t.StringOffsets.Num1)
	load(&t.PhraseAnalyzed, t.StringOffsets.Num2)
	load(&t.UnknownString4, t.StringOffsets.UnknownString4)
	load(&t.Message, t.StringOffsets.Message)
	load(&t.KuvoPublic, t.StringOffsets.KuvoPublic)
	load(&t.AutoloadHotcues, t.StringOffsets.AutoloadHotcues)
	load(&t.UnknownString5, t.StringOffsets.UnknownString5)
	load(&t.UnknownString6, t.StringOffsets.UnknownString6)
	load(&t.DateAdded, t.StringOffsets.DateAdded)
	load(&t.ReleaseDate, t.StringOffsets.ReleaseDate)
	load(&t.MixName, t.StringOffsets.MixName)
	load(&t.UnknownString7, t.StringOffsets.UnknownString7)
	load(&t.AnalyzePath, t.StringOffsets.AnalyzePath)
	load(&t.AnalyzeDate, t.StringOffsets.AnalyzeDate)
	load(&t.Comment, t.StringOffsets.Comment)
	load(&t.Title, t.StringOffsets.Title)
	load(&t.UnknownString8, t.StringOffsets.UnknownString8)
	load(&t.Filename, t.StringOffsets.Filename)
	load(&t.FilePath, t.StringOffsets.FilePath)

	return err
}

func (t *Track) MarshalBinary() ([]byte, error) {
	var err error
	buf := &bytes.Buffer{}
	stringHeap := &bytes.Buffer{}

	// Not figured out yet
	t.Header.Unnamed0 = 0x24 // Always this value
	t.Header.Bitmask = 0xC0700

	// t.Header.Checksum = 0x0ef38622
	t.Header.Unnamed26 = 0x29 // Always this value

	// Omitting this value will cause Rekordbox to crash, and leave a "brokendb" file behind.
	// It seems to be present in all rows.
	t.Header.Unnamed30 = 0x3

	t.KeyAnalyzed = "1"
	t.PhraseAnalyzed = "1"
	// When nums are 1, Unnamed7-8 are 0x526d, 0x5fb6.
	// When nums are 2, Unnamed7-8 are 0xb8eb, 0x575b.
	// But also we see                 0x758a, 0x57a2  for both 1 and 2!!!
	t.Header.Unnamed7 = 0x758a
	t.Header.Unnamed8 = 0x57a2

	// Usually ON
	t.AutoloadHotcues = "ON"

	err = binary.Write(buf, binary.LittleEndian, t.Header)
	if err != nil {
		return nil, err
	}
	recordLen := uint16(buf.Len() + 42)

	// Writes an encoded string to the string heap
	// and returns the data's relative position on the string heap.
	write := func(input encoding.BinaryMarshaler) uint16 {
		ret := uint16(stringHeap.Len())
		data, e := input.MarshalBinary()
		if e != nil {
			panic(e)
		}
		_, err = stringHeap.Write(data)
		if e != nil {
			panic(e)
		}
		return ret
	}

	t.StringOffsets.Isrc = write(dstring.IsrcString(t.Isrc)) + recordLen
	t.StringOffsets.Composer = write(dstring.New(t.Composer)) + recordLen
	t.StringOffsets.Num1 = write(dstring.New(t.KeyAnalyzed)) + recordLen
	t.StringOffsets.Num2 = write(dstring.New(t.PhraseAnalyzed)) + recordLen
	t.StringOffsets.UnknownString4 = write(dstring.New(t.UnknownString4)) + recordLen
	t.StringOffsets.Message = write(dstring.New(t.Message)) + recordLen
	t.StringOffsets.KuvoPublic = write(dstring.New(t.KuvoPublic)) + recordLen
	t.StringOffsets.AutoloadHotcues = write(dstring.New(t.AutoloadHotcues)) + recordLen
	t.StringOffsets.UnknownString5 = write(dstring.New(t.UnknownString5)) + recordLen
	t.StringOffsets.UnknownString6 = write(dstring.New(t.UnknownString6)) + recordLen
	t.StringOffsets.DateAdded = write(dstring.New(t.DateAdded)) + recordLen
	t.StringOffsets.ReleaseDate = write(dstring.New(t.ReleaseDate)) + recordLen
	t.StringOffsets.MixName = write(dstring.New(t.MixName)) + recordLen
	t.StringOffsets.UnknownString7 = write(dstring.New(t.UnknownString7)) + recordLen
	t.StringOffsets.AnalyzePath = write(dstring.New(t.AnalyzePath)) + recordLen
	t.StringOffsets.AnalyzeDate = write(dstring.New(t.AnalyzeDate)) + recordLen
	t.StringOffsets.Comment = write(dstring.New(t.Comment)) + recordLen

	// for some reason, writing three null bytes directly into the string heap allows the test to complete.
	// perhaps the database engine is trying to align fields on byte boundaries?
	// this change resulted in Title being written to offset 0x90.
	// stringHeap.Write([]byte{0, 0, 0})

	t.StringOffsets.Title = write(dstring.New(t.Title)) + recordLen
	t.StringOffsets.UnknownString8 = write(dstring.New(t.UnknownString8)) + recordLen
	t.StringOffsets.Filename = write(dstring.New(t.Filename)) + recordLen
	t.StringOffsets.FilePath = write(dstring.New(t.FilePath)) + recordLen

	err = binary.Write(buf, binary.LittleEndian, t.StringOffsets)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(buf, stringHeap)

	return buf.Bytes(), err
}

func (t *Track) SetIndexShift(shift uint16) {
	t.IndexShift = shift
}
