package unic

import (
	"io"

	"go.osspkg.com/unic/internal/decode"
	"go.osspkg.com/unic/internal/node"
)

type Decoder struct {
	root *node.Block
}

func NewDecoder(r io.Reader) (*Decoder, error) {
	parser := decode.New(r)
	if err := parser.Decode(); err != nil {
		return nil, err
	}
	dec := &Decoder{root: parser.GetBlock()}
	return dec, nil
}

func (dec *Decoder) Decode(v interface{}) error {
	return nil
}
