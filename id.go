// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml

import "fmt"

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
