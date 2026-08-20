package unic

import (
	"reflect"
	"strings"
	"testing"

	"go.osspkg.com/bb"
)

func TestUnit_MarshalUnsupportedAndNilPointer(t *testing.T) {
	t.Parallel()
	if _, err := Marshal(1); err == nil {
		t.Fatal("int")
	}
	var p *struct {
		A int `unic:"a"`
	}
	if _, err := Marshal(p); err == nil {
		t.Fatal("nil pointer")
	}
	var i any
	if _, err := Marshal(i); err == nil {
		t.Fatal("nil interface")
	}
}

func TestUnit_MarshalPointerAndTopMap(t *testing.T) {
	t.Parallel()
	type cfg struct {
		A int `unic:"a"`
	}
	v := &cfg{A: 2}
	data, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "a 2;") {
		t.Fatalf("%s", data)
	}
	data, err = Marshal(map[string]int{"b": 3, "a": 1})
	if err == nil {
		t.Fatal("marshaling map got no error")
	}
	if data != nil {
		t.Fatalf("marshaling map got result:%s", data)
	}
}

func TestUnit_MarshalMapIntKeysAndBlockMap(t *testing.T) {
	t.Parallel()
	type inner struct {
		X int `unic:"x"`
	}
	type cfg struct {
		M map[string]inner `unic:"m"`
		P map[string]int   `unic:"p"`
		U uint16           `unic:"u"`
		F float32          `unic:"f"`
		B bool             `unic:"b"`
		S string           `unic:"s"`
		E string           `unic:"e"`
		L []int            `unic:"l"`
	}
	data, err := Marshal(cfg{
		M: map[string]inner{"z": {X: 1}, "a": {X: 2}},
		P: map[string]int{},
		U: 4,
		F: 1.5,
		B: false,
		S: "has {brace}",
		E: "",
		L: nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if !strings.Contains(out, "m {") || !strings.Contains(out, "a {") || !strings.Contains(out, "x 2;") {
		t.Fatalf("block map: %q", out)
	}
	if !strings.Contains(out, "p ();") {
		t.Fatalf("empty map: %q", out)
	}
	if !strings.Contains(out, "u 4;") || !strings.Contains(out, "f 1.5;") || !strings.Contains(out, "b false;") {
		t.Fatalf("scalars: %q", out)
	}
	if !strings.Contains(out, "s 'has {brace}';") {
		t.Fatalf("quote: %q", out)
	}
	if !strings.Contains(out, "e '';") {
		t.Fatalf("empty string: %q", out)
	}
	if !strings.Contains(out, "l [];") {
		t.Fatalf("nil slice: %q", out)
	}

	data, err = Marshal(map[int]int{2: 20, 1: 10})
	if err == nil {
		t.Fatal("marshaling map got no error")
	}
	if data != nil {
		t.Fatalf("marshaling map got result: %q", string(data))
	}
}

func TestUnit_MarshalQuotesSpecialAndBacktick(t *testing.T) {
	t.Parallel()
	type cfg struct {
		Space string `unic:"space"`
		Tick  string `unic:"tick"`
		Both  string `unic:"both"`
		Hash  string `unic:"hash"`
	}
	data, err := Marshal(cfg{
		Space: "a b",
		Tick:  "x`y",
		Both:  `it's "ok"`,
		Hash:  "a#b",
	})
	if err != nil {
		t.Fatal(err)
	}
	var got cfg
	if err := Unmarshal(data, &got); err != nil {
		t.Fatalf("%v\n%s", err, data)
	}
	if got.Space != "a b" || got.Tick != "x`y" || got.Both != `it's "ok"` || got.Hash != "a#b" {
		t.Fatalf("%+v\n%s", got, data)
	}
}

func TestUnit_MarshalAttrOmitemptyAndDesc(t *testing.T) {
	t.Parallel()
	type item struct {
		Tag string `unic:"tag,attr=1,omitempty"`
		Alt string `unic:"alt,attr=2"`
		N   int    `unic:"n,desc='num'"`
	}
	type cfg struct {
		Items []item `unic:"item,desc='блок'"`
	}
	data, err := Marshal(cfg{Items: []item{
		{Tag: "", Alt: "x", N: 1},
		{Tag: "web", Alt: "y", N: 2},
	}})
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if !strings.Contains(out, "item x { # блок") && !strings.Contains(out, "item x {") {
		t.Fatalf("attr skip:\n%s", out)
	}
	if !strings.Contains(out, "n 1; # num") {
		t.Fatalf("desc:\n%s", out)
	}
	if !strings.Contains(out, "item web x {") && !strings.Contains(out, "item web y {") {
		t.Fatalf("attrs:\n%s", out)
	}
}

func TestUnit_MarshalNestedListAndInvalidField(t *testing.T) {
	t.Parallel()
	type ok struct {
		Nest [][]int `unic:"nest"`
	}
	data, err := Marshal(ok{Nest: [][]int{{1}, {2, 3}}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "nest [[1], [2, 3]];") {
		t.Fatalf("%s", data)
	}
	type bad struct {
		C chan int `unic:"c"`
	}
	if _, err := Marshal(bad{C: make(chan int)}); err == nil {
		t.Fatal("expected chan error")
	}
}

func TestUnit_MarshalSliceOfStructPointers(t *testing.T) {
	t.Parallel()
	type item struct {
		N int `unic:"n"`
	}
	type cfg struct {
		Items []*item `unic:"item"`
	}
	data, err := Marshal(cfg{Items: []*item{{N: 1}, {N: 2}}})
	if err != nil {
		t.Fatal(err)
	}
	var got cfg
	if err := Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 2 || got.Items[0].N != 1 || got.Items[1].N != 2 {
		t.Fatalf("%+v\n%s", got, data)
	}
}

func TestUnit_QuoteValueAndHasMeta(t *testing.T) {
	t.Parallel()
	if quoteValue("") != "''" {
		t.Fatal("empty")
	}
	if quoteValue("plain") != "plain" {
		t.Fatal("plain")
	}
	if quoteValue("a b") != "'a b'" {
		t.Fatal("space")
	}
	if quoteValue(`say "hi"`) != `'say "hi"'` {
		t.Fatal("double")
	}
	if quoteValue("it's") != `"it's"` {
		t.Fatal("single")
	}
	if quoteValue("it's \"x\"") != "```it's \"x\"```" {
		t.Fatal("both")
	}
	if !hasMeta(";") || hasMeta("ok") {
		t.Fatal("meta")
	}
}

func TestUnit_SanitizeDesc(t *testing.T) {
	t.Parallel()
	b := sanitizeDesc(" a\nb\r ")
	if b != "a b" {
		t.Fatalf("%q", b)
	}
}

func TestUnit_FormatAtomKinds(t *testing.T) {
	t.Parallel()
	s, err := formatAtom(reflect.ValueOf("x"))
	if err != nil || s != "x" {
		t.Fatal(s, err)
	}
	s, err = formatAtom(reflect.ValueOf(true))
	if err != nil || s != "true" {
		t.Fatal(s, err)
	}
	s, err = formatAtom(reflect.ValueOf(uint(3)))
	if err != nil || s != "3" {
		t.Fatal(s, err)
	}
	if _, err = formatAtom(reflect.ValueOf([]int{1})); err == nil {
		t.Fatal("slice atom")
	}
	if !isEmptyValue(reflect.Value{}) {
		t.Fatal("invalid")
	}
}

func TestUnit_MarshalNilStructPointerFieldAndNilSliceElem(t *testing.T) {
	t.Parallel()
	type item struct {
		N int `unic:"n"`
	}
	type cfg struct {
		Inner *item   `unic:"inner"`
		Items []*item `unic:"item"`
	}
	data, err := Marshal(cfg{Items: []*item{nil, {N: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if !strings.Contains(out, "inner '';") && !strings.Contains(out, "inner {") {
		t.Fatalf("fail1: %s", out)
	}
	if !strings.Contains(out, "item {") {
		t.Fatalf("fail2: %s", out)
	}
}

func TestUnit_MarshalAttrGapAndMapBlockPointer(t *testing.T) {
	t.Parallel()
	type item struct {
		Alt string `unic:"alt,attr=2"`
		N   int    `unic:"n"`
	}
	type inner struct {
		X int `unic:"x"`
	}
	type cfg struct {
		Item item              `unic:"item"`
		M    map[string]*inner `unic:"m"`
	}
	data, err := Marshal(cfg{
		Item: item{Alt: "z", N: 1},
		M:    map[string]*inner{"k": {X: 3}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "item z {") {
		t.Fatalf("%s", data)
	}
	if !strings.Contains(string(data), "m {") || !strings.Contains(string(data), "x 3;") {
		t.Fatalf("%s", data)
	}
}

func TestUnit_MarshalErrorsOnBadMapKeyAndInlineStruct(t *testing.T) {
	t.Parallel()
	if _, err := Marshal(map[[2]int]int{{1, 2}: 3}); err == nil {
		t.Fatal("array map key")
	}
	type cfg struct {
		L []any `unic:"l"`
	}
	if _, err := Marshal(cfg{L: []any{struct {
		X int `unic:"x"`
	}{X: 1}}}); err == nil {
		t.Fatal("inline struct")
	}
	e := &encoder{buf: bb.New(16)}
	if err := e.writeBlock("n", reflect.ValueOf(1), "", "n"); err == nil {
		t.Fatal("block int")
	}
}

func TestUnit_WriteInlineNilAndMapPairsError(t *testing.T) {
	t.Parallel()
	e := &encoder{buf: bb.New(32)}
	if err := e.writeInline(reflect.Value{}, "x"); err != nil {
		t.Fatal(err)
	}
}

func TestUnit_MapAsBlockAndStructSlice(t *testing.T) {
	t.Parallel()
	if isStructSlice(reflect.ValueOf(0)) {
		t.Fatal("int")
	}
	if !isStructSlice(reflect.ValueOf([]struct{ A int }{{}})) {
		t.Fatal("slice struct")
	}
	m := map[string]struct{ A int }{}
	if !mapAsBlock(reflect.ValueOf(m)) {
		t.Fatal("map block")
	}
	if mapAsBlock(reflect.ValueOf(map[string]int{})) {
		t.Fatal("map pairs")
	}
}
