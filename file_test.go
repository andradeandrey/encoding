// +build file
// This test is only run with the "file" build tag,
// as it touches the filesystem and needs a test file.

package ebml

import "testing"

func TestFile(t *testing.T) {
	fm, err := NewFileMarshaler("/tmp/test.in")
	if err != nil {
		t.Fatal("NewFileMarshaler:", err)
	}
	x := &struct {
		EbmlId Id        `ebml:"81"`
		F      Marshaler `ebml:"82"`
	}{F: fm}
	b, err := Marshal(x)
	if err != nil {
		t.Fatal("Marshal:", err)
	}

	fu, err := NewFileUnmarshaler("/tmp/test.out")
	if err != nil {
		t.Fatal("NewFileUnmarshaler:", err)
	}
	y := &struct {
		EbmlId Id          `ebml:"81"`
		F      Unmarshaler `ebml:"82"`
	}{F: fu}
	err = Unmarshal(b, y)
	if err != nil {
		t.Fatal("Unmarshal:", err)
	}
}
