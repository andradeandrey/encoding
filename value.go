package ebml

import (
	"bytes"
)

type Int struct {
	id uint32
	*bytes.Buffer
}

func (i *Int) ID() uint32 { return i.id }

func (i *Int) Size() int64 { return int64(i.Len()) }

func NewInt(id uint32, v interface{}) *Int {
	var i int64
	switch x := v.(type) {
	case int:
		i = int64(x)
	case int32:
		i = int64(x)
	case int64:
		i = int64(x)
	default:
		return nil
	}

	var s int
	switch {
	case i < 0x8F, i > -0x8F:
		s = 1
	case i < 0x8FFF, i > -0x8FFF:
		s = 2
	case i < 0x8FFFFF, i > -0x8FFFFF:
		s = 3
	case i < 0x8FFFFFFF, i > -0x8FFFFFFF:
		s = 4
	case i < 0x8FFFFFFFFF, i > -0x8FFFFFFFFF:
		s = 5
	case i < 0x8FFFFFFFFFFF, i > -0x8FFFFFFFFFFF:
		s = 6
	case i < 0x8FFFFFFFFFFFFF, i > -0x8FFFFFFFFFFFFF:
		s = 7
	default:
		s = 8
	}
	b := make([]byte, s)
	for s > 1 {
		s--
		b[s] = byte(i)
		i >>= 8
	}
	b[0] = byte(i)

	return &Int{id, bytes.NewBuffer(b)}
}

type Uint struct {
	id uint32
	*bytes.Buffer
}

func (u *Uint) ID() uint32 { return u.id }

func (u *Uint) Size() int64 { return int64(u.Len()) }

func NewUint(id uint32, v interface{}) *Uint {
	var i uint64
	switch x := v.(type) {
	case uint:
		i = uint64(x)
	case uint32:
		i = uint64(x)
	case uint64:
		i = uint64(x)
	default:
		return nil
	}

	var s int
	switch {
	case i < 0xFF:
		s = 1
	case i < 0xFFFF:
		s = 2
	case i < 0xFFFFFF:
		s = 3
	case i < 0xFFFFFFFF:
		s = 4
	case i < 0xFFFFFFFFFF:
		s = 5
	case i < 0xFFFFFFFFFFFF:
		s = 6
	case i < 0xFFFFFFFFFFFFFF:
		s = 7
	default:
		s = 8
	}
	b := make([]byte, s)
	for s > 1 {
		s--
		b[s] = byte(i)
		i >>= 8
	}
	b[0] = byte(i)
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

func (s *String) Size() int64 { return int64(s.Len()) }

func NewString(id uint32, s string) *String {
	return &String{id, bytes.NewBufferString(s)}
}
