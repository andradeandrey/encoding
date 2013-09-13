package ebml

import "fmt"

func ExampleNewId() {
	id := NewId(0x1a45dfa3)
	fmt.Printf("%b, %d, %x, %v", id, id, id, id)
	// Output:
	// 00011010.01000101.11011111.10100011, 440786851, 1a45dfa3, 1a45dfa3
}
