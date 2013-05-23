package ebml

import "io"

const (
	HeaderID = 0xa45dfa3
	DocTypeID = 0x4282
	DocTypeVersionID = 0x4287
	DocTypeReadVersionID = 0x4285
)

type Element interface {
	ID() uint32
	Size() uint64
	io.Reader
}

type Container interface {
	ID() uint32
	Size() uint64
	Next() Element
}

type header struct {
	docType            *String
	docTypeVersion     *Uint
	docTypeReadVersion *Uint
}

func newHeader(docType string, docTypeVersion, docTypeReadVersion uint64) *header {
	return &header{
		NewString(DocTypeID, docType),
		NewUint(DocTypeVersionID, docTypeVersion),
		NewUint(DocTypeReadVersionID, docTypeReadVersion),
	}
}

func (h *header) ID() uint32 { return HeaderID }

func (h *header) Size() (n uint64) {
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
