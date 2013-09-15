// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml

import (
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sync"
)

var indent string

type EncoderElement interface {
	io.WriterTo
	Size() int64
}

type containerElement struct {
	// These []bytes can probably be merged
	id       Id
	size     int64
	header   []byte
	elements []EncoderElement
}

func (ce *containerElement) Append(e EncoderElement) {
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
			return
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

// sliceElement is a EncoderElement for holding groups of elements, such as a group
// elements of the same type that occur more than once in a container.
type sliceElement []EncoderElement

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

// An UnsupportedTypeError is returned by Marshal when attempting
// to encode an unsupported value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "ebml: unsupported type: " + e.Type.String()
}

type MarshalerError struct {
	Type reflect.Type
	Err  error
}

func (e *MarshalerError) Error() string {
	return "ebml: error marshaling type " + e.Type.String() + ": " + e.Err.Error()
}

func marshal(id Id, v reflect.Value) (E EncoderElement, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return reflectValue(id, v)
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

func reflectValue(id Id, v reflect.Value) (EncoderElement, error) {
	if m, ok := v.Interface().(Marshaler); ok {
		wt, size := m.MarshalEBML()
		header := append(id.Bytes(), marshalSize(size)...)
		return &marshalerElement{id, size, header, wt}, nil
	}

	switch v.Kind() {
	case reflect.Struct:
		return marshalStruct(id, v)

	case reflect.Slice:
		if v.IsNil() || v.Len() == 0 {
			return nil, nil
		}
		s := make(sliceElement, v.Len())
		for i := 0; i < v.Len(); i++ {
			child, err := reflectValue(id, v.Index(i))
			if err != nil {
				return nil, &MarshalerError{v.Type(), err}
			}
			s[i] = child
		}
		return s, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x := v.Int()
		return marshalInt(id, x), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x := v.Uint()
		return marshalUint(id, x), nil

	case reflect.String:
		return marshalString(id, v.String()), nil

	case reflect.Interface, reflect.Ptr:
		return reflectValue(id, v.Elem())
	}
	return nil, &UnsupportedTypeError{v.Type()}
}

func marshalStruct(id Id, v reflect.Value) (EncoderElement, error) {
	E := &containerElement{id: id}
	for _, f := range cachedTypeFields(v.Type()) {
		fv := fieldByIndex(v, f.index)
		if !fv.IsValid() || isEmptyValue(fv) {
			continue
		}
		e, err := reflectValue(f.id, fv)
		if err != nil {
			return nil, &MarshalerError{v.Type(), err}
		}
		if e == nil {
			continue
		}
		E.Append(e)
	}
	if len(E.elements) == 0 {
		return nil, nil
	}
	return E, nil
}

func fieldByIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v
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
		panic(fmt.Sprintf("%x overflows element size", x))

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

func marshalInt(id Id, x int64) EncoderElement {
	var xl int
	switch {
	case x < 0x8F, x > -0x8F:
		xl = 1
	case x < 0x8FFF, x > -0x8FFF:
		xl = 2
	case x < 0x8FFFFF, x > -0x8FFFFF:
		xl = 3
	case x < 0x8FFFFFFF, x > -0x8FFFFFFF:
		xl = 4
	case x < 0x8FFFFFFFFF, x > -0x8FFFFFFFFF:
		xl = 5
	case x < 0x8FFFFFFFFFFF, x > -0x8FFFFFFFFFFF:
		xl = 6
	case x < 0x8FFFFFFFFFFFFF, x > -0x8FFFFFFFFFFFFF:
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

	i := l - 1

	b[i] = byte(x)
	for i > p {
		x >>= 8
		b[i] = byte(x)
		i--
	}
	return b
}

func marshalUint(id Id, x uint64) EncoderElement {
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

	i := l - 1
	b[i] = byte(x)
	for i > p {
		x >>= 8
		b[i] = byte(x)
		i--
	}
	return b
}

func marshalString(id Id, s string) EncoderElement {
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

// A field represents a single field found in a struct.
type field struct {
	id    Id
	index []int
	typ   reflect.Type
}

// typeFields returns a list of fields that EBML should recognize for the given type.
func typeFields(t reflect.Type) []field {
	// Anonymous fields to explore at the current level and the next.
	current := []field{}
	next := []field{{typ: t}}

	// Count of queued names for current level and the next.
	count := map[reflect.Type]int{}
	nextCount := map[reflect.Type]int{}

	// Types already visited at an earlier level.
	visited := map[reflect.Type]bool{}

	// Fields found.
	var fields []field

	for len(next) > 0 {
		current, next = next, current[:0]
		count, nextCount = nextCount, map[reflect.Type]int{}

		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true

			// Scan f.typ for fields to include.
			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)
				if sf.Name == "EbmlId" {
					continue
				}
				tag := sf.Tag.Get("ebml")
				if tag == "" {
					continue
				}
				id, err := NewIdFromString(tag)
				if err != nil {
					panic(err.Error())
				}
				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i

				ft := sf.Type
				//if ft.Kind() == reflect.Ptr {
				//	// Follow pointer
				//	ft = ft.Elem()
				//}

				// Record found field and index sequence.
				fields = append(fields, field{id, index, ft})
				if count[f.typ] > 1 {
					// If there were multipe instances, add a second,
					// so that the annihilation code will see a dulicate.
					// It only cares about the distinction between 1 or 2,
					// so don't bother generating and more copies.
					fields = append(fields, fields[len(fields)-1])
				}
			}
		}
	}
	return fields
}

var fieldCache struct {
	sync.RWMutex
	m map[reflect.Type][]field
}

// cachedTypeFields is like typeFields but uses a cache to avoid repeated work.
func cachedTypeFields(t reflect.Type) []field {
	fieldCache.RLock()
	f := fieldCache.m[t]
	fieldCache.RUnlock()
	if f != nil {
		return f
	}

	// Compute fields without lock.
	// Might dulpicate effort but won't hold other computations back.
	f = typeFields(t)
	if f == nil {
		f = []field{}
	}

	fieldCache.Lock()
	if fieldCache.m == nil {
		fieldCache.m = map[reflect.Type][]field{}
	}
	fieldCache.m[t] = f
	fieldCache.Unlock()
	return f
}
