package abc

// Go-only semantic, limit, malformed-input, and fuzz tests for parseCore.c.

import (
	"errors"
	"strings"
	"testing"
)

func rankedFormulaValue(t *testing.T, formula string, assignment uint64) bool {
	t.Helper()
	parsed, err := ParseRankedBooleanFormula(formula, []string{"a", "b", "c"}, 0, RankedBooleanParseOptions{})
	if err != nil {
		t.Fatal(formula, err)
	}
	value, err := parsed.Table.Get(assignment)
	if err != nil {
		t.Fatal(formula, assignment, err)
	}
	return value
}

func TestRankedBooleanOperatorsAndPrecedence(t *testing.T) {
	tests := []struct {
		formula string
		model   func(bool, bool, bool) bool
	}{
		{"a&b", func(a, b, _ bool) bool { return a && b }},
		{"a*b", func(a, b, _ bool) bool { return a && b }},
		{"a^b", func(a, b, _ bool) bool { return a != b }},
		{"a<+>b", func(a, b, _ bool) bool { return a != b }},
		{"a+b", func(a, b, _ bool) bool { return a || b }},
		{"a|b", func(a, b, _ bool) bool { return a || b }},
		{"a<=>b", func(a, b, _ bool) bool { return a == b }},
		{"a=>b", func(a, b, _ bool) bool { return !a || b }},
		{"a<=b", func(a, b, _ bool) bool { return a || !b }},
		{"!a", func(a, _, _ bool) bool { return !a }},
		{"~a", func(a, _, _ bool) bool { return !a }},
		{"a'", func(a, _, _ bool) bool { return !a }},
		{"!!a", func(a, _, _ bool) bool { return a }},
		{"a b", func(a, b, _ bool) bool { return a && b }},
		{"a(b)", func(a, b, _ bool) bool { return a && b }},
		{"(a)b", func(a, b, _ bool) bool { return a && b }},
		{"a!b", func(a, b, _ bool) bool { return a && !b }},
		{"a+b*c", func(a, b, c bool) bool { return a || b && c }},
		{"a^b+c", func(a, b, c bool) bool { return (a != b) || c }},
		{"a=>b<=c", func(a, b, c bool) bool { return (!a || b) || !c }},
		{" ( a + b ) ' ", func(a, b, _ bool) bool { return !(a || b) }},
		{"0+a", func(a, _, _ bool) bool { return a }},
		{"1&a", func(a, _, _ bool) bool { return a }},
	}
	for _, test := range tests {
		for assignment := uint64(0); assignment < 8; assignment++ {
			a, b, c := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
			if got, want := rankedFormulaValue(t, test.formula, assignment), test.model(a, b, c); got != want {
				t.Fatalf("%q assignment %03b: got %t want %t", test.formula, assignment, got, want)
			}
		}
	}
}

