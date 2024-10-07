package ref

import (
	"fmt"
	"reflect"
	"strconv"
)

func decSwitch(t reflect.Type, s string) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Bool:
		d, err := strconv.ParseBool(s)
		return reflect.ValueOf(d), err

	case reflect.Float32:
		d, err := strconv.ParseFloat(s, 32)
		return reflect.ValueOf(float32(d)), err
	case reflect.Float64:
		d, err := strconv.ParseFloat(s, 64)
		return reflect.ValueOf(d), err

	case reflect.Int:
		d, err := strconv.ParseInt(s, 10, 0)
		return reflect.ValueOf(int(d)), err
	case reflect.Int8:
		d, err := strconv.ParseInt(s, 10, 8)
		return reflect.ValueOf(int8(d)), err
	case reflect.Int16:
		d, err := strconv.ParseInt(s, 10, 16)
		return reflect.ValueOf(int16(d)), err
	case reflect.Int32:
		d, err := strconv.ParseInt(s, 10, 32)
		return reflect.ValueOf(int32(d)), err
	case reflect.Int64:
		d, err := strconv.ParseInt(s, 10, 64)
		return reflect.ValueOf(d), err

	case reflect.Uint:
		d, err := strconv.ParseUint(s, 10, 0)
		return reflect.ValueOf(uint(d)), err
	case reflect.Uint8:
		d, err := strconv.ParseUint(s, 10, 8)
		return reflect.ValueOf(uint8(d)), err
	case reflect.Uint16:
		d, err := strconv.ParseUint(s, 10, 16)
		return reflect.ValueOf(uint16(d)), err
	case reflect.Uint32:
		d, err := strconv.ParseUint(s, 10, 32)
		return reflect.ValueOf(uint32(d)), err
	case reflect.Uint64:
		d, err := strconv.ParseUint(s, 10, 64)
		return reflect.ValueOf(d), err

	case reflect.String:
		return reflect.ValueOf(s), nil
	}

	return reflect.ValueOf(nil), fmt.Errorf("unsupported value: %v", s)
}
