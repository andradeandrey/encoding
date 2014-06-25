// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml

// A ebmlError is used to distinguish errors (panics) generated in this package.
type ebmlError string

func (e ebmlError) Error() string { return string(e) }

func encError(s string) {
	panic(ebmlError("ebml encoder: " + s))
}

func decError(s string) {
	panic(ebmlError("ebml decoder: " + s))
}
