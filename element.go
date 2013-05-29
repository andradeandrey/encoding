package ebml

import "io"

type Element interface {
	ID() uint32
	Size() int64
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
	docType            *String
	docTypeVersion     *Uint
	docTypeReadVersion *Uint
}

func newHeader(docType string, docTypeVersion, docTypeReadVersion interface{}) *header {
	return &header{
		NewString(DocTypeID, docType),
		NewUint(DocTypeVersionID, docTypeVersion),
		NewUint(DocTypeReadVersionID, docTypeReadVersion),
	}
}

func (h *header) ID() uint32 { return EBMLID }

func (h *header) Size() (n int64) {
	return h.docType.Size() + h.docTypeVersion.Size() + h.docTypeReadVersion.Size()
}

func (h *header) Next() (e Element) {
	switch {
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
