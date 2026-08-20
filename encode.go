/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unic

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"go.osspkg.com/bb"
)

type encoder struct {
	buf   *bb.Buffer
	depth int
}

//nolint:unused
func (e *encoder) writeRoot(v reflect.Value, path string) error {
	v = marshalDeref(v)
	if !v.IsValid() {
		return fmt.Errorf("unic: Marshal(nil)")
	}
	switch v.Kind() {
	case reflect.Struct:
		return e.writeStructBody(v, path)
	case reflect.Map:
		return e.writeMapFields(v, path)
	default:
		return fmt.Errorf("unic: Marshal(unsupported type %s)", v.Type())
	}
}

func (e *encoder) writeStructBody(v reflect.Value, path string) error {
	meta, err := inspectStruct(v.Type())
	if err != nil {
		return errAt(path, err)
	}
	for _, f := range meta.fields {
		if f.attr > 0 {
			continue
		}
		fv := valueAt(v, f.index)
		if f.omitempty && isEmptyValue(fv) {
			continue
		}
		if err = e.writeNamed(f, fv, joinPath(path, f.name)); err != nil {
			return err
		}
	}
	return nil
}

func (e *encoder) writeNamed(f *boundField, v reflect.Value, path string) (err error) {
	v = marshalDeref(v)
	if isStructSlice(v) {
		for i := 0; i < v.Len(); i++ {
			if err = e.writeBlock(f.name, v.Index(i), f.desc, joinPath(path, f.name)); err != nil {
				return err
			}
		}
		return nil
	}
	// Только структуры и карты со значениями-структурами/картами идут в блок
	if v.IsValid() && (v.Kind() == reflect.Struct || (v.Kind() == reflect.Map && mapAsBlock(v))) {
		return e.writeBlock(f.name, v, f.desc, path)
	}

	if err = e.writeIndent(); err != nil {
		return err
	}
	if err = e.writeToken(f.name); err != nil {
		return err
	}
	if err = e.writeByte(' '); err != nil {
		return err
	}
	if err = e.writeInline(v, path); err != nil {
		return err
	}
	if err = e.writeByte(';'); err != nil {
		return err
	}
	if err = e.writeDesc(f.desc); err != nil {
		return err
	}
	return e.writeByte('\n')
}

func (e *encoder) writeEmptyBlock(name string) (err error) {
	if err = e.writeIndent(); err != nil {
		return err
	}
	if err = e.writeToken(name); err != nil {
		return err
	}
	if err = e.writeString(" {}"); err != nil {
		return err
	}
	return e.writeByte('\n')
}

func (e *encoder) writeMergedBlock(name string, vals []interface{}, path string) (err error) {
	if err = e.writeIndent(); err != nil {
		return err
	}
	if err = e.writeToken(name); err != nil {
		return err
	}
	if err = e.writeString(" {"); err != nil {
		return err
	}
	if err = e.writeByte('\n'); err != nil {
		return err
	}
	e.depth++
	seen := make(map[string]struct{}, len(vals)*2)
	for _, val := range vals {
		v := reflect.ValueOf(val)
		v = marshalDeref(v)
		if !v.IsValid() || v.Kind() != reflect.Struct {
			continue
		}
		var meta *structMeta
		meta, err = inspectStruct(v.Type())
		if err != nil {
			return errAt(path, err)
		}
		for _, f := range meta.fields {
			if f.attr > 0 {
				continue
			}
			if _, ok := seen[f.name]; ok {
				continue
			}
			fv := valueAt(v, f.index)
			if f.omitempty && isEmptyValue(fv) {
				continue
			}
			if err = e.writeNamed(f, fv, joinPath(path, f.name)); err != nil {
				return err
			}
			seen[f.name] = struct{}{}
		}
	}
	e.depth--
	if err = e.writeIndent(); err != nil {
		return err
	}
	return e.writeString("}\n")
}

