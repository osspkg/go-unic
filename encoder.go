package unic

import "io"

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Done() error {
	return nil
}

func (e *Encoder) Encode(v interface{}) error {
	return nil
}
