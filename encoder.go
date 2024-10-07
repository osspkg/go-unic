package unic

import (
	"io"

	"go.osspkg.com/unic/internal/encode"
	"go.osspkg.com/unic/internal/node"
	"go.osspkg.com/unic/internal/ref"
)

type Encoder struct {
	w    io.Writer
	tree *ref.Tree
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:    w,
		tree: &ref.Tree{},
	}
}

func (e *Encoder) Build() error {
	b := node.NewBlock()
	if err := e.tree.Root().Export(b); err != nil {
		return err
	}
	encode.New(e.w).Encode(b)
	return nil
}

func (e *Encoder) Encode(v interface{}) error {
	reference, err := ref.NewNonPointer(v)
	if err != nil {
		return err
	}
	return reference.Build(e.tree.Root())
}
