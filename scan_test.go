package unic

import (
	"io"
	"strings"
	"testing"

	"go.osspkg.com/bb"
)

func TestUnit_ScannerTokensAndErrors(t *testing.T) {
	t.Parallel()
	buf := bb.FromBytes([]byte("k { } [ ] ( ) ; ,"))
	if _, err := buf.Seek(0, bb.SeekStart); err != nil {
		t.Fatal(err)
	}
	s := newScanner(buf)
	var kinds []tokenKind
	for {
		tok, err := s.next()
		if err != nil {
			t.Fatal(err)
		}
		kinds = append(kinds, tok.kind)
		if tok.kind == tokEOF {
			break
		}
	}
	want := []tokenKind{tokIdent, tokLBrace, tokRBrace, tokLBrack, tokRBrack, tokLParen, tokRParen, tokSemi, tokComma, tokEOF}
	if len(kinds) != len(want) {
		t.Fatalf("%v", kinds)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("i=%d %v", i, kinds)
		}
	}
}

func TestUnit_ScannerPeekTwice(t *testing.T) {
	t.Parallel()
	buf := bb.FromBytes([]byte("ab"))
	if _, err := buf.Seek(0, bb.SeekStart); err != nil {
		t.Fatal(err)
	}
	s := newScanner(buf)
	a, err := s.peek()
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.peek()
	if err != nil {
		t.Fatal(err)
	}
	if a != b || a.val != "ab" {
		t.Fatalf("%+v %+v", a, b)
	}
	n, err := s.next()
	if err != nil || n.val != "ab" {
		t.Fatalf("%+v %v", n, err)
	}
}

func TestUnit_ScannerRawEmbeddedTick(t *testing.T) {
	t.Parallel()
	type cfg struct {
		S string `unic:"s"`
	}
	var got cfg
	if err := Unmarshal([]byte("s ```a`b```;"), &got); err != nil {
		t.Fatal(err)
	}
	if got.S != "a`b" {
		t.Fatalf("%q", got.S)
	}
}

func TestUnit_ScannerErrors(t *testing.T) {
	t.Parallel()
	type cfg struct {
		S string `unic:"s"`
	}
	var c cfg
	for _, in := range []string{
		"s 'oops;",
		"s \"oops;",
		"s ```oops;",
		"s `nope;",
		"s ``;",
	} {
		if err := Unmarshal([]byte(in), &c); err == nil {
			t.Fatalf("expected error for %q", in)
		}
	}
}

func TestUnit_ScannerCommentEOF(t *testing.T) {
	t.Parallel()
	type cfg struct {
		A int `unic:"a"`
	}
	var got cfg
	if err := Unmarshal([]byte("a 1; # trailing"), &got); err != nil {
		t.Fatal(err)
	}
	if got.A != 1 {
		t.Fatal(got.A)
	}
}

func TestUnit_ScannerWrap(t *testing.T) {
	t.Parallel()
	s := newScanner(bb.New(8))
	err := s.wrap(io.EOF)
	if err == nil || !strings.Contains(err.Error(), "unic:1:1") {
		t.Fatalf("%v", err)
	}
}
