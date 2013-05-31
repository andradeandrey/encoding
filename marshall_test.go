package ebml

import (
	"bytes"
	"testing"
)

var goldenIDs = []struct {
	num   uint32
	bytes []byte
}{
	{0x1A45DFA3, []byte{0x1A, 0x45, 0xDF, 0xA3}}, // EBML
	{0x4286, []byte{0x42, 0x86}},                 // EBMLVersion
	{0x42F7, []byte{0x42, 0xF7}},                 // EBMLReadVersion
	{0x42F2, []byte{0x42, 0xF2}},                 // EBMLMaxIDLength
	{0x42F3, []byte{0x42, 0xF3}},                 // EBMLMaxSizeLength
	{0x4282, []byte{0x42, 0x82}},                 // DocType
	{0x4287, []byte{0x42, 0x87}},                 // DocTypeVersion
	{0x4285, []byte{0x42, 0x85}},                 // DocTypeReadVersion
}

func TestGoldenIDs(t *testing.T) {
	for _, g := range goldenIDs {
		i, _ := UnmarshallID(g.bytes)
		if i != g.num {
			t.Errorf("failed to unmarshal ID %v, wanted %d, got %d", g.bytes, g.num, i)
		}
	}
}

var goldenSizes = []struct {
	num   uint64
	bytes []byte
}{
	{0, []byte{0}},
	{1, []byte{129}},
	{2, []byte{130}},
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
		b := MarshallSize(g.num)
		if !bytes.Equal(b, g.bytes) {
			t.Errorf("failed to marshal size %d, wanted %v, got %v", g.num, g.bytes, b)
		}

		s, _ := UnmarshallSize(g.bytes)
		if s != g.num {
			t.Errorf("failed to unmarshal size %v, wanted %d, got %d", g.bytes, g.num, s)
		}
	}
}

func BenchmarkSizeMarshalling(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := MarshallSize(uint64(i))
		UnmarshallSize(b)
	}
}
