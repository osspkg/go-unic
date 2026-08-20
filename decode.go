/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unic

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type structMeta struct {
	fields []*boundField
	byName map[string]int
}

type boundField struct {
	fieldTag
	index []int
	typ   reflect.Type
}

func decodeValue(dv reflect.Value, n *node, path string) error {
	if n == nil || n.kind == nodeEmpty {
		return nil
	}
	dv, err := alloc(dv)
	if err != nil {
		return errAt(path, err)
	}

	if dv.Kind() == reflect.Interface && dv.NumMethod() == 0 {
		var val any
		val, err = n.asAny()
		if err != nil {
			return errAt(path, err)
		}
		dv.Set(reflect.ValueOf(val))
		return nil
	}

	switch n.kind {
	case nodeScalar:
		return assignScalar(dv, n.scalar, path)
	case nodeList:
		return assignList(dv, n, path)
	case nodeMap:
		return assignMap(dv, n, path)
	case nodeBlock:
		return assignBlock(dv, n, path)
	default:
		return errAt(path, errUnknownNode)
	}
}

func assignBlock(dv reflect.Value, n *node, path string) error {
	switch dv.Kind() {
	case reflect.Struct:
		return decodeStruct(dv, n, path)
	case reflect.Map:
		return decodeBlockMap(dv, n, path)
	case reflect.Slice, reflect.Array:
		return decodeValue(dv, &node{kind: nodeList, list: []*node{n}, line: n.line, col: n.col}, path)
	default:
		return errAt(path, fmt.Errorf("cannot unmarshal block into %s", dv.Type()))
	}
}

func decodeStruct(dv reflect.Value, n *node, path string) error {
	meta, err := inspectStruct(dv.Type())
	if err != nil {
		return errAt(path, err)
	}

	if err = decodeStructAttrs(dv, n, meta, path); err != nil {
		return err
	}

	groups, order := buildFieldGroups(n)
	seen, err := decodeStructFields(dv, meta, groups, order, path)
	if err != nil {
		return err
	}

	return decodeStructRemaining(dv, meta, seen, path)
}

