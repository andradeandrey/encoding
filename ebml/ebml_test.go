// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

/*
func ExampleHeader() {
	var headerA, headerB Header
	headerA.EBMLVersion = 1
	headerA.EBMLReadVersion = 1
	headerA.EBMLMaxIDLength = 1
	headerA.EBMLMaxSizeLength = 1
	headerA.DocType = "matroska"
	headerA.DocTypeVersion = 1
	headerA.DocTypeReadVersion = 1

	b, _ := Marshal(headerA)

	src := bytes.NewReader(b)
	dec := NewDecoder(src)

	dec.Decode(&headerB)

	fmt.Printf("0x%x DeepEqual: %v", b, reflect.DeepEqual(headerA, headerB))
	// Output:
	// 0x1a45dfa3a34286810142f7810142f2810142f381014282886d6174726f736b614287810142858101 DeepEqual: true
}

func ExampleMarshal() {
	DoDad := new(struct {
		EbmlId      Id `ebml:"3f0000"`
		DoHickey    uint    `ebml:"4242"`
		ThingaMabob string  `ebml:"4243"`
		HumDinger   int     `ebml:"4244"`
	})

	DoDad.DoHickey = 70000
	DoDad.ThingaMabob = "huzah"
	DoDad.HumDinger = -92387

	b, err := Marshal(*DoDad)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("0x%x\n", b)

	// Output:
	// 0x3f00009442428301117042438568757a6168424483fe971d
}

func ExampleUnmarshal() {
	data := []byte{
		0x3f, 0x00, 0x00, 0x94,
		0x42, 0x42, 0x83, 0x01, 0x11, 0x70,
		0x42, 0x43, 0x85, 0x68, 0x75, 0x7a, 0x61, 0x68,
		0x42, 0x44, 0x83, 0xfe, 0x97, 0x1d,
	}

	DoDad := new(struct {
		EbmlId      Id `ebml:"3f0000"`
		DoHickey    uint    `ebml:"4242"`
		ThingaMabob string  `ebml:"4243"`
		HumDinger   int     `ebml:"4244"`
	})

	err := Unmarshal(data, DoDad)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%v\n", DoDad)
	// Output:
	// &{0 70000 huzah -92387}
}
*/

type intTestStruct struct {
	EbmlId Id  `ebml:"81"`
	I      int `ebml:"88"`
}

func TestInt(t *testing.T) {
	var in, out intTestStruct
	in.I = rand.Int()

	buf, err := Marshal(in)
	if err != nil {
		t.Fatalf("marshal int %d: %s\n buf: %x", in.I, err, buf)
		return
	}

	err = Unmarshal(buf, &out)
	if err != nil && err != io.EOF {
		t.Fatalf("unmarshal int %d: %s\n buf: %x", in.I, err, buf)
		return
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("marshal and unmarshal int: in=%0.2x out=%0.2x", in.I, out.I)
		return
	}

}

func BenchmarkMarshalInt(b *testing.B) {
	var (
		in  intTestStruct
		buf []byte
		err error
	)
	in.I = rand.Int()
	for i := 0; i < b.N; i++ {
		_, err = Marshal(in)
		if err != nil {
			b.Fatalf("marshal int %d: %s\n buf: %x", in.I, err, buf)
		}
	}
}

func BenchmarkUnmarshalInt(b *testing.B) {
	var (
		in, out intTestStruct
		buf     []byte
		err     error
	)
	in.I = rand.Int()
	buf, err = Marshal(in)
	if err != nil {
		b.Fatalf("marshal int %d: %s\n buf: %x", in.I, err, buf)
		return
	}
	for i := 0; i < b.N; i++ {
		err = Unmarshal(buf, &out)
		if err != nil && err != io.EOF {
			b.Fatalf("unmarshal int %d: %s\n buf: %x", in.I, err, buf)
			return
		}
	}
}

type uintTestStruct struct {
	EbmlId Id   `ebml:"81"`
	I      uint `ebml:"88"`
}

func TestUint(t *testing.T) {
	var in, out uintTestStruct

	in.I = uint(rand.Int63())
	buf, err := Marshal(in)
	if err != nil {
		t.Fatalf("marshal uint %d: %s\n buf: %x", in.I, err, buf)
	}

	err = Unmarshal(buf, &out)
	if err != nil && err != io.EOF {
		t.Fatalf("unmarshal uint %d: %s\n buf: %x", in.I, err, buf)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("marshal and unmarshal uint: in=%0.2x out=%0.2x", in.I, out.I)
	}
}

func BenchmarkMarshalUint(b *testing.B) {
	var (
		in  uintTestStruct
		err error
	)

	in.I = uint(rand.Int63())
	for i := 0; i < b.N; i++ {
		if _, err = Marshal(in); err != nil {
			b.Fatalf("marshal uint %d: %s\n buf: %x", in.I, err)
		}
	}
}

