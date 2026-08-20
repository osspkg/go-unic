/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unic

import (
	"reflect"
)

//nolint:unused
func mergeValues(a, b any) (any, error) {
	if a == nil && b == nil {
		return nil, nil
	}
	if a == nil {
		return b, nil
	}
	if b == nil {
		return a, nil
	}

	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	wasPtrA := va.Kind() == reflect.Ptr
	wasPtrB := vb.Kind() == reflect.Ptr

	baseA := derefValue(va)
	baseB := derefValue(vb)
	if !baseA.IsValid() || !baseB.IsValid() {
		if !baseA.IsValid() {
			return b, nil
		}
		return a, nil
	}

	// Если оба — структуры, сливаем их поля через карты
	if baseA.Kind() == reflect.Struct && baseB.Kind() == reflect.Struct {
		mapA, err := structToMap(baseA)
		if err != nil {
			return nil, err
		}
		mapB, err := structToMap(baseB)
		if err != nil {
			return nil, err
		}
		mergedMap, err := mergeMapValues(mapA, mapB)
		if err != nil {
			return nil, err
		}
		return mergedMap, nil
	}

	// Если типы не совпадают (и не обе структуры) -> список
	if baseA.Type() != baseB.Type() {
		return []any{a, b}, nil
	}

	var result any
	var err error
	switch baseA.Kind() {
	case reflect.Struct:
		result, err = mergeStruct(baseA, baseB)
	case reflect.Map:
		result, err = mergeMap(baseA, baseB)
	case reflect.Slice:
		result, err = mergeSlice(baseA, baseB)
	case reflect.Array:
		result, err = mergeArray(baseA, baseB)
	default:
		return b, nil // скаляры перезаписываются
	}
	if err != nil {
		return nil, err
	}

	if wasPtrA && wasPtrB {
		resPtr := reflect.New(baseA.Type())
		resPtr.Elem().Set(reflect.ValueOf(result))
		return resPtr.Interface(), nil
	}
	return result, nil
}

//nolint:unused
func derefValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

//nolint:unused
func mergeStruct(va, vb reflect.Value) (any, error) {
	t := va.Type()
	result := reflect.New(t).Elem()
	result.Set(va)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}
		fvA := va.Field(i)
		fvB := vb.Field(i)

		if fvA.CanInterface() && fvB.CanInterface() {
			merged, err := mergeValues(fvA.Interface(), fvB.Interface())
			if err != nil {
				return nil, err
			}

			if result.Field(i).CanSet() {
				result.Field(i).Set(reflect.ValueOf(merged))
			}
		}
	}
	return result.Interface(), nil
}

//nolint:unused
func mergeMap(va, vb reflect.Value) (any, error) {
	t := va.Type()
	result := reflect.MakeMap(t)

	for _, key := range va.MapKeys() {
		valA := va.MapIndex(key)
		valB := vb.MapIndex(key)
		if valB.IsValid() {
			merged, err := mergeValues(valA.Interface(), valB.Interface())
			if err != nil {
				return nil, err
			}
			result.SetMapIndex(key, reflect.ValueOf(merged))
		} else {
			result.SetMapIndex(key, valA)
		}
	}

	for _, key := range vb.MapKeys() {
		if !va.MapIndex(key).IsValid() {
			result.SetMapIndex(key, vb.MapIndex(key))
		}
	}
	return result.Interface(), nil
}

//nolint:unused,unparam
func mergeSlice(va, vb reflect.Value) (any, error) {
	totalLen := va.Len() + vb.Len()
	result := reflect.MakeSlice(va.Type(), totalLen, totalLen)
	reflect.Copy(result, va)
	reflect.Copy(result.Slice(va.Len(), totalLen), vb)
	return result.Interface(), nil
}

//nolint:unused
func mergeArray(va, vb reflect.Value) (any, error) {
	if va.Len() != vb.Len() {
		return vb.Interface(), nil
	}
	t := va.Type()
	result := reflect.New(t).Elem()
	for i := 0; i < va.Len(); i++ {
		merged, err := mergeValues(va.Index(i).Interface(), vb.Index(i).Interface())
		if err != nil {
			return nil, err
		}
		result.Index(i).Set(reflect.ValueOf(merged))
	}
	return result.Interface(), nil
}

//nolint:unused
func structToMap(v reflect.Value) (map[string]any, error) {
	v = derefValue(v)
	if v.Kind() != reflect.Struct {
		return nil, errNotStruct
	}
	t := v.Type()
	result := make(map[string]any)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}
		tag, ok := field.Tag.Lookup("unic")
		if !ok {
			continue // пропускаем поля без тега
		}
		ft, keep, err := parseUnicTag(tag)
		if err != nil {
			return nil, err
		}
		if !keep {
			continue
		}
		fv := v.Field(i)
		if !fv.CanInterface() {
			continue
		}
		if fv.Kind() == reflect.Struct {
			subMap, err := structToMap(fv)
			if err != nil {
				return nil, err
			}
			result[ft.name] = subMap
		} else {
			result[ft.name] = fv.Interface()
		}
	}
	return result, nil
}

//nolint:unused
func mergeMapValues(a, b map[string]any) (map[string]any, error) {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)
	if va.Kind() != reflect.Map || vb.Kind() != reflect.Map {
		return nil, errNotMaps
	}
	mergedVal, err := mergeMap(va, vb)
	if err != nil {
		return nil, err
	}
	return mergedVal.(map[string]any), nil
}

func hasAttrs(t reflect.Type) (bool, error) {
	meta, err := inspectStruct(t)
	if err != nil {
		return false, err
	}
	for _, f := range meta.fields {
		if f.attr > 0 {
			return true, nil
		}
	}
	return false, nil
}
