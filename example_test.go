// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml_test

import (
	"fmt"
	"git.gitorious.org/go-ebml/ebml.git"
)

func ExampleNewId() {
	id := ebml.NewId(0x1a45dfa3)
	fmt.Printf("%b, %d, %x, %v", id, id, id, id)
	// Output:
	// 00011010.01000101.11011111.10100011, 440786851, 1a45dfa3, 1a45dfa3
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
