package ebml

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"runtime"
	"strconv"
	"sync"
)

// Unmarshal parses the EBML-encoded data and stores the result
// in the value pointed to by v.
func Unmarshal(data []byte, v interface{}) error {
	d := decodeState{r: bytes.NewReader(data)}
	return d.unmarshal(v)
}

// Unmarshalar is the interface implemented by objects that can
// unmarshal themselves from a an EBML stream fed into an io.Writer.
// The Unmarshaler is free to block on Write and Close to pause decoding.
type Unmarshaler interface {
	UnmarshalEBML() io.WriteCloser
}

// An InvalidUnmarshalError describes an invalid argument passed to Unmarshal.
// (The argument to Unmarshal must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "ebml: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "ebml: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "ebml: Unmarshal(nil " + e.Type.String() + ")"
}

func (d *decodeState) unmarshal(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	for {
		id := d.readId()
		if id == 0 {
			break
		}
		size := d.readSize()
		rv = reflect.Indirect(rv)
		if rv.Type().NumMethod() > 0 {
			if w, ok := rv.Interface().(io.Writer); ok {
				_, err = io.CopyN(w, d.r, size)
				if err != nil {
					d.error(err)
				}
			}
		} else {
			d.readToStruct(id, size, rv)
		}
	}
	return d.savedError
}

type decodeState struct {
	r          io.ReadSeeker
	savedError error
}

// error aborts the decoding by panicking with err.
func (d *decodeState) error(err error) {
	panic(err)
}

// saveError saves the first err it is called with,
// for reporting at the end of the unmarshal.
func (d *decodeState) saveError(err error) {
	if d.savedError == nil {
		d.savedError = err
	}
}

func (d *decodeState) readId() uint32 {
	b := make([]byte, 4)
	_, err := d.r.Read(b[:1])
	if err != nil {
		if err == io.EOF {
			return 0
		} else {
			d.error(err)
		}
	}
	x := uint32(b[0])
	switch {
	case x >= 0x80:
		return x
	case x >= 0x40:
		b = b[1:2]
	case x >= 0x20:
		b = b[1:3]
	case x >= 0x10:
		b = b[1:4]
	default:
		d.error(errors.New("positioned at an invalid Id or EBMLMaxIDLength > 4"))
		return 0
	}
	_, err = d.r.Read(b)
	if err != nil {
		d.error(err)
	}
	for _, c := range b {
		x <<= 8
		x += uint32(c)
	}
	return x
}

func (d *decodeState) readSize() int64 {
	b := make([]byte, 8)
	_, err := d.r.Read(b[:1])
	if err != nil {
		d.error(err)
	}
	x := int64(b[0])

	switch {
	case x >= 0x80:
		x -= 0x80
		return x
	case x >= 0x40:
		x -= 0x40
		b = b[1:2]
	case x >= 0x20:
		x -= 0x20
		b = b[1:3]
	case x >= 0x10:
		x -= 0x10
		b = b[1:4]
	case x >= 0x08:
		x -= 0x08
		b = b[1:5]
	case x >= 0x04:
		x -= 0x04
		b = b[1:6]
	case x >= 0x02:
		x -= 0x02
		b = b[1:7]
	case x >= 0x01:
		x -= 0x01
		b = b[1:]
	}
	_, err = d.r.Read(b)
	if err != nil {
		d.error(err)
	}
	for _, c := range b {
		x <<= 8
		x += int64(c)
	}
	return x
}

var i int

func (d *decodeState) readValue(size int64, v reflect.Value) {
	i++
	if !v.IsValid() {
		_, err := d.r.Seek(size, 1)
		if err != nil {
			d.error(err)
		}
		return
	}

	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		b := make([]byte, size)
		_, err := d.r.Read(b)
		if err != nil {
			d.error(err)
		}
		v.SetUint(unmarshalUint(b))

	case reflect.Struct:
		if size == 0 {
			return
		}
		d.readStruct(size, v)

	case reflect.Interface:
		if m, ok := v.Interface().(Unmarshaler); ok {
			w := m.UnmarshalEBML()
			_, err := io.CopyN(w, d.r, size)
			if err != nil {
				d.error(err)
			}
		} else {
			d.error(&UnsupportedTypeError{v.Type()})
		}

	case reflect.Slice:
		d.readToSlice(size, v)

	case reflect.String:
		b := make([]byte, size)
		d.r.Read(b)
		v.SetString(string(b))

	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		d.readValue(size, v.Elem())

	default:
		d.error(&UnsupportedTypeError{v.Type()})
	}
}

