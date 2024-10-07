package unic

import (
	"io"

	"go.osspkg.com/unic/internal/decode"
	"go.osspkg.com/unic/internal/node"
	"go.osspkg.com/unic/internal/ref"
)

type Decoder struct {
	block *node.Block
}

func NewDecoder(r io.Reader) (*Decoder, error) {
	parser := decode.New(r)
	if err := parser.Decode(); err != nil {
		return nil, err
	}
	return &Decoder{block: parser.GetBlock()}, nil
}

func (d *Decoder) Decode(v interface{}) error {
	reference, err := ref.NewPointer(v)
	if err != nil {
		return err
	}
	tree := &ref.Tree{}
	if err = reference.Build(tree); err != nil {
		return err
	}
	return tree.Import(d.block)
}
