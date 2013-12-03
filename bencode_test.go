package bencode

import (
	"bytes"
	"reflect"
	"testing"
)

type test struct {
	in  string
	ptr interface{}
	out interface{}
	err error
}

type Ambig struct {
	// Given "hello", the first match should win.
	First  int `bencode:"HELLO"`
	Second int `bencode:"Hello"`
}

var tests = []test{
	// basic types
	//{in: `i1e`, ptr: new(bool), out: true},
	{in: `i1e`, ptr: new(int), out: 1},
	{in: `i2e`, ptr: new(int64), out: int64(2)},
	{in: `i-5e`, ptr: new(int16), out: int16(-5)},
	{in: `i2e`, ptr: new(interface{}), out: int64(2)},
	{in: "i0e", ptr: new(interface{}), out: int64(0)},
	{in: "i0e", ptr: new(int), out: 0},
	{in: "1:a", ptr: new(string), out: "a"},
	{in: "2:a\"", ptr: new(string), out: "a\""},
	{in: "11:0123456789a", ptr: new(interface{}), out: []byte("0123456789a")},
	{in: "le", ptr: new([]int64), out: []int64{}},
	{in: "li1ei2ee", ptr: new([]int), out: []int{1, 2}},
	{in: "l3:abc3:defe", ptr: new([]string), out: []string{"abc", "def"}},
	//{in: "li42e3:abce", ptr: new([]interface{}), out: []interface{}{42, []byte("abc")}},
	{in: "de", ptr: new(map[string]interface{}), out: make(map[string]interface{})},
	{in: "d3:cati1e3:dogi2ee", ptr: new(map[string]int), out: map[string]int{"cat": 1, "dog": 2}},
}

func TestMarshal(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	for i, tt := range tests {
		buf.Reset()
		var scan scanner
		in := []byte(tt.in)
		if err := checkValid(in, &scan); err != nil {
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("#%d: checkValid: %#v", i, err)
				continue
			}
		}
		if err := enc.Encode(tt.out); err != nil {
			t.Errorf("#%d: %q Error: %s", i, tt.in, err)
			continue
		}

		out := buf.String()
		if out != tt.in {
			t.Errorf("#%d: Want %q, got %q", i, tt.in, out)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	for i, tt := range tests {
		var scan scanner
		in := []byte(tt.in)
		if err := checkValid(in, &scan); err != nil {
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("#%d: checkValid: %#v", i, err)
				continue
			}
		}
		if tt.ptr == nil {
			continue
		}
		v := reflect.New(reflect.TypeOf(tt.ptr).Elem())
		dec := NewDecoder(bytes.NewBuffer(in))
		if err := dec.Decode(v.Interface()); !reflect.DeepEqual(err, tt.err) {
			t.Errorf("#%d: %q %v want %v", i, tt.in, err, tt.err)
			continue
		}
		if !reflect.DeepEqual(v.Elem().Interface(), tt.out) {
			t.Errorf("#%d: mismatch\nhave: %#+v\nwant: %#+v", i, v.Elem().Interface(), tt.out)
		}
	}
}

type benchmarkStruct struct {
	Q      string      `bencode:"q"`
	AQ     string      `bencode:"aq,omitempty"`
	Cookie string      `bencode:"cookie,omitempty"`
	Hash   string      `bencode:"hash,omitempty"`
	Args   interface{} `bencode:"args,omitempty"`
	Txid   string      `bencode:"txid"`
}

var benchmarkTest = []byte("d1:q4:auth2:aq4:ping6:cookie10:13536270564:hash64:d1e4881e30895e2ee3e14c9bbce4537288a72a242dbd1786e8f1cc512e4cb4674:txid8:37199054e")

func BenchmarkUnmarshal(b *testing.B) {
	x := new(benchmarkStruct)
	var err error
	for i := 0; i < b.N; i++ {
		if err = Unmarshal(benchmarkTest, x); err != nil {
			b.Fatal(err.Error())
		}
	}
}
