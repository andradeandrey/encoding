package ebml

import (
	"io"
)

type Element interface {
	ID() []byte
	Size() uint64
}

type Container interface {
	Element
	Next() Element
}

type Value interface {
	Element
	io.Reader
}

type header struct {
	version            *Uint
	readVersion        *Uint
	maxIDLength        *Uint
	maxSizeLength      *Uint
	docType            *String
	docTypeVersion     *Uint
	docTypeReadVersion *Uint
}

func newHeader(docType string, docTypeVersion, docTypeReadVersion interface{}) *header {
	return &header{
		NewUint(EBMLVersionID, uint32(1)),
		NewUint(EBMLReadVersionID, uint32(1)),
		NewUint(EBMLMaxIDLengthID, uint32(4)),
		NewUint(EBMLMaxSizeLengthID, uint32(8)),
		NewString(DocTypeID, docType),
		NewUint(DocTypeVersionID, docTypeVersion),
		NewUint(DocTypeReadVersionID, docTypeReadVersion),
	}
}

func (h *header) ID() []byte { return EBMLID }

func (h *header) Size() uint64 {
	return h.docType.Size() + h.docTypeVersion.Size() + h.docTypeReadVersion.Size()
}

func (h *header) Next() (e Element) {
	switch {
	case h.version != nil:
		e = h.version
		h.version = nil

	case h.readVersion != nil:
		e = h.readVersion
		h.readVersion = nil

	case h.maxIDLength != nil:
		e = h.maxIDLength
		h.maxIDLength = nil

	case h.maxSizeLength != nil:
		e = h.maxSizeLength
		h.maxSizeLength = nil

	case h.docType != nil:
		e = h.docType
		h.docType = nil

	case h.docTypeVersion != nil:
		e = h.docTypeVersion
		h.docTypeVersion = nil

	case h.docTypeReadVersion != nil:
		e = h.docTypeReadVersion
		h.docTypeReadVersion = nil
	}
	return
}
