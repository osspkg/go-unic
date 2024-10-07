package ref

import (
	"fmt"
	"reflect"
	"strconv"

	"go.osspkg.com/unic/internal/node"
)

func encSwitch(v reflect.Value, b *node.Block) error {
	switch v.Kind() {
	case reflect.Bool:
		b.Key().Set(strconv.FormatBool(v.Bool()))

	case reflect.Float32, reflect.Float64:
		b.Key().Set(strconv.FormatFloat(v.Float(), 'f', -1, 64))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.Key().Set(strconv.FormatInt(v.Int(), 10))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		b.Key().Set(strconv.FormatUint(v.Uint(), 10))

	case reflect.String:
		if v.Len() > 0 {
			b.Key().Set(v.String())
		} else {
			b.Key().Set("``")
		}

	case reflect.Interface:
		return encSwitch(v.Elem(), b)

	case reflect.Slice:
		if v.Len() == 0 {
			return nil
		}
		if value, ok := v.Interface().([]byte); ok {
			b.Key().Set(string(value))
			return nil
		}
		for i := 0; i < v.Len(); i++ {
			if err := encSwitch(v.Index(i), b); err != nil {
				return err
			}
		}

	case reflect.Map:
		if v.Len() == 0 {
			return nil
		}
		for _, key := range v.MapKeys() {
			b = b.Next()
			if err := encSwitch(key, b); err != nil {
				return err
			}
			if err := encSwitch(v.MapIndex(key), b); err != nil {
				return err
			}
			b = b.Previous()
		}

	case reflect.Struct:

	default:
		return fmt.Errorf("unsupported type: %s", v.Type().String())
	}

	return nil
}