func (e *encoder) writeBlock(name string, v reflect.Value, desc, path string) (err error) {
	v = marshalDeref(v)
	if !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil()) {
		if err = e.writeIndent(); err != nil {
			return err
		}
		if err = e.writeToken(name); err != nil {
			return err
		}
		if err = e.writeString(" {}"); err != nil {
			return err
		}
		if err = e.writeDesc(desc); err != nil {
			return err
		}
		return e.writeByte('\n')
	}

	if err = e.writeIndent(); err != nil {
		return err
	}
	if err = e.writeToken(name); err != nil {
		return err
	}
	if v.Kind() == reflect.Struct {
		if err = e.writeAttrs(v, path); err != nil {
			return err
		}
	}
	if err = e.writeString(" {"); err != nil {
		return err
	}
	if err = e.writeDesc(desc); err != nil {
		return err
	}
	if err = e.writeByte('\n'); err != nil {
		return err
	}
	e.depth++
	switch v.Kind() {
	case reflect.Struct:
		err = e.writeStructBody(v, path)
	case reflect.Map:
		err = e.writeMapFields(v, path)
	default:
		err = errAt(path, fmt.Errorf("cannot marshal %s as block", v.Type()))
	}
	e.depth--
	if err != nil {
		return err
	}
	if err = e.writeIndent(); err != nil {
		return err
	}
	return e.writeString("}\n")
}

func (e *encoder) writeAttrs(v reflect.Value, path string) error {
	meta, err := inspectStruct(v.Type())
	if err != nil {
		return errAt(path, err)
	}
	maxIdx := 0
	byAttr := make(map[int]*boundField, 2)
	for _, f := range meta.fields {
		if f.attr > 0 {
			byAttr[f.attr] = f
			if f.attr > maxIdx {
				maxIdx = f.attr
			}
		}
	}
	for i := 1; i <= maxIdx; i++ {
		f, ok := byAttr[i]
		if !ok {
			continue
		}
		fv := marshalDeref(valueAt(v, f.index))
		if f.omitempty && isEmptyValue(fv) {
			continue
		}
		var atom string
		if atom, err = formatAtom(fv); err != nil {
			return errAt(joinPath(path, f.name), err)
		}
		if err = e.writeByte(' '); err != nil {
			return err
		}
		if err = e.writeString(atom); err != nil {
			return err
		}
	}
	return nil
}

