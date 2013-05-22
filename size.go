package ebml

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

func (e *Encoder) EncodeSize(x uint64) (err error) {
	var buf []byte
	switch {
	case x == 0:
		_, err = e.Write([]byte{byte(0)})
		return err

	case x < o1:
		b := byte(x)
		buf = []byte{b | 0x80}

	case x < o2:
		buf = make([]byte, 2)
		buf[1] = byte(x)
		x >>= 8
		buf[0] = byte(x) | 0x40

	case x < o3:
		buf = make([]byte, 3)
		buf[2] = byte(x)
		x >>= 8
		buf[1] = byte(x)
		x >>= 8
		buf[0] = byte(x) | 0x20

	case x < o4:
		buf = make([]byte, 4)
		for i := 3; i > 0; i-- {
			buf[i] = byte(x)
			x >>= 8
		}
		buf[0] = byte(x) | 0x10

	case x < o5:
		buf = make([]byte, 5)
		for i := 4; i > 0; i-- {
			buf[i] = byte(x)
			x >>= 8
		}
		buf[0] = byte(x) | 0x08

	case x < o6:
		buf = make([]byte, 6)
		for i := 5; i > 0; i-- {
			buf[i] = byte(x)
			x >>= 8
		}
		buf[0] = byte(x) | 0x04

	case x < o7:
		buf = make([]byte, 7)
		for i := 6; i > 0; i-- {
			buf[i] = byte(x)
			x >>= 8
		}
		buf[0] = byte(x) | 0x02

	case x < o8:
		buf = make([]byte, 8)
		for i := 7; i > 0; i-- {
			buf[i] = byte(x)
			x >>= 8
		}
		buf[0] = 0x01

	default:
		panic("element size overflow")
	}

	_, err = e.Write(buf)
	return
}
