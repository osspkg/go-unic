package unic

import "testing"

func TestUnitMergeStructKeepsType(t *testing.T) {
	type config struct {
		Name string
		Port int
	}

	got, err := mergeValues(config{Name: "old"}, config{Name: "old", Port: 8080})
	if err != nil {
		t.Fatal(err)
	}
	merged, ok := got.(config)
	if !ok || merged.Name != "old" || merged.Port != 8080 {
		t.Fatalf("merged=%#v", got)
	}
}

func TestUnitMergePointers(t *testing.T) {
	type config struct{ Value int }
	a, b := &config{Value: 1}, &config{Value: 2}

	got, err := mergeValues(a, b)
	if err != nil {
		t.Fatal(err)
	}
	merged, ok := got.(*config)
	if !ok || merged.Value != 2 {
		t.Fatalf("merged=%#v", got)
	}
}

func TestUnitMergeTypedMapWithAny(t *testing.T) {
	a := map[string]any{"nested": map[string]any{"a": 1}}
	b := map[string]any{"nested": map[string]any{"b": 2}}

	got, err := mergeValues(a, b)
	if err != nil {
		t.Fatal(err)
	}
	merged := got.(map[string]any)
	nested := merged["nested"].(map[string]any)
	if nested["a"] != 1 || nested["b"] != 2 {
		t.Fatalf("merged=%#v", got)
	}
}
