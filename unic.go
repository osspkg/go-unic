package unic

import "bytes"

func Marshal(in interface{}) ([]byte, error) {
	var b bytes.Buffer
	enc := NewEncoder(&b)
	if err := enc.Encode(in); err != nil {
		return nil, err
	}
	if err := enc.Build(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func Unmarshal(in []byte, out interface{}) error {
	dec, err := NewDecoder(bytes.NewBuffer(in))
	if err != nil {
		return err
	}
	return dec.Decode(out)
}
