Go-EBML
=======
A marshaler/unmarshaler for [EBML](http://ebml.sourceforge.net/).

It's quite simple to use:
```go

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

	b, _ := ebml.Marshal(*DoDad)
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

	ebml.Unmarshal(data, DoDad)
	fmt.Printf("%v\n", DoDad)
	// Output:
	// &{0 70000 huzah -92387}
}

```


Caveats
-------
If an EBML DTD specifies a value as an unsigned integer, you must use a uint 
type. For example: the uint32 8388608 and int -8388608 may encode to identical
0x800000 24 bit arrays in an EBML stream.

EBML does not have a boolean type, but booleans are commonly defined in DTDs 
as zero or non-zero usigned integers.


Limitations
-----------
Default values are not implemented yet, as they are defined in the EBML RFC:
> Every non-container MAY be assigned a default value. If so, its
> value will be added to the interpretation of the EBML data if no
> element with another value exists in the data.
 
> As an example, consider this EBML DTD:
 
> Weight := 4101 {
>   WeightValue := 41a1 uint;
>   WeightUnit  := 41a2 string [ def:"kilogram" ];
> }
 
> If the Weight element only contains the WeightValue element, the
> WeightUnit element with value "kilogram" will be added when the
> information is semantically processed. A WeightUnit element with
> another value would of course override the default.
 
> The default value can also be a symbol referring back to a
> previously seen symbol. If however no such symbol has been seen,
> i.e. it has not been encoded into the EBML data and has no default
> value, the element will not be added as a child on the semantic
> level.
 
> Weight := 4101 {
>   WeightValue := 41a1 uint;
>   WeightUnit  := 41a2 string [ def:WeightUnit ];
> }
