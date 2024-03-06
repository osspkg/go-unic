package coma

import (
	"io"
)

type Decoder struct {
	root *block
}

func NewDecoder(r io.Reader) (*Decoder, error) {
	parser := newNodeDecoder(r)
	if err := parser.Decode(); err != nil {
		return nil, err
	}
	dec := &Decoder{root: parser.RootNode()}
	return dec, nil
}

func (dec *Decoder) Decode(v interface{}) error {
	return nil
}
