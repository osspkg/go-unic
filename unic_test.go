package unic

import (
	"reflect"
	"strings"
	"testing"
)

const readmeConfig = `
log_level 1;
servers {
	domains ['localhost', 'local.host'];
	server web { # веб
		port 80; # номер порта
		host '127.0.0.1'; # IP или домен
		ttl [1, 2, 3];
		ssl ['/etc/ssl/host1.pem', '/etc/ssl/host1.pem']; # пути для сертификатов
	}
	server admin {
		port 80; # номер порта
		host '127.0.0.2'; # IP или домен
		auth (user1, passwd1, user2, passwd2);
	}
	route admin {
		prefix /api/admin/v1; # префикс методов api для админки
		middleware [log, oauth]; # набор миделвар
	}
}
`

type readmeConfigStruct struct {
	LogLevel int `unic:"log_level,default=1,desc='уровень логирования'"`
	Servers  struct {
		Domains []string `unic:"domains,default='localhost;local.host',desc='список доменов'"`
		Servers []struct {
			Tag  string            `unic:"tag,attr=1,desc='веб'"`
			Port int               `unic:"port,default=80,desc='номер порта'"`
			Host string            `unic:"host,default='127.0.0.1',desc='IP или домен'"`
			Ttl  []int             `unic:"ttl,omitempty"`
			Ssl  []string          `unic:"ssl,omitempty,desc='пути для сертификатов'"`
			Auth map[string]string `unic:"auth,omitempty"`
		} `unic:"server"`
		Routes []struct {
			Tag        string   `unic:"tag,attr=1"`
			Prefix     string   `unic:"prefix"`
			Middleware []string `unic:"middleware"`
		} `unic:"route"`
	} `unic:"servers,desc='настройки серверов'"`
}

