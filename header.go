// Copyright (c) 2013, Emery Hemingway. All rights reserved.
// Actully most of it comes from encoding/json, courtesy of
// The Go Authors

package ebml

// Header is a struct for encoding and decoding EBML streams.
//
// If nesting in a struct, it's field should be listed first 
// and should should have an ebml tag set to 1a45dfa3.
// 
// The following could be a valid top-level struct for 
// representing Matroska streams:
//	type Matroska struct {
//		ebml.Header `ebml:"1a45dfa3"`
//		Segment     []Segment `ebml:"18538067"`
//	}
//
// You will however need to populate field values in Header
// to form a valid EBML document.
type Header struct {
	EBMLVersion        uint   `ebml:"4286"`
	EBMLReadVersion    uint   `ebml:"42f7"`
	EBMLMaxIDLength    uint   `ebml:"42f2"`
	EBMLMaxSizeLength  uint   `ebml:"42f3"`
	DocType            string `ebml:"4282"`
	DocTypeVersion     uint   `ebml:"4287"`
	DocTypeReadVersion uint   `ebml:"4285"`
}
