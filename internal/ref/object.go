package ref

import (
	"fmt"
	"reflect"
	"strings"
)

type Object struct {
	obj reflect.Value
}

func NewPointer(obj any) (*Object, error) {
	refValue := reflect.ValueOf(obj)

	switch refValue.Kind() {
	case reflect.Ptr:
		if refValue.IsNil() {
			return nil, fmt.Errorf("got nil-pointer object")
		}
		refType := refValue.Type().Elem()
		if refType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("object must be a struct: %s", refValue.Type().String())
		}
		return &Object{obj: refValue}, nil
	case reflect.Struct:
		return nil, fmt.Errorf("object must be ponter of struct")
	default:
		return nil, fmt.Errorf("got unsupported type: %s", refValue.Type().String())
	}
}

func NewNonPointer(obj any) (*Object, error) {
	refValue := reflect.ValueOf(obj)

	switch refValue.Kind() {
	case reflect.Ptr:
		return nil, fmt.Errorf("object must be non ponter struct")
	case reflect.Struct:
		return &Object{obj: refValue}, nil
	default:
		return nil, fmt.Errorf("got unsupported type: %s", refValue.Type().String())
	}
}

func (v *Object) Build(data *Tree) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("fail: %+v", e)
		}
	}()

	switch v.obj.Kind() {
	case reflect.Pointer:
		parsePointer(v.obj, data)
	case reflect.Struct:
		parseStruct(v.obj, data)
	default:
		fail("invalid type of field %s", v.obj.Type())
	}
	return
}

func parsePointer(ref reflect.Value, data *Tree) {
	data.CanNil = true

	if ref.IsNil() {
		if ref.CanSet() {
			ref.Set(reflect.New(ref.Type().Elem()))
		} else {
			return
		}
	}

	ref = ref.Elem()

	switch ref.Kind() {
	case reflect.Struct:
		parseStruct(ref, data)
	default:
		parseOther(ref, data)
	}
}

func parseOther(ref reflect.Value, data *Tree) {
	data.Ref = &ref

	switch ref.Kind() {
	case reflect.Bool, reflect.String,
		reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// data.Ref = &ref

	case reflect.Slice:
		data.CanNil = true
		if ref.Len() == 0 {
			ref.Set(reflect.MakeSlice(ref.Type(), 0, 2))
		}

		switch ref.Type().Elem().Kind() {
		case reflect.Bool, reflect.String,
			reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

		case reflect.Struct:
			for i := 0; i < ref.Len(); i++ {
				data = data.Next("-", false)
				parseStruct(ref.Index(i), data)
				data = data.Previous()
			}

		default:
			fail("invalid slice value type %s", ref.Type())
		}

	case reflect.Map:
		data.CanNil = true
		if ref.Len() == 0 {
			ref.Set(reflect.MakeMap(ref.Type()))
		}

		switch ref.Type().Key().Kind() {
		case reflect.Bool, reflect.String,
			reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		default:
			fail("invalid map key type %s", ref.Type())
		}

		switch ref.Type().Elem().Kind() {
		case reflect.Bool, reflect.String,
			reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return

		case reflect.Struct:
			keys := ref.MapKeys()
			sortReflectValues(keys)
			for _, key := range keys {
				data = data.Next(fmt.Sprintf("%+v", key), true)
				parseStruct(ref.MapIndex(key), data)
				data = data.Previous()
			}

		default:
			fail("invalid map value type %s", ref.Type())
		}

	default:
		fail("invalid type of field %s", ref.Type())
	}
}

func parseStruct(ref reflect.Value, data *Tree) {
	refType := ref.Type()
	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)

		if !field.IsExported() {
			continue
		}

		tagValue, ok := field.Tag.Lookup(tagName)
		if !ok {
			continue
		}

		fieldTag, ok := parseTag(tagValue)
		if !ok {
			fail("invalid tag of field %s", field.Name)
		}

		data = data.Next(fieldTag, true)
		data.Key = fieldTag

		switch field.Type.Kind() {
		case reflect.Struct:
			parseStruct(ref.Field(i), data)
		case reflect.Pointer:
			parsePointer(ref.Field(i), data)
		default:
			parseOther(ref.Field(i), data)
		}

		data = data.Previous()
	}
}

func parseTag(tag string) (fieldName string, isValid bool) {
	isValid = true

	vs := strings.Split(tag, ",")
	switch len(vs) {
	case 0:
	default:
		fieldName = vs[0]
	}

	fieldName = strings.TrimSpace(fieldName)
	if len(fieldName) == 0 {
		isValid = false
	}
	return
}

func fail(format string, a ...any) {
	panic(fmt.Sprintf(format, a...))
}
