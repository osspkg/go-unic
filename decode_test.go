package unic

import (
	"reflect"
	"strings"
	"testing"
)

func TestUnit_FieldByIndexAndAlloc(t *testing.T) {
	t.Parallel()
	type inner struct{ X int }
	type wrap struct{ Inner *inner }
	var w wrap
	fv, err := fieldByIndex(reflect.ValueOf(&w).Elem(), []int{0, 0})
	if err != nil {
		t.Fatal(err)
	}
	fv.SetInt(9)
	if w.Inner == nil || w.Inner.X != 9 {
		t.Fatalf("%+v", w)
	}
	if _, err = fieldByIndex(reflect.ValueOf(1), []int{0}); err == nil {
		t.Fatal("expected not a struct")
	}
	var p *int
	if _, err = alloc(reflect.ValueOf(p)); err == nil {
		t.Fatal("expected unsettable pointer")
	}
}

func TestUnit_DecodeUnknownNodeAndStructLikePointer(t *testing.T) {
	t.Parallel()
	var n int
	if err := decodeValue(reflect.ValueOf(&n).Elem(), &node{kind: 99}, "n"); err == nil {
		t.Fatal("unknown node")
	}
	type cfg struct {
		Inner *struct {
			X int `unic:"x,default=4"`
		} `unic:"inner"`
	}
	var got cfg
	if err := Unmarshal([]byte(""), &got); err != nil {
		t.Fatal(err)
	}
	if got.Inner != nil {
		t.Fatalf("%+v", got)
	}
}

func TestUnit_HelpersEmptyNodeAndAsAny(t *testing.T) {
	t.Parallel()
	if !isEmptyNode(&node{}) {
		t.Fatal("nil")
	}
	if !isEmptyNode(nil) {
		t.Fatal("nil")
	}
	if !isEmptyNode(&node{kind: nodeMap}) || !isEmptyNode(&node{kind: nodeBlock}) {
		t.Fatal("empty containers")
	}
	if isEmptyNode(&node{kind: 99}) {
		t.Fatal("unknown")
	}
	if _, err := (&node{kind: 99}).asAny(); err == nil {
		t.Fatal("asAny unknown")
	}
	if err := decodeValue(reflect.ValueOf(0), nil, ""); err != nil {
		t.Fatal(err)
	}
	if err := errAt("", ioStrError("x")); err == nil || !strings.Contains(err.Error(), "unic:") {
		t.Fatalf("%v", err)
	}
}

type ioStrError string

func (e ioStrError) Error() string { return string(e) }
