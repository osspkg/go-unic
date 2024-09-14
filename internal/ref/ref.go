package ref

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	tagName = "unic"
)

type Ref struct {
	obj reflect.Value
}

func New(obj any) (*Ref, error) {
	refValue := reflect.ValueOf(obj)
	if refValue.Kind() != reflect.Ptr || refValue.IsNil() {
		return nil, fmt.Errorf("got non-pointer or nil-pointer object")
	}

	refType := refValue.Type().Elem()
	if refType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("pointer must be a strict")
	}

	return &Ref{obj: refValue}, nil
}

func (v *Ref) Marshal() error {
	data := &Data{}
	if err := v.parseRef(v.obj, data); err != nil {
		return err
	}
	data.Root().Dump(1)
	return nil
}

func (v *Ref) Unmarshal() error {
	return nil
}

func (v *Ref) parseRef(ref reflect.Value, data *Data) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %+v", e)
		}
	}()

	switch ref.Kind() {
	case reflect.Pointer:
		return v.parsePointer(ref, data)
	case reflect.Struct:
		return v.parseStruct(ref, data)
	default:
		return fmt.Errorf("invalid type of field %s", ref.Type())
	}
}

func (v *Ref) parsePointer(ref reflect.Value, data *Data) error {
	if ref.IsZero() {
		ref.Set(reflect.New(ref.Type().Elem()))
	}

	ref = ref.Elem()
	data.IsOmit = true

	switch ref.Kind() {
	case reflect.Struct:
		return v.parseStruct(ref, data)
	default:
		return v.parseOther(ref, data)
	}
}

func (v *Ref) parseOther(ref reflect.Value, data *Data) error {
	switch ref.Kind() {
	case reflect.Bool, reflect.String, reflect.Float32, reflect.Float64,
		reflect.Array, reflect.Slice, reflect.Map,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		data.Ref = &ref
		return nil

	default:
		return fmt.Errorf("invalid type of field %s", ref.Type())
	}
}

func (v *Ref) parseStruct(ref reflect.Value, data *Data) error {
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

		fieldTag, ok := v.decodeTag(tagValue)
		if !ok {
			return fmt.Errorf("invalid tag of field %s", field.Name)
		}

		data = data.Next()
		data.Key = fieldTag

		var err error
		switch field.Type.Kind() {
		case reflect.Struct:
			err = v.parseStruct(ref.Field(i), data)
		case reflect.Pointer:
			err = v.parsePointer(ref.Field(i), data)
		default:
			err = v.parseOther(ref.Field(i), data)
		}
		if err != nil {
			return err
		}

		data = data.Previous()
	}
	return nil
}

func (v *Ref) decodeTag(tag string) (fieldName string, isValid bool) {
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