func TestRankedBooleanRanksAndOwnership(t *testing.T) {
	names := []string{"a", "b"}
	parsed, err := ParseRankedBooleanFormula("a & [a+b] & [[b]]", names, 2, RankedBooleanParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Variables != 2 || parsed.Ranks != 2 || parsed.Table.Variables() != 6 || parsed.Tokens != 13 || parsed.Nodes == 0 {
		t.Fatal(parsed.Variables, parsed.Ranks, parsed.Table.Variables(), parsed.Tokens, parsed.Nodes)
	}
	for assignment := uint64(0); assignment < 64; assignment++ {
		want := assignment&(1<<4) != 0 && (assignment&(1<<2) != 0 || assignment&(1<<3) != 0) && assignment&(1<<1) != 0
		got, getErr := parsed.Table.Get(assignment)
		if getErr != nil || got != want {
			t.Fatalf("assignment %06b: got %t want %t err %v", assignment, got, want, getErr)
		}
	}
	names[0] = "changed"
	if parsed.VariableNames[0] != "a" {
		t.Fatal("variable names alias caller storage")
	}
	words := parsed.Table.Words()
	words[0] = 0
	if value, getErr := parsed.Table.Get(58); getErr != nil || !value {
		t.Fatal("truth words alias caller storage", value, getErr)
	}
}

func TestRankedBooleanDuplicateNamesChooseFirst(t *testing.T) {
	parsed, err := ParseRankedBooleanFormula("x", []string{"x", "x"}, 0, RankedBooleanParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < 4; assignment++ {
		got, _ := parsed.Table.Get(assignment)
		if got != (assignment&1 != 0) {
			t.Fatal(assignment, got)
		}
	}
}

func TestRankedBooleanRejectsMalformedInput(t *testing.T) {
	bad := []string{
		"", " ", "a+", "+a", "'a", "a1", "1a", "a<", "a<+", "a=", "a=b",
		"(a", "a)", "[a", "a]", "([a)]", "a?", "unknown", "a\x00b", "()", "[]",
	}
	for _, formula := range bad {
		if _, err := ParseRankedBooleanFormula(formula, []string{"a", "b"}, 1, RankedBooleanParseOptions{}); err == nil {
			t.Errorf("accepted malformed formula %q", formula)
		}
	}
	if _, err := ParseRankedBooleanFormula("[[a]]", []string{"a"}, 1, RankedBooleanParseOptions{}); err == nil {
		t.Fatal("accepted rank underflow")
	}
	for _, names := range [][]string{{""}, {"1a"}, {"a-b"}} {
		if _, err := ParseRankedBooleanFormula("a", names, 0, RankedBooleanParseOptions{}); !errors.Is(err, ErrDimension) {
			t.Fatal(names, err)
		}
	}
	if _, err := ParseRankedBooleanFormula("a", []string{"a"}, -1, RankedBooleanParseOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestRankedBooleanLimitsAndCancellation(t *testing.T) {
	tests := []struct {
		formula string
		options RankedBooleanParseOptions
	}{
		{"a+b", RankedBooleanParseOptions{MaxTokens: 2}},
		{"((a))", RankedBooleanParseOptions{MaxDepth: 1}},
		{"a+b", RankedBooleanParseOptions{MaxNodes: 2}},
		{"a", RankedBooleanParseOptions{MaxWords: 1}},
	}
	for _, test := range tests {
		_, err := ParseRankedBooleanFormula(test.formula, []string{"a", "b", "c", "d", "e", "f", "g"}, 0, test.options)
		if !errors.Is(err, ErrLimit) {
			t.Errorf("%q: got %v", test.formula, err)
		}
	}
	calls := 0
	_, err := ParseRankedBooleanFormula(strings.Repeat("(", 100)+"a"+strings.Repeat(")", 100), []string{"a"}, 0, RankedBooleanParseOptions{Stop: func() bool {
		calls++
		return calls == 8
	}})
	if !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err = ParseRankedBooleanFormula("a", []string{"a"}, 0, RankedBooleanParseOptions{MaxTokens: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func rankedFuzzLeaf(selector byte) (string, func(uint64) bool) {
	index := int(selector % 6)
	names := []string{"a", "b"}
	rank := index / 2
	name := names[index%2]
	formula := strings.Repeat("[", 2-rank) + name + strings.Repeat("]", 2-rank)
	return formula, func(assignment uint64) bool { return assignment&(uint64(1)<<uint(index)) != 0 }
}

func FuzzParseRankedBooleanFormula(f *testing.F) {
	for _, seed := range [][]byte{{}, {1}, {2, 3, 4}, {5, 6, 7, 8, 9}, {255, 0, 31, 17}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 40 {
			data = data[:40]
		}
		formula, model := rankedFuzzLeaf(byte(len(data)))
		operators := []string{"&", "*", "^", "<+>", "+", "|", "<=>", "=>", "<="}
		for index, value := range data {
			rightFormula, rightModel := rankedFuzzLeaf(value)
			if value&0x10 != 0 {
				rightFormula = "!" + rightFormula
				previous := rightModel
				rightModel = func(assignment uint64) bool { return !previous(assignment) }
			}
			operator := operators[int(value)%len(operators)]
			leftModel := model
			model = func(assignment uint64) bool {
				left, right := leftModel(assignment), rightModel(assignment)
				switch operator {
				case "&", "*":
					return left && right
				case "^", "<+>":
					return left != right
				case "+", "|":
					return left || right
				case "<=>":
					return left == right
				case "=>":
					return !left || right
				default:
					return left || !right
				}
			}
			formula = "(" + formula + operator + rightFormula + ")"
			if index&7 == 7 {
				formula += "'"
				previous := model
				model = func(assignment uint64) bool { return !previous(assignment) }
			}
		}
		parsed, err := ParseRankedBooleanFormula(formula, []string{"a", "b"}, 2, RankedBooleanParseOptions{MaxTokens: 512, MaxDepth: 128, MaxNodes: 512, MaxWords: 1})
		if err != nil {
			t.Fatal(formula, err)
		}
		if err = parsed.Table.Check(); err != nil {
			t.Fatal(err)
		}
		for assignment := uint64(0); assignment < 64; assignment++ {
			got, getErr := parsed.Table.Get(assignment)
			if getErr != nil || got != model(assignment) {
				t.Fatalf("%q assignment %06b: got %t want %t err %v", formula, assignment, got, model(assignment), getErr)
			}
		}
	})
}
