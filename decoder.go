package ebml

import (
	"bufio"
	"fmt"
	"io"
)

const (
	SupportedEBMLReadVersion = 1
)

type DecoderError string

func (e DecoderError) Error() string {
	return string(e)
}

type DecoderConfig struct {
	DocType            string
	DocTypeReadVersion uint
}

// A Decoder decodes an ebml stream.
type Decoder struct {
	*bufio.Reader
}

// NewDecoder returns a new decoder that reads and decodes from r.
func NewDecoder(r io.Reader, c *DecoderConfig) (*Decoder, error) {
	d := &Decoder{bufio.NewReader(r)}

	header, err := d.Next()
	if err != nil {
		return nil, err
	}

	if header.ID != EBMLIDUint {
		return nil, DecoderError(fmt.Sprintf("invalid EBML header ID %x", header.ID))
	}

	fields := make(map[uint32][]byte)

	var b []byte
	for e, err := header.Next(); e != nil; e, err = header.Next() {
		if err != nil {
			return nil, err
		}

		b, err = e.Bytes()
		if err != nil {
			return nil, err
		}
		fields[e.ID] = b
	}

	b, ok := fields[EBMLReadVersionIDUint]
	if ok {
		rv := DecodeUint(b)
		if rv > SupportedEBMLReadVersion {
			return nil, DecoderError(fmt.Sprintf("stream requires an EBML version of %d or greater", rv))
		}
	}

	b, ok = fields[DocTypeIDUint]
	if !ok {
		return nil, DecoderError("did not find mandatory DocType element")
	}
	dt := string(b)
	if dt != c.DocType {
		return nil, DecoderError(fmt.Sprintf("stream is of DocType %v, not %v", dt, c.DocType))
	}

	b, ok = fields[DocTypeReadVersionIDUint]
	if !ok {
		return nil, DecoderError("did not find mandatory DocTypeReadVersion element")
	}
	dtrv := DecodeUint(b)
	if dtrv > c.DocTypeReadVersion {
		return nil, DecoderError(fmt.Sprintf("stream requires a %v reader version of %d or greater", dt, dtrv))
	}

	return d, nil
}

func (d *Decoder) Next() (*DecoderElement, error) {
	id, err := d.decodeID()
	if err != nil {
		return nil, err
	}

	size, err := d.decodeSize()
	if err != nil {
		return nil, err
	}

	return &DecoderElement{d, int64(size), id, 0}, nil
}

func (d *Decoder) decodeID() (id uint32, err error) {
	var c byte
	c, err = d.ReadByte()
	if err != nil {
		return
	}

	id = uint32(c)
	var s int
	switch {

	case id >= 0x80: // 1xxx xxxx
		return

	case id >= 0x40: // 01xx xxxx  xxxx xxxx
		s = 1

	case id >= 0x20: // 001x xxxx  xxxx xxxx  xxxx xxxx
		s = 2

	case id >= 0x10: // 0001 xxxx  xxxx xxxx  xxxx xxxx  xxxx xxxx
		s = 3

	default:
		// TODO make an error
		return 0, DecoderError("An EBML ID element not at current position, or EBMLMaxIDLength is unknown")
	}

	for i := 0; i < s; i++ {
		id <<= 8
		c, err = d.ReadByte()
		if err != nil {
			return
		}
		id += uint32(c)
	}
	return
}

func (d *Decoder) decodeSize() (size uint64, err error) {
	var c byte
	c, err = d.ReadByte()
	if err != nil {
		return
	}
	size = uint64(c)
	var s int
	switch {
	case size >= 0x80: // 1xxx xxxx
		size -= 0x80
		return

	case size >= 0x40: // 01xx xxxx  xxxx xxxx
		size -= 0x40
		s = 1

	case size >= 0x20: // 001x xxxx  xxxx xxxx  xxxx xxxx
		size -= 0x20
		s = 2

	case size >= 0x10: // 0001 xxxx  xxxx xxxx  xxxx xxxx  xxxx xxxx
		size -= 0x10
		s = 3

	case size >= 0x08: // 0000 1xxx  xxxx xxxx  xxxx xxxx  xxxx xxxx  xxxx xxxx
		size -= 0x08
		s = 4

	case size >= 0x04: // 0000 01xx  xxxx xxxx  xxxx xxxx  xxxx xxxx  xxxx xxxx  xxxx xxxx
		size -= 0x04
		s = 5

	case size >= 0x02: // 0000 001x  xxxx xxxx  xxxx xxxx  xxxx xxxx  xxxx xxxx  xxxx xxxx  xxxx xxxx
		size -= 0x02
		s = 6

	case size == 0x01: // 0000 0001  xxxx xxxx  xxxx xxxx  xxxx xxxx  xxxx xxxx  xxxx xxxx  xxxx xxxx  xxxx xxxx
		size = 0
		s = 7

	default:
		// TODO make an error
		return 0, DecoderError("An EBML size element is not at the current position, or EBMLMaxSizeLength is unknown")
	}

	for i := 0; i < s; i++ {
		size <<= 8
		c, err = d.ReadByte()
		if err != nil {
			return
		}
		size += uint64(c)
	}
	return
}

type DecoderElement struct {
	R    *Decoder
	N    int64
	ID   uint32
	size uint64
}

func (e *DecoderElement) Read(p []byte) (n int, err error) {
	if e.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > e.N {
		p = p[0:e.N]
	}
	n, err = e.R.Read(p)
	e.N -= int64(n)
	return
}

func (e *DecoderElement) Bytes() ([]byte, error) {
	b := make([]byte, e.N)
	n, err := e.R.Read(b)
	e.N -= int64(n)
	return b, err
}

func (e *DecoderElement) Next() (nx *DecoderElement, err error) {
	if e.N <= 0 {
		return nil, nil
	}

	nx, err = e.R.Next()
	if nx != nil {
		e.N -= nx.N
	}
	return
}

func DecodeUint(b []byte) (v uint) {
	l := len(b)

	v = uint(b[0])
	for i := 1; i < l; i++ {
		v <<= 8
		v += uint(b[i])
	}
	return
}
