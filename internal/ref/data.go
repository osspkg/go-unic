package ref

import (
	"fmt"
	"reflect"

	"go.osspkg.com/unic/internal/node"
)

type Tree struct {
	Key    string
	CanNil bool
	Ref    *reflect.Value
	child  []*Tree
	parent *Tree
}

func (v *Tree) Root() (d *Tree) {
	d = v
	for {
		if d.parent == nil {
			break
		}
		d = v.parent
	}
	return
}

func (v *Tree) Next(key string, uniq bool) *Tree {
	if uniq {
		for _, data := range v.child {
			if data.Key == key {
				return data
			}
		}
	}
	d := &Tree{Key: key}
	d.parent = v
	v.child = append(v.child, d)
	return d
}

func (v *Tree) Previous() (d *Tree) {
	d = v
	if d.parent != nil {
		d = v.parent
	}
	return
}

func (v *Tree) Export(b *node.Block) error {
	if v.Ref == nil && len(v.child) == 0 {
		return nil
	}

	b.Key().Set(v.Key)

	if v.Ref != nil {
		if err := encSwitch(*v.Ref, b); err != nil {
			return err
		}
	}

	for _, data := range v.child {
		b = b.Next()
		if err := data.Export(b); err != nil {
			return err
		}
		b = b.Previous()
	}

	return nil
}

func (v *Tree) Import(b *node.Block) error {

	for _, data := range v.child {
		block := node.Search(b, data.Key)

		if block == nil {
			if data.CanNil {
				data.Ref.SetZero()
			}
			continue
		}

		fmt.Println(data.Key, data.CanNil, data.Ref, block.Key())

		if data.Ref == nil {
			continue
		}

		switch data.Ref.Kind() {
		case reflect.Bool, reflect.String,
			reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if !block.Key().NoValue() {
				rv, err := decSwitch(data.Ref.Type().Elem(), block.Key().Values()[0])
				if err != nil {
					return err
				}
				data.Ref.Set(rv)
			}

		case reflect.Slice:
			switch data.Ref.Type().Elem().Kind() {
			case reflect.Bool, reflect.String,
				reflect.Float32, reflect.Float64,
				reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				for _, s := range block.Key().Values() {
					rv, err := decSwitch(data.Ref.Type().Elem(), s)
					if err != nil {
						return err
					}
					data.Ref.Set(reflect.Append(*data.Ref, rv))
				}

			default:
				fmt.Println("--->", data.Ref.Type().Elem())
			}

		case reflect.Map:
			fmt.Println("--->", data.Ref.Type().Key(), data.Ref.Type().Elem())

		default:
		}

		if err := data.Import(block); err != nil {
			return err
		}

	}

	return nil
}
