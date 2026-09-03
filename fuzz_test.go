package unic

import (
	"testing"

	"go.osspkg.com/bb"
)

func FuzzParseDocument(f *testing.F) {
	for _, seed := range []string{
		"",
		"name value;",
		"server web 80 { host localhost; }",
		"items [1, 2, [3, 4],];",
		"labels (one, 1, two, 2);",
		"text ```line 1\nline 2```;",
		"broken { [ ( ;",
	} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		buf := bb.FromBytes(data)
		if _, err := buf.Seek(0, bb.SeekStart); err != nil {
			t.Fatal(err)
		}
		_, _ = parseDocument(buf)
	})
}

func FuzzScanDocument(f *testing.F) {
	f.Add([]byte{0xff, 0xfe, 0xfd})
	for _, seed := range []string{
		"a;",
		"# comment\na 'value';",
		"a ```quoted ` value```;",
		"'key with spaces' [true, false, 1.5];",
	} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		buf := bb.FromBytes(data)
		if _, err := buf.Seek(0, bb.SeekStart); err != nil {
			t.Fatal(err)
		}
		s := newScanner(buf)
		for i := 0; i < 10000; i++ {
			tok, err := s.next()
			if err != nil || tok.kind == tokEOF {
				return
			}
		}
		t.Fatal("scanner did not terminate")
	})
}

func FuzzUnmarshalDocument(f *testing.F) {
	for _, seed := range []string{
		"value 1;",
		"nested { value 'text'; }",
		"values [1, 2, 3];",
		"data (name, service, enabled, true);",
	} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var dst struct {
			Value  int `unic:"value"`
			Nested struct {
				Value string `unic:"value"`
			} `unic:"nested"`
			Values []int          `unic:"values"`
			Data   map[string]any `unic:"data"`
		}
		_ = Unmarshal(data, &dst)
	})
}
