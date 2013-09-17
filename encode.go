// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml

import (
	"fmt"
	"io"
	"reflect"
)

var indent string

type encoder interface {
	io.WriterTo
	Size() int64
}

type containerElement struct {
	// These []bytes can probably be merged
	id       Id
	size     int64
	header   []byte
	elements []encoder
}

func (ce *containerElement) Append(e encoder) {
	ce.elements = append(ce.elements, e)
	ce.size += e.Size()
}

func (ce *containerElement) Size() (n int64) {
	ce.header = append(ce.id.Bytes(), marshalSize(ce.size)...)
	return int64(len(ce.header)) + ce.size
}

func (ce *containerElement) WriteTo(w io.Writer) (n int64, err error) {
	if len(ce.header) == 0 {
		ce.Size() // top level elements don't get Size called
	}
	var N int
	var nn int64
	N, err = w.Write(ce.header)
	if err != nil {
		return
	}
	n += int64(N)

	for _, e := range ce.elements {
		nn, err = e.WriteTo(w)
		n += nn
		if err != nil {
			break
		}
	}
	return
}

type simpleElement []byte

func (b simpleElement) Size() int64 { return int64(len(b)) }
func (b simpleElement) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(b)
	return int64(n), err
}

// sliceElement is a encoder for holding groups of elements, such as a group
// elements of the same type that occur more than once in a container.
type sliceElement []encoder

func (se sliceElement) Size() (n int64) {
	for _, e := range se {
		n += e.Size()
	}
	return
}

func (se sliceElement) WriteTo(w io.Writer) (n int64, err error) {
	var nn int64
	for _, e := range se {
		nn, err = e.WriteTo(w)
		n += nn
		if err != nil {
			break
		}
	}
	return
}

type marshalerElement struct {
	id     Id
	size   int64
	header []byte
	wt     io.WriterTo
}

func (me *marshalerElement) Size() int64 { return me.size }

func (me *marshalerElement) WriteTo(w io.Writer) (n int64, err error) {
	var N int
	var nn int64
	N, err = w.Write(me.header)
	n = int64(N)
	if err != nil {
		return
	}
	nn, err = me.wt.WriteTo(w)
	n += nn
	return
}

func unsupportedTypeError(typ reflect.Type) {
	encError("unsupported type: " + typ.String())
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func newEncoder(id Id, v reflect.Value) encoder {
	if m, ok := v.Interface().(Marshaler); ok {
		wt, size := m.MarshalEBML()
		header := append(id.Bytes(), marshalSize(size)...)
		return &marshalerElement{id, size, header, wt}
	}

	switch v.Kind() {
	case reflect.Struct:
		return newStructEncoder(id, v)

	case reflect.Slice:
		if v.IsNil() || v.Len() == 0 {
			return nil
		}
		s := make(sliceElement, v.Len())
		for i := 0; i < v.Len(); i++ {
			s[i] = newEncoder(id, v.Index(i))
		}
		return s

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x := v.Int()
		return marshalInt(id, x)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x := v.Uint()
		return marshalUint(id, x)

	case reflect.String:
		return marshalString(id, v.String())

	case reflect.Interface, reflect.Ptr:
		return newEncoder(id, v.Elem())
	}
	unsupportedTypeError(v.Type())
	return nil
}

const (
	o1 = 1<<7 - 2
	o2 = 1<<14 - 2
	o3 = 1<<21 - 2
	o4 = 1<<28 - 2
	o5 = 1<<35 - 2
	o6 = 1<<42 - 2
	o7 = 1<<49 - 2
	o8 = 1<<56 - 2
)

// MarshalSize returns the EBML variable width representation
// of an element's size
func marshalSize(x int64) []byte {
	var s int
	var m byte

	switch {
	case x == 0:
		return []byte{byte(0x80)}
	case x < o1:
		return []byte{byte(0x80) | byte(x)}
	case x < o2:
		s = 2
		m = 0x40
	case x < o3:
		s = 3
		m = 0x20
	case x < o4:
		s = 4
		m = 0x10
	case x < o5:
		s = 5
		m = 0x08
	case x < o6:
		s = 6
		m = 0x04
	case x < o7:
		s = 7
		m = 0x02
	case x < o8:
		s = 8
		m = 0x01
	default:
		encError(fmt.Sprintf("element size %d overflows", x))
	}
	b := make([]byte, s)
	s--
	for s > 0 {
		b[s] = byte(x)
		x >>= 8
		s--
	}
	b[0] = byte(x) | m
	return b
}

func marshalInt(id Id, x int64) encoder {
	var xl int
	if x > 0 {
		switch {
		case x < 0x80:
			xl = 1
		case x < 0x8000:
			xl = 2
		case x < 0x800000:
			xl = 3
		case x < 0x80000000:
			xl = 4
		case x < 0x8000000000:
			xl = 5
		case x < 0x800000000000:
			xl = 6
		case x < 0x80000000000000:
			xl = 7
		default:
			xl = 8
		}
	} else {
		switch {
		case x > -0x7F:
			xl = 1
		case x > -0x7FFF:
			xl = 2
		case x > -0x7FFFFF:
			xl = 3
		case x > -0x7FFFFFFF:
			xl = 4
		case x > -0x7FFFFFFFFF:
			xl = 5
		case x > -0x7FFFFFFFFFFF:
			xl = 6
		case x > -0x7FFFFFFFFFFFFF:
			xl = 7
		default:
			xl = 8
		}
	}

	idBuf := id.Bytes()
	l := len(idBuf) + 1 + xl
	b := make(simpleElement, l)
	p := copy(b, idBuf)
	b[p] = 0x80 | byte(xl)
	p++

	l--
	b[l] = byte(x)
	for l > p {
		l--
		x >>= 8
		b[l] = byte(x)
	}
	return b
}

func marshalUint(id Id, x uint64) encoder {
	var xl int
	switch {
	case x < 0xFF:
		xl = 1
	case x < 0xFFFF:
		xl = 2
	case x < 0xFFFFFF:
		xl = 3
	case x < 0xFFFFFFFF:
		xl = 4
	case x < 0xFFFFFFFFFF:
		xl = 5
	case x < 0xFFFFFFFFFFFF:
		xl = 6
	case x < 0xFFFFFFFFFFFFFF:
		xl = 7
	default:
		xl = 8
	}
	idBuf := id.Bytes()

	l := len(idBuf) + 1 + xl
	b := make(simpleElement, l)
	p := copy(b, idBuf)
	b[p] = 0x80 | byte(xl)
	p++

	l--
	b[l] = byte(x)
	for l > p {
		l--
		x >>= 8
		b[l] = byte(x)
	}
	return b
}

func marshalString(id Id, s string) encoder {
	sb := []byte(s)
	l := len(sb)
	sz := marshalSize(int64(l))
	idBuf := id.Bytes()
	b := make(simpleElement, len(idBuf)+len(sz)+l)
	n := copy(b, idBuf)
	n += copy(b[n:], sz)
	copy(b[n:], sb)
	return b
}

func newStructEncoder(id Id, v reflect.Value) encoder {
	e := &containerElement{id: id}
	for fid, i := range cachedFieldIdMap(v.Type()) {
		fv := v.Field(i)
		if !fv.IsValid() || isEmptyValue(fv) {
			continue
		}
		fe := newEncoder(fid, fv)
		if e != nil {
			e.Append(fe)
		}
	}
	return e
}