func TestUnit_UnmarshalREADME(t *testing.T) {
	var cfg readmeConfigStruct
	if err := Unmarshal([]byte(readmeConfig), &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.LogLevel != 1 {
		t.Fatalf("LogLevel=%d", cfg.LogLevel)
	}
	if got := cfg.Servers.Domains; len(got) != 2 || got[0] != "localhost" || got[1] != "local.host" {
		t.Fatalf("Domains=%v", got)
	}
	if len(cfg.Servers.Servers) != 2 {
		t.Fatalf("servers=%d", len(cfg.Servers.Servers))
	}
	web := cfg.Servers.Servers[0]
	if web.Tag != "web" || web.Port != 80 || web.Host != "127.0.0.1" {
		t.Fatalf("web=%+v", web)
	}
	if len(web.Ttl) != 3 || web.Ttl[0] != 1 || web.Ttl[2] != 3 {
		t.Fatalf("ttl=%v", web.Ttl)
	}
	if len(web.Ssl) != 2 || web.Ssl[0] != "/etc/ssl/host1.pem" {
		t.Fatalf("ssl=%v", web.Ssl)
	}
	admin := cfg.Servers.Servers[1]
	if admin.Tag != "admin" || admin.Host != "127.0.0.2" {
		t.Fatalf("admin=%+v", admin)
	}
	if admin.Auth["user1"] != "passwd1" || admin.Auth["user2"] != "passwd2" {
		t.Fatalf("auth=%v", admin.Auth)
	}
	if len(cfg.Servers.Routes) != 1 {
		t.Fatalf("routes=%d", len(cfg.Servers.Routes))
	}
	rt := cfg.Servers.Routes[0]
	if rt.Tag != "admin" || rt.Prefix != "/api/admin/v1" {
		t.Fatalf("route=%+v", rt)
	}
	if len(rt.Middleware) != 2 || rt.Middleware[0] != "log" || rt.Middleware[1] != "oauth" {
		t.Fatalf("middleware=%v", rt.Middleware)
	}
}

func TestUnit_UnmarshalDefaults(t *testing.T) {
	type cfg struct {
		LogLevel int      `unic:"log_level,default=1"`
		Name     string   `unic:"name,default='svc'"`
		Flags    []string `unic:"flags,default='a;b'"`
	}
	var got cfg
	if err := Unmarshal([]byte(""), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, cfg{}) {
		t.Fatalf("%+v", got)
	}
}

func TestUnit_UnmarshalQuotes(t *testing.T) {
	type cfg struct {
		A string `unic:"a"`
		B string `unic:"b"`
		C string `unic:"c"`
	}
	in := "a 'hello \" world';\nb \"hello ' world\";\nc ```hello '\n\t world\"```;\n"
	var got cfg
	if err := Unmarshal([]byte(in), &got); err != nil {
		t.Fatal(err)
	}
	if got.A != `hello " world` {
		t.Fatalf("a=%q", got.A)
	}
	if got.B != `hello ' world` {
		t.Fatalf("b=%q", got.B)
	}
	if got.C != "hello '\n\t world\"" {
		t.Fatalf("c=%q", got.C)
	}
}

func TestUnit_UnmarshalMyConfig(t *testing.T) {
	type MyConfig struct {
		ServiceName string `unic:"service_name"`
		Port        int    `unic:"port"`
		Features    []bool `unic:"features"`
	}
	var cfg MyConfig
	in := "service_name api-gateway;\nport 8080;\nfeatures [true, false];\n"
	if err := Unmarshal([]byte(in), &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.ServiceName != "api-gateway" || cfg.Port != 8080 {
		t.Fatalf("%+v", cfg)
	}
	if len(cfg.Features) != 2 || !cfg.Features[0] || cfg.Features[1] {
		t.Fatalf("features=%v", cfg.Features)
	}
}

func TestUnit_UnmarshalErrors(t *testing.T) {
	var n int
	if err := Unmarshal([]byte("x 1;"), n); err == nil {
		t.Fatal("expected pointer error")
	}
	type cfg struct {
		Port int `unic:"port"`
	}
	var c cfg
	if err := Unmarshal([]byte("port abc;"), &c); err == nil {
		t.Fatal("expected type error")
	}
	if err := Unmarshal([]byte("foo {"), &c); err == nil {
		t.Fatal("expected unclosed block")
	}
}

func TestUnit_UnmarshalOmitempty(t *testing.T) {
	type cfg struct {
		Ttl  []int `unic:"ttl,omitempty"`
		Port int   `unic:"port,default=80"`
	}
	var got cfg
	if err := Unmarshal([]byte("ttl [];"), &got); err != nil {
		t.Fatal(err)
	}
	if got.Ttl != nil {
		t.Fatalf("ttl=%v", got.Ttl)
	}
	if got.Port != 80 {
		t.Fatalf("port=%d", got.Port)
	}
}

func TestUnit_UnmarshalMapTopLevel(t *testing.T) {
	out := map[string]any{}
	if err := Unmarshal([]byte("a 1;\nb [x, y];\n"), &out); err == nil {
		t.Fatal("unmarshalling map got no error")
	}
}

func TestUnit_MarshalRoundTripREADME(t *testing.T) {
	var cfg readmeConfigStruct
	if err := Unmarshal([]byte(readmeConfig), &cfg); err != nil {
		t.Fatal(err)
	}
	data, err := Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var got readmeConfigStruct
	if err := Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal marshaled: %v\n%s", err, data)
	}
	if !reflect.DeepEqual(cfg, got) {
		t.Fatalf("round trip mismatch\nmarshaled:\n%s\nwant=%+v\ngot=%+v", data, cfg, got)
	}
	out := string(data)
	for _, frag := range []string{
		"log_level 1;",
		"servers {",
		"server web {",
		"server admin {",
		"route admin {",
		"auth (user1, passwd1, user2, passwd2);",
	} {
		if !strings.Contains(out, frag) {
			t.Fatalf("missing %q in:\n%s", frag, out)
		}
	}
}

func TestUnit_MarshalQuotes(t *testing.T) {
	type cfg struct {
		A string `unic:"a"`
		B string `unic:"b"`
		C string `unic:"c"`
	}
	data, err := Marshal(cfg{
		A: `hello " world`,
		B: `hello ' world`,
		C: "hello '\n\t world\"",
	})
	if err != nil {
		t.Fatal(err)
	}
	var got cfg
	if err := Unmarshal(data, &got); err != nil {
		t.Fatalf("%v\n%s", err, data)
	}
	if got.A != `hello " world` || got.B != `hello ' world` || got.C != "hello '\n\t world\"" {
		t.Fatalf("got=%+v\n%s", got, data)
	}
}

func TestUnit_MarshalMultiModel(t *testing.T) {
	type cfg1 struct {
		Port int `unic:"port"`
	}
	type cfg2 struct {
		Ttl []int `unic:"ttl,omitempty"`
	}

	type full1 struct {
		Serv cfg1 `unic:"serv"`
	}
	type full2 struct {
		Serv cfg2 `unic:"serv"`
	}

	data, err := Marshal(
		full1{Serv: cfg1{Port: 123}},
		full2{Serv: cfg2{Ttl: []int{1, 2, 3}}},
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := "serv {\n\tport 123;\n\tttl [1, 2, 3];\n}\n"

	if string(data) != expected {
		t.Fatalf("MarshalMultiModel: got=%q\nwant=%q", string(data), expected)
	}
}

func TestUnit_MarshalMultiModelWithAttr(t *testing.T) {
	type cfg1 struct {
		Tag  string `unic:"tag,attr=1"`
		Port int    `unic:"port"`
	}
	type cfg2 struct {
		Tag string `unic:"tag,attr=1"`
		Ttl []int  `unic:"ttl,omitempty"`
	}

	type full1 struct {
		Serv cfg1 `unic:"serv"`
	}
	type full2 struct {
		Serv cfg2 `unic:"serv"`
	}

	data, err := Marshal(
		full1{Serv: cfg1{Port: 123, Tag: "A"}},
		full2{Serv: cfg2{Ttl: []int{1, 2, 3}, Tag: "B"}},
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `serv A {
	port 123;
}
serv B {
	ttl [1, 2, 3];
}
`

	if string(data) != expected {
		t.Fatalf("TestUnit_MarshalMultiModelWithAttr: got=%q\nwant=%q", string(data), expected)
	}
}

func TestUnit_MarshalOmitempty(t *testing.T) {
	type cfg struct {
		Ttl  []int `unic:"ttl,omitempty"`
		Port int   `unic:"port"`
	}
	data, err := Marshal(cfg{Port: 80})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "ttl") {
		t.Fatalf("omitempty ttl present:\n%s", data)
	}
	if !strings.Contains(string(data), "port 80;") {
		t.Fatalf("missing port:\n%s", data)
	}
}

func TestUnit_MarshalMyConfig(t *testing.T) {
	type MyConfig struct {
		ServiceName string `unic:"service_name"`
		Port        int    `unic:"port"`
		Features    []bool `unic:"features"`
	}
	data, err := Marshal(MyConfig{
		ServiceName: "api-gateway",
		Port:        8080,
		Features:    []bool{true, false},
	})
	if err != nil {
		t.Fatal(err)
	}
	var got MyConfig
	if err := Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.ServiceName != "api-gateway" || got.Port != 8080 || len(got.Features) != 2 || !got.Features[0] || got.Features[1] {
		t.Fatalf("%+v\n%s", got, data)
	}
}

func TestUnit_MarshalNil(t *testing.T) {
	if _, err := Marshal(nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestUnit_DecodeArrayFromRepeatedKeys(t *testing.T) {
	t.Parallel()
	type cfg struct {
		A [2]int `unic:"a"`
	}
	var got cfg
	if err := Unmarshal([]byte("a 1; a 2;"), &got); err != nil {
		t.Fatal(err)
	}
	if got.A != [2]int{1, 2} {
		t.Fatalf("%v", got.A)
	}
	if err := Unmarshal([]byte("a 1; a 2; a 3;"), &got); err == nil {
		t.Fatal("expected overflow")
	}
}

func TestUnit_ParseValueBlockInListAndMissingSemi(t *testing.T) {
	t.Parallel()
	type item struct {
		X int `unic:"x"`
	}
	type cfg struct {
		Items []item `unic:"items"`
		A     int    `unic:"a"`
	}
	var got cfg
	if err := Unmarshal([]byte("items [{ x 1; }, { x 2; }];"), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 2 || got.Items[1].X != 2 {
		t.Fatalf("%+v", got)
	}
	if err := Unmarshal([]byte("a [1, 2]"), &got); err == nil {
		t.Fatal("expected missing semicolon")
	}
}

func TestUnit_UnmarshalSingleBlockIntoSlice(t *testing.T) {
	t.Parallel()
	type item struct {
		N int `unic:"n"`
	}
	type cfg struct {
		Items []item `unic:"item"`
	}
	var got cfg
	if err := Unmarshal([]byte("item { n 5; }"), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 || got.Items[0].N != 5 {
		t.Fatalf("%+v", got)
	}
}

func TestUnit_UnmarshalNilAndPointer(t *testing.T) {
	t.Parallel()
	if err := Unmarshal(nil, nil); err != nil {
		t.Fatal("nil target")
	}
	var p *struct {
		A int `unic:"a"`
	}
	if err := Unmarshal([]byte("a 1;"), p); err == nil {
		t.Fatal("nil pointer")
	}
	var n int
	if err := Unmarshal([]byte("a 1;"), &n); err == nil {
		t.Fatal("non-struct")
	}
}

func TestUnit_ParseSyntaxErrors(t *testing.T) {
	t.Parallel()
	type cfg struct {
		A int `unic:"a"`
	}
	var c cfg
	cases := []string{
		"}",
		"a",
		"a [1,2",
		"a [1 2];",
		"a (x, 1",
		"a (x 1);",
		"a (only);",
		"a { b 1;",
		"; a 1;",
		"a , 1;",
		"a [ {; ];",
		"a 1 extra;",
	}
	for _, in := range cases {
		if err := Unmarshal([]byte(in), &c); err == nil {
			t.Fatalf("expected error for %q", in)
		}
	}
}

func TestUnit_ParseListsMapsEmptyAndTrailingComma(t *testing.T) {
	t.Parallel()
	type cfg struct {
		L []int          `unic:"l"`
		M map[string]int `unic:"m"`
		E []int          `unic:"e"`
		N []int          `unic:"n"`
	}
	var got cfg
	in := "l [1, 2,];\nm (a, 1, b, 2,);\ne [];\nn 3 4;\n"
	if err := Unmarshal([]byte(in), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.L, []int{1, 2}) {
		t.Fatalf("L=%v", got.L)
	}
	if got.M["a"] != 1 || got.M["b"] != 2 {
		t.Fatalf("M=%v", got.M)
	}
	if got.E == nil || len(got.E) != 0 {
		t.Fatalf("E=%v", got.E)
	}
	if !reflect.DeepEqual(got.N, []int{3, 4}) {
		t.Fatalf("N=%v", got.N)
	}
}

func TestUnit_ParseQuotedKeyEmptyValueNested(t *testing.T) {
	t.Parallel()
	type cfg struct {
		Weird string           `unic:"weird key"`
		Empty string           `unic:"empty"`
		Nest  [][]int          `unic:"nest"`
		Maps  []map[string]int `unic:"maps"`
	}
	in := "'weird key' ok;\nempty;\nnest [[1], [2, 3]];\nmaps [(a, 1)];\n"
	var got cfg
	if err := Unmarshal([]byte(in), &got); err != nil {
		t.Fatal(err)
	}
	if got.Weird != "ok" || got.Empty != "" {
		t.Fatalf("%+v", got)
	}
	if len(got.Nest) != 2 || got.Nest[1][1] != 3 {
		t.Fatalf("nest=%v", got.Nest)
	}
	if got.Maps[0]["a"] != 1 {
		t.Fatalf("maps=%v", got.Maps)
	}
}

func TestUnit_UnmarshalScalarsPointersArrays(t *testing.T) {
	t.Parallel()
	type inner struct {
		X int `unic:"x"`
	}
	type cfg struct {
		B        bool    `unic:"b"`
		U        uint    `unic:"u"`
		F        float64 `unic:"f"`
		P        *int    `unic:"p"`
		S        *inner  `unic:"s"`
		Arr      [2]int  `unic:"arr"`
		One      int     `unic:"one"`
		Untagged int
		Skip     string `unic:"-"`
		hidden   int    `unic:"hidden"`
	}
	in := "b false;\nu 8;\nf 1.5;\np 7;\ns { x 4; }\narr [9, 10];\none [11];\n"
	var got cfg
	if err := Unmarshal([]byte(in), &got); err != nil {
		t.Fatal(err)
	}
	if got.B || got.U != 8 || got.F != 1.5 || got.P == nil || *got.P != 7 {
		t.Fatalf("%+v p=%v", got, got.P)
	}
	if got.S == nil || got.S.X != 4 {
		t.Fatalf("s=%+v", got.S)
	}
	if got.Arr != [2]int{9, 10} || got.One != 11 {
		t.Fatalf("arr=%v one=%d", got.Arr, got.One)
	}
	if got.Untagged != 0 || got.Skip != "" || got.hidden != 0 {
		t.Fatalf("skipped fields set: %+v", got)
	}
}

func TestUnit_UnmarshalBoolZeroAndTrue(t *testing.T) {
	t.Parallel()
	type cfg struct {
		A bool `unic:"a"`
		B bool `unic:"b"`
	}
	var got cfg
	if err := Unmarshal([]byte("a 1;\nb 0;"), &got); err != nil {
		t.Fatal(err)
	}
	if !got.A || got.B {
		t.Fatalf("%+v", got)
	}
}

func TestUnit_UnmarshalNestedDefaultsAndAttrDefault(t *testing.T) {
	t.Parallel()
	type item struct {
		Tag string `unic:"tag,attr=1,default=web"`
		N   int    `unic:"n"`
	}
	type cfg struct {
		Inner struct {
			X int `unic:"x,default=3"`
		} `unic:"inner"`
		Item item `unic:"item"`
	}
	var got cfg
	if err := Unmarshal([]byte("item { n 2; }"), &got); err != nil {
		t.Fatal(err)
	}
	if got.Inner.X != 3 {
		t.Fatalf("inner default %d", got.Inner.X)
	}
	if got.Item.Tag != "web" || got.Item.N != 2 {
		t.Fatalf("%+v", got.Item)
	}
}

func TestUnit_UnmarshalBlockMapAndParenMap(t *testing.T) {
	t.Parallel()
	type cfg struct {
		Block map[string]int `unic:"block"`
		Paren map[int]int    `unic:"paren"`
	}
	var got cfg
	in := "block { a 1; b 2; }\nparen (1, 10, 2, 20);\n"
	if err := Unmarshal([]byte(in), &got); err != nil {
		t.Fatal(err)
	}
	if got.Block["a"] != 1 || got.Block["b"] != 2 {
		t.Fatalf("block=%v", got.Block)
	}
	if got.Paren[1] != 10 || got.Paren[2] != 20 {
		t.Fatalf("paren=%v", got.Paren)
	}
}

func TestUnit_UnmarshalInterfaceAndAnyTree(t *testing.T) {
	t.Parallel()
	type cfg struct {
		V any `unic:"v"`
		W any `unic:"w"`
		M any `unic:"m"`
	}
	var got cfg
	in := "v { a 1; a 2; a 3; }\nw 1.25;\nm (k, v);\n"
	if err := Unmarshal([]byte(in), &got); err != nil {
		t.Fatal(err)
	}
	m := got.V.(map[string]any)
	sl := m["a"].([]any)
	if len(sl) != 3 || sl[0].(int64) != 1 || sl[2].(int64) != 3 {
		t.Fatalf("v=%v", got.V)
	}
	if got.W.(float64) != 1.25 {
		t.Fatalf("w=%v", got.W)
	}
	mp := got.M.(map[string]any)
	if mp["k"] != "v" {
		t.Fatalf("m=%v", got.M)
	}
}

func TestUnit_UnmarshalTypeErrors(t *testing.T) {
	t.Parallel()
	type cfgUint struct {
		U uint `unic:"u"`
	}
	type cfgFloat struct {
		F float32 `unic:"f"`
	}
	type cfgDup struct {
		A int `unic:"a"`
	}
	type cfgArr struct {
		A [1]int `unic:"a"`
	}
	type cfgMapKey struct {
		M map[int]int `unic:"m"`
	}
	type cfgDupTag struct {
		A int `unic:"x"`
		B int `unic:"x"`
	}
	type cfgBlock struct {
		N int `unic:"n"`
	}
	cases := []struct {
		dst any
		in  string
	}{
		{&cfgUint{}, "u -1;"},
		{&cfgFloat{}, "f nope;"},
		{&cfgDup{}, "a 1; a 2;"},
		{&cfgArr{}, "a [1, 2];"},
		{&cfgMapKey{}, "m { a 1; }"},
		{&cfgDupTag{}, "x 1;"},
		{&cfgBlock{}, "n { x 1; }"},
		{&cfgDup{}, "a [1, 2];"},
	}
	for _, tt := range cases {
		if err := Unmarshal([]byte(tt.in), tt.dst); err == nil {
			t.Fatalf("expected error for %q -> %T", tt.in, tt.dst)
		}
	}
}

func TestUnit_UnmarshalOmitemptyEmptyMapAndScalar(t *testing.T) {
	t.Parallel()
	type cfg struct {
		S string            `unic:"s,omitempty"`
		M map[string]string `unic:"m,omitempty"`
		B struct {
			X int `unic:"x"`
		} `unic:"b,omitempty"`
	}
	var got cfg
	if err := Unmarshal([]byte("s ;\nm ();\nb {}"), &got); err != nil {
		t.Fatal(err)
	}
	if got.S != "" || got.M != nil || got.B.X != 0 {
		t.Fatalf("%+v", got)
	}
}

func TestUnit_UnmarshalUnknownFieldsIgnored(t *testing.T) {
	t.Parallel()
	type cfg struct {
		A int `unic:"a"`
	}
	var got cfg
	if err := Unmarshal([]byte("a 1;\nzzz 2;"), &got); err != nil {
		t.Fatal(err)
	}
	if got.A != 1 {
		t.Fatal(got)
	}
}

/*
goos: linux
goarch: amd64
pkg: go.osspkg.com/unic
cpu: 12th Gen Intel(R) Core(TM) i9-12900KF
Benchmark_Unmarshal
Benchmark_Unmarshal-24    	  283897	      5598 ns/op	   10318 B/op	     185 allocs/op
PASS
*/
func Benchmark_Unmarshal(b *testing.B) {
	configData := []byte(readmeConfig)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var cfg readmeConfigStruct
			if err := Unmarshal(configData, &cfg); err != nil {
				b.Fatal(err)
			}
		}
	})
}

/*
goos: linux
goarch: amd64
pkg: go.osspkg.com/unic
cpu: 12th Gen Intel(R) Core(TM) i9-12900KF
Benchmark_Marshal
Benchmark_Marshal-24      	 1685868	       707.7 ns/op	    1363 B/op	      35 allocs/op
PASS
*/
func Benchmark_Marshal(b *testing.B) {
	configData := []byte(readmeConfig)
	var cfg readmeConfigStruct
	if err := Unmarshal(configData, &cfg); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := Marshal(cfg); err != nil {
				b.Fatal(err)
			}
		}
	})
}
