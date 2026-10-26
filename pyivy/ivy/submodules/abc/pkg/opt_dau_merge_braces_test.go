package abc

// Go-only semantic and fuzz checks for DSD associative brace removal.

import (
	"errors"
	"slices"
	"testing"
)

func TestDauDSDRemoveBracesForms(t *testing.T) {
	tests := map[string]string{
		"a":                "a",
		"(a(bc))":          "(abc)",
		"(a!(bc))":         "(a!(bc))",
		"[a[bc]]":          "[abc]",
		"[a![bc]]":         "[a!bc]",
		"[!a![!bc]]":       "[!abc]",
		"<a[bc](de)>":      "<a[bc](de)>",
		"96{ab(cd)}":       "96{ab(cd)}",
		"!!(a(b[cd]))":     "(ab[cd])",
		"!!![a[bc]]":       "![abc]",
		"<abc>{d(e(fg))h}": "<abc>{d(efg)h}",
	}
	for input, want := range tests {
		got, err := DauDSDRemoveBraces(input, DauDSDOptions{})
		if err != nil || got != want {
			t.Fatalf("DauDSDRemoveBraces(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
}

func TestDauDSDRemoveBracesSemanticsAndBounds(t *testing.T) {
	inputs := []string{"(a(bc))", "(a!(bc))", "[a[bc]]", "[a![bc]]", "[!a![!bc]]", "<a[bc](de)>", "96{ab(cd)}"}
	for _, input := range inputs {
		before, err := ParseDauDSD(input, DauDSDOptions{})
		if err != nil {
			t.Fatal(input, err)
		}
		result, err := DauDSDRemoveBraces(input, DauDSDOptions{})
		if err != nil {
			t.Fatal(input, err)
		}
		after, err := ParseDauDSD(result, DauDSDOptions{})
		if err != nil {
			t.Fatal(input, result, err)
		}
		variables := max(before.VariableCount(), after.VariableCount())
		left, err := before.TruthTable(variables, 0, nil)
		if err != nil {
			t.Fatal(err)
		}
		right, err := after.TruthTable(variables, 0, nil)
		if err != nil || !slices.Equal(left.Words(), right.Words()) {
			t.Fatal("semantic change", input, result, err)
		}
		again, err := DauDSDRemoveBraces(result, DauDSDOptions{})
		if err != nil || again != result {
			t.Fatal("not idempotent", input, result, again, err)
		}
	}
	if result, err := DauDSDRemoveBraces("(abc)", DauDSDOptions{MaxBytes: 4}); result != "" || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := DauDSDRemoveBraces("(abc)", DauDSDOptions{Stop: func() bool { return true }}); result != "" || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	if result, err := DauDSDRemoveBraces("(a)", DauDSDOptions{}); result != "" || err == nil {
		t.Fatal(result, err)
	}
}

func FuzzDauDSDRemoveBraces(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5})
	f.Add([]byte{7, 8, 9, 10, 11, 12, 13})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 128 {
			data = data[:128]
		}
		input := fuzzDauFormula(data)
		before, err := ParseDauDSD(input, DauDSDOptions{})
		if err != nil {
			return
		}
		result, err := DauDSDRemoveBraces(input, DauDSDOptions{MaxBytes: 2000, MaxNodes: 2000, MaxDepth: 256})
		if err != nil {
			t.Fatal(input, err)
		}
		after, err := ParseDauDSD(result, DauDSDOptions{})
		if err != nil {
			t.Fatal(input, result, err)
		}
		variables := max(before.VariableCount(), after.VariableCount())
		left, err := before.TruthTable(variables, 0, nil)
		if err != nil {
			t.Fatal(err)
		}
		right, err := after.TruthTable(variables, 0, nil)
		if err != nil || !slices.Equal(left.Words(), right.Words()) {
			t.Fatal("semantic change", input, result, err)
		}
		again, err := DauDSDRemoveBraces(result, DauDSDOptions{})
		if err != nil || again != result {
			t.Fatal("not idempotent", input, result, again, err)
		}
	})
}
