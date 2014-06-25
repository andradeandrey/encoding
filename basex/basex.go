// Copyright (c) 2013 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// Package basex implements encoding to an alphabet with an arbitrary length.
package basex

import (
	"bytes"
	"fmt"
	"math/big"
	//"math"
)

// An Encoding is an encoding/decoding scheme defined by an alphabet.
type Encoding struct {
	alphabet []byte
	radix    *big.Int
}

const encodeBitcoin = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// NewEncoding returns a new Encoding defined by the given alphabet.
func NewEncoding(alphabet string) *Encoding {
	b := []byte(alphabet)
	return &Encoding{b, big.NewInt(int64(len(b)))}
}

var BitcoinEncoding = NewEncoding(encodeBitcoin)

var bigZero = big.NewInt(0)

// EncodedLen returns a guaranteed maximum length in bytes an encoding can consume for
// an input buffer of length n.
//
// It estimates very poorly.
func (enc *Encoding) EncodedLen(n int) int {
	if n == 0 {
		return 0
	}

	// not right
	return (0xff / len(enc.alphabet)) * n
}

// EncodedLen returns a guaranteed maximum length in bytes an encoding can represent for
// an input buffer of length n.
//
// It estimates very poorly.
func (enc *Encoding) DecodedLen(n int) int {
	if n == 0 {
		return 0
	}
	// not right
	return (len(enc.alphabet) * n)

}

// Encode encodes src using the encoding enc, writing
// EncodedLen(len(src)) bytes to dst.
func (enc *Encoding) Encode(dst, src []byte) {
	if len(src) == 0 {
		return
	}

	x := new(big.Int)
	x.SetBytes(src)

	var i int
	for x.Cmp(bigZero) > 0 {
		mod := new(big.Int)
		x.DivMod(x, enc.radix, mod)
		dst[i] = enc.alphabet[mod.Int64()]
		i++
	}

	// reverse
	for j := 0; j < i; j++ {
		i--
		dst[j], dst[i] = dst[i], dst[j]
	}
}

// EncodeToString returns the encoded form of src.
func (enc *Encoding) EncodeToString(src []byte) string {
	buf := make([]byte, enc.EncodedLen(len(src)))
	enc.Encode(buf, src)
	if i := bytes.IndexByte(buf, 0x00); i != -1 {
		return string(buf[:i])
	}
	return string(buf)
}

// Decode decodes src using the encoding enc.  It writes at most
// DecodedLen(len(src)) bytes to dst and returns the number of bytes
// written. If src contains invalid encodings, it will return CorruptInputError.
func (enc *Encoding) Decode(dst, src []byte) (n int, err error) {
	answer := big.NewInt(0)
	j := big.NewInt(1)

	for i := len(src) - 1; i >= 0; i-- {
		tmp := bytes.IndexByte(enc.alphabet, src[i])
		if tmp == -1 {
			if src[i] == 0x00 {
				continue
			}
			return 0, fmt.Errorf("character %q not in alphabet", src[i])
		}
		idx := big.NewInt(int64(tmp))
		tmp1 := big.NewInt(0)
		tmp1.Mul(j, idx)

		answer.Add(answer, tmp1)
		j.Mul(j, enc.radix)
	}

	tmp := answer.Bytes()
	var numZeros int
	for ; numZeros < len(src); numZeros++ {
		if src[numZeros] != enc.alphabet[0] {
			break
		}
	}
	return copy(dst[numZeros:], tmp) + numZeros, nil
}

func (enc *Encoding) DecodeString(s string) ([]byte, error) {
	src := []byte(s)
	dst := make([]byte, enc.DecodedLen(len(src)))
	n, err := enc.Decode(dst, src)
	return dst[:n], err
}
