package ebml

import (
	"bufio"
	"io"
)

// An Encoder writes EBML data to an output stream.
type Encoder struct {
	*bufio.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{Writer: bufio.NewWriter(w)}
}

// Encode writes a value that conforms to the Container 
// or Element interface.
func (e *Encoder) Encode(v interface{}) (err error) {
	if V, ok := v.(Container); ok {
		e.EncodeID(V.ID())
		e.EncodeSize(V.Size())
		for element := V.Next(); element != nil; element = V.Next() {
			err = e.Encode(element)
			if err != nil {
				return
			}
		}
		return
	}
	if V, ok := v.(Element); ok {
		e.EncodeID(V.ID())
		e.EncodeSize(V.Size())
		_, err = io.Copy(e, V)
		return
	}
	panic("Could not encode value")
}

// PutHeader writes a EBML header to the encoder stream
// docType is an ASCII string that identifies the type of the document.
// docTypeVersion is the version of document type to which the
// document conforms to.
// docTypeReadVersion is the minimum DocType version an interpreter
// has to support in order to read the document.
func (e *Encoder) PutHeader(docType string, docTypeVersion, docTypeReadVersion uint64) error {
	return e.Encode(newHeader(docType, docTypeVersion, docTypeReadVersion))
}

// EncodeID writes an element ID to the encoder stream.
//
// See the Encode convenience function.
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

// EncodeID writes an element size to the encoder stream.
//
// See the Encode convenience function.
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
