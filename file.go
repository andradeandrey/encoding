// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml

import (
	"bufio"
	"io"
	"os"
)

type fileAbstraction struct {
	filename string
	info     os.FileInfo
}

// NewFileMarshaler returns a Marshaler that reads from a file
// during encoding.
func NewFileMarshaler(filename string) (Marshaler, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	return &fileAbstraction{filename, info}, nil
}

func (f *fileAbstraction) MarshalEBML() (size int64, wt io.WriterTo) {
	return f.info.Size(), f
}

func (f *fileAbstraction) WriteTo(w io.Writer) (written int64, err error) {
	var file *os.File
	file, err = os.Open(f.filename)
	if err != nil {
		return
	}
	defer file.Close()
	// let bufio handle the details
	return bufio.NewReader(file).WriteTo(w)
}

// NewFileUnmarshaler returns an Unmarshaler that writes to a file
// during decoding.
func NewFileUnmarshaler(filename string) (Unmarshaler, error) {
	info, err := os.Stat(filename)
	if err == nil {
		return &fileAbstraction{filename, info}, nil
	}

	// create the file to check write permission before decoding.
	if !os.IsNotExist(err) {
		return nil, err
	}

	var file *os.File
	file, err = os.OpenFile(filename, os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// try again
	info, err = os.Stat(filename)
	if err != nil {
		return nil, err
	}
	return &fileAbstraction{filename, info}, nil
}

func (f *fileAbstraction) UnmarshalEBML(n int64) io.ReaderFrom {
	if f.info.Size() != n {
		err := os.Truncate(f.filename, n)
		if err != nil {
			panic(err)
		}
	}
	return f
}

func (f *fileAbstraction) ReadFrom(r io.Reader) (readed int64, err error) {
	var file *os.File
	file, err = os.OpenFile(f.filename, os.O_WRONLY, 0)
	if err != nil {
		return
	}
	defer file.Close()
	// let bufio handle the details
	return bufio.NewWriter(file).ReadFrom(r)
}

var (
	_ Marshaler   = (*fileAbstraction)(nil)
	_ Unmarshaler = (*fileAbstraction)(nil)
)
