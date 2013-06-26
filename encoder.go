// Copyright (c) 2013, Emery Hemingway. All rights reserved.
// Actully most of it comes from encoding/json, courtesy of
// The Go Authors

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
