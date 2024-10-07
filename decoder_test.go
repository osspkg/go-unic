package unic_test

import (
	"bytes"
	"fmt"
	"testing"

	"go.osspkg.com/casecheck"
	"go.osspkg.com/unic"
)

type (
	testModel1 struct {
		Server     string                `unic:"server"`
		STag       []string              `unic:"s-tag"`
		ITag       []int                 `unic:"i-tag"`
		ServerOpts testServOpt           `unic:"server"`
		Tags1      map[int64]testServOpt `unic:"tags1"`
		Tags2      []testServOpt         `unic:"tags2"`
	}
	testServOpt struct {
		Host string `unic:"host"`
		Port int    `unic:"port"`
	}
)

func TestUnit_NewDecoder(t *testing.T) {
	var b bytes.Buffer
	enc := unic.NewEncoder(&b)
	err := enc.Encode(testModel1{
		Server: "main",
		STag:   []string{"1", "2", "3"},
		ITag:   []int{5, 6, 7},
		Tags1: map[int64]testServOpt{
			99: {
				Host: "0.0.0.0",
				Port: 1,
			},
			5: {
				Host: "0.0.0.0",
				Port: 2,
			},
		},
		Tags2: []testServOpt{
			{
				Host: "0.0.0.0",
				Port: 3,
			},
			{
				Host: "0.0.0.0",
				Port: 4,
			},
		},
		ServerOpts: testServOpt{
			Host: "0.0.0.0",
			Port: 128,
		},
	})
	casecheck.NoError(t, err)
	err = enc.Build()
	casecheck.NoError(t, err)

	fmt.Println(b.String())

	dec, err := unic.NewDecoder(bytes.NewBuffer(b.Bytes()))
	casecheck.NoError(t, err)
	model := testModel1{}
	err = dec.Decode(&model)
	casecheck.NoError(t, err)

	fmt.Printf("\n--------------\n%#v\n--------------\n", model)

}
