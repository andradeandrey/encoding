package bencode

import (
	"encoding"
	"errors"
)

// RawMessage is a raw encoded bencode object.
// It is intedended to delay decoding or precomute an encoding.
type RawMessage []byte

// MarshalText returns *m as the bencode encoding of m.
func (m *RawMessage) MarshalText() ([]byte, error) {
	return *m, nil
}

// UnmarshalText sets *m to a copy of data.
func (m *RawMessage) UnmarshalText(text []byte) error {
	if m == nil {
		return errors.New("bencode.RawMessage: UnmarshalText on nil pointer")
	}
	*m = append((*m)[0:0], text...)
	return nil
}

var _ encoding.TextMarshaler = (*RawMessage)(nil)
var _ encoding.TextUnmarshaler = (*RawMessage)(nil)
