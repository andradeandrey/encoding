// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"time"
)

// readIdFrom reads an Id from a Reader and returns the number of bytes read and the Id
func (d *Decoder) readId() (int, Id) {
	n, err := d.r.Read(d.buf[:1])
	if err != nil {
		decError(err.Error())
	}

	id := Id(d.buf[0])
	var buf []byte
	switch {
	case id >= 0x80:
		return n, id
	case id >= 0x40:
		buf = d.buf[:1]
	case id >= 0x20:
		buf = d.buf[:2]
	case id >= 0x10:
		buf = d.buf[:3]
	default:
		p, err := d.r.Seek(-1, 1)
		if err != nil {
			decError(err.Error())
		}
		d.r.Read(buf)
		decError(fmt.Sprintf("invalid Id at reader position 0x%x or EBMLMaxIDLength > 4, next 8 bytes 0x%0.2x", p, buf))
	}
	var nn int
	nn, err = d.r.Read(buf)
	n += nn
	if err != nil {
		decError(err.Error())
	}
	for _, c := range buf {
		id <<= 8
		id += Id(c)
	}
	return n, id
}

// readSizeFrom reads a size from a Reader and returns the number of bytes read and the size
func (d *Decoder) readSize() (int, int64) {
	n, err := d.r.Read(d.buf[:1])
	if err != nil {
		decError(err.Error())
	}
	size := int64(d.buf[0])
	var buf []byte
	switch {
	case size >= 0x80:
		size -= 0x80
		return n, size
	case size >= 0x40:
		size -= 0x40
		buf = d.buf[:1]
	case size >= 0x20:
		size -= 0x20
		buf = d.buf[:2]
	case size >= 0x10:
		size -= 0x10
		buf = d.buf[:3]
	case size >= 0x08:
		size -= 0x08
		buf = d.buf[:4]
	case size >= 0x04:
		size -= 0x04
		buf = d.buf[:5]
	case size >= 0x02:
		size -= 0x02
		buf = d.buf[:6]
	case size >= 0x01:
		size -= 0x01
		buf = d.buf[:7]
	}
	var nn int
	nn, err = d.r.Read(buf)
	n += nn
	if err != nil {
		decError(err.Error())
	}
	for _, c := range buf {
		size <<= 8
		size += int64(c)
	}
	return n, size
}

type decoderFunc func(d *Decoder, id Id, size int64, v reflect.Value)

/* Sadly this using this table results in 'initialization loops' during building
var decoderFuncTable = [...]decoderFunc{
	reflect.Uint:    decodeUint,
	reflect.Uint8:   decodeUint,
	reflect.Uint16:  decodeUint,
	reflect.Uint32:  decodeUint,
	reflect.Uint64:  decodeUint,
	reflect.Float32: decodeFloat,
	reflect.Float64: decodeFloat,
	reflect.Slice:   decodeSlice,
	reflect.String:  decodeString,
	reflect.Struct:  decodeStruct,
}
*/

func decodeValue(d *Decoder, id Id, size int64, v reflect.Value) {
	if um, ok := v.Interface().(Unmarshaler); ok {
		if v.Kind() == reflect.Ptr && v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
			um = v.Interface().(Unmarshaler)
		}

		rf := um.UnmarshalEBML(size)
		r := io.LimitReader(d.r, size)
		rf.ReadFrom(r)
		return
	}

	// If we got an interface or a pointer, dereference it.
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	// look up the function to decode to value v
	var fn decoderFunc
	// I wanted to use an array of functions indexed by reflect.Kind,
	// but kept getting initialization loop build errors
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		fn = decodeInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fn = decodeUint
	case reflect.Float32:
		fn = decodeFloat32
	case reflect.Float64:
		fn = decodeFloat64
	case reflect.Slice:
		fn = decodeSlice
	case reflect.String:
		fn = decodeString
	case reflect.Struct:
		if _, ok := v.Interface().(time.Time); ok {
			fn = decodeTime
		} else {
			fn = decodeStruct
		}
	default:
		decError(fmt.Sprintf("unsupported type %v", v.Type()))
	}
	fn(d, id, size, v)
}

func decodeInt(d *Decoder, id Id, size int64, v reflect.Value) {
	_, err := d.r.Read(d.buf[:size])
	x := int64(int8(d.buf[0]))
	for _, c := range d.buf[1:size] {
		x <<= 8
		x += int64(c)
	}
	if x == 0 {
		return
	}
	if v.OverflowInt(x) {
		decError(fmt.Sprintf("element %s value %d overflows %s", id, x, v.Type()))
	}
	v.SetInt(x)
	if err != nil {
		decError(err.Error())
	}
}

