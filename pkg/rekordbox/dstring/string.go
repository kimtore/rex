package dstring

import (
	`bytes`
	`encoding`
	`encoding/binary`
	`fmt`
	`io`
	`unicode`

	`github.com/ambientsound/rex/pkg/marshal`
	unicode_enc "golang.org/x/text/encoding/unicode"
)

/**
 * Header of DeviceSQL long strings.
 */
type StringHeader struct {
	Encoding StringEncoding
	Length   uint16
	Padding  uint8
}

/*
 * From https://djl-analysis.deepsymmetry.org/rekordbox-export-analysis/exports.html#devicesql-strings:
 *
 * The details of this analysis are somewhat speculative because the only bit patterns we have seen in practice
 * when S is zero are 0b01000000 for long-form ascii strings and 0b10010000 for long-form utf16le strings
 * (Rekordbox probably just does not use the other supported string formats).
 */
type StringEncoding uint8

const (
	StringEncodingShortAscii  StringEncoding = 0b00000001 // note: 7 leftmost bits are length
	StringEncodingLongAscii   StringEncoding = 0b01000000
	StringEncodingLongUTF16LE StringEncoding = 0b10010000
)

type ShortAsciiString string

type LongAsciiString string

type UnicodeString string

type IsrcString string

func New(s string) encoding.BinaryMarshaler {
	if isASCII(s) {
		if len(s) < 127 {
			return ShortAsciiString(s)
		}
		return LongAsciiString(s)
	}
	return UnicodeString(s)
}

func UnmarshalBinary(data []byte) (string, error) {
	var err error

	r := bytes.NewReader(data)
	header := &StringHeader{}

	err = marshal.UnpackFrom(r, &header.Encoding)
	if err != nil {
		return "", err
	}

	// Short-circuit short ascii strings, they do not contain the whole header.
	if header.Encoding&StringEncodingShortAscii == StringEncodingShortAscii {
		header.Length = uint16(header.Encoding>>1) - 1
		header.Encoding = StringEncodingShortAscii
	} else {
		err = r.UnreadByte()
		if err != nil {
			return "", err
		}
		err = marshal.UnpackFrom(r, header)
		if err != nil {
			return "", err
		}
		header.Length -= 4
	}

	// Try to read the string data.
	out := make([]byte, header.Length)
	_, err = io.ReadFull(r, out)
	if err != nil {
		return "", err
	}

	switch header.Encoding {
	case StringEncodingShortAscii:
		return string(out), nil
	case StringEncodingLongAscii:
		return string(out), nil
	case StringEncodingLongUTF16LE:
		decoder := unicode_enc.UTF16(unicode_enc.LittleEndian, unicode_enc.IgnoreBOM).NewDecoder()
		out, err = decoder.Bytes(out)
		return string(out), err
	default:
		return "", fmt.Errorf("string encoding %x not known", header.Encoding)
	}
}

/*
 * The flag byte described above is labeled lk (lengthAndKind) below.
 * If S (the low-order bit of lk) is set, it means the string field holds a short ASCII string.
 * The length of such a field can be extracted by right-shifting lk once (or, equivalently, dividing it by two).
 * This length is for the entire string field, including lk itself, so the maximum length of actual string data is 126 bytes.
 */
func (s ShortAsciiString) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}
	strlen := uint8(len(s)+1) << 1
	strlen = strlen | uint8(StringEncodingShortAscii)
	err := binary.Write(buf, binary.LittleEndian, strlen)
	if err != nil {
		return nil, err
	}
	_, err = buf.WriteString(string(s))
	return buf.Bytes(), err
}

func (s LongAsciiString) MarshalBinary() ([]byte, error) {
	var err error
	buf := &bytes.Buffer{}

	err = binary.Write(buf, binary.LittleEndian, StringHeader{
		Encoding: StringEncodingLongAscii,
		Length:   uint16(len(s) + 4),
	})
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString(string(s))

	return buf.Bytes(), err
}

func (s UnicodeString) MarshalBinary() ([]byte, error) {
	var err error
	var out string
	buf := &bytes.Buffer{}

	encoder := unicode_enc.UTF16(unicode_enc.LittleEndian, unicode_enc.IgnoreBOM).NewEncoder()
	out, err = encoder.String(string(s))
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, StringHeader{
		Encoding: StringEncodingLongUTF16LE,
		Length:   uint16(len(out) + 4),
	})
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString(out)

	return buf.Bytes(), err
}

// ISRC Strings
//
// When an International Standard Recording Code is present as the first string pointer in a track row,
// it is marked with kind 90 but does not actually hold a UTF-16-LE string. Instead, the first byte after
// the pad value following the length is the value 03 and then there are (len-7)
// bytes of ASCII, followed by a null byte. Crate Digger does not yet attempt to cope with this.
func (s IsrcString) MarshalBinary() ([]byte, error) {
	var err error
	buf := &bytes.Buffer{}

	err = binary.Write(buf, binary.LittleEndian, StringHeader{
		Encoding: StringEncodingLongUTF16LE,
		Length:   uint16(len(s) + 6),
	})
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, uint8(0x03))
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString(string(s))
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, uint8(0x00))
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func isASCII(s string) bool {
	for _, c := range s {
		if c > unicode.MaxASCII {
			return false
		}
	}

	return true
}
