package ebml

import (
	"fmt"
	"io"
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

type EncoderError string

func (e EncoderError) Error() string {
	return string(e)
}

type EncoderConfig struct {
	DocType            string
	DocTypeVersion     uint
	DocTypeReadVersion uint
}

// An Encoder writes EBML data to an output stream.
type Encoder struct {
	io.Writer
}

// NewEncoder returns a new encoder that writes to w.
//
// NewEncoder will write an EBML header to w with values from
// c before returning.,
func NewEncoder(w io.Writer, c *EncoderConfig) (*Encoder, error) {
	e := &Encoder{w}
	return e, e.Encode(newHeader(c.DocType, c.DocTypeVersion, c.DocTypeReadVersion))
}

// Encode writes a value that conforms to the Container
// or Element interface.
func (e *Encoder) Encode(v Element) (err error) {
	switch V := v.(type) {
	case Container:
		err = e.EncodeID(V.ID())
		if err != nil {
			return
		}
		err = e.EncodeSize(V.Size())
		if err != nil {
			return
		}
		for element := V.Next(); element != nil; element = V.Next() {
			err = e.Encode(element)
			if err != nil {
				return
			}
		}
	case Value:
		err = e.EncodeID(V.ID())
		if err != nil {
			return
		}
		err = e.EncodeSize(V.Size())
		if err != nil {
			return
		}
		_, err = io.Copy(e, V)

	default:
		err = EncoderError(fmt.Sprintf("%T conforms to neither the ebml.Container nor the ebml.Value interface", v))
	}
	return
}

// PutHeader writes a EBML header to the encoder stream
// docType is an ASCII string that identifies the type of the document.
// docTypeVersion is the version of document type to which the
// document conforms to.
// docTypeReadVersion is the minimum DocType version an interpreter
// has to support in order to read the document.
func (e *Encoder) PutHeader(docType string, docTypeVersion, docTypeReadVersion int64) error {
	return e.Encode(newHeader(docType, docTypeVersion, docTypeReadVersion))
}

// PutUint writes an unsigned interger with ebml ID id to the encoder strem.
func (e *Encoder) PutUint(id uint32, v interface{}) error {
	i := NewUint(id, v)
	return e.Encode(i)
}

// EncodeID writes an element ID to the encoder stream.
//
// See the Encode convenience function.
func (e *Encoder) EncodeID(x uint32) (err error) {
	var s int
	switch {

	case x < 0x10:
		panic("invalid element ID")
	case x < 0x400:
		s = 1
	case x < 0x8000:
		s = 2
	case x < 0x400000:
		s = 3
	case x < 0x20000000:
		s = 4
	default:
		return EncoderError(fmt.Sprintf("%x overflows element ID", x))
	}

	buf := make([]byte, s)
	s--
	for s > 0 {
		buf[s] = byte(x)
		x >>= 8
		s--
	}
	buf[0] = byte(x)

	_, err = e.Write(buf)
	return

}

// EncodeID writes an element size to the encoder stream.
//
// See the Encode convenience function.
func (e *Encoder) EncodeSize(x int64) (err error) {
	var s int
	var m byte
	switch {
	case x == 0:
		_, err = e.Write([]byte{byte(0)})
		return err
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
		return EncoderError(fmt.Sprintf("%x overflows element size", x))
	}

	buf := make([]byte, s)
	s--
	for s > 0 {
		buf[s] = byte(x)
		x >>= 8
		s--
	}
	buf[0] = byte(x) | m

	_, err = e.Write(buf)
	return
}
