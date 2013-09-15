// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml

import (
	"bytes"
	"io"
	"reflect"
)

// An Encoder writes EBML data to an output stream.
type Encoder struct {
	w   io.Writer
	err error
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the EBML binary encoding of v to an Encoder stream.
func (enc *Encoder) Encode(element interface{}) (err error) {
	if enc.err != nil {
		return enc.err
	}

	v := reflect.ValueOf(element)
	id := getId(v)

	e, err := marshal(id, v)
	if err == nil {
		_, err = e.WriteTo(enc.w)
	}
	return
}

// Marshal returns an EBML representation of element.
//
// Marshal first determines the Id of element from the field named 'EbmlId',
// then recursively traverses element. Any exported struct field of element
// with an `ebml` tag will be including in marshalling, with the exception
// of fields tagged with `ebml:"-"`.
//
// The ebml tag should contain a valid EBML id, see the EBML documention for
// what constitutes a valid id.
func Marshal(element interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := NewEncoder(buf)
	err := encoder.Encode(element)
	return buf.Bytes(), err
}

// BUG(Emery): no documentation here
func Unmarshal(data []byte, element interface{}) error {
	return NewDecoder(bytes.NewReader(data)).Decode(element)
}

// Marshaler is the interface implemented by objects that
// can marshal themselves into an EBML stream. The WriterTo
// should only Write element data, and not the id and size
// header of the element. size is the length of the data that
// shall be written and is used to build the element header
// and compute the size of the parent element before it is
// writen to an EBML stream.
//
// If a struct both implements Marshaler and contains ebml
// tagged fields, the fields will be ignored. This implies
// if a Marshaler is an embedded field, the parent struct
// will inherit it's interface, and the marshaler will take
// the place of the parent in the encoder.
type Marshaler interface {
	// BUG(Emery): an embedded Marshaler will trample on a struct
	MarshalEBML() (wt io.WriterTo, size int64)
}

// Unmarshaler is the interface implemented by objects that
// can unmarshal themselves from an EBML stream. The data
// read into ReaderFrom will contain the data for the element
// being unmarshaled, and not an id or size header. n shall be
// the size of the element data, but an Unmarshaler does not
// need use this number to limit the length of data that is read.
//
// If a struct both implements Unmarshaler and contains ebml
// tagged fields, the fields will be ignored. This implies
// that if an Unmarshaler is an embedded field, the parent
// struct will inherit it's interface, and the marshaler will
// take the place of the parent in the decoder.
type Unmarshaler interface {
	// BUG(Emery): an embedded Unmarshaler will trample on a struct
	UnmarshalEBML(n int64) io.ReaderFrom
}

// MarshalUnmarshaler is an interface that
// combines both the Marshaler and Unmarshaler.
type MarshalUnmarshaler interface {
	Marshaler
	Unmarshaler
}
