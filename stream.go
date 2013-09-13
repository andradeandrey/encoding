// Copyright (c) 2013, Emery Hemingway.
// See the LICENSE file for terms and conditions.

package ebml

import (
	"io"
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

// Encode writes the EBML binary encoding of v to the stream.
func (enc *Encoder) Encode(v interface{}) error {
	if enc.err != nil {
		return enc.err
	}
	e := &encodeState{w: enc.w}
	err := e.marshal(v)
	if err != nil {
		enc.err = err
		return err
	}
	return nil
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
// can marshal themselves into an EBML stream. If the Marshaler
// is not a container it should not contain and id and size
// header. N will be the size used to compute the size of the
// element that will contain marshaler, and only n bytes will
// be read from r.
type Marshaler interface {
	MarshalEBML() (r io.Reader, n int64)
}

// Unmarshaler is the interface implemented by objects that
// can unmarshal themselves from an EBML stream. The data
// written to W will contain the data for the element being
//  unmarshaled, and not an id or size header. n shall be
// the size of the data at w.
type Unmarshaler interface {
	UnmarshalEBML(n int64) (w io.Writer)
}

// MarshalUnmarshaler is an interface that
// combines both the Marshaler and Unmarshaler.
type MarshalUnmarshaler interface {
	Marshaler
	Unmarshaler
}
