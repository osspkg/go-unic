package unic

import (
	"io"

	"go.osspkg.com/unic/internal/encode"
	"go.osspkg.com/unic/internal/node"
	"go.osspkg.com/unic/internal/ref"
)

type Encoder struct {
	enc *encode.Encoder
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		enc: encode.New(w),
	}
}

func (e *Encoder) Encode(v interface{}) error {
	reference, err := ref.New(v)
	if err != nil {
		return err
	}

	b := node.NewBlock()

	if err = reference.Marshal(b); err != nil {
		return err
	}

	return e.enc.Encode(b.Root())
}