func (e *encoder) writeMapFields(v reflect.Value, path string) error {
	keys, err := sortedMapKeys(v)
	if err != nil {
		return errAt(path, err)
	}
	for _, k := range keys {
		var name string
		if name, err = mapFieldName(k); err != nil {
			return errAt(path, err)
		}
		fv := v.MapIndex(k)
		fv = marshalDeref(fv)
		if fv.IsValid() && (fv.Kind() == reflect.Struct || fv.Kind() == reflect.Map) {
			if err = e.writeBlock(name, fv, "", joinPath(path, name)); err != nil {
				return err
			}
		} else {
			f := &boundField{fieldTag: fieldTag{name: name}}
			if err = e.writeNamed(f, v.MapIndex(k), joinPath(path, name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func mapFieldName(k reflect.Value) (string, error) {
	k = marshalDeref(k)
	if k.Kind() == reflect.String {
		return k.String(), nil
	}
	return formatAtom(k)
}

func (e *encoder) writeInline(v reflect.Value, path string) error {
	v = marshalDeref(v)
	if !v.IsValid() {
		return e.writeString("''")
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return e.writeList(v, path)
	case reflect.Map:
		return e.writeMapPairs(v, path)
	case reflect.Struct:
		return errAt(path, errNestedStructBlock)
	default:
		s, err := formatAtom(v)
		if err != nil {
			return errAt(path, err)
		}
		return e.writeString(s)
	}
}

func (e *encoder) writeList(v reflect.Value, path string) (err error) {
	if err = e.writeByte('['); err != nil {
		return err
	}
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			if err = e.writeString(", "); err != nil {
				return err
			}
		}
		if err = e.writeInline(v.Index(i), indexPath(path, i)); err != nil {
			return err
		}
	}
	return e.writeByte(']')
}

func (e *encoder) writeMapPairs(v reflect.Value, path string) (err error) {
	if err = e.writeByte('('); err != nil {
		return err
	}
	var keys []reflect.Value
	if keys, err = sortedMapKeys(v); err != nil {
		return err
	}
	for i, k := range keys {
		if i > 0 {
			if err = e.writeString(", "); err != nil {
				return err
			}
		}
		var ks string

		if ks, err = formatAtom(k); err != nil {
			return errAt(path, err)
		}
		if err = e.writeString(ks); err != nil {
			return err
		}
		if err = e.writeString(", "); err != nil {
			return err
		}
		val := marshalDeref(v.MapIndex(k))
		if val.IsValid() && val.Kind() == reflect.Struct {
			if err = e.writeInlineBlock(val, joinPath(path, ks)); err != nil {
				return err
			}
		} else if err = e.writeInline(val, joinPath(path, ks)); err != nil {
			return err
		}
	}
	return e.writeByte(')')
}

func (e *encoder) writeInlineBlock(v reflect.Value, path string) (err error) {
	if err = e.writeByte(' '); err != nil {
		return err
	}
	if err = e.writeByte('{'); err != nil {
		return err
	}
	e.depth++
	if err = e.writeByte('\n'); err != nil {
		return err
	}
	var meta *structMeta
	if meta, err = inspectStruct(v.Type()); err != nil {
		return errAt(path, err)
	}
	for _, f := range meta.fields {
		if f.attr > 0 {
			continue
		}
		fv := valueAt(v, f.index)
		if f.omitempty && isEmptyValue(fv) {
			continue
		}
		if err = e.writeNamed(f, fv, joinPath(path, f.name)); err != nil {
			return err
		}
	}
	e.depth--
	if err = e.writeIndent(); err != nil {
		return err
	}
	return e.writeByte('}')
}

func (e *encoder) writeDesc(desc string) error {
	if len(desc) == 0 {
		return nil
	}
	desc = sanitizeDesc(desc)
	if len(desc) == 0 {
		return nil
	}
	if err := e.writeString(" # "); err != nil {
		return err
	}
	return e.writeString(desc)
}

func (e *encoder) writeIndent() (err error) {
	for i := 0; i < e.depth; i++ {
		if err = e.writeByte('\t'); err != nil {
			return err
		}
	}
	return nil
}

func (e *encoder) writeToken(s string) error {
	return e.writeString(quoteValue(s))
}

func (e *encoder) writeString(s string) error {
	_, err := e.buf.WriteString(s)
	return err
}

func (e *encoder) writeByte(c byte) error {
	return e.buf.WriteByte(c)
}

func mapAsBlock(v reflect.Value) bool {
	t := v.Type().Elem()
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct || t.Kind() == reflect.Map
}

func isStructSlice(v reflect.Value) bool {
	if !v.IsValid() || (v.Kind() != reflect.Slice && v.Kind() != reflect.Array) {
		return false
	}
	t := v.Type().Elem()
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct
}

func marshalDeref(v reflect.Value) reflect.Value {
	for v.IsValid() && (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

func valueAt(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		v = marshalDeref(v)
		if !v.IsValid() || v.Kind() != reflect.Struct {
			return reflect.Value{}
		}
		v = v.Field(i)
	}
	return v
}

func isEmptyValue(v reflect.Value) bool {
	v = marshalDeref(v)
	if !v.IsValid() {
		return true
	}
	return v.IsZero()
}

func sortedMapKeys(v reflect.Value) ([]reflect.Value, error) {
	keys := v.MapKeys()
	var first error
	sort.Slice(keys, func(i, j int) bool {
		a, err1 := formatAtom(keys[i])
		b, err2 := formatAtom(keys[j])
		if err1 != nil {
			first = err1
			return false
		}
		if err2 != nil {
			first = err2
			return true
		}
		return a < b
	})
	return keys, first
}

func formatAtom(v reflect.Value) (string, error) {
	v = marshalDeref(v)
	if !v.IsValid() {
		return "''", nil
	}
	switch v.Kind() {
	case reflect.String:
		return quoteValue(v.String()), nil
	case reflect.Bool:
		if v.Bool() {
			return "true", nil
		}
		return "false", nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, v.Type().Bits()), nil
	default:
		return "", fmt.Errorf("cannot marshal %s as scalar", v.Type())
	}
}

func quoteValue(s string) string {
	if s == "" {
		return "''"
	}
	var hasSingle, hasDouble, hasTick, hasNL, hasMetaChar bool
	for _, r := range s {
		switch r {
		case '\'':
			hasSingle = true
		case '"':
			hasDouble = true
		case '`':
			hasTick = true
		case '\n', '\r':
			hasNL = true
			goto LabelEnd
		default:
			if unicode.IsSpace(r) || strings.ContainsRune("{}[]()#;,", r) {
				hasMetaChar = true
			}
		}
	}
LabelEnd:
	if !hasSingle && !hasDouble && !hasTick && !hasNL && !hasMetaChar {
		return s
	}
	if hasNL || hasTick || (hasSingle && hasDouble) {
		return "```" + s + "```"
	}
	if hasSingle {
		return `"` + s + `"`
	}
	return "'" + s + "'"
}

//nolint:unused
func hasMeta(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) {
			return true
		}
		switch r {
		case '{', '}', '[', ']', '(', ')', '#', ';', ',', '"', '\'':
			return true
		}
	}
	return false
}

func sanitizeDesc(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}
