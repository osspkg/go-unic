package coma

import "bytes"

func Marshal(in interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(in); err != nil {
		return nil, err
	}
	if err := enc.Done(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Unmarshal(in []byte, out interface{}) error {
	dec, err := NewDecoder(bytes.NewBuffer(in))
	if err != nil {
		return err
	}
	return dec.Decode(out)
}
