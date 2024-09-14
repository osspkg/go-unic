package ref

import (
	"fmt"
	"reflect"
	"strings"
)

type Data struct {
	Key    string
	IsOmit bool
	Ref    *reflect.Value
	child  []*Data
	parent *Data
}

func (v *Data) Root() (d *Data) {
	d = v
	for {
		if d.parent == nil {
			break
		}
		d = v.parent
	}
	return
}

func (v *Data) Next() *Data {
	d := &Data{}
	d.parent = v
	v.child = append(v.child, d)
	return d
}

func (v *Data) Previous() (d *Data) {
	d = v
	if d.parent != nil {
		d = v.parent
	}
	return
}

func (v *Data) Dump(level int) {
	fmt.Println(strings.Repeat("..", level), v.Key, v.IsOmit, func() string {
		if v.Ref == nil {
			return ""
		}
		return fmt.Sprintf("type=%s val=%#v", v.Ref.Type(), v.Ref.Interface())
	}())
	for _, data := range v.child {
		data.Dump(level + 1)
	}
}
