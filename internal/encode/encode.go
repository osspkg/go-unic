package encode

import (
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

func (v *Encoder) Encode(b *node.Block) {
	for _, c := range b.Root().Child() {
		node.DrawBlock(v.w, 0, c)
	}
	return
}
