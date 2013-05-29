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
	{0x1A45DFA3, []byte{0x1A, 0x45, 0xDF, 0xA3}}, // EBML
	{0x4286, []byte{0x42, 0x86}},                 // EBMLVersion
	{0x42F7, []byte{0x42, 0xF7}},                 // EBMLReadVersion
	{0x42F2, []byte{0x42, 0xF2}},                 // EBMLMaxIDLength
	{0x42F3, []byte{0x42, 0xF3}},                 // EBMLMaxSizeLength
	{0x4282, []byte{0x42, 0x82}},                 // DocType
	{0x4287, []byte{0x42, 0x87}},                 // DocTypeVersion
	{0x4285, []byte{0x42, 0x85}},                 // DocTypeReadVersion
	{0xBF, []byte{0xBF}},                         // CRC-32
	{0x3FFF, []byte{0x3F, 0XFF}},
}

func TestGoldenIDs(t *testing.T) {
	for _, g := range goldenIDs {
		buf := new(bytes.Buffer)
		enc := NewEncoder(buf)
		enc.EncodeID(g.in)
		got := buf.Bytes()
		if !bytes.Equal(got, g.out) {
			t.Errorf("failed to marshal ID %d, wanted %v, got %v", g.in, g.out, got)
		}
	}
}

func BenchmarkEncodeID(b *testing.B) {
	enc := NewEncoder(ioutil.Discard)
	for i := 0x10; i < b.N; i++ {
		enc.EncodeID(uint32(i))
	}
}

var goldenSizes = []struct {
	in  int64
	out []byte
}{
	{0, []byte{0}},
	{1, []byte{129}},
	{127, []byte{64, 127}},
	{128, []byte{64, 128}},
	{256, []byte{65, 0}},
	{16382, []byte{32, 63, 254}},
	{16383, []byte{32, 63, 255}},
	{16384, []byte{32, 64, 0}},
	{16385, []byte{32, 64, 1}},
	{32768, []byte{32, 128, 0}},
	{2097150, []byte{16, 31, 255, 254}},
	{2097151, []byte{16, 31, 255, 255}},
	{2097152, []byte{16, 32, 0, 0}},
	{268435454, []byte{8, 15, 255, 255, 254}},
	{268435455, []byte{8, 15, 255, 255, 255}},
	// http://en.wikipedia.org/wiki/Big_Endian#Examples_of_storing_the_value_0A0B0C0Dh_in_memory
	{0x0A0B0C0D, []byte{16 | 0x0A, 0x0B, 0x0C, 0x0D}},
	{1<<56 - 3, []byte{1, 255, 255, 255, 255, 255, 255, 253}},
}

func TestGoldenSizes(t *testing.T) {
	for _, g := range goldenSizes {
		buf := new(bytes.Buffer)
		enc := NewEncoder(buf)
		enc.EncodeSize(g.in)
		got := buf.Bytes()
		if !bytes.Equal(got, g.out) {
			t.Errorf("failed to marshal size %d, wanted %v, got %v", g.in, g.out, got)
		}
	}
}

func BenchmarkEncodeSize(b *testing.B) {
	enc := NewEncoder(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		enc.EncodeSize(int64(i))
	}
}
