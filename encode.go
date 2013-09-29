// Copyright © 2013 Emery Hemingway

package bencode

import (
	"bytes"
	"io"
	"reflect"
)

// Encoder bencodes data to a Reader.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new Encoder that bencodes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode bencodes data to the wrapped stream.
func (e *Encoder) Encode(x interface{}) (err error) {
	return writeValue(e.w, reflect.ValueOf(x))
}

// Marshal returns a bencoded form of x.
//
// Marshal traverses the value x recursively.
//
// Marshal uses the following type-dependent encodings:
//
// Floating point, integer, and Number values encode as bencode numbers.
//
// String values encode as bencode strings.
//
// Array and slice values encode as bencode arrays.
//
// Struct values encode as bencode maps. Each exported struct field
// becomes a member of the object.
// The object's default key string is the struct field name
// but can be specified in the struct field's tag value. The text of
// the struct field's tag value is the key name. Examples:
//
//   // Field appears in bencode as key "Field".
//   Field int
//
//   // Field appears in bencode as key "myName".
//   Field int `bencode:"myName"`
//
// Anonymous struct fields are ignored.
//
// Map values encode as bencode objects.
// The map's key type must be string; the object keys are used directly
// as map keys.
//
// Boolean, Pointer, Interface, Channel, complex, and function values cannot
// be encoded in bencode.
// Attempting to encode such a value causes Marshal to return
// a MarshalError.
//
// Bencode cannot represent cyclic data structures and Marshal does not
// handle them.  Passing cyclic structures to Marshal will result in
// an infinite recursion.
func Marshal(x interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := NewEncoder(buf)
	err := encoder.Encode(x)
	return buf.Bytes(), err
}
