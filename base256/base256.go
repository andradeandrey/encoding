// Copyright 2014 Emery Hemingway. All rights reserved.
// Use of this source code is governed by the GPLv3+.

// Package base256 implements radix 256
// encoding and decoding using UTF-8.
package base256

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

type Encoding []byte

var InvalidEncoding = errors.New("Invalid Encoding")

type InvalidZero rune

func (e InvalidZero) Error() string {
	return fmt.Sprintf("base256: invalid zero rune %q", rune(e))
}

// NewEncoding generates a new Encoding using r as 0x00.
func NewEncoding(r rune) (Encoding, error) {
	if !utf8.ValidRune(r) {
		return nil, InvalidEncoding
	}

	b := make([]byte, utf8.UTFMax)
	l := utf8.EncodeRune(b, r)
	b = b[:l]
	if b[2] != 0x80 {
		return nil, InvalidZero(r)
	}
	return b, nil
}

// EncodedLen returns the length of an encoding of n source bytes.
func (e Encoding) EncodedLen(n int) int { return n * len(e) }

// Encode encodes src into EncodedLen(len(src))
// bytes of dst.  As a convenience, it returns the number
// of bytes written to dst, but this value is always
// EncodedLen(len(src)).
// Encode implements Base256 encoding.
func (e Encoding) Encode(dst, src []byte) (n int) {
	var i int
	for i = 0; i < len(dst); i += len(e) {
		copy(dst[i:], e)
	}

	var j = len(e) - 2
	var k = len(e) - 1
	for i = 0; i < len(src); i++ {
		dst[j] |= (src[i] >> 6)
		dst[k] |= (src[i]) & 0x3F
		j += len(e)
		k += len(e)
	}
	return len(src) * 3
}

// InvalidRuneError values describe errors resulting from an invalid byte in a Base256 string.
type InvalidRuneError rune

func (e InvalidRuneError) Error() string {
	return fmt.Sprintf("base256: invalid rune: %#U", rune(e))
}

var ErrLength = errors.New("base256: bad length")

func (e Encoding) DecodedLen(x int) int { return x / len(e) }

// Decode has weak error checking, use utf8.Valid before
// decoding if corruption is a concern.
func (e Encoding) Decode(dst, src []byte) (int, error) {
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
func (e Encoding) EncodeToString(src []byte) string {
	dst := make([]byte, e.EncodedLen(len(src)))
	e.Encode(dst, src)
	return string(dst)
}

// DecodeString returns the bytes represented by the base256 string s.
func (e Encoding) DecodeString(s string) ([]byte, error) {
	src := []byte(s)
	dst := make([]byte, e.DecodedLen(len(src)))
	_, err := e.Decode(dst, src)
	if err != nil {
		return nil, err
	}
	return dst, nil
}
