package ebml

import (
	"bytes"
	"io/ioutil"
	"testing"
)

var goldenIDs = []struct {
	in  uint32
	out []byte
}{
	{0, []byte{0}},
	{1, []byte{129}},
	{127, []byte{64, 127}},
	{128, []byte{64, 128}},
	{256, []byte{65, 0}},
	{16382, []byte{127, 254}},
	{16383, []byte{32, 63, 255}},
	{16384, []byte{32, 64, 0}},
	{16385, []byte{32, 64, 1}},
	{32768, []byte{32, 128, 0}},
	{2097150, []byte{63, 255, 254}},
	{2097151, []byte{16, 31, 255, 255}},
	{2097152, []byte{16, 32, 0, 0}},
	{268435453, []byte{31, 255, 255, 253}},
	{268435454, []byte{31, 255, 255, 254}},
}

func TestGoldenIDs(t *testing.T) {
	for _, g := range goldenIDs {
		buf := new(bytes.Buffer)
		enc := NewEncoder(buf)
		enc.EncodeID(g.in)
		enc.Flush()
		got := buf.Bytes()
		if !bytes.Equal(got, g.out) {
			t.Errorf("failed to marshal ID %d, wanted %v, got %v", g.in, g.out, got)
		}
	}
}

func BenchmarkEncodeID(b *testing.B) {
	enc := NewEncoder(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		enc.EncodeID(uint32(i))
	}
}
