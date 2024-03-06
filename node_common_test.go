package coma

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestUnit_Node_DecoderEncoder(t *testing.T) {
	b, err := os.ReadFile("testdata/example1.conf")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	expected, err := os.ReadFile("testdata/example1.golden.conf")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	dec := newNodeDecoder(bytes.NewBuffer(b))
	if err = dec.Decode(); err != nil {
		t.Error(err)
		t.FailNow()
	}

	var buf bytes.Buffer
	enc := newNodeEncoder(&buf)
	if err = enc.Encode(dec.RootNode()); err != nil {
		t.Error(err)
		t.FailNow()
	}

	actual := buf.Bytes()
	if !bytes.Equal(expected, actual) {
		fmt.Println("--- want: ---")
		fmt.Println(string(expected))
		fmt.Println("--- got: ---")
		fmt.Println(string(actual))
		t.FailNow()
	}
}
