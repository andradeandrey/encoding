// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Represents bencode data structure using native Go types: booleans, floats,
// strings, slices, and maps.

package bencode

import (
	"bytes"
	"io"
)

// A Decoder decodes bencoded data from a stream.
type Decoder struct {
	r io.Reader
}

// NewDecoder returns a new decoder that decodes from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r}
}

// Decode decodes data from the wrapped stream into x.
func (d *Decoder) Decode(x interface{}) error {
	return unmarshalValue(d.r, x)
}

type decoder struct {
	// A value being constructed.
	value interface{}
	// Container entity to flush into.  Can be either []interface{} or
	// map[string]interface{}.
	container interface{}
	// The index into the container interface.  Either int or string.
	index interface{}
}

func newDecoder(container interface{}, key interface{}) *decoder {
	return &decoder{container: container, index: key}
}

func (j *decoder) Int64(i int64) { j.value = int64(i) }

func (j *decoder) Uint64(i uint64) { j.value = uint64(i) }

func (j *decoder) Float64(f float64) { j.value = float64(f) }

func (j *decoder) String(s string) { j.value = s }

func (j *decoder) Bool(b bool) { j.value = b }

func (j *decoder) Null() { j.value = nil }

func (j *decoder) Array() { j.value = make([]interface{}, 0, 8) }

func (j *decoder) Map() { j.value = make(map[string]interface{}) }

func (j *decoder) Elem(i int) builder {
	v, ok := j.value.([]interface{})
	if !ok {
		v = make([]interface{}, 0, 8)
		j.value = v
	}
	/* XXX There is a bug in here somewhere, but append() works fine.
	lens := len(v)
	if cap(v) <= lens {
		news := make([]interface{}, 0, lens*2)
		copy(news, j.value.([]interface{}))
		v = news
	}
	v = v[0 : lens+1]
	*/
	v = append(v, nil)
	j.value = v
	return newDecoder(v, i)
}

func (j *decoder) Key(s string) builder {
	m, ok := j.value.(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		j.value = m
	}
	return newDecoder(m, s)
}

func (j *decoder) Flush() {
	switch c := j.container.(type) {
	case []interface{}:
		index := j.index.(int)
		c[index] = j.Copy()
	case map[string]interface{}:
		index := j.index.(string)
		c[index] = j.Copy()
	}
}

// Get the value built by this builder.
func (j *decoder) Copy() interface{} {
	return j.value
}

// Unmarshal reads and parses the bencode syntax data from r and fills in
// an arbitrary struct or slice pointed at by val.
// It uses the reflect package to assign to fields
// and arrays embedded in val.  Well-formed data that does not fit
// into the struct is discarded.
//
// For example, given these definitions:
//
//	type Email struct {
//		Where string;
//		Addr string;
//	}
//
//	type Result struct {
//		Name string;
//		Phone string;
//		Email []Email
//	}
//
//	var r = new(Result)
//
// unmarshalling the bencode syntax string
//
//	"d5:emailld5:where4:home4:addr15:gre@example.come\
//  d5:where4:work4:addr12:gre@work.comee4:name14:Gr\
//  ace R. Emlin7:address15:123 Main Streete"
//
// via Unmarshal(s, r) is equivalent to assigning
//
//	r = Result{
//		"Grace R. Emlin",	// name
//		"phone",		// no phone given
//		[]Email{
//			Email{ "home", "gre@example.com" },
//			Email{ "work", "gre@work.com" }
//		}
//	}
//
// Note that the field r.Phone has not been modified and
// that the bencode field "address" was discarded.
//
// Because Unmarshal uses the reflect package, it can only
// assign to upper case fields.  Unmarshal uses a case-insensitive
// comparison to match bencode field names to struct field names.
//
// If you provide a "bencode" key in the tag string for a struct member, the
// tag string will be used as the bencode dictionary key for that member. If
// the field tag is "-" the field is ignored.
//
// To unmarshal a top-level bencode array, pass in a pointer to an empty
// slice of the correct type.
//
func Unmarshal(data []byte, x interface{}) error {
	return unmarshalValue(bytes.NewReader(data), x)
}
