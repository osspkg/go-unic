package tests

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"testing"

	"go.osspkg.com/unic/internal/decode"
	"go.osspkg.com/unic/internal/encode"
	"go.osspkg.com/unic/internal/node"
)

func TestUnit_Node_DecoderEncoder(t *testing.T) {
	b, err := os.ReadFile("./example1.conf")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	expected, err := os.ReadFile("./example1.golden.conf")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	dec := decode.New(bytes.NewBuffer(b))
	if err = dec.Decode(); err != nil {
		t.Error(err)
		t.FailNow()
	}

	val := node.Search(dec.GetBlock(), "block1", "sub_block2", "sub_block_with_value")
	if val == nil {
		t.Errorf("search is empty")
		t.FailNow()
	}

	if !reflect.DeepEqual(val.Values(), []string{"data1", "data2"}) {
		t.Errorf("search value not equal: %#v", val)
		t.FailNow()
	}

	var buf bytes.Buffer
	enc := encode.New(&buf)
	enc.Encode(dec.GetBlock())

	actual := buf.Bytes()
	if !bytes.Equal(expected, actual) {
		fmt.Println("--- want: ---")
		fmt.Println(string(expected))
		fmt.Println("--- got: ---")
		fmt.Println(string(actual))
		t.FailNow()
	}
}
