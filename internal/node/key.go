package node

type Key struct {
	key    string
	values []string
}

func NewKey() *Key {
	return &Key{}
}

func (v *Key) IsValid() bool {
	return len(v.key) == 0
}

func (v *Key) Set(s string) {
	if len(v.key) == 0 {
		v.key = s
	} else {
		v.values = append(v.values, s)
	}
}

func (v *Key) Key() string {
	return v.key
}

func (v *Key) Values() []string {
	return append([]string{}, v.values...)
}
