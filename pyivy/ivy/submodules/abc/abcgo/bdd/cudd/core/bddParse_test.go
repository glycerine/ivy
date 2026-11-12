package core

import (
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/parse"
)

func TestBDDFormulaParserParity(t *testing.T) {
	dd := NewBDDManager(3)
	for _, formula := range []string{"a*b + !c", "a ^ b ^ c", "(a+b)*(!a+c)", "1", "0", "a => b"} {
		expression, err := parse.ParseFormula(formula, []string{"a", "b", "c"})
		if err != nil {
			t.Fatal(err)
		}
		var diagnostics strings.Builder
		result := ParseFormula(&diagnostics, formula, 3, 0, []string{"a", "b", "c"}, dd, dd.Vars)
		if result == nil {
			t.Fatal(diagnostics.String())
		}
		for bits := 0; bits < 8; bits++ {
			assignment := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			want, err := expression.Evaluate(assignment)
			if err != nil {
				t.Fatal(err)
			}
			got, err := dd.Evaluate(result, assignment)
			if err != nil || got != want {
				t.Fatalf("%s at %d: got %v want %v (%v)", formula, bits, got, want, err)
			}
		}
	}
}
func TestBDDFormulaRankShift(t *testing.T) {
	dd := NewBDDManager(6)
	names := []string{"a", "b"}
	vars := dd.Vars[4:6]
	var diagnostics strings.Builder
	result := ParseFormula(&diagnostics, "[[a]] + [b]", 2, 2, names, dd, vars)
	expected := Cudd_bddOr(dd, dd.Vars[0], dd.Vars[3])
	if result != expected {
		t.Fatal("rank mismatch", diagnostics.String())
	}
	if ParseFormula(&diagnostics, "[[[a]]]", 2, 2, names, dd, vars) != nil {
		t.Fatal("out-of-range rank accepted")
	}
}
