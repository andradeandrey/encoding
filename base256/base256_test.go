package base256

import "testing"

func newEncT(t *testing.T) Encoding {
	e, err := NewEncoding('⤀')
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func newEncB(t *testing.B) Encoding {
	e, err := NewEncoding('⤀')
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestIterative(t *testing.T) {
	e := newEncT(t)

	var in, out byte
	enc := make([]byte, e.EncodedLen(1))
	dec := make([]byte, 1)
	for in = 0; in < 0xFF; in++ {
		dec[0] = in
		e.Encode(enc, dec)
		e.Decode(dec, enc)
		out = dec[0]

		if in != out {
			t.Errorf("Wanted %08b, got %08b", in, out)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	e := newEncB(b)

	dec := make([]byte, 255)
	enc := make([]byte, e.EncodedLen(255))

	var c byte
	for c = 0; c < 255; c++ {
		dec[c] = c
	}
	for i := 0; i < b.N; i++ {
		e.Encode(enc, dec)
		b.SetBytes(255)
	}
}

func BenchmarkDecode(b *testing.B) {
	e := newEncB(b)

	dec := make([]byte, 255)
	enc := make([]byte, e.EncodedLen(255))

	var c byte
	for c = 0; c < 255; c++ {
		dec[c] = c
	}
	e.Encode(enc, dec)

	for i := 0; i < b.N; i++ {
		e.Decode(dec, enc)
		b.SetBytes(255)
	}
}
