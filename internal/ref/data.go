package ref

import (
	"fmt"
	"reflect"
	"strings"

	"go.osspkg.com/unic/internal/node"
)

type Data struct {
	Key    string
	CanNil bool
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

func (v *Data) Dump(level int, b *node.Block) error {
	fmt.Println(strings.Repeat("..", level), v.Key, v.CanNil, func() string {
		if v.Ref == nil {
			return "<nil>"
		}
		return fmt.Sprintf("kind=%s type=%s val=%#v", v.Ref.Kind(), v.Ref.Type(), v.Ref.Interface())
	}())

	if v.Ref == nil && len(v.child) == 0 {
		return nil
	}

	b.Key().Set(v.Key)

	if v.Ref != nil {
		switch v.Ref.Kind() {
		case reflect.Bool,
			reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			b.Key().Set(fmt.Sprintf("%v", v.Ref.Interface()))

		case reflect.String:
			if !v.Ref.IsZero() {
				b.Key().Set(v.Ref.String())
			}

		case reflect.Slice:
			if v.Ref.Len() > 0 {
				if value, ok := v.Ref.Interface().([]byte); ok {
					b.Key().Set(string(value))
				} else {
					for i := 0; i < v.Ref.Len(); i++ {
						sliceValue := v.Ref.Index(i)
						if sliceValue.Kind() == reflect.Interface {
							sliceValue = sliceValue.Elem()
						}
						switch sliceValue.Kind() {
						case reflect.Bool, reflect.String,
							reflect.Float32, reflect.Float64,
							reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
							reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
							b.Key().Set(fmt.Sprintf("%v", v.Ref.Index(i).Interface()))

						default:
							return fmt.Errorf("unsupported type for `%s`: %s", v.Key, sliceValue.Elem().Type().String())
						}
					}
				}
			}

		case reflect.Map:
			if v.Ref.Len() > 0 {
				for _, mapKey := range v.Ref.MapKeys() {
					mapValue := v.Ref.MapIndex(mapKey)
					b = b.NextBlock()

					if mapKey.Kind() == reflect.Interface {
						mapKey = mapKey.Elem()
					}
					switch mapKey.Kind() {
					case reflect.Bool, reflect.String,
						reflect.Float32, reflect.Float64,
						reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
						reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						b.Key().Set(fmt.Sprintf("%v", mapKey.Interface()))

					default:
						return fmt.Errorf("unsupported type map key for `%s`: %s", v.Key, mapKey.Elem().Type().String())
					}

					if mapValue.Kind() == reflect.Interface {
						mapValue = mapValue.Elem()
					}
					switch mapValue.Kind() {
					case reflect.Bool, reflect.String,
						reflect.Float32, reflect.Float64,
						reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
						reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						b.Key().Set(fmt.Sprintf("%v", mapValue.Interface()))

					default:
						return fmt.Errorf("unsupported type map value for `%s`: %s", v.Key, mapValue.Elem().Type().String())
					}

					b = b.PreviousBlock()
				}
			}
		default:
			return fmt.Errorf("unsupported type for `%s`: %s", v.Key, v.Ref.Type().String())
		}
	}

	for _, data := range v.child {
		b = b.NextBlock()
		if err := data.Dump(level+1, b); err != nil {
			return err
		}
		b = b.PreviousBlock()
	}

	return nil
}
