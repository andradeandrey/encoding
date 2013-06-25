// Copyright (c) 2013, Emery Hemingway. All rights reserved.
// Actully most of it comes from encoding/json, courtesy of
// The Go Authors

package ebml

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strconv"
	"sync"
)

type encElement struct {
	body     []byte
	elements []*encElement
	reader   io.Reader
	size     int64
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

type encodeState struct {
	w        io.Writer
	elements []*encElement
}

func (es *encodeState) marshal(x interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	v := reflect.ValueOf(x)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for _, f := range cachedTypeFields(v.Type()) {
		fv := fieldByIndex(v, f.index)

		if !fv.IsValid() || isEmptyValue(fv) {
			continue
		}

		e, err := reflectValue(f.id, fv)
		if err != nil {
			return err
		}
		// TODO this append can go away and instead increase the
		// es.elements capacity to the amount of cachedTypeFields,
		// then have a moving index
		es.elements = append(es.elements, e)
	}
	for _, e := range es.elements {
		err = es.push(e)
		if err != nil {
			return
		}
	}
	return nil
}

func (es *encodeState) push(e *encElement) (err error) {
	_, err = es.w.Write(e.body)
	if err != nil {
		return
	}
	for _, se := range e.elements {
		err = es.push(se)
		if err != nil {
			return
		}
	}
	if e.reader != nil {
		_, err = io.Copy(es.w, e.reader)
		if err != nil {
			return
		}
	}
	return nil
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

func reflectValue(id []byte, v reflect.Value) (*encElement, error) {
	if id == nil {
		panic(fmt.Sprintf("nil id for value %v", v.Type()))
	}

	/*m, ok := v.Interface().(Marshaler)
	if !ok {
		// v dosen't match the interface. Check against *v too.
		if v.Kind() != reflect.Ptr && v.CanAddr() {
			m, ok = v.Addr().Interface().(Marshaler)
			if ok {
				v = v.Addr()
			}
		}
	}
	if ok && (v.Kind() != reflect.Ptr || !v.IsNil()) {
		r, size := m.MarshalEBML()
		return &encElement{reader: r, size: size}, nil
	}*/

	switch v.Kind() {
	case reflect.Struct:
		var children []*encElement
		var size int64
		for _, f := range cachedTypeFields(v.Type()) {
			fv := fieldByIndex(v, f.index)
			if !fv.IsValid() || isEmptyValue(fv) {
				continue
			}

			child, err := reflectValue(f.id, fv)
			if err != nil {
				return nil, &MarshalerError{v.Type(), err}
			}
			if child == nil {
				continue
			}
			children = append(children, child)
			size += child.size
		}

		sz := MarshalSize(size)
		l := len(id) + len(sz)
		b := make([]byte, l)
		p := copy(b, id)
		copy(b[p:], sz)
		size += int64(l)
		return &encElement{body: b, elements: children, size: size}, nil

	case reflect.Slice:
		if v.IsNil() || v.Len() == 0 {
			return nil, nil
		}
		var size int64
		children := make([]*encElement, v.Len())
		for i := 0; i < v.Len(); i++ {
			child, err := reflectValue(id, v.Index(i))
			if err != nil {
				return nil, &MarshalerError{v.Type(), err}
			}
			children[i] = child
			size += child.size
		}
		// in the case of the slice, do not note the Id, nor marshal the size,
		// slice don't represent containers, only structs do.
		return &encElement{elements: children, size: size}, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x := v.Int()
		return marshalInt(id, x), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x := v.Uint()
		return marshalUint(id, x), nil

	case reflect.String:
		return marshalString(id, v.String()), nil

	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return nil, nil
		}
		return reflectValue(id, v.Elem())
	}
	return nil, &UnsupportedTypeError{v.Type()}
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

func parseIdTag(s string) ([]byte, error) {
	x, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return nil, err
	}
	var xl int
	switch {
	case x < 0x10:
		return nil, errors.New("invalid element ID " + s)
	case x < 0x400:
		xl = 1
	case x < 0x8000:
		xl = 2
	case x < 0x400000:
		xl = 3
	case x < 0x20000000:
		xl = 4
	default:
		return nil, errors.New(s + " overflows element ID")
	}
	buf := make([]byte, xl)
	for xl > 1 {
		xl--
		buf[xl] = byte(x)
		x >>= 8
	}
	buf[0] = byte(x)
	return buf, nil
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
func MarshalSize(x int64) []byte {
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

func marshalInt(id []byte, x int64) *encElement {
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

	l := len(id) + 1 + xl
	b := make([]byte, l)
	p := copy(b, id)
	b[p] = 0x80 | byte(xl)
	p++

	i := l - 1
	b[i] = byte(x)
	for i > p {
		x >>= 8
		b[i] = byte(x)
		i--
	}
	return &encElement{body: b, size: int64(len(b))}
}

func marshalUint(id []byte, x uint64) *encElement {
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

	l := len(id) + 1 + xl
	b := make([]byte, l)
	p := copy(b, id)
	b[p] = 0x80 | byte(xl)
	p++

	i := l - 1
	b[i] = byte(x)
	for i > p {
		x >>= 8
		b[i] = byte(x)
		i--
	}
	return &encElement{body: b, size: int64(len(b))}
}

func marshalString(id []byte, s string) *encElement {
	sb := []byte(s)
	l := len(sb)
	sz := MarshalSize(int64(l))
	b := make([]byte, len(id)+len(sz)+l)
	n := copy(b, id)
	n += copy(b[n:], sz)
	copy(b[n:], sb)
	return &encElement{body: b, size: int64(len(b))}
}

// A field represents a single field found in a struct.
type field struct {
	id    []byte
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
				tag := sf.Tag.Get("ebml")
				if tag == "" {
					continue
				}
				id, err := parseIdTag(tag)
				if err != nil {
					panic(err.Error())
				}
				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i

				ft := sf.Type
				if ft.Kind() == reflect.Ptr {
					// Follow pointer
					ft = ft.Elem()
				}

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

const singletonField = 0
