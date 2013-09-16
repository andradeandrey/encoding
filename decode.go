// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml

import (
	//"bytes"
	//"errors"
	"fmt"
	"io"
	"reflect"
	//"unsafe"
)

// readIdFrom reads an Id from a Reader and returns the number of bytes read and the Id
func readIdFrom(r io.Reader) (int, Id) {
	buf := make([]byte, 8)
	n, err := r.Read(buf[:1])
	if err != nil {
		encError(err.Error())
	}

	id := Id(buf[0])
	switch {
	case id >= 0x80:
		return n, id
	case id >= 0x40:
		buf = buf[:1]
	case id >= 0x20:
		buf = buf[:2]
	case id >= 0x10:
		buf = buf[:3]
	default:
		encError("positioned at an invalid Id or EBMLMaxIDLength > 4")
	}
	var nn int
	nn, err = r.Read(buf)
	n += nn
	if err != nil {
		encError(err.Error())
	}
	for _, c := range buf {
		id <<= 8
		id += Id(c)
	}
	return n, id
}

// readSizeFrom reads a size from a Reader and returns the number of bytes read and the size
func readSizeFrom(r io.Reader) (int, int64) {
	buf := make([]byte, 8)
	n, err := r.Read(buf[:1])
	if err != nil {
		encError(err.Error())
	}
	size := int64(buf[0])

	switch {
	case size >= 0x80:
		size -= 0x80
		return n, size
	case size >= 0x40:
		size -= 0x40
		buf = buf[:1]
	case size >= 0x20:
		size -= 0x20
		buf = buf[:2]
	case size >= 0x10:
		size -= 0x10
		buf = buf[:3]
	case size >= 0x08:
		size -= 0x08
		buf = buf[:4]
	case size >= 0x04:
		size -= 0x04
		buf = buf[:5]
	case size >= 0x02:
		size -= 0x02
		buf = buf[:6]
	case size >= 0x01:
		size -= 0x01
		buf = buf[:7]
	}
	var nn int
	nn, err = r.Read(buf)
	n += nn
	if err != nil {
		encError(err.Error())
	}
	for _, c := range buf {
		size <<= 8
		size += int64(c)
	}
	return n, size
}

type decoderFunc func(d *Decoder, id Id, size int64, v reflect.Value)

func decodeValue(d *Decoder, id Id, size int64, v reflect.Value) {
	if um, ok := v.Interface().(Unmarshaler); ok {
		rf := um.UnmarshalEBML(size)
		rf.ReadFrom(d.r)
	}

	// If we got an interface or a pointer, dereference it.
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// I wanted to use an array of functions indexed by reflect.Kind, 
	// but kept getting initialization loop build errors
	var fn decoderFunc
	switch v.Kind() {
	case reflect.Uint:
		fn = decodeUint
	case reflect.Uint8:
		fn = decodeUint
	case reflect.Uint16:
		fn = decodeUint
	case reflect.Uint32:
		fn = decodeUint
	case reflect.Uint64:
		fn = decodeUint
	case reflect.String:
		fn = decodeString
	case reflect.Struct:
		fn = decodeStruct
	default:
		encError(fmt.Sprintf("unsupported type %v", v.Type()))
	}
	fn(d, id, size, v)
}

func decodeUint(d *Decoder, id Id, size int64, v reflect.Value) {
	_, err := d.r.Read(d.buf[:size])
	x := uint64(d.buf[0])
	for _, c := range d.buf[1:] {
		x += uint64(c)
	}
	if v.OverflowUint(x) {
		decError(fmt.Sprintf("element %s value %d overflows %s", id, x, v.Type()))
	}
	v.SetUint(x)
	if err != nil {
		decError(err.Error())
	}
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
	// get Id -> field mappings
	idField := cachedFieldIdMap(t)
	fieldFunc := cachedFieldDecoderTable(t)

	var n int
	var subId Id
	var subSize int64
	var ok bool
	for size > 0 {
		// read and and size
		n, subId = readIdFrom(d.r)
		size -= int64(n)
		n, subSize = readSizeFrom(d.r)
		size -= int64(n)

		// look up if the subId should decode into a field
		if n, ok = idField[subId]; !ok {
			continue
		}
		// use the cached decoder funtion for field
		fieldFunc[n](d, subId, subSize, v.Field(n))
		size -= subSize
	}
}
