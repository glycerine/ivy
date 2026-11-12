package parse

import (
	"strings"
	"testing"
)

func TestFormulaOperatorsAndPrecedence(t *testing.T) {
	for _, tc := range []struct {
		s string
		f func(bool, bool, bool) bool
	}{
		{"!a * b + c", func(a, b, c bool) bool { return !a && b || c }},
		{"a & b ^ c", func(a, b, c bool) bool { return (a && b) != c }},
		{"a <=> b", func(a, b, c bool) bool { return a == b }},
		{"a <+> b", func(a, b, c bool) bool { return a != b }},
		{"a => b", func(a, b, c bool) bool { return !a || b }},
		{"a <= b", func(a, b, c bool) bool { return a || !b }},
		{"a' (b + ~c)", func(a, b, c bool) bool { return !a && (b || !c) }},
		{"(a | b)'", func(a, b, c bool) bool { return !(a || b) }},
		{"a b", func(a, b, c bool) bool { return a && b }},
		{"0 + 1 * a", func(a, b, c bool) bool { return a }},
	} {
		e, err := ParseFormula(tc.s, []string{"a", "b", "c"})
		if err != nil {
			t.Fatal(tc.s, err)
		}
		for bits := 0; bits < 8; bits++ {
			v := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			got, err := e.Evaluate(v)
			if err != nil || got != tc.f(v[0], v[1], v[2]) {
				t.Fatal(tc.s, bits, got, err)
			}
		}
	}
}
func TestEquationGrammar(t *testing.T) {
	e, err := ParseEquation("!a * b ^ c + 1", []string{"a", "b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		v := []bool{i&1 != 0, i&2 != 0, i&4 != 0}
		got, err := e.Evaluate(v)
		if err != nil || !got {
			t.Fatal(got, err)
		}
	}
	for _, s := range []string{"a b", "a(b)", "~a", "a'", "[a]", "a&b", "a|b", "a=>b"} {
		if _, err := ParseEquation(s, []string{"a", "b"}); err == nil {
			t.Fatal("accepted", s)
		}
	}
}
func TestFormulaErrors(t *testing.T) {
	for _, s := range []string{"", "a +", "*a", "(a", "a)", "[a)", "unknown", "a <> b", "a == b"} {
		if _, err := ParseFormula(s, []string{"a", "b"}); err == nil {
			t.Fatal("accepted", s)
		}
	}
	if _, err := ParseFormula("a", []string{"a", "a"}); err == nil {
		t.Fatal("duplicate")
	}
	e, _ := ParseFormula("b", []string{"a", "b"})
	if _, err := e.Evaluate([]bool{true}); err == nil {
		t.Fatal("missing assignment")
	}
}
func TestRankCompilation(t *testing.T) {
	e, err := ParseFormula("[[a]] + b", []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	ops := Operations[string]{Constant: func(v bool) string {
		if v {
			return "1"
		}
		return "0"
	}, Variable: func(i int) string { return []string{"a", "b"}[i] }, Or: func(a, b string) string { return "(" + a + "+" + b + ")" }, Shift: func(v string, n int) (string, error) {
		if n != -2 {
			t.Fatal(n)
		}
		return "shift(" + v + ")", nil
	}}
	got, err := Compile(e, ops, 2)
	if err != nil || got != "(shift(shift(a))+b)" {
		t.Fatal(got, err)
	}
	if _, err := e.Evaluate([]bool{true, false}); err == nil || !strings.Contains(err.Error(), "rank") {
		t.Fatal(err)
	}
}

func TestOriginalEntryPoints(t *testing.T) {
	ops := Operations[bool]{Constant: func(v bool) bool { return v }, Variable: func(i int) bool { return []bool{true, false}[i] }, Not: func(v bool) bool { return !v }, And: func(a, b bool) bool { return a && b }, Or: func(a, b bool) bool { return a || b }, Xor: func(a, b bool) bool { return a != b }}
	if !Parse_FormulaParser(nil, "a + b", 2, 0, []string{"a", "b"}, ops, []bool{true, false}) {
		t.Fatal("formula adapter")
	}
	if !Parse_FormulaParserEqn(nil, "a + b", []string{"a", "b"}, ops) {
		t.Fatal("equation adapter")
	}
	var output strings.Builder
	if Parse_FormulaParserEqn(&output, "missing", []string{"a"}, ops) || !strings.Contains(output.String(), "unknown variable") {
		t.Fatal(output.String())
	}
}