func (d *decodeState) readStruct(size int64, v reflect.Value) {
	t := v.Type()
	fieldMap := cachedIdFields(t)
	pos, err := d.r.Seek(0, 1)
	if err != nil {
		d.error(err)
	}
	end := pos + size
	for pos < end {
		subId := d.readId()
		if subId == 0 {
			return
		}
		subSize := d.readSize()
		if subSize > 0 {
			fi, ok := fieldMap[subId]
			if !ok {
				pos, err = d.r.Seek(subSize, 1)
				if err != nil {
					d.error(err)
				}
				continue
			}
			subv := v
			if subv.Kind() == reflect.Ptr {
				if subv.IsNil() {
					subv.Set(reflect.New(subv.Type().Elem()))
				}
				subv = subv.Elem()
			}
			subv = subv.Field(fi)
			switch subv.Kind() {
			case reflect.Struct:
				d.readStruct(subSize, subv)
			case reflect.Slice:
				d.readToSlice(subSize, subv)
			default:
				d.readValue(subSize, subv)
			}
		}
		pos, err = d.r.Seek(0, 1)
		if err != nil {
			d.error(err)
		}

	}
}

// readToSlice decodes an EBML element from d.r and appends it to the
// slice at v.
func (d *decodeState) readToSlice(size int64, v reflect.Value) {
	//var subv reflect.Value
	i := v.Len()
	if i >= v.Cap() {
		newcap := v.Cap() + v.Cap()/2
		if newcap < 4 {
			newcap = 4
		}
		newv := reflect.MakeSlice(v.Type(), v.Len(), newcap)
		reflect.Copy(newv, v)
		v.Set(newv)
	}
	v.SetLen(i + 1)
	d.readValue(size, v.Index(i))
}

// GET RID OF THIS
// readToStruct decodes an EBML element from d.r into the struct at v,
// if that element is tagged to a field of v.
func (d *decodeState) readToStruct(id uint32, size int64, v reflect.Value) {
	// Calling this each time may be inefficient
	fieldMap := cachedIdFields(v.Type())
	fi, ok := fieldMap[id]
	if !ok {
		_, err := d.r.Seek(size, 1)
		if err != nil {
			d.error(err)
		}
		return
	}

	subv := v
	if subv.Kind() == reflect.Ptr {
		if subv.IsNil() {
			subv.Set(reflect.New(subv.Type().Elem()))
		}
		subv = subv.Elem()
	}
	subv = subv.Field(fi)
	d.readValue(size, subv)
}

// indirect walks down v allocating pointers as needed,
// until it gets to a non-pointer.
func indirect(v reflect.Value) reflect.Value {
	// If v is a named type and is addressable,
	// start with its addres, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() && e.Elem().Kind() == reflect.Ptr {
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		if v.Elem().Kind() != reflect.Ptr && v.CanSet() {
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

func unmarshalUint(b []byte) uint64 {
	x := uint64(b[0])
	for _, c := range b[1:] {
		x <<= 8
		x += uint64(c)
	}
	return x
}

func parseFieldIds(t reflect.Type) (m map[uint32]int) {
	m = make(map[uint32]int)
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		tag := sf.Tag.Get("ebml")
		if tag == "" {
			continue
		}
		id, err := strconv.ParseUint(tag, 16, 32)
		if err != nil {
			panic(err.Error())
		}
		m[uint32(id)] = i
	}
	return
}

var fieldIdCache struct {
	sync.RWMutex
	m map[reflect.Type]map[uint32]int
}

func cachedIdFields(t reflect.Type) map[uint32]int {
	fieldIdCache.RLock()
	m, ok := fieldIdCache.m[t]
	fieldIdCache.RUnlock()
	if !ok {
		if fieldIdCache.m == nil {
			fieldIdCache.m = map[reflect.Type]map[uint32]int{}
		}

		m = parseFieldIds(t)
		fieldIdCache.Lock()
		fieldIdCache.m[t] = m
		fieldIdCache.Unlock()
	}
	return m
}
