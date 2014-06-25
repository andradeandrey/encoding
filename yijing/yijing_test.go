package yijing

import (
	"testing"
)

type encDecTest struct {
	enc string
	dec []byte
}

var encDecTests = []encDecTest{
	{"", []byte{}},
	{"䷀ǀǀ", []byte{0x00}},
	{"䷿¦¦", []byte{0xFF}},
	{"䷀䷀☰ǀ", []byte{0x00, 0x00}},
	{"䷿䷿☷¦", []byte{0xFF, 0xFF}},
	{"䷀䷀䷀䷀", []byte{0x00, 0x00, 0x00}},
	{"䷕䷕䷕䷕", []byte{0x55, 0x55, 0x55}},
	{"䷪䷪䷪䷪", []byte{0xAA, 0xAA, 0xAA}},
	{"䷿䷿䷿䷿", []byte{0xFF, 0xFF, 0xFF}},
	{"䷀䷀䷀䷀䷀䷀䷀䷀", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"䷿䷿䷿䷿䷿䷿䷿䷿", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
}

func TestEncode(t *testing.T) {
	for i, test := range encDecTests {
		dst := make([]byte, EncodedLen(len(test.dec)))
		n := Encode(dst, test.dec)
		if n != len(dst) {
			t.Errorf("#%d: bad return value: got: %d want: %d", i, n, len(dst))
		}
		if string(dst) != test.enc {
			t.Errorf("#%d: got: %q want: %q", i, dst, test.enc)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	var c byte
	for i := 0; i < b.N; i++ {
		srcBuf := make([]byte, 27)
		dstBuf := make([]byte, EncodedLen(27))
		for l := 0; l < 28; l++ {
			src := srcBuf[:l]
			for ; c < 255; c++ {
				for j := 0; j < l; j++ {
					src[j] = c
					Encode(dstBuf, src)
				}
			}
		}
	}
}