func decodeStructAttrs(dv reflect.Value, n *node, meta *structMeta, path string) error {
	for _, f := range meta.fields {
		if f.attr <= 0 {
			continue
		}
		fv, err := fieldByIndex(dv, f.index)
		if err != nil {
			return errAt(joinPath(path, f.name), err)
		}
		if f.attr <= len(n.attrs) {
			if err = decodeValue(fv, n.attrs[f.attr-1], joinPath(path, f.name)); err != nil {
				return err
			}
			continue
		}
		if f.hasDefault {
			if err = applyDefault(fv, f, joinPath(path, f.name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func buildFieldGroups(n *node) (map[string][]*node, []string) {
	groups := make(map[string][]*node, len(n.fields))
	order := make([]string, 0, len(n.fields))
	for _, fl := range n.fields {
		if _, ok := groups[fl.key]; !ok {
			order = append(order, fl.key)
		}
		groups[fl.key] = append(groups[fl.key], fl.val)
	}
	return groups, order
}

func decodeStructFields(
	dv reflect.Value, meta *structMeta, groups map[string][]*node,
	order []string, path string,
) (map[string]struct{}, error) {
	seen := make(map[string]struct{}, len(order))
	for _, key := range order {
		idx, ok := meta.byName[key]
		if !ok || meta.fields[idx].attr > 0 {
			continue
		}
		f := meta.fields[idx]
		seen[key] = struct{}{}
		fv, err := fieldByIndex(dv, f.index)
		if err != nil {
			return nil, errAt(joinPath(path, key), err)
		}
		vals := groups[key]
		if len(vals) == 1 && f.omitempty && isEmptyNode(vals[0]) {
			continue
		}
		if isSliceOrArray(fv) && f.attr == 0 {
			if err = decodeSliceField(fv, vals, joinPath(path, key)); err != nil {
				return nil, err
			}
			continue
		}
		if len(vals) > 1 {
			return nil, errAt(joinPath(path, key), errDuplicateField)
		}
		if err = decodeValue(fv, vals[0], joinPath(path, key)); err != nil {
			return nil, err
		}
	}
	return seen, nil
}

func decodeStructRemaining(
	dv reflect.Value, meta *structMeta,
	seen map[string]struct{}, path string,
) error {
	for _, f := range meta.fields {
		if f.attr > 0 {
			continue
		}
		if _, ok := seen[f.name]; ok {
			continue
		}

		fv, err := fieldByIndex(dv, f.index)
		if err != nil {
			return errAt(joinPath(path, f.name), err)
		}
		if f.omitempty && !f.hasDefault {
			continue
		}
		if f.hasDefault {
			if err = applyDefault(fv, f, joinPath(path, f.name)); err != nil {
				return err
			}
			continue
		}
		if structLike(fv.Type()) {
			if err = decodeValue(fv, &node{kind: nodeBlock}, joinPath(path, f.name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeSliceField(fv reflect.Value, vals []*node, path string) error {
	if len(vals) == 1 && vals[0].kind == nodeList {
		return decodeValue(fv, vals[0], path)
	}

	if fv.Kind() == reflect.Array {
		if len(vals) > fv.Len() {
			return errAt(path, errArrayManyValues)
		}
		for i, val := range vals {
			if err := decodeValue(fv.Index(i), val, indexPath(path, i)); err != nil {
				return err
			}
		}
		return nil
	}

	sl := reflect.MakeSlice(fv.Type(), 0, len(vals))
	for i, val := range vals {
		el := reflect.New(fv.Type().Elem()).Elem()
		if err := decodeValue(el, val, indexPath(path, i)); err != nil {
			return err
		}
		sl = reflect.Append(sl, el)
	}
	fv.Set(sl)
	return nil
}

func decodeBlockMap(dv reflect.Value, n *node, path string) error {
	if dv.Type().Key().Kind() != reflect.String {
		return errAt(path, errMapKeyString)
	}
	if dv.IsNil() {
		dv.Set(reflect.MakeMapWithSize(dv.Type(), len(n.fields)))
	}
	elemType := dv.Type().Elem()
	for _, fl := range n.fields {
		el := reflect.New(elemType).Elem()
		if err := decodeValue(el, fl.val, joinPath(path, fl.key)); err != nil {
			return err
		}
		dv.SetMapIndex(reflect.ValueOf(fl.key), el)
	}
	return nil
}

func assignList(dv reflect.Value, n *node, path string) error {
	switch dv.Kind() {
	case reflect.Slice:
		sl := reflect.MakeSlice(dv.Type(), len(n.list), len(n.list))
		for i, item := range n.list {
			if err := decodeValue(sl.Index(i), item, indexPath(path, i)); err != nil {
				return err
			}
		}
		dv.Set(sl)
		return nil
	case reflect.Array:
		if len(n.list) > dv.Len() {
			return errAt(path, errArrayManyValues)
		}
		for i, item := range n.list {
			if err := decodeValue(dv.Index(i), item, indexPath(path, i)); err != nil {
				return err
			}
		}
		return nil
	default:
		if len(n.list) == 1 {
			return decodeValue(dv, n.list[0], path)
		}
		return errAt(path, fmt.Errorf("cannot unmarshal list into %s", dv.Type()))
	}
}

func assignMap(dv reflect.Value, n *node, path string) error {
	if dv.Kind() != reflect.Map {
		return errAt(path, fmt.Errorf("cannot unmarshal map into %s", dv.Type()))
	}
	if dv.IsNil() {
		dv.Set(reflect.MakeMapWithSize(dv.Type(), len(n.pairs)))
	}
	kt, vt := dv.Type().Key(), dv.Type().Elem()
	for i, p := range n.pairs {
		key := reflect.New(kt).Elem()
		if err := decodeValue(key, p.key, indexPath(path, i)+".key"); err != nil {
			return err
		}
		val := reflect.New(vt).Elem()
		if err := decodeValue(val, p.val, indexPath(path, i)+".val"); err != nil {
			return err
		}
		dv.SetMapIndex(key, val)
	}
	return nil
}

func assignScalar(dv reflect.Value, s, path string) error {
	switch dv.Kind() {
	case reflect.String:
		dv.SetString(s)
	case reflect.Bool:
		b, err := parseBool(s)
		if err != nil {
			return errAt(path, err)
		}
		dv.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, dv.Type().Bits())
		if err != nil {
			return errAt(path, fmt.Errorf("invalid integer %q", s))
		}
		dv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(s, 10, dv.Type().Bits())
		if err != nil {
			return errAt(path, fmt.Errorf("invalid unsigned integer %q", s))
		}
		dv.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, dv.Type().Bits())
		if err != nil {
			return errAt(path, fmt.Errorf("invalid float %q", s))
		}
		dv.SetFloat(n)
	default:
		return errAt(path, fmt.Errorf("cannot unmarshal %q into %s", s, dv.Type()))
	}
	return nil
}

func applyDefault(fv reflect.Value, f *boundField, path string) error {
	if isSliceOrArray(fv) {
		parts := splitDefaultList(f.defaultVal)
		items := make([]*node, 0, len(parts))
		for _, p := range parts {
			items = append(items, &node{kind: nodeScalar, scalar: p})
		}
		return decodeValue(fv, &node{kind: nodeList, list: items}, path)
	}
	return decodeValue(fv, &node{kind: nodeScalar, scalar: f.defaultVal}, path)
}

var cacheStructMeta sync.Map

func inspectStruct(t reflect.Type) (*structMeta, error) {
	v, ok := cacheStructMeta.Load(t)
	if ok {
		return v.(*structMeta), nil
	}

	meta := &structMeta{
		byName: make(map[string]int, t.NumField()),
		fields: make([]*boundField, 0, t.NumField()),
	}
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous {
			continue
		}
		raw, ok := sf.Tag.Lookup("unic")
		if !ok {
			continue
		}
		ft, keep, err := parseUnicTag(raw)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", sf.Name, err)
		}
		if !keep {
			continue
		}
		if !isExported(sf.Name) {
			continue
		}
		bf := &boundField{fieldTag: ft, index: sf.Index, typ: sf.Type}
		if _, exists := meta.byName[ft.name]; exists {
			return nil, fmt.Errorf("duplicate unic tag %q", ft.name)
		}
		meta.fields = append(meta.fields, bf)
		meta.byName[ft.name] = len(meta.fields) - 1
	}

	cacheStructMeta.Store(t, meta)

	return meta, nil
}

func fieldByIndex(v reflect.Value, index []int) (reflect.Value, error) {
	for i, x := range index {
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				if !v.CanSet() {
					return reflect.Value{}, errCanSetEmbeddedPtr
				}
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return reflect.Value{}, errNotStruct
		}
		v = v.Field(x)
		if i == len(index)-1 {
			return alloc(v)
		}
	}
	return v, nil
}

func alloc(v reflect.Value) (reflect.Value, error) {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			if !v.CanSet() {
				return reflect.Value{}, errCantAllocatePtr
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v, nil
}

func (n *node) asAny() (any, error) {
	switch n.kind {
	case nodeScalar:
		if i, err := strconv.ParseInt(n.scalar, 10, 64); err == nil {
			return i, nil
		}
		if f, err := strconv.ParseFloat(n.scalar, 64); err == nil {
			return f, nil
		}
		if b, err := parseBool(n.scalar); err == nil {
			return b, nil
		}
		return n.scalar, nil
	case nodeList:
		out := make([]any, 0, len(n.list))
		for _, item := range n.list {
			v, err := item.asAny()
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case nodeMap:
		out := make(map[string]any, len(n.pairs))
		for _, p := range n.pairs {
			k, err := p.key.asAny()
			if err != nil {
				return nil, err
			}
			ks, ok := k.(string)
			if !ok {
				ks = fmt.Sprint(k)
			}
			v, err := p.val.asAny()
			if err != nil {
				return nil, err
			}
			out[ks] = v
		}
		return out, nil
	case nodeBlock:
		out := make(map[string]any)
		for _, f := range n.fields {
			v, err := f.val.asAny()
			if err != nil {
				return nil, err
			}
			if prev, ok := out[f.key]; ok {
				if sl, ok := prev.([]any); ok {
					out[f.key] = append(sl, v)
				} else {
					out[f.key] = []any{prev, v}
				}
			} else {
				out[f.key] = v
			}
		}
		return out, nil
	default:
		return nil, errUnknownNode
	}
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool %q", s)
	}
}

func isSliceOrArray(v reflect.Value) bool {
	k := v.Kind()
	return k == reflect.Slice || k == reflect.Array
}

func structLike(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct
}

func isEmptyNode(n *node) bool {
	if n == nil || n.kind == nodeEmpty {
		return true
	}
	switch n.kind {
	case nodeScalar:
		return n.scalar == ""
	case nodeList:
		return len(n.list) == 0
	case nodeMap:
		return len(n.pairs) == 0
	case nodeBlock:
		return len(n.fields) == 0 && len(n.attrs) == 0
	default:
		return false
	}
}

func joinPath(base, name string) string {
	if base == "" {
		return name
	}
	return base + "." + name
}

func indexPath(base string, i int) string {
	return base + "[" + strconv.FormatInt(int64(i), 10) + "]"
}

func errAt(path string, err error) error {
	if path == "" {
		return fmt.Errorf("unic: %w", err)
	}
	return fmt.Errorf("unic: %s: %w", path, err)
}
