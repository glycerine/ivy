package abc

// Go-only enumeration, bounds and fuzz checks for opt_dau_enum.go.

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestDauEnumLiftAndCombine(t *testing.T) {
	if got, err := DauEnumLift("*<a!bc>", 3, 0); err != nil || got != "*<d!ef>" {
		t.Fatalf("lift=%q, %v", got, err)
	}
	if got, err := DauEnumCombineTwo(true, true, "*a", "*(ab)", 1, false, false, 0); err != nil || got != "*[a(bc)]" {
		t.Fatalf("two=%q, %v", got, err)
	}
	if got, err := DauEnumCombineThree(true, "*a", "*(ab)", "*a", 1, 3, true, false, true, 0); err != nil || got != "*<!a(bc)!d>" {
		t.Fatalf("three=%q, %v", got, err)
	}
	if _, err := DauEnumLift("z", 1, 0); !errors.Is(err, ErrDimension) {
		t.Fatalf("letter overflow: %v", err)
	}
	if _, err := DauEnumLift("abc", 0, 2); !errors.Is(err, ErrLimit) {
		t.Fatalf("byte limit: %v", err)
	}
}

func TestDauEnumerateDSDOrderAndBounds(t *testing.T) {
	sets, err := DauEnumerateDSD(3, DauEnumOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sets) != 4 || len(sets[0]) != 0 || !slices.Equal(sets[1], []string{"*a"}) || !slices.Equal(sets[2], []string{"(ab)", "*[ab]"}) {
		t.Fatalf("small sets: %#v", sets)
	}
	want3 := []string{"(a(bc))", "(a!(bc))", "*[a(bc)]", "(a[bc])", "*[a[bc]]", "*<abc>"}
	if !slices.Equal(sets[3], want3) {
		t.Fatalf("three vars:\n got %q\nwant %q", sets[3], want3)
	}
	formatted := FormatDauEnumeration(sets)
	if !strings.HasPrefix(formatted, "VARIABLE NUMBER 0:\nVARIABLE NUMBER 1:\n*a\n") || !strings.HasSuffix(formatted, "*<abc>\n") {
		t.Fatalf("format=%q", formatted)
	}
	if limited, err := DauEnumerateDSD(3, DauEnumOptions{MaxResults: 2}); limited != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("result limit: %#v, %v", limited, err)
	}
	if cancelled, err := DauEnumerateDSD(3, DauEnumOptions{Stop: func() bool { return true }}); cancelled != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel: %#v, %v", cancelled, err)
	}
	for _, variables := range []int{-1, 27} {
		if _, err := DauEnumerateDSD(variables, DauEnumOptions{}); !errors.Is(err, ErrDimension) {
			t.Errorf("variables %d: %v", variables, err)
		}
	}
}

func TestDauEnumerateDSDFourVariableSyntax(t *testing.T) {
	sets, err := DauEnumerateDSD(4, DauEnumOptions{MaxResults: 100000})
	if err != nil {
		t.Fatal(err)
	}
	if len(sets[4]) == 0 {
		t.Fatal("empty four-variable enumeration")
	}
	for index, marked := range sets[4] {
		formula := strings.TrimPrefix(marked, "*")
		parsed, err := ParseDauDSD(formula, DauDSDOptions{})
		if err != nil {
			t.Fatalf("formula %d %q: %v", index, formula, err)
		}
		if parsed.VariableCount() != 4 {
			t.Fatalf("formula %d %q has %d variables", index, formula, parsed.VariableCount())
		}
		truth, err := parsed.TruthTable(4, 1, nil)
		if err != nil || truth.Check() != nil {
			t.Fatalf("formula %d %q truth: %v / %v", index, formula, err, truth.Check())
		}
	}
	sets[1][0] = "corrupt"
	again, err := DauEnumerateDSD(1, DauEnumOptions{})
	if err != nil || again[1][0] != "*a" {
		t.Fatalf("result storage aliased: %#v, %v", again, err)
	}
	if result, err := DauEnumerateDSD(2, DauEnumOptions{MaxFormulaBytes: 3}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("formula limit: %#v, %v", result, err)
	}
}

func FuzzDauEnumeration(t *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, ^uint64(0)} {
		t.Add(seed)
	}
	sets, err := DauEnumerateDSD(4, DauEnumOptions{MaxResults: 100000})
	if err != nil {
		t.Fatal(err)
	}
	t.Fuzz(func(t *testing.T, seed uint64) {
		formulas := sets[1+int(seed%4)]
		if len(formulas) == 0 {
			return
		}
		formula := strings.TrimPrefix(formulas[int(seed%uint64(len(formulas)))], "*")
		parsed, err := ParseDauDSD(formula, DauDSDOptions{})
		if err != nil {
			t.Fatalf("enumerated %q: %v", formula, err)
		}
		truth, err := parsed.TruthTable(parsed.VariableCount(), 0, nil)
		if err != nil || truth.Check() != nil {
			t.Fatalf("truth %q: %v / %v", formula, err, truth.Check())
		}
	})
}
