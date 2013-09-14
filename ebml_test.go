// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml

import (
	"bytes"
	"reflect"
	"testing"
)

func TestMarshal(t *testing.T) {
	var headerA Header
	headerA.EBMLVersion = 1
	headerA.EBMLReadVersion = 1
	headerA.EBMLMaxIDLength = 1
	headerA.EBMLMaxSizeLength = 1
	headerA.DocType = "matroska"
	headerA.DocTypeVersion = 1
	headerA.DocTypeReadVersion = 1

	b, err := Marshal(headerA)
	if err != nil {
		t.Fatal("Marshal:", err)
	}

	src := bytes.NewReader(b)
	dec := NewDecoder(src)

	var headerB Header
	err = dec.Decode(&headerB)
	if err != nil {
		t.Fatal("Unmarshal:", err)
	}

	if !reflect.DeepEqual(headerA, headerB) {
		t.Fatalf("Marshal -> Unmarshal: marshaled %v to %x, but unmarshaled %v", headerA, b, headerB)
	}
}
