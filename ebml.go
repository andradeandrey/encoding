// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

// Package ebml marshals and unmarshals Go objects struct to and from
// the Exensible Binary Markup Langauge.
package ebml

import (
	"fmt"
	"strconv"
)

// Header is a struct for encoding and decoding EBML streams.
//
// If nesting in a struct, it's field should be listed first
// and should should have an ebml tag set to 1a45dfa3.
//
// The following could be a valid top-level struct for
// representing Matroska streams:
//	type Matroska struct {
//		ebml.Header `ebml:"1a45dfa3"`
//		Segment     []Segment `ebml:"18538067"`
//	}
//
// You will however need to populate field values in Header
// to form a valid EBML document.
type Header struct {
	EbmlId             Id     `ebml:"1a45dfa3"`
	EBMLVersion        uint8  `ebml:"4286"`
	EBMLReadVersion    uint8  `ebml:"42f7"`
	EBMLMaxIDLength    uint8  `ebml:"42f2"`
	EBMLMaxSizeLength  uint8  `ebml:"42f3"`
	DocType            string `ebml:"4282"`
	DocTypeVersion     uint8  `ebml:"4287"`
	DocTypeReadVersion uint8  `ebml:"4285"`
}

// Id is a type that identifies an ebml element.
type Id []byte

//NewId generates a new Id for a given ebml element ID expressed as a uint32.
func NewId(x uint32) Id {
	var l int
	switch {
	case x < 0x10:
		panic("invalid element ID")
	case x < 0x400:
		l = 1
	case x < 0x8000:
		l = 2
	case x < 0x400000:
		l = 3
	case x < 0x20000000:
		l = 4
	default:
		panic("invalid element ID")
	}
	id := make(Id, l)
	for l > 1 {
		l--
		id[l] = byte(x)
		x >>= 8
	}
	id[0] = byte(x)
	return id
}

func NewIdFromString(s string) (Id, error) {
	x, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return nil, err
	}
	return NewId(uint32(x)), nil
}

// Format returns
func (id Id) Format(s fmt.State, c rune) {
	if len(id) == 0 {
		fmt.Fprint(s, "<nil>")
		return
	}

	switch c {
	case 'b':
		fmt.Fprintf(s, "%08b", id[0])
		for _, b := range id[1:] {
			fmt.Fprintf(s, ".%08b", b)
		}

	case 'd':
		i := uint32(id[0])
		for _, b := range id[1:] {
			i <<= 8
			i += uint32(b)
		}
		fmt.Fprint(s, i)

	case 'v', 's', 'x':
		fmt.Fprint(s, id.String())

	default:
		fmt.Fprintf(s, "%%!%c(ebml.Id=%s)", c, id.String())
	}
}

func (id Id) String() string {
	s := fmt.Sprintf("%x", id[0])
	for _, b := range id[1:] {
		s += fmt.Sprintf("%02x", b)
	}
	return s
}
