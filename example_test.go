// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml_test

import (
	"bytes"
	"fmt"
	"git.gitorious.org/go-ebml/ebml.git"
	"reflect"
)

func ExampleHeader() {
	var headerA, headerB ebml.Header
	headerA.EBMLVersion = 1
	headerA.EBMLReadVersion = 1
	headerA.EBMLMaxIDLength = 1
	headerA.EBMLMaxSizeLength = 1
	headerA.DocType = "matroska"
	headerA.DocTypeVersion = 1
	headerA.DocTypeReadVersion = 1

	b, _ := ebml.Marshal(headerA)

	src := bytes.NewReader(b)
	dec := ebml.NewDecoder(src)

	dec.Decode(&headerB)

	fmt.Printf("%x DeepEqual: %v", b, reflect.DeepEqual(headerA, headerB))
	// Output:
	// 1a45dfa3a34286810142f7810142f2810142f381014282886d6174726f736b614287810142858101 DeepEqual: true
}

func ExampleMarshal() {
	DoDad := new(struct {
		EbmlId      ebml.Id `ebml:"3f0000"`
		DoHickey    int     `ebml:"4242"`
		ThingaMabob string  `ebml:"4243"`
	})

	DoDad.DoHickey = 14
	DoDad.ThingaMabob = "huzah"

	b, _ := ebml.Marshal(*DoDad)
	fmt.Printf("%x", b)
	// Output:
	// 3f00008c4242810e42438568757a6168
}
