package ebml

import (
	"fmt"
)

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

type MarshallError string

func (e MarshallError) Error() string {
	return string(e)
}

// MarshallSize returns an EBML binary representation of a size
func MarshallSize(x uint64) []byte {
	var s int
	var m byte
	switch {
	case x == 0:
		return []byte{byte(0)}
	case x < o1:
		s = 1
		m = 0x80
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

type UnmarshallError string

func (e UnmarshallError) Error() string {
	return string(e)
}

// UnmarshallID returns an EBML ID element at the start of buffer b
func UnmarshallID(b []byte) (uint32, error) {
	x := uint32(b[0])
	var i int
	switch {
	case x > 0x80: // 1xxx xxxx
		return x, nil

	case x > 0x40: // 01xx xxxx  xxxx xxxx
		i = 2

	case x > 0x20: // 001x xxxx  xxxx xxxx  xxxx xxxx
		i = 3

	case x > 0x10: // 0001 xxxx  xxxx xxxx  xxxx xxxx  xxxx xxxx
		i = 4

	default:
		return 0, UnmarshallError("not positioned at an ID element or document MaxIDLength is not supported")
	}

	for j := 1; j < i; j++ {
		x <<= 8
		x += uint32(b[j])
	}

	return x, nil
}

// UnmarshallSize returns an EBML size element at the start of buffer b
func UnmarshallSize(b []byte) (uint64, error) {
	x := uint64(b[0])
	var i int
	switch {
	case x >= 0x80: // 1xxx xxxx
		return x - 0x80, nil

	case x >= 0x40: // 01xx xxxx  xxxx xxxx
		i = 2
		x -= 0x40

	case x >= 0x20: // 001x xxxx  xxxx xxxx  xxxx xxxx
		i = 3
		x -= 0x20

	case x >= 0x10: // 0001 xxxx  xxxx xxxx  xxxx xxxx  xxxx xxxx
		i = 4
		x -= 0x10

	case x >= 0x08: 
		i = 5
		x -= 0x08

	case x >= 0x04: 
		i = 6
		x -= 0x04

	case x >= 0x02: 
		i = 7
		x -= 0x02

	case x >= 0x01: 
		i = 8
		x -= 0x01

	default:
		return 0, UnmarshallError("not positioned at an size element or document MaxSizeLength is not supported")
	}

	for j := 1; j < i; j++ {
		x <<= 8
		x += uint64(b[j])
	}

	return x, nil
}
