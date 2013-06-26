// Copyright (c) 2013, Emery Hemingway. All rights reserved.
// Actully most of it comes from encoding/json and encoding/gob,
// courtesy of The Go Authors

package ebml

import (
	"io"
)

type Decoder struct {
	r   io.ReadSeeker
	err error
}

// NewDecoder returns as new decoder that reads from r.
func NewDecoder(r io.ReadSeeker) *Decoder {
	return &Decoder{r: r}
}

// Decode reads EBML elements into from the intov.
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
