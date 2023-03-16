package marshal

// Misc marshalling shortcuts, useful for writing binary formats.

import (
	`bytes`
	`encoding`
	`encoding/binary`
	`io`

	`github.com/lunixbochs/struc`
)

// Write binary data from a marshaler into a writer.
func Into(dst io.Writer, src encoding.BinaryMarshaler) error {
	data, err := src.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = dst.Write(data)
	return err
}

// Shorthand for little-endian packing into a byte slice.
func Pack(data any) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := struc.PackWithOptions(buf, data, &struc.Options{
		Order: binary.LittleEndian,
	})
	return buf.Bytes(), err
}

// Shorthand for little-endian unpacking from a byte slice.
func Unpack(destination any, data []byte) error {
	buf := bytes.NewReader(data)
	return struc.UnpackWithOptions(buf, destination, &struc.Options{
		Order: binary.LittleEndian,
	})
}

// Shorthand for little-endian packing into a writer.
func PackInto(w io.Writer, data any) error {
	return struc.PackWithOptions(w, data, &struc.Options{
		Order: binary.LittleEndian,
	})
}

// Consume bytes from io.Reader `r` to fill struct `data`.
func UnpackFrom(r io.Reader, destination any) error {
	return struc.UnpackWithOptions(r, destination, &struc.Options{
		Order: binary.LittleEndian,
	})
}
