// idtest is a utility for checking the validity of EBML ids
package main

import (
	"fmt"
	"github.com/3M3RY/go-ebml"
	"os"
)

func main() {
	for _, s := range os.Args[1:] {
		id, err := ebml.NewIdFromString(s)
		if err != nil {
			fmt.Println(err)
			continue
		}

		b := id.Bytes()
		c := b[0]
		var good bool
		switch l := len(b); l {
		case 1:
			good = c <= 0xff && c > 0x80
		case 2:
			good = c <= 0x7f && c > 0x40
		case 3:
			good = c <= 0x3f && c > 0x20
		case 4:
			good = c <= 0x1f && c > 0x10
		}

		if good {
			fmt.Print("Id is good: ")
		} else {
			fmt.Print("Id is bad:  ")
		}
		fmt.Printf("% 11s ", id)
		for _, c := range b {
			fmt.Printf(".%08b", c)
		}
		fmt.Println()
	}
}
