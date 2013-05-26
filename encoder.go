package ebml

import (
	"io"
)

// An Encoder writes EBML data to an output stream.
type Encoder struct {
	io.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

// Encode writes a value that conforms to the Container
// or Element interface.
func (e *Encoder) Encode(v Element) (err error) {
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
	if V, ok := v.(Value); ok {
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

// PutUint writes an unsigned interger with ebml ID id to the encoder strem.
func (e *Encoder) PutUint(id uint32, v uint64) error {
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
		panic("element ID overflow")
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
		panic("element size overflow")
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
