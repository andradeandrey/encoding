// Copyright 2014 Emery Hemingway. All rights reserved.
// Use of this source code is governed by the GPLv3+.

// Package yijing64 implements radix 64 Yijing encoding and decoding.
package yijing

// ䷀ 0xE4 0xB7 0x80
// ䷿ 0xE4 0xB7 0xBF

// ☰ 0xE2 0x98 0xb0
// ☷ 0xE2 0x98 0xb8

// ǀ 0xC7 0x80
// ¦ 0xC2 0xa6

// full block is 0xE2, 0x96, 0x88

// EncodedLen returns the length of an encoding of n source bytes.
func EncodedLen(n int) (l int) {
	if n == 0 {
		return 0
	}

	l = ((n * 4) / 3) * 3

	switch n % 3 {
	case 0:
	case 1:
		l += 4
	case 2:
		l += 5
	}
	return l
}

// Encode encodes src into EncodedLen(len(src))
// bytes of dst.  As a convenience, it returns the number
// of bytes written to dst, but this value is always EncodedLen(len(src)).
// Encode implements Yi Jing encoding.
func Encode(dst, src []byte) (n int) {
	for len(src) > 0 {
		dst[0] = 0xE4
		dst[1] = 0xB7
		dst[2] = 0x80 | (src[0] >> 2)

		switch len(src) {

		case 2:
			dst[3] = 0xE4
			dst[4] = 0xB7
			dst[5] = 0x80 | ((src[0] & 0x03) << 4) | (src[1] >> 4)

			dst[6] = 0xE2
			dst[7] = 0x98
			dst[8] = 0xB0 | ((src[1] & 0x0F) >> 1)

			if (src[1] & 0xFE) == 0 {
				dst[9] = 0xC7
				dst[10] = 0x80
			} else {
				dst[9] = 0xC2
				dst[10] = 0xA6
			}
			return n + 11

		case 1:
			if (src[0] & 0x02) == 0 {
				dst[3] = 0xC7
				dst[4] = 0x80
			} else {
				dst[3] = 0xC2
				dst[4] = 0xA6
			}

			if (src[0] & 0x01) == 0 {
				dst[5] = 0xC7
				dst[6] = 0x80
			} else {
				dst[5] = 0xC2
				dst[6] = 0xA6
			}
			return n + 7
		}

		dst[3] = 0xE4
		dst[4] = 0xB7
		dst[5] = 0x80 | ((src[0] & 0x03) << 4) | (src[1] >> 4)

		dst[6] = 0xE4
		dst[7] = 0xB7
		dst[8] = 0x80 | ((src[1] & 0x0F) << 2) | (src[2] >> 4)

		dst[9] = 0xE4
		dst[10] = 0xB7
		dst[11] = 0x80 | (src[2] & 0x3F)

		n += 12

		src = src[3:]
		dst = dst[12:]
	}
	return n
}

// EncodeToString returns the Yi Jing encoding of src.
func EncodeToString(src []byte) string {
	dst := make([]byte, EncodedLen(len(src)))
	Encode(dst, src)
	return string(dst)
}
