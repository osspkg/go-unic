package unic

import "testing"

func TestUnit_KindAndTokenName(t *testing.T) {
	t.Parallel()
	names := map[tokenKind]string{
		tokEOF:    "EOF",
		tokIdent:  "identifier",
		tokString: "string",
		tokLBrace: "'{'",
		tokRBrace: "'}'",
		tokLBrack: "'['",
		tokRBrack: "']'",
		tokLParen: "'('",
		tokRParen: "')'",
		tokSemi:   "';'",
		tokComma:  "','",
		99:        "token",
	}
	for k, want := range names {
		if got := kindName(k); got != want {
			t.Fatalf("kindName(%d)=%q want %q", k, got, want)
		}
	}
	if tokenName(token{kind: tokEOF}) != "EOF" {
		t.Fatal("eof")
	}
	if tokenName(token{kind: tokIdent, val: "x"}) != "x" {
		t.Fatal("val")
	}
	if tokenName(token{kind: tokIdent}) != "identifier" {
		t.Fatal("empty val")
	}
}
