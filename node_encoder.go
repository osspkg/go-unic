package coma

import (
	"fmt"
	"io"
)

type nodeEncoder struct {
	w io.Writer
}

func newNodeEncoder(w io.Writer) *nodeEncoder {
	return &nodeEncoder{
		w: w,
	}
}

func (v *nodeEncoder) Encode(b *block) error {
	_, err := fmt.Fprintf(v.w, b.toString(0))
	return err
}