func BenchmarkUnmarshalUint(b *testing.B) {
	var (
		in, out uintTestStruct
		buf     []byte
		err     error
	)

	in.I = uint(rand.Int63())
	buf, err = Marshal(in)
	if err != nil {
		b.Fatalf("marshal uint %d: %s\n buf: %x", in.I, err, buf)
	}
	for i := 0; i < b.N; i++ {
		if err = Unmarshal(buf, &out); err != nil && err != io.EOF {
			b.Fatalf("unmarshal uint %d: %s\n buf: %x", in.I, err, buf)
		}
	}
}

type floatTestStruct struct {
	EbmlId Id      `ebml:"81"`
	F      float32 `ebml:"88"`
	FF     float64 `ebml:"89"`
}

func TestFloat(t *testing.T) {
	var in, out floatTestStruct

	in.F, in.FF = rand.Float32(), rand.Float64()
	buf, err := Marshal(in)
	if err != nil {
		t.Fatalf("encode floats %f %f: %s\nbuf: %x", in.F, in.FF, err, buf)
		return
	}
	if err = Unmarshal(buf, &out); err != nil {
		t.Fatalf("decode floats %f %f: %s\nbuf: %x", in.FF, in.FF, err, buf)
		return
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("encode/decode float mismatch: in=%f,%f out=%f,%f\n%x", in.F, in.FF, out.F, out.FF, buf)
		return
	}
}

func BenchmarkMarshalFloat(b *testing.B) {
	var (
		in  floatTestStruct
		err error
	)
	in.F, in.FF = rand.Float32(), rand.Float64()

	for i := 0; i < b.N; i++ {
		if _, err = Marshal(in); err != nil {
			b.Fatalf("encode floats %f %f: %s\nbuf: %x", in.F, in.FF, err)
		}
	}
}

func BenchmarkUnmarshalFloat(b *testing.B) {
	var (
		in, out floatTestStruct
		buf     []byte
		err     error
	)
	in.F, in.FF = rand.Float32(), rand.Float64()
	if buf, err = Marshal(in); err != nil {
		b.Fatalf("encode floats %f %f: %s\nbuf: %x", in.F, in.FF, err, buf)
		return
	}
	for i := 0; i < b.N; i++ {
		if err = Unmarshal(buf, &out); err != nil {
			b.Fatalf("decode floats %f %f: %s", in.F, in.FF, err)
		}
	}
}

type dateTestStruct struct {
	EbmlId Id        `ebml:"81"`
	D      time.Time `ebml:"82"`
}

func TestDate(t *testing.T) {
	now := time.Now()
	I := rand.Int63n(1 << 62)

	var in, out dateTestStruct

	for i := 0; i < 32; i++ {
		date := now.Add(time.Duration(I))

		in.D = date
		buf, err := Marshal(in)
		if err != nil {
			t.Fatalf("marshal time.Time %v: %s\n buf: %x", date, err, buf)
		}

		err = Unmarshal(buf, &out)
		if err != nil && err != io.EOF {
			t.Fatalf("unmarshal time.Time %v: %s\n buf: %x", date, err, buf)
		}

		if !in.D.Equal(out.D) {
			t.Fatalf("marshal and unmarshal time.Time: in=%v out=%v", in.D, out.D)
		}
	}
}

type stringTest struct {
	EbmlId Id     `ebml:"81"`
	S      string `ebml:"82"`
}

func TestString(t *testing.T) {
	var in, out stringTest
	in.S = "wrong me not good sister, nor wrong yourself"
	buf, err := Marshal(in)
	if err != nil {
		t.Fatal("marshal string failed:", err)
	}
	if err = Unmarshal(buf, &out); err != nil {
		t.Fatalf("unmarshal string failed: %s\n %x", err, buf)
	}

	if in.S != out.S {
		t.Fatalf("%s != %s\n%x", in.S, out.S, buf)
	}
}

type benchTestStruct struct {
	EbmlId Id      `ebml:"81"`
	A      uint    `ebml:"4011"`
	B      int64   `ebml:"200011"`
	C      float64 `ebml:"10000011"`
	D      string  `ebml:"3FFFFFFE"`
}

func BenchmarkEncoding(b *testing.B) {
	var test benchTestStruct
	test.A = uint(rand.Int63())
	test.B = rand.Int63()
	test.C = rand.Float64()
	test.D = "of ships and sails and whether pigs have wings"

	var err error
	b.ResetTimer()
	enc := NewEncoder(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		err = enc.Encode(test)
		if err != nil {
			b.Fatal(err.Error())
		}
	}
}

func BenchmarkDecoding(b *testing.B) {
	control := new(benchTestStruct)

	control.A = uint(rand.Int63())
	control.B = rand.Int63()
	control.C = rand.Float64()
	control.D = "of ships and sails and whether pigs have wings"

	buf, err := Marshal(control)
	r := bytes.NewReader(buf)
	b.ResetTimer()

	dec := NewDecoder(r)
	for i := 0; i < b.N; i++ {
		var test benchTestStruct
		err = dec.Decode(&test)
		if err != nil {
			b.Fatal(err.Error())
		}

		if !reflect.DeepEqual(control, test) {
			b.Fatalf("not equal\n%v\n%v", control, test)
		}
		r.Seek(0, 0)
	}
}
