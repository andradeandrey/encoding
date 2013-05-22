package ebml

import "io"

type Element struct {
	ID   uint32
	Size uint64
	Data io.Reader
}
