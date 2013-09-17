// Copyright © 2013 Emery Hemingway
// Released under the terms of the GNU Public License version 3

package ebml_test

import (
	"bytes"
	"fmt"
	"github.com/3M3RY/go-ebml"
	"io"
	"io/ioutil"
	"math/rand"
	"reflect"
	"testing"
	"time"
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

	fmt.Printf("0x%x DeepEqual: %v", b, reflect.DeepEqual(headerA, headerB))
	// Output:
	// 0x1a45dfa3a34286810142f7810142f2810142f381014282886d6174726f736b614287810142858101 DeepEqual: true
}

func ExampleMarshal() {
	DoDad := new(struct {
		EbmlId      ebml.Id `ebml:"3f0000"`
		DoHickey    uint    `ebml:"4242"`
		ThingaMabob string  `ebml:"4243"`
		HumDinger   int     `ebml:"4244"`
	})

	DoDad.DoHickey = 70000
	DoDad.ThingaMabob = "huzah"
	DoDad.HumDinger = -92387

	b, err := ebml.Marshal(*DoDad)
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
		EbmlId      ebml.Id `ebml:"3f0000"`
		DoHickey    uint    `ebml:"4242"`
		ThingaMabob string  `ebml:"4243"`
		HumDinger   int     `ebml:"4244"`
	})

	err := ebml.Unmarshal(data, DoDad)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%v\n", DoDad)
	// Output:
	// &{0 70000 huzah -92387}
}

type intTestStruct struct {
	EbmlId ebml.Id `ebml:"81"`
	I      int     `ebml:"88"`
}

func TestInt(t *testing.T) {
	var in, out intTestStruct
	var buf []byte
	var err error
	I := 2
	for i := 0; i < 107; i++ {
		I = int(float64(I) * -1.5)
		in.I = I
		buf, err = ebml.Marshal(in)
		if err != nil {
			t.Fatalf("marshal int %d: %s\n buf: %x", I, err, buf)
		}

		err = ebml.Unmarshal(buf, &out)
		if err != nil && err != io.EOF {
			t.Fatalf("unmarshal int %d: %s\n buf: %x", I, err, buf)
		}

		if !reflect.DeepEqual(in, out) {
			t.Fatalf("marshal and unmarshal int %d: in=%0.2x out=%0.2x", I, in.I, out.I)
		}
	}
}

type uintTestStruct struct {
	EbmlId ebml.Id `ebml:"81"`
	I      uint    `ebml:"88"`
}

func TestUint(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var in, out uintTestStruct
	var buf []byte
	var err error
	var I uint
	for i := 0; i < 256; i++ {
		I = uint(rand.Int63())
		in.I = I
		buf, err = ebml.Marshal(in)
		if err != nil {
			t.Fatalf("marshal uint %d: %s\n buf: %x", I, err, buf)
		}

		err = ebml.Unmarshal(buf, &out)
		if err != nil && err != io.EOF {
			t.Fatalf("unmarshal uint %d: %s\n buf: %x", I, err, buf)
		}

		if !reflect.DeepEqual(in, out) {
			t.Fatalf("marshal and unmarshal uint %d: in=%0.2x out=%0.2x", I, in.I, out.I)
		}
	}
}

type dateTestStruct struct {
	EbmlId ebml.Id   `ebml:"81"`
	D      time.Time `ebml:"82"`
}

func TestDate(t *testing.T) {
	now := time.Now()
	rand.Seed(now.UnixNano())
	I := rand.Int63()

	var in, out dateTestStruct
	

	for i := 0; i < 256; i++ {
		I = int64(float32(I) * 1.5)
		date := now.Add(time.Duration(I))

		in.D = date
		buf, err := ebml.Marshal(in)
		if err != nil {
			t.Fatalf("marshal time.Time %v: %s\n buf: %x", date, err, buf)
		}

		err = ebml.Unmarshal(buf, &out)
		if err != nil && err != io.EOF {
			t.Fatalf("unmarshal time.Time %v: %s\n buf: %x", date, err, buf)
		}

		if !in.D.Equal(out.D) {
			t.Fatalf("marshal and unmarshal time.Time: in=%v out=%v", in.D, out.D)
		}
	}
}

type benchTestStruct struct {
	EbmlId ebml.Id `ebml:"81"`
	A      uint    `ebml:"4011"`
	B      int64   `ebml:"200011"`
	C      int64   `ebml:"10000011"`
	D      string  `ebml:"800000000"`
}

func BenchmarkEncoding(b *testing.B) {
	var test benchTestStruct
	rand.Seed(time.Now().UnixNano())

	test.A = uint(rand.Int63())
	test.B = rand.Int63()
	test.C = 1 - rand.Int63()
	test.D = "of ships and sails and whether pigs have wings"

	var err error
	b.ResetTimer()
	enc := ebml.NewEncoder(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		err = enc.Encode(test)
		if err != nil {
			b.Fatal(err.Error())
		}
	}
}

func BenchmarkDecoding(b *testing.B) {
	var control benchTestStruct
	rand.Seed(time.Now().UnixNano())

	control.A = uint(rand.Int63())
	control.B = rand.Int63()
	control.C = 1 - rand.Int63()
	control.D = "of ships and sails and whether pigs have wings"

	buf, err := ebml.Marshal(control)
	r := bytes.NewReader(buf)
	b.ResetTimer()

	dec := ebml.NewDecoder(r)
	for i := 0; i < b.N; i++ {
		var test benchTestStruct
		err = dec.Decode(&test)
		if err != nil {
			b.Fatal(err.Error())
		}

		if !reflect.DeepEqual(control, test) {
			b.Fatal("not equal")
		}
		r.Seek(0, 0)
	}
}
