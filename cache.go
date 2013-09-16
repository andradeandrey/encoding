// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	fieldIdMap        = make(map[reflect.Type]map[Id]int)
	fieldIdMutex      sync.RWMutex
	fieldDecoderMap   = make(map[reflect.Type][]decoderFunc)
	fieldDecoderMutex sync.RWMutex
)

// cachedFieldIdMap returns a map that contains Id to field number
// mappings for typ
func cachedFieldIdMap(typ reflect.Type) map[Id]int {
	fieldIdMutex.RLock()
	m, ok := fieldIdMap[typ]
	fieldIdMutex.RUnlock()
	if ok {
		return m
	}

	// Didn't get a map, lock down the map
	fieldIdMutex.Lock()
	defer fieldIdMutex.Unlock()
	m, ok = fieldIdMap[typ]
	if ok { // the work is already done
		return m
	}

	m = make(map[Id]int)
	// look over the fields
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Name == "EbmlId" {
			continue
		}
		tag := f.Tag.Get("ebml")
		if tag == "" || tag == "-" {
			continue
		}
		id, err := NewIdFromString(tag)
		if err != nil {
			panic(ebmlError("ebml: could not parse Id from struct " + typ.String() + " field " + f.Name + ", " + err.Error()))
		}
		m[id] = i
	}
	fieldIdMap[typ] = m
	return m
}

// cachedFieldIpDecoderTable returns a slice that contains decoder functions
// indexed by field number
func cachedFieldDecoderTable(typ reflect.Type) []decoderFunc {
	fieldDecoderMutex.RLock()
	s, ok := fieldDecoderMap[typ]
	fieldDecoderMutex.RUnlock()
	if ok {
		return s
	}

	// Didn't get a slice, lock down the map
	fieldDecoderMutex.Lock()
	defer fieldDecoderMutex.Unlock()
	s, ok = fieldDecoderMap[typ]
	if ok { // the work is already done
		return s
	}

	// a struct should have its fields parsed by cachedFieldIdMap already,
	// so this shouldn't be too sloppy to just pull it
	m := fieldIdMap[typ]

	// create a map with pre-nil'ed functions for typ
	s = make([]decoderFunc, typ.NumField())
	// look over the fields that could be decoded into
	for _, i := range m {
		f := typ.Field(i)
		k := f.Type.Kind()
		for k == reflect.Ptr {
			k = f.Type.Elem().Kind()
		}

		// I wanted to use an array of functions indexed by reflect.Kind,
		// but kept getting initialization loop build errors
		switch k {
		case reflect.Uint:
			s[i] = decodeUint
		case reflect.Uint8:
			s[i] = decodeUint
		case reflect.Uint16:
			s[i] = decodeUint
		case reflect.Uint32:
			s[i] = decodeUint
		case reflect.Uint64:
			s[i] = decodeUint
		case reflect.String:
			s[i] = decodeString
		case reflect.Struct:
			s[i] = decodeStruct
		default:
			decError(fmt.Sprintf("cannot decode to %s field %s, %s decoding is unsupported ",
				typ, f.Name, f.Type))
		}
	}
	fieldDecoderMap[typ] = s
	return s
}
