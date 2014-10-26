package base256_test

import (
	"crypto/sha256"
	"fmt"
	"github.com/ehmry/encoding/base256"
)

func ExampleNewEncoding() {
	h := sha256.New()
	digest := h.Sum(nil)

	fmt.Printf("%x\n", digest)

	e, _ := base256.NewEncoding('⠀') // U+2800 BRAILLE PATTERN BLANK
	fmt.Println(e.EncodeToString(digest))

	e, _ = base256.NewEncoding('⤀')
	fmt.Println(e.EncodeToString(digest))

	// Output:
	// e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
	// ⣣⢰⣄⡂⢘⣼⠜⠔⢚⣻⣴⣈⢙⡯⢹⠤⠧⢮⡁⣤⡤⢛⢓⡌⢤⢕⢙⠛⡸⡒⢸⡕
	// ⧣⦰⧄⥂⦘⧼⤜⤔⦚⧻⧴⧈⦙⥯⦹⤤⤧⦮⥁⧤⥤⦛⦓⥌⦤⦕⦙⤛⥸⥒⦸⥕

}
