package ebml

import (
	"bufio"
	"io"
)

// An Encoder writes EBML data to an output stream.
type Encoder struct {
	*bufio.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{Writer: bufio.NewWriter(w)}
}

