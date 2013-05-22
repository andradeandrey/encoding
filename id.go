package ebml

func (e *Encoder) EncodeID(x uint32) (err error) {
	var buf []byte
	switch {
	case x == 0:
		_, err = e.Write([]byte{byte(0)})
		return err

	case x < 127:
		b := byte(x)
		buf = []byte{b | 0x80}

	case x < 16383:
		buf = make([]byte, 2)
		buf[1] = byte(x)
		x >>= 8
		buf[0] = byte(x) | 0x40

	case x < 2097151:
		buf = make([]byte, 3)
		buf[2] = byte(x)
		x >>= 8
		buf[1] = byte(x)
		x >>= 8
		buf[0] = byte(x) | 0x20

	case x < 268435455:
		buf = make([]byte, 4)
		buf[3] = byte(x)
		x >>= 8
		buf[2] = byte(x)
		x >>= 8
		buf[1] = byte(x)
		x >>= 8
		buf[0] = byte(x) | 0x10

	default:
		panic("element ID overflow")
	}

	_, err = e.Write(buf)
	return
}
