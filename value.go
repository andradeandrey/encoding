package ebml

import (
	"bytes"
)

type Int struct {
	id uint32
	*bytes.Buffer
}

func (i *Int) ID() uint32 { return i.id }

func (i *Int) Size() uint64 { return uint64(i.Len()) }

func NewInt(id uint32, v int64) (i *Int) {
	var s int
	switch {
	case v < 0x8F, v > -0x8F:
		s = 1
	case v < 0x8FFF, v > -0x8FFF:
		s = 2
	case v < 0x8FFFFF, v > -0x8FFFFF:
		s = 3
	case v < 0x8FFFFFFF, v > -0x8FFFFFFF:
		s = 4
	case v < 0x8FFFFFFFFF, v > -0x8FFFFFFFFF:
		s = 5
	case v < 0x8FFFFFFFFFFF, v > -0x8FFFFFFFFFFF:
		s = 6
	case v < 0x8FFFFFFFFFFFFF, v > -0x8FFFFFFFFFFFFF:
		s = 7
	default:
		s = 8
	}
	b := make([]byte, s)
	for s > 1 {
		s--
		b[s] = byte(v)
		v = v >> 8
	}
	b[s] = byte(v)

	return &Int{id, bytes.NewBuffer(b)}
}

type Uint struct {
	id uint32
	*bytes.Buffer
}

func (u *Uint) ID() uint32 { return u.id }

func (u *Uint) Size() uint64 { return uint64(u.Len()) }

func NewUint(id uint32, v uint64) *Uint {
	var s int
	switch {
	case v < 0xFF:
		s = 1
	case v < 0xFFFF:
		s = 2
	case v < 0xFFFFFF:
		s = 3
	case v < 0xFFFFFFFF:
		s = 4
	case v < 0xFFFFFFFFFF:
		s = 5
	case v < 0xFFFFFFFFFFFF:
		s = 6
	case v < 0xFFFFFFFFFFFFFF:
		s = 7
	default:
		s = 8
	}
	b := make([]byte, s)
	for s > 1 {
		s--
		b[s] = byte(v)
		v = v >> 8
	}

	return &Uint{id, bytes.NewBuffer(b)}
}

type Float struct {
	s int
	b []byte
}

type String struct {
	id uint32
	*bytes.Buffer
}

func (s *String) ID() uint32 { return s.id }

func (s *String) Size() uint64 { return uint64(s.Len()) }

func NewString(id uint32, s string) *String {
	return &String{id, bytes.NewBufferString(s)}
}
