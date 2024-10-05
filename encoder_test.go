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
		Type   string     `unic:"Type"`
		BaseP  *typeBase  `unic:"BaseP"`
		Base   typeBase   `unic:"Base"`
		SliceP *testSlice `unic:"SliceP"`
		Slice  testSlice  `unic:"Slice"`
		MapPE  *testMap   `unic:"MapPE"`
		MapP   *testMap   `unic:"MapP"`
		Map    testMap    `unic:"Map"`
	}
	typeBase struct {
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
	testSlice struct {
		BytesP *[]byte `unic:"BytesP"`
		BytesE []byte  `unic:"Bytes"`
		Bytes  []byte  `unic:"Bytes"`
		IntsP  *[]int  `unic:"IntsP"`
		Ints   []int   `unic:"Ints"`
		AnysP  *[]any  `unic:"AnysP"`
		Anys   []any   `unic:"Anys"`
	}
	testMap struct {
		MapIntIntP *map[int]int `unic:"MapIntIntP"`
		MapIntInt  map[int]int  `unic:"MapIntInt"`
		MapAnyAny  map[any]any  `unic:"MapAnyAny"`
	}
)

func TestUnit_NewEncoder(t *testing.T) {
	buf := bytes.NewBuffer(nil)

	data := testType1{
		Type:  "TestUnit NewEncoder",
		BaseP: nil,
		Base: typeBase{
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
			Float32:  1.3,
			Float64P: nil,
			Float64:  -1.4,
			StringP:  nil,
			StringE:  "",
			String:   "123",
		},
		SliceP: nil,
		Slice: testSlice{
			BytesP: nil,
			BytesE: []byte{},
			Bytes:  []byte("qwer"),
			IntsP:  nil,
			Ints:   []int{1, 2, 3},
			AnysP:  nil,
			Anys:   []any{"aaa", 1, true},
		},
		MapP: &testMap{
			MapIntInt: map[int]int{1: 2, 3: 4},
			MapAnyAny: map[any]any{"s": 123, true: 1.3},
		},
		Map: testMap{
			MapIntIntP: nil,
			MapIntInt:  map[int]int{1: 2, 3: 4},
			MapAnyAny:  map[any]any{"s": 123, true: 1.3},
		},
	}

	enc := unic.NewEncoder(buf)

	err := enc.Encode(data)
	casecheck.NoError(t, err)

	err = enc.Encode(data)
	casecheck.NoError(t, err)

	fmt.Println(buf.String())
}
