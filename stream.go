// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml

import (
	"bytes"
	"errors"
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

	var id Id
	t := reflect.TypeOf(element)
	if f, ok := t.FieldByName("EbmlId"); ok {
		id = v.FieldByIndex(f.Index).Interface().(Id)
		if id == nil {
			id, err = NewIdFromString(f.Tag.Get("ebml"))
			if err != nil {
				return
			}
		}
	} else {
		return errors.New("cannot resolve EBML Id for " + t.Name())
	}

	E, err := marshal(id, v)
	if err != nil {
		return err
	}

	_, err = io.Copy(enc.w, E)
	if err != nil {
		return err
	}
	return nil
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

// A decoder reads data from an EBML stream.
type Decoder struct {
	r   io.ReadSeeker
	err error
}

// NewDecoder returns as new decoder that reads from r.
func NewDecoder(r io.ReadSeeker) *Decoder {
	return &Decoder{r: r}
}

// Decode decodes EBML elements into v, where v
// is a struct containing tagged fields.
// See ebml.Header for an example.
func (dec *Decoder) Decode(v interface{}) error {
	if dec.err != nil {
		return dec.err
	}
	d := decodeState{r: dec.r}
	err := d.unmarshal(v)
	if err != nil {
		dec.err = err
		return err
	}
	return nil
}

// Marshaler is the interface implemented by objects that
// can marshal themselves into an EBML stream. r should only
// contain element data, and not the id and size header of
// the Marshaler element. n is the length of the data in r
// and will be used to compute the size of the element above
// Marshaler. Only n bytes will be read from r.
//
// If a struct both implements Marshaler and contains ebml
// tagged fields, the fields will be ignored.
type Marshaler interface {
	MarshalEBML() (r io.Reader, n int64)
}

// Unmarshaler is the interface implemented by objects that
// can unmarshal themselves from an EBML stream. The data
// written to W will contain the data for the element being
// unmarshaled, and not an id or size header. n shall be the
// size of the data at w.
//
// If a struct both implements Unmarshaler and contains ebml
// tagged fields, the fields will be ignored.
type Unmarshaler interface {
	UnmarshalEBML(n int64) (w io.Writer)
}

// MarshalUnmarshaler is an interface that
// combines both the Marshaler and Unmarshaler.
type MarshalUnmarshaler interface {
	Marshaler
	Unmarshaler
}
