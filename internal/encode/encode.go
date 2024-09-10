package encode

import (
	"bytes"
	"io"

	"go.osspkg.com/unic/internal/node"
)

type Encoder struct {
	w io.Writer
}

func New(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

func (v *Encoder) Encode(b *node.Block) error {
	buff := bytes.NewBuffer(nil)
	for _, c := range b.Root().Child() {
		node.DrawBlock(buff, 0, c)
	}
	_, err := buff.WriteTo(v.w)
	return err
}
