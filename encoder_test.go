package unic_test

import (
	"bytes"
	"fmt"
	"testing"

	"go.osspkg.com/casecheck"
	"go.osspkg.com/unic"
)

type (
	testType1 struct {
		Type   string      `unic:"Type"`
		BaseP  *typeBase1  `unic:"BaseP"`
		Base   typeBase1   `unic:"Base"`
		SliceP *testSlice1 `unic:"SliceP"`
		Slice  testSlice1  `unic:"Slice"`
		MapPE  *testMap1   `unic:"MapPE"`
		MapP   *testMap1   `unic:"MapP"`
		Map    testMap1    `unic:"Map"`
	}
	typeBase1 struct {
		BoolP    *bool    `unic:"BoolP"`
		Bool     bool     `unic:"Bool"`
		IntP     *int     `unic:"IntP"`
		Int      int      `unic:"Int"`
		Int8     int8     `unic:"Int8"`
		Int16    int16    `unic:"Int16"`
		Int32    int32    `unic:"Int32"`
		Int64    int64    `unic:"Int64"`
		UIntP    *uint    `unic:"UIntP"`
		UInt     uint     `unic:"UInt"`
		UInt8    uint8    `unic:"UInt8"`
		UInt16   uint16   `unic:"UInt16"`
		UInt32   uint32   `unic:"UInt32"`
		UInt64   uint64   `unic:"UInt64"`
		Float32P *float32 `unic:"Float32P"`
		Float32  float32  `unic:"Float32"`
		Float64P *float64 `unic:"Float64P"`
		Float64  float64  `unic:"Float64"`
		StringP  *string  `unic:"StringP"`
		StringE  string   `unic:"StringE"`
		String   string   `unic:"String"`
	}
	testSlice1 struct {
		BytesP *[]byte `unic:"BytesP"`
		BytesE []byte  `unic:"BytesE"`
		Bytes  []byte  `unic:"Bytes"`
		IntsP  *[]int  `unic:"IntsP"`
		Ints   []int   `unic:"Ints"`
	}
	testMap1 struct {
		MapIntIntP *map[int]int `unic:"MapIntIntP"`
		MapIntInt  map[int]int  `unic:"MapIntInt"`
	}
)

func TestUnit_NewEncoder(t *testing.T) {
	data := testType1{
		Type:  "TestUnit NewEncoder",
		BaseP: nil,
		Base: typeBase1{
			BoolP:    nil,
			Bool:     true,
			IntP:     nil,
			Int:      -10,
			Int8:     -11,
			Int16:    -12,
			Int32:    -13,
			Int64:    14,
			UIntP:    nil,
			UInt:     10,
			UInt8:    11,
			UInt16:   12,
			UInt32:   13,
			UInt64:   14,
			Float32P: nil,
			Float32:  1.5,
			Float64P: nil,
			Float64:  -1.5,
			StringP:  nil,
			StringE:  "",
			String:   "123",
		},
		SliceP: nil,
		Slice: testSlice1{
			BytesP: nil,
			BytesE: []byte{},
			Bytes:  []byte("qwer"),
			IntsP:  nil,
			Ints:   []int{1, 2, 3},
		},
		MapP: &testMap1{
			MapIntInt: map[int]int{1: 2},
		},
		Map: testMap1{
			MapIntIntP: nil,
			MapIntInt:  map[int]int{3: 4},
		},
	}

	var b bytes.Buffer
	enc := unic.NewEncoder(&b)

	err := enc.Encode(data)
	casecheck.NoError(t, err)

	err = enc.Encode(testMap1{
		MapIntIntP: nil,
		MapIntInt:  map[int]int{3: 2},
	})
	casecheck.NoError(t, err)

	err = enc.Build()
	casecheck.NoError(t, err)

	fmt.Printf("\n------------------------\n%#v\n------------------------\n", b.String())

	casecheck.Equal(t, "Type `TestUnit NewEncoder`;\nBase {\n    Bool true;\n    Int -10;\n    Int8 -11;\n    Int16 -12;\n    Int32 -13;\n    Int64 14;\n    UInt 10;\n    UInt8 11;\n    UInt16 12;\n    UInt32 13;\n    UInt64 14;\n    Float32 1.5;\n    Float64 -1.5;\n    StringE ``;\n    String 123;\n}\nSlice {\n    Bytes qwer;\n    Ints 1 2 3;\n}\nMapP {\n    MapIntInt {\n        1 2;\n    }\n}\nMap {\n    MapIntInt {\n        3 4;\n    }\n}\nMapIntInt {\n    3 2;\n}\n", b.String())
}
