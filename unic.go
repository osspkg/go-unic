/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unic

import (
	"bytes"
	"fmt"
	"reflect"

	"go.osspkg.com/bb"

	"go.osspkg.com/unic/internal/pool"
)

// Unmarshal parses UNIC data and stores the result in v.
// v must be a non-nil pointer to a struct.
func Unmarshal(data []byte, args ...any) error {
	if len(data) == 0 || len(args) == 0 {
		return nil
	}

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if _, err := buf.Write(data); err != nil {
		return fmt.Errorf("unic: initialize buffer: %w", err)
	}
	if _, err := buf.Seek(0, bb.SeekStart); err != nil {
		return fmt.Errorf("unic: seek buffer: %w", err)
	}

	doc, err := parseDocument(buf)
	if err != nil {
		return err
	}

	for _, arg := range args {
		rv, err := decodeTarget(arg)
		if err != nil {
			return err
		}

		if _, err = inspectStruct(rv.Type()); err != nil {
			return fmt.Errorf("unic: %w", err)
		}

		if err = decodeValue(rv, doc, ""); err != nil {
			return err
		}
	}

	return nil
}

func decodeTarget(v any) (reflect.Value, error) {
	if v == nil {
		return reflect.Value{}, fmt.Errorf("unic: Unmarshal(nil)")
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return reflect.Value{}, fmt.Errorf("unic: Unmarshal(non-nil pointer required)")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("unic: Unmarshal(non-struct pointer required)")
	}

	return rv, nil
}

// Marshal returns the UNIC encoding of v.
// v must be a struct or a non-nil pointer to struct.
func Marshal(args ...any) ([]byte, error) {
	if len(args) == 0 {
		return nil, nil
	}

	fieldsMap := make(map[string][]interface{})
	for _, arg := range args {
		rv, err := encodeTarget(arg)
		if err != nil {
			return nil, err
		}

		meta, err := inspectStruct(rv.Type())
		if err != nil {
			return nil, err
		}

		for _, bf := range meta.fields {
			if bf.attr > 0 {
				continue
			}
			fv := valueAt(rv, bf.index)
			if bf.omitempty && isEmptyValue(fv) {
				continue
			}
			fieldsMap[bf.name] = append(fieldsMap[bf.name], fv.Interface())
		}
	}

	mergedGroups := make(map[string][]interface{})
	for name, vals := range fieldsMap {
		if len(vals) > 1 {
			allStructs := true
			hasAnyAttr := false
			for _, val := range vals {
				fv := marshalDeref(reflect.ValueOf(val))
				if fv.Kind() != reflect.Struct {
					allStructs = false
					break
				}
				has, err := hasAttrs(fv.Type())
				if err != nil {
					return nil, err
				}
				if has {
					hasAnyAttr = true
					break
				}
			}
			if allStructs && !hasAnyAttr {
				mergedGroups[name] = vals
				delete(fieldsMap, name)
				continue
			}
		}
	}

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	e := &encoder{buf: buf}

	for name, vals := range mergedGroups {
		if err := e.writeMergedBlock(name, vals, ""); err != nil {
			return nil, err
		}
	}

	for name, vals := range fieldsMap {
		for _, val := range vals {
			fv := reflect.ValueOf(val)
			if fv.Kind() == reflect.Ptr && fv.IsNil() {
				if err := e.writeEmptyBlock(name); err != nil {
					return nil, err
				}
				continue
			}
			fv = marshalDeref(fv)
			if !fv.IsValid() {
				continue
			}
			f := &boundField{fieldTag: fieldTag{name: name}}
			if err := e.writeNamed(f, fv, ""); err != nil {
				return nil, err
			}
		}
	}

	return e.buf.Bytes(), nil
}

func encodeTarget(v any) (reflect.Value, error) {
	if v == nil {
		return reflect.Value{}, fmt.Errorf("unic: Marshal(nil)")
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return reflect.Value{}, fmt.Errorf("unic: Marshal(nil)")
		}
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("unic: Marshal(non-struct pointer required)")
	}

	return rv, nil
}
