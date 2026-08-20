package unic

import (
	"reflect"
	"testing"
)

func TestUnit_ParseUnicTag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      string
		want    fieldTag
		keep    bool
		wantErr bool
	}{
		{in: "", keep: false},
		{in: "-", keep: false},
		{in: "  ", keep: false},
		{
			in:   "port,default=80,desc='номер порта'",
			keep: true,
			want: fieldTag{name: "port", hasDefault: true, defaultVal: "80", desc: "номер порта"},
		},
		{
			in:   `host,default="127.0.0.1",omitempty`,
			keep: true,
			want: fieldTag{name: "host", hasDefault: true, defaultVal: "127.0.0.1", omitempty: true},
		},
		{
			in:   "tag,attr=1,unknown=x",
			keep: true,
			want: fieldTag{name: "tag", attr: 1},
		},
		{
			in:   "flags,default='a;b'",
			keep: true,
			want: fieldTag{name: "flags", hasDefault: true, defaultVal: "a;b"},
		},
		{in: "x,attr", wantErr: true},
		{in: "x,attr=0", wantErr: true},
		{in: "x,attr=-1", wantErr: true},
		{in: "x,attr=nope", wantErr: true},
		{in: "x,desc='oops", wantErr: true},
	}
	for _, tt := range tests {
		got, keep, err := parseUnicTag(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("parseUnicTag(%q) err=nil", tt.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("parseUnicTag(%q): %v", tt.in, err)
		}
		if keep != tt.keep || !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("parseUnicTag(%q)=(%+v,%v) want (%+v,%v)", tt.in, got, keep, tt.want, tt.keep)
		}
	}
}

func TestUnit_SplitDefaultList(t *testing.T) {
	t.Parallel()
	if splitDefaultList("") != nil {
		t.Fatal("empty")
	}
	got := splitDefaultList(" a ; ; b ;")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("%v", got)
	}
}

func TestUnit_ParseUnicTagEmptyNameAndCommas(t *testing.T) {
	t.Parallel()
	_, keep, err := parseUnicTag(",omitempty")
	if err != nil || keep {
		t.Fatalf("keep=%v err=%v", keep, err)
	}
	ft, keep, err := parseUnicTag("port,,desc='x'")
	if err != nil || !keep || ft.name != "port" || ft.desc != "x" {
		t.Fatalf("%+v %v %v", ft, keep, err)
	}
}
