// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"time"
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
	ce.header = append(ce.id.bytes(), marshalSize(ce.size)...)
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

func encode(id Id, v reflect.Value) encoder {
	if m, ok := v.Interface().(Marshaler); ok {
		size, wt := m.MarshalEBML()
		header := append(id.bytes(), marshalSize(size)...)
		return &marshalerElement{id, size, header, wt}
	}

	if t, ok := v.Interface().(time.Time); ok {
		return encodeTime(id, t)
	}

	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x := v.Int()
		return marshalInt(id, x)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x := v.Uint()
		return marshalUint(id, x)

	case reflect.Float32:
		return float32Element{id, v.Interface().(float32)}

	case reflect.Float64:
		return float64Element{id, v.Interface().(float64)}

	case reflect.String:
		return encodeString(id, v)

	case reflect.Slice:
		return encodeSlice(id, v)

	case reflect.Struct:
		return encodeStruct(id, v)
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

	idBuf := id.bytes()
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
	idBuf := id.bytes()

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

type float32Element struct {
	id Id
	f  float32
}

func (e float32Element) Size() int64 { return e.id.len() + 5 }
func (e float32Element) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(e.id.bytes())
	if err == nil {
		_, err = w.Write([]byte{0x84})
		if err == nil {
			err = binary.Write(w, binary.BigEndian, e.f)
			n += 5
		}
	}
	return int64(n), err
}

type float64Element struct {
	id Id
	f  float64
}

func (e float64Element) Size() int64 { return e.id.len() + 9 }
func (e float64Element) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(e.id.bytes())
	if err == nil {
		_, err = w.Write([]byte{0x88})
		if err == nil {
			n++
			err = binary.Write(w, binary.BigEndian, e.f)
		} else {
			n += 8
		}
	}
	return int64(n), err
}

func encodeSlice(id Id, v reflect.Value) encoder {
	if bs, ok := v.Interface().([]byte); ok {
		idBuf := id.bytes()
		sizeBuf := marshalSize(int64(len(bs)))
		buf := make(simpleElement, len(idBuf)+len(sizeBuf)+len(bs))
		n := copy(buf, idBuf)
		n += copy(buf[n:], sizeBuf)
		copy(buf[n:], bs)
		return buf
	}

	l := v.Len()
	s := make(sliceElement, l)
	for i := 0; i < l; i++ {
		s[i] = encode(id, v.Index(i))
	}
	return s
}

func encodeString(id Id, v reflect.Value) encoder {
	sb := []byte(v.String())
	l := len(sb)
	sz := marshalSize(int64(l))
	idBuf := id.bytes()
	b := make(simpleElement, len(idBuf)+len(sz)+l)
	n := copy(b, idBuf)
	n += copy(b[n:], sz)
	copy(b[n:], sb)
	return b
}

func encodeStruct(id Id, v reflect.Value) encoder {
	e := &containerElement{id: id}
	for fid, i := range cachedFieldIdMap(v.Type()) {
		fv := v.Field(i)
		if !fv.IsValid() || isEmptyValue(fv) {
			continue
		}
		fe := encode(fid, fv)
		if e != nil {
			e.Append(fe)
		}
	}
	return e
}

func encodeTime(id Id, t time.Time) encoder {
	d := t.Sub(epoch) // epoch defined in ebml.go

	idb := id.bytes()
	b := make(simpleElement, len(idb)+9)
	n := copy(b, idb)
	b[n] = 0x88 // length will be 8
	for i := len(b) - 1; i > n; i-- {
		b[i] = byte(d)
		d >>= 8
	}
	return b
}
