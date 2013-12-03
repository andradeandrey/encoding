// Copyright © 2013 Emery Hemingway

package bencode

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sort"
)

// Encoder writes bencode data to an output stream..
type Encoder struct {
	w   io.Writer
	e   encodeState
	err error
}

// NewEncoder returns a new Encoder that bencodes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes bencode data to the wrapped stream.
//
// See the documentation for Marshal for details about the
// conversion of Go values to bencode.
func (enc *Encoder) Encode(v interface{}) error {
	if enc.err != nil {
		return enc.err
	}
	enc.e.Reset()
	err := enc.e.marshal(v)
	if err != nil {
		return err
	}

	if _, err = enc.w.Write(enc.e.Bytes()); err != nil {
		enc.err = err
	}
	return err
}

// Marshal returns a bencoded form of x.
//
// Marshal traverses the value x recursively.
//
// Marshal uses the following type-dependent encodings:
//
// Floating point, integer, and Number values encode as bencode numbers.
//
// String values encode as bencode strings.
//
// Array and slice values encode as bencode arrays.
//
// Struct values encode as bencode maps. Each exported struct field
// becomes a member of the object.
// The object's default key string is the struct field name
// but can be specified in the struct field's tag value. The text of
// the struct field's tag value is the key name. Examples:
//
//   // Field appears in bencode as key "Field".
//   Field int
//
//   // Field appears in bencode as key "myName".
//   Field int `bencode:"myName"`
//
// Anonymous struct fields are ignored.
//
// Map values encode as bencode objects.
// The map's key type must be string; the object keys are used directly
// as map keys.
//
// Boolean, Pointer, Interface, Channel, complex, and function values cannot
// be encoded in bencode.
// Attempting to encode such a value causes Marshal to return
// a MarshalError.
//
// Bencode cannot represent cyclic data structures and Marshal does not
// handle them.  Passing cyclic structures to Marshal will result in
// an infinite recursion.
func Marshal(v interface{}) ([]byte, error) {
	e := &encodeState{}
	err := e.marshal(v)
	if err != nil {
		return nil, err
	}
	return e.Bytes(), err
}

// An UnsupportedTypeError is returned by Marshal when attempting
// to encode an unsupported value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "bencode: unsupported type: " + e.Type.String()
}

type UnsupportedValueError struct {
	Value reflect.Value
	Str   string
}

func (e *UnsupportedValueError) Error() string {
	return "bencode: unsupported value: " + e.Str
}

// An encodeState encodes bencode into a bytes.Buffer.
type encodeState struct {
	bytes.Buffer // accumulated output
	scratch      [64]byte
}

func (e *encodeState) marshal(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()
	e.reflectValue(reflect.ValueOf(v))
	return nil
}

func (e *encodeState) error(err error) {
	panic(err)
}

var byteSliceType = reflect.TypeOf([]byte(nil))

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// reflectValue writes the value in v to the output.
func (e *encodeState) reflectValue(v reflect.Value) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fmt.Fprintf(e, "i%de", v.Int())

	case reflect.String:
		s := v.String()
		fmt.Fprintf(e, "%d:%s", len(s), s)

	case reflect.Struct:
		e.WriteByte('d')
		for _, f := range cachedTypeFields(v.Type()) {
			fv := fieldByIndex(v, f.index)
			if !fv.IsValid() || f.omitEmpty && isEmptyValue(fv) {
				continue
			}
			fmt.Fprintf(e, "%d:%s", len(f.name), f.name)
			e.reflectValue(fv)
		}
		e.WriteByte('e')

	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			e.error(&UnsupportedTypeError{v.Type()})
		}
		e.WriteByte('d')
		if !v.IsNil() {
			var keys stringValues = v.MapKeys()
			sort.Sort(keys)
			for _, k := range keys {
				fmt.Fprintf(e, "%d:%s", k.Len(), k)
				e.reflectValue(v.MapIndex(k))
			}
		}
		e.WriteByte('e')

	case reflect.Slice, reflect.Array:
		if t := v.Type(); t == byteSliceType || t.Elem().Kind() == reflect.Uint8 {
			fmt.Fprintf(e, "%d:", v.Len())
			e.Write(v.Bytes())
			return
		}
		e.WriteByte('l')
		n := v.Len()
		for i := 0; i < n; i++ {
			e.reflectValue(v.Index(i))
		}
		e.WriteByte('e')
		return

	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return
		}
		e.reflectValue(v.Elem())

	default:
		e.error(&UnsupportedTypeError{v.Type()})
	}
}
