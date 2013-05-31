package ebml

import (
	"fmt"
	"io"
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
		_, err = e.Write(V.ID())
		if err != nil {
			return
		}
		_, err = e.Write(MarshallSize(V.Size()))
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
		_, err = io.Copy(e, V)
		if err != nil {
			return
		}

	default:
		panic(fmt.Sprintf("%T conforms to neither the ebml.Container nor the ebml.Value interface", v))
	}
	return
}

// PutUint writes an unsigned interger with ebml ID id to the encoder strem.
func (e *Encoder) PutUint(id []byte, v interface{}) error {
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
