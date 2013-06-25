// Copyright (c) 2013, Emery Hemingway. All rights reserved.
// Actully most of it comes from encoding/json, courtesy of
// The Go Authors

package ebml

import (
	"io"
)

// An Encoder writes EBML data to an output stream.
type Encoder struct {
	w   io.Writer
	err error
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the EBML binary encoding of v to the stream.
func (enc *Encoder) Encode(v interface{}) error {
	if enc.err != nil {
		return enc.err
	}
	e := &encodeState{w: enc.w}
	err := e.marshal(v)
	if err != nil {
		enc.err = err
		return err
	}
	return nil
}

/*
// Decode reads EBML elements into from the intov.
func (dec *Decoder) Decode(v interface{}) error {
	if dec.err != nil {
		return dec.err
	}
	d := decodeState{r: dec.r}
	err := d.unmarshal(v)
	if err != nil {
		dec.err = err
		return err
	}
	return nil
}
*/

/*
func (dec *Decoder) decodeStruct(v reflect.Value) error {
	for dec.off < dec.end {
		id, err := dec.readId()
		if err != nil {
			dec.err = err
			return err
		}
		size, err := readSize(dec.r)
		if err != nil {
			dec.err = err
			return err
		}
		fmt.Printf("got to %x, size %d\n", id, size)

		t := v.Elem().Type()
		i := indexById(t, unMarshallId(id))
		fmt.Println("id was", i)
		if i < 0 {
			fmt.Println("i was < 0, continueing")
			continue
		}
		fv := v.Field(i)
		switch fv.Kind() {
		case reflect.Struct:
			fmt.Println("got to a struct within a struct")

			off, err := dec.r.Seek(0, 1)
			if err != nil {
				dec.err = err
				return err
			}

			if v.Type() == ioLimitedReaderType {
				off, err := dec.r.Seek(0, 1)
				if err != nil {
					dec.err = err
					return err
				}
				// this won't work because the mutex is lost or confused
				v.Set(reflect.ValueOf(&io.LimitedReader{R: &seekedReader{dec, off}, N: size}))
				_, err = dec.r.Seek(size, 1)
				if err != nil {
					dec.err = err
					return err
				}
				continue
			}

			subDec := &Decoder{r: dec.r, off: off, end: off + size}
			err = subDec.decodeStruct(fv)
			if err != nil {
				dec.err = err
				return err
			}

		case reflect.Slice:
			fmt.Println("went into reflect.Slice")
							i := v.Len()
							l := i + 1
							newv := reflect.MakeSlice(v.Type(), l, l)
							err = dec.decode(newv.Index(i))
			 		if err != nil {
								dec.err = err
									return err
									}

									reflect.Copy(newv, v)
									v.Set(newv)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fmt.Println("went into reflect.Uint")
			b := make([]byte, size)

			_, err := dec.r.Read(b)
			if err != nil {
				dec.err = err
				return err
			}
			x := unMarshallUint(b)
			fv.SetUint(x)

		case reflect.String:
			fmt.Println("went into reflect.String")
			b := make([]byte, size)
			_, err := dec.r.Read(b)
			if err != nil {
				dec.err = err
				return err
			}
			fv.SetString(string(b))

		}
		return &UnsupportedTypeError{v.Type()}
	}
	return nil
}
*/
