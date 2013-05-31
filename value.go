package ebml

import (
	"io"
)

type Int struct {
	id  []byte
	buf []byte
	off int
}

func (i *Int) ID() []byte { return i.id }

func (i *Int) Size() uint64 { return uint64(len(i.buf)) }

func (i *Int) Read(p []byte) (n int, err error) {
	if i.off >= len(i.buf) {
		if len(p) == 0 {
			return
		}
		return 0, io.EOF
	}
	n = copy(p, i.buf[i.off:])
	i.off += n
	return
}

func NewInt(id []byte, v interface{}) *Int {
	var x int64
	switch V := v.(type) {
	case int:
		x = int64(V)
	case int32:
		x = int64(V)
	case int64:
		x = int64(V)
	default:
		return nil
	}

	var i int
	switch {
	case x < 0x8F, x > -0x8F:
		i = 1
	case x < 0x8FFF, x > -0x8FFF:
		i = 2
	case x < 0x8FFFFF, x > -0x8FFFFF:
		i = 3
	case x < 0x8FFFFFFF, x > -0x8FFFFFFF:
		i = 4
	case x < 0x8FFFFFFFFF, x > -0x8FFFFFFFFF:
		i = 5
	case x < 0x8FFFFFFFFFFF, x > -0x8FFFFFFFFFFF:
		i = 6
	case x < 0x8FFFFFFFFFFFFF, x > -0x8FFFFFFFFFFFFF:
		i = 7
	default:
		i = 8
	}

	j := len(id) + 1 + i
	b := make([]byte, j)
	k := copy(b, id)
	// Too much type conversion
	k = copy(b[k:], MarshallSize(uint64(i)))

	for j > k {
		j--
		b[j] = byte(x)
		x >>= 8
	}
	b[j] = byte(x)

	return &Int{id, b, 0}
}

type Uint struct {
	id  []byte
	buf []byte
	off int
}

func (u *Uint) ID() []byte { return u.id }

func (u *Uint) Size() uint64 { return uint64(len(u.buf)) }

func (u *Uint) Read(p []byte) (n int, err error) {
	if u.off >= len(u.buf) {
		if len(p) == 0 {
			return
		}
		return 0, io.EOF
	}
	n = copy(p, u.buf[u.off:])
	u.off += n
	return
}

func NewUint(id []byte, v interface{}) *Uint {
	var x uint64
	switch V := v.(type) {
	case uint:
		x = uint64(V)
	case uint32:
		x = uint64(V)
	case uint64:
		x = uint64(V)
	default:
		return nil
	}

	var i int
	switch {
	case x < 0xFF:
		i = 1
	case x < 0xFFFF:
		i = 2
	case x < 0xFFFFFF:
		i = 3
	case x < 0xFFFFFFFF:
		i = 4
	case x < 0xFFFFFFFFFF:
		i = 5
	case x < 0xFFFFFFFFFFFF:
		i = 6
	case x < 0xFFFFFFFFFFFFFF:
		i = 7
	default:
		i = 8
	}

	j := len(id) + 1 + i
	b := make([]byte, j)
	k := copy(b, id)
	k += copy(b[k:], MarshallSize(uint64(i)))

	for j > k {
		j--
		b[j] = byte(x)
		x >>= 8
	}
	b[j] = byte(i)
	return &Uint{id, b, 0}
}

type String struct {
	id  []byte
	buf []byte
	off int
}

func (s *String) ID() []byte { return s.id }

func (s *String) Size() uint64 { return uint64(len(s.buf)) }

func (s *String) Read(p []byte) (n int, err error) {
	if s.off >= len(s.buf) {
		if len(p) == 0 {
			return
		}
		return 0, io.EOF
	}
	n = copy(p, s.buf[s.off:])
	s.off += n
	return
}

func NewString(id []byte, s string) *String {
	sb := []byte(s)
	sz := MarshallSize(uint64(len(sb)))
	buf := make([]byte, len(id)+len(sz)+len(sb))

	n := copy(buf, id)
	n += copy(buf[n:], sz)
	copy(buf[n:], sb)

	return &String{id, buf, 0}
}
