package encoding

import (
	"bytes"
	"encoding/gob"
	"errors"
)

// EncodeGOB serializes an object to binary format using encoding/gob.
func EncodeGOB(input any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(input); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeGOB deserializes binary gob data into the provided output object.
// The out parameter should be a pointer to the target struct.
func DecodeGOB(data []byte, out any) error {
	if len(data) == 0 {
		return errors.New("input data is empty")
	}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(out)
}