func decodeUint(d *Decoder, id Id, size int64, v reflect.Value) {
	_, err := d.r.Read(d.buf[:size])
	x := uint64(d.buf[0])
	for _, c := range d.buf[1:size] {
		x <<= 8
		x += uint64(c)
	}
	if x == 0 {
		return
	}
	if v.OverflowUint(x) {
		decError(fmt.Sprintf("element %s value %d overflows %s", id, x, v.Type()))
	}
	v.SetUint(x)
	if err != nil {
		decError(err.Error())
	}
}

func decodeFloat32(d *Decoder, id Id, size int64, v reflect.Value) {
	var x float32
	if size != 4 {
		decError(fmt.Sprintf("cannot decode a float of len %d to a float32", size))
	}

	buf := d.buf[:size]
	_, err := d.r.Read(buf)
	if err != nil && err != io.EOF {
		decError(err.Error())
	}
	err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &x)
	if err != nil {
		println(fmt.Sprintf("%x", buf))
		decError(err.Error())
	}
	v.SetFloat(float64(x))
}

func decodeFloat64(d *Decoder, id Id, size int64, v reflect.Value) {
	var x float64
	if size != 8 {
		decError(fmt.Sprintf("cannot decode a float of len %d to a float64", size))
	}

	buf := d.buf[:size]
	_, err := d.r.Read(buf)
	if err != nil && err != io.EOF {
		decError(err.Error())
	}
	err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &x)
	if err != nil {
		println(fmt.Sprintf("%x", buf))
		decError(err.Error())
	}
	v.SetFloat(x)
}

func decodeSlice(d *Decoder, id Id, size int64, v reflect.Value) {
	if _, ok := v.Interface().([]byte); ok {
		var buf []byte
		if size <= 8 {
			buf = d.buf[:size]
		} else {
			buf = make([]byte, int(size))
		}
		_, err := d.r.Read(buf)
		v.Set(reflect.ValueOf(buf))
		if err != nil {
			decError(err.Error())
		}
		return
	}

	// TODO(Emery): would be nice to use reflect.Append()
	n := v.Len()
	if n >= v.Cap() {
		newcap := v.Cap() + v.Cap()/2
		if newcap < 4 {
			newcap = 4
		}
		newv := reflect.MakeSlice(v.Type(), n+1, newcap)
		reflect.Copy(newv, v)
		v.Set(newv)
	} else {
		v.SetLen(n + 1)
	}
	decodeValue(d, id, size, v.Index(n))
}

func decodeString(d *Decoder, id Id, size int64, v reflect.Value) {
	buf := make([]byte, size)
	_, err := d.r.Read(buf)
	v.SetString(string(buf))
	if err != nil {
		decError(err.Error())
	}
}

func decodeStruct(d *Decoder, id Id, size int64, v reflect.Value) {
	t := v.Type()
	// get Id to field mappings
	idField := cachedFieldIdMap(t)
	// BUG(Emery): not caching decoder funtions for struct fields is suboptimal
	//fieldFunc := cachedFieldDecoderTable(t)

	var n int
	var subId Id
	var subSize int64
	var ok bool
	for size > 0 {
		// read and and size
		n, subId = d.readId()
		size -= int64(n)
		n, subSize = d.readSize()
		size -= int64(n)

		// look up if the subId should decode into a field
		if n, ok = idField[subId]; ok {
			decodeValue(d, subId, subSize, v.Field(n))
			/*
				subV = v.Field(n)
				// Derefence pointer
				for subV.Kind() == reflect.Ptr {
					subV = subV.Elem()
				}

				// use the cached decoder funtion for field
				fieldFunc[n]
			*/
		} else {
			d.r.Seek(size, 0)
		}
		size -= subSize
	}
}

func decodeTime(d *Decoder, id Id, size int64, v reflect.Value) {
	if size != 8 {
		decError(fmt.Sprintf("%d is an invalid length for a date value", size))
	}
	_, err := d.r.Read(d.buf)

	date := time.Duration(int8(d.buf[0]))
	for _, c := range d.buf[1:] {
		date <<= 8
		date += time.Duration(c)
	}
	v.Set(reflect.ValueOf(epoch.Add(date))) // epoch defined in ebml.go

	if err != nil {
		decError(err.Error())
	}
}
