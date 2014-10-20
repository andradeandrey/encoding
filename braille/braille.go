// Copyright 2014 Emery Hemingway. All rights reserved.
// Use of this source code is governed by the GPLv3+.

// Package braille implements radix 256 encoding and decoding using
// the UTF-8 Braille encoding.
package braille

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// EncodedLen returns the length of an encoding of n source bytes.
func EncodedLen(n int) int { return n * 3 }

// Encode encodes src into EncodedLen(len(src))
// bytes of dst.  As a convenience, it returns the number
// of bytes written to dst, but this value is always
// EncodedLen(len(src)).
// Encode implements Braille encoding.
func Encode(dst, src []byte) (n int) {
	var i int
	for i = 0; i < len(dst); i += 3 {
		dst[i] = 0xE2
	}

	var j = 1
	var k = 2
	for i = 0; i < len(src); i++ {
		dst[j] = 0xA0 | (src[i] >> 6)
		dst[k] = 0x80 | (src[i])&0x3F
		j += 3
		k += 3
	}
	return len(src) * 3
}

// InvalidRuneError values describe errors resulting from an invalid byte in a Braille string.
type InvalidRuneError rune

func (e InvalidRuneError) Error() string {
	return fmt.Sprintf("braille: invalid rune: %#U", rune(e))
}

var ErrLength = errors.New("braille: bad length")

func DecodedLen(x int) int { return x / 3 }

func Decode(dst, src []byte) (int, error) {
	if len(dst) < len(src)/3 {
		return 0, ErrLength
	}

	var i int
	for i = 0; i < len(src); i += 3 {
		if src[i] != 0xE2 {
			r, _ := utf8.DecodeRune(src[i:])

			return 0, InvalidRuneError(r)
		}
	}
	var j = 1
	var k = 2
	for i = 0; i < len(dst); i++ {
		dst[i] = (src[j] << 6) | (src[k] ^ 0x80)
		j += 3
		k += 3
	}
	return len(src) / 3, nil
}

// EncodeToString returns the bralle encoding of src.
func EncodeToString(src []byte) string {
	dst := make([]byte, EncodedLen(len(src)))
	Encode(dst, src)
	return string(dst)
}

// DecodeString returns the bytes represented by the braille string s.
func DecodeString(s string) ([]byte, error) {
	src := []byte(s)
	dst := make([]byte, DecodedLen(len(src)))
	_, err := Decode(dst, src)
	if err != nil {
		return nil, err
	}
	return dst, nil
}
