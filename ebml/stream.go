// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"runtime"
)

// An Encoder writes EBML data to an output stream.
type Encoder struct {
	w   io.Writer
	err error
}

// NewEncoder returns a new Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the EBML binary encoding of element to an Encoder stream.
func (enc *Encoder) Encode(element interface{}) (err error) {
	if enc.err != nil {
		return enc.err
	}

	// encoding doesn't use error internally, but panics if there is a
	// problem and then unwinds up to here.
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	v := reflect.ValueOf(element)
	id := getId(v)

	elem := encode(id, v)
	if elem != nil {
		_, err = elem.WriteTo(enc.w)
	}
	return
}

// A Decoder decodes EBML data from a ReadSeeker
type Decoder struct {
	r   io.ReadSeeker
	buf []byte // this is a resuable buffer for decoding
	err error
}

// NewDecoder returns a new decoder that decodes from r.
func NewDecoder(r io.ReadSeeker) *Decoder {
	// TODO(Emery): just ask for a io.Reader and try and make it
	// an io.ReadSeeker
	return &Decoder{r: r, buf: make([]byte, 8)}
}

// Decode decodes a EBML stream into v.
func (d *Decoder) Decode(element interface{}) (err error) {
	if d.err != nil {
		return d.err
	}

	// decoding doesn't use error internally, but panics if there is a
	// problem and then unwinds up to here.
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	v := reflect.ValueOf(element)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	id := getId(v)
	if n, curId := d.readId(); id != curId {
		d.r.Seek(int64(-n), 0)
		return fmt.Errorf("ebml: read stream positioned at element %s not %s", curId, id)
	}
	_, size := d.readSize()

	decodeValue(d, id, size, v)
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

// Unmarshal unmarshals EBML data into element.
//
// Unmarshal first determines the Id of element from the field named 'EbmlId',
// then recursively traverses element. Any exported struct field of element
// with an `ebml` tag will be including in unmarshalling, with the exception
// of fields tagged with `ebml:"-"`.
//
// The ebml tag should contain a valid EBML id, see the EBML documention for
// what constitutes a valid id.
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
type Marshaler interface {
	// BUG(Emery): an embedded Marshaler will trample on a struct
	MarshalEBML() (size int64, wt io.WriterTo)
}

// Unmarshaler is the interface implemented by objects that
// can unmarshal themselves from an EBML stream. The data
// read into ReaderFrom will contain the data for the element
// being unmarshaled, and not an id or size header.
//
// n shall be the size of the element data, and it is not the
// resposibility of an Unmarshaler to limit reading to n.
//
// An Unmarshaler is usually sent to the decoding engine as a nil pointer
// in a struct and created when a tagged element is encountered, for this
// reason the UnmarshalEBML method should behave as if the Unmarshaler is
// at a zero value state.
type Unmarshaler interface {
	// BUG(Emery): an embedded Unmarshaler will trample on a struct
	UnmarshalEBML(n int64) io.ReaderFrom
}
