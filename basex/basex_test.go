// Copyright (c) 2013 Conformal Systems LLC.
// Use of this source code is governed by an ISC

package basex

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"testing"
)

var stringTests = []struct {
	in  string
	out string
}{
	{"", ""},
	{" ", "Z"},
	{"-", "n"},
	{"0", "q"},
	{"1", "r"},
	{"-1", "4SU"},
	{"11", "4k8"},
	{"abc", "ZiCa"},
	{"1234598760", "3mJr7AoUXx2Wqd"},
	{"abcdefghijklmnopqrstuvwxyz", "3yxU3u1igY8WkgtjK92fbJQCd4BZiiT1v25f"},
	{"00000000000000000000000000000000000000000000000000000000000000", "3sN2THZeE9Eh9eYrwkvZqNstbHGvrxSAM7gXUXvyFQP8XvQLUqNCS27icwUeDT7ckHm4FUHM2mTVh1vbLmk7y"},
}

var invalidStringTests = []struct {
	in  string
	out string
}{
	{"0", ""},
	{"O", ""},
	{"I", ""},
	{"l", ""},
	{"3mJr0", ""},
	{"O3yxU", ""},
	{"3sNI", ""},
	{"4kl8", ""},
	{"0OIl", ""},
	{"!@#$%^&*()-_=+~`", ""},
}

var hexTests = []struct {
	in  string
	out string
}{
	{"61", "2g"},
	{"626262", "a3gV"},
	{"636363", "aPEr"},
	{"73696d706c792061206c6f6e6720737472696e67", "2cFupjhnEsSn59qHXstmK2ffpLv2"},
	{"00eb15231dfceb60925886b67d065299925915aeb172c06647", "1NS17iag9jJgTHD1VXjvLCEnZuQ3rJDE9L"},
	{"516b6fcd0f", "ABnLTmg"},
	{"bf4f89001e670274dd", "3SEo3LWLoPntC"},
	{"572e4794", "3EFU7m"},
	{"ecac89cad93923c02321", "EJDM8drfXA6uyA"},
	{"10c8511e", "Rt5zm"},
	{"00000000000000000000", "1111111111"},
}

func TestBitcoin(t *testing.T) {
	// BitcoinEncode tests
	for x, test := range stringTests {
		res := BitcoinEncoding.EncodeToString([]byte(test.in))
		if string(res) != test.out {
			t.Errorf("BitcoinEncode test #%d failed: got: %q want: %q",
				x, res, test.out)
			continue
		}
	}

	// BitcoinDecode tests
	for x, test := range hexTests {
		b, err := hex.DecodeString(test.in)
		if err != nil {
			t.Errorf("hex.DecodeString failed failed #%d: got: %s", x, test.in)
			continue
		}
		if res, _ := BitcoinEncoding.DecodeString(test.out); bytes.Equal(res, b) != true {
			t.Errorf("BitcoinDecode test #%d failed: got: %q want: %q",
				x, res, test.in)
			continue
		}
	}

	// BitcoinDecode with invalid input
	for x, test := range invalidStringTests {
		if res, _ := BitcoinEncoding.DecodeString(test.in); string(res) != test.out {
			t.Errorf("BitcoinDecode invalidString test #%d failed: got: %q want: %q",
				x, res, test.out)
			continue
		}
	}
}

func TestEncodeMax(t *testing.T) {
	encoding := NewEncoding("abc")

	for i := 0; i < 4; i++ {
		src := make([]byte, i)
		dst := make([]byte, encoding.EncodedLen(i))
		for j := 0; j < i; j++ {
			src[j] = 0xff
		}
		encoding.Encode(dst, src)
		//t.Errorf("%q", dst)
	}
}

func TestDecodeMax(t *testing.T) {
	encoding := NewEncoding("abc")

	for i := 0; i < 4; i++ {
		src := make([]byte, i)
		dst := make([]byte, encoding.DecodedLen(i))
		for j := 0; j < i; j++ {
			src[j] = 'c'
		}
		encoding.Decode(dst, src)
		//t.Errorf("%q", dst)
	}
}

func TestRandom(t *testing.T) {
	control := make([]byte, 1024)
	rand.Read(control)
	raw := make([]byte, 1024)
	encoded := make([]byte, BitcoinEncoding.EncodedLen(1024))

	BitcoinEncoding.Encode(encoded, control)
	BitcoinEncoding.Decode(raw, encoded)
	if !bytes.Equal(raw, control) {
		t.Fatal("random test failed")
	}
}

func TestRandomBase3(t *testing.T) {
	encoding := NewEncoding("abc")

	control := make([]byte, 64)
	rand.Read(control)
	raw := make([]byte, 64)
	encoded := make([]byte, encoding.EncodedLen(64))

	encoding.Encode(encoded, control)
	encoding.Decode(raw, encoded)
	if !bytes.Equal(raw, control) {
		t.Fatal("random test failed")
	}
}

func BenchmarkBitcoinEncoding(b *testing.B) {
	src := make([]byte, 1024)
	dst := make([]byte, BitcoinEncoding.EncodedLen(1024))
	rand.Read(src)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BitcoinEncoding.Encode(dst, src)
	}
}

func BenchmarkBitcoinDecodeing(b *testing.B) {
	control := make([]byte, 1024)
	rand.Read(control)
	dst := make([]byte, 1024)
	src := make([]byte, BitcoinEncoding.EncodedLen(1024))
	BitcoinEncoding.Encode(src, control)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BitcoinEncoding.Decode(dst, src)
	}
}
