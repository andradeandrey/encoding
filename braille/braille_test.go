package braille

import (
	"bytes"
	"testing"
)

type encDecTest struct {
	enc string
	dec []byte
}

var encDecTests = []encDecTest{
	{"⠀", []byte{0x00}},
	{"⠁", []byte{0x01}},
	{"⠂", []byte{0x02}},
	{"⠃", []byte{0x03}},
	{"⠄", []byte{0x04}},
	{"⣿", []byte{0xFF}},
	{"⠀⠀", []byte{0x00, 0x00}},
	{"⣿⣿", []byte{0xFF, 0xFF}},
	{"⣿⣿", []byte{0xFF, 0xFF}},
	{"⠀⠀⠀", []byte{0x00, 0x00, 0x00}},
	{"⣿⣿⣿⣿⣿⣿", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
}

func TestEncode(t *testing.T) {
	for i, test := range encDecTests {
		dst := make([]byte, EncodedLen(len(test.dec)))
		n := Encode(dst, test.dec)
		if n != len(dst) {
			t.Errorf("#%d: bad return value: got: %d want: %d", i, n, len(dst))
		}
		if string(dst) != test.enc {
			t.Errorf("#%d: %X: got: %X want: %X - %q %q", i, test.dec, dst, test.enc, dst, test.enc)
		}
	}
}

func TestDecode(t *testing.T) {
	for i, test := range encDecTests {
		dst := make([]byte, DecodedLen(len(test.enc)))
		n, err := Decode(dst, []byte(test.enc))
		if err != nil {
			t.Error(err)
		}
		if n != len(dst) {
			t.Errorf("#%d: bad return value: got: %d want: %d", i, n, len(dst))
		}
		if !bytes.Equal(dst, test.dec) {
			t.Errorf("#%d: %X: got: %X want: %X - %q %q", i, test.dec, dst, test.enc, dst, test.enc)
		}
	}
}

func TestIterative(t *testing.T) {
	var in, out byte
	enc := make([]byte, 3)
	dec := make([]byte, 1)
	for in = 0; in < 0xFF; in++ {
		dec[0] = in
		Encode(enc, dec)
		Decode(dec, enc)
		out = dec[0]

		if in != out {
			t.Errorf("Wanted %08b, got %08b", in, out)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	dec := make([]byte, 255)
	enc := make([]byte, EncodedLen(255))

	var c byte
	for c = 0; c < 255; c++ {
		dec[c] = c
	}
	for i := 0; i < b.N; i++ {
		Encode(enc, dec)
		b.SetBytes(255)
	}
}

func BenchmarkDecode(b *testing.B) {
	dec := make([]byte, 255)
	enc := make([]byte, EncodedLen(255))

	var c byte
	for c = 0; c < 255; c++ {
		dec[c] = c
	}
	Encode(enc, dec)

	for i := 0; i < b.N; i++ {
		Decode(dec, enc)
		b.SetBytes(255)
	}
}
