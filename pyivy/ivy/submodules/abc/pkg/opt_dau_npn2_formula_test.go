package abc

// Go-only parser, Boolean-model, library and fuzz tests for dauNpn2 formulas.

import (
	"errors"
	"slices"
	"testing"
)

const dauNPN2Example = "~((~~d&~(~~b&c))^(~(~a&~d)&~(~c^~b)))"

func TestDauLogicFormulaExample(t *testing.T) {
	formula, err := ParseDauLogicFormula(dauNPN2Example, DauLogicFormulaOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if formula.VariableCount() != 4 {
		t.Fatal(formula.VariableCount())
	}
	table, err := formula.TruthTable(6, 0, nil)
	if err != nil || table.words[0] != 0x037d037d037d037d {
		t.Fatal(table, err)
	}
	for assignment := 0; assignment < 64; assignment++ {
		inputs := make([]bool, 6)
		for variable := range inputs {
			inputs[variable] = assignment&(1<<uint(variable)) != 0
		}
		a, b, c, d := inputs[0], inputs[1], inputs[2], inputs[3]
		left := d && !(b && c)
		right := (a || d) && !(c != b)
		want := left == right
		got, err := formula.Evaluate(inputs)
		if err != nil || got != want {
			t.Fatal(assignment, got, want, err)
		}
	}
	normalized := formula.String()
	again, err := ParseDauLogicFormula(normalized, DauLogicFormulaOptions{})
	if err != nil || again.String() != normalized {
		t.Fatal(normalized, err)
	}
	againTable, _ := again.TruthTable(6, 0, nil)
	if !slices.Equal(table.Words(), againTable.Words()) {
		t.Fatal("normalization changed semantics")
	}
}

func TestDauLogicFormulaErrorsAndLimits(t *testing.T) {
	for _, text := range []string{"", "g", "a&", "&a", "(a", "a)", "()", "a|b", "a&b&c", "~"} {
		if formula, err := ParseDauLogicFormula(text, DauLogicFormulaOptions{}); formula != nil || !errors.Is(err, ErrFormat) {
			t.Fatal(text, formula, err)
		}
	}
	for _, test := range []struct {
		options DauLogicFormulaOptions
		want    error
	}{
		{DauLogicFormulaOptions{MaxBytes: -1}, ErrDimension},
		{DauLogicFormulaOptions{MaxBytes: 2}, ErrLimit},
		{DauLogicFormulaOptions{MaxNodes: 1}, ErrLimit},
		{DauLogicFormulaOptions{MaxDepth: 1}, ErrLimit},
		{DauLogicFormulaOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if formula, err := ParseDauLogicFormula("(a&b)", test.options); formula != nil || !errors.Is(err, test.want) {
			t.Fatal(test.options, formula, err)
		}
	}
	formula, _ := ParseDauLogicFormula("a", DauLogicFormulaOptions{})
	if value, err := formula.Evaluate(nil); value || !errors.Is(err, ErrDimension) {
		t.Fatal(value, err)
	}
	if table, err := formula.TruthTable(7, 0, nil); table != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(table, err)
	}
}

func TestDauLogicLibraryVerification(t *testing.T) {
	data := []byte("00000000 = (a&~a)\r\naaaaaaaa = a\n66666666 = (a^b)\n037d037d = " + dauNPN2Example + "\n")
	report, err := VerifyDauLogicLibrary(data, 5, DauLogicFormulaOptions{})
	if err != nil || report.Failures() != 0 || len(report.Entries()) != 4 {
		t.Fatal(report, err)
	}
	entries := report.Entries()
	entries[0].Expected = 1
	if !report.Entries()[0].Matches() {
		t.Fatal("report accessor aliases storage")
	}
	mismatch, err := VerifyDauLogicLibrary([]byte("00000001 = a\n"), 5, DauLogicFormulaOptions{})
	if err != nil || mismatch.Failures() != 1 || mismatch.Entries()[0].Matches() {
		t.Fatal(mismatch, err)
	}
	for _, bad := range [][]byte{[]byte("x = a\n"), []byte("00000000 a\n"), []byte("00000000 = z\n"), []byte("100000000 = a\n")} {
		if report, err := VerifyDauLogicLibrary(bad, 5, DauLogicFormulaOptions{}); len(report.Entries()) != 0 || err == nil {
			t.Fatal(string(bad), report, err)
		}
	}
}

func FuzzDauLogicFormula(f *testing.F) {
	f.Add(dauNPN2Example)
	f.Add("~(a^(b&~c))")
	f.Add("~~f")
	f.Fuzz(func(t *testing.T, text string) {
		formula, err := ParseDauLogicFormula(text, DauLogicFormulaOptions{MaxBytes: 256, MaxNodes: 128, MaxDepth: 32})
		if err != nil {
			return
		}
		normalized := formula.String()
		again, err := ParseDauLogicFormula(normalized, DauLogicFormulaOptions{})
		if err != nil || again.String() != normalized {
			t.Fatal(text, normalized, err)
		}
		variables := max(formula.VariableCount(), 1)
		first, err := formula.TruthTable(variables, 0, nil)
		if err != nil {
			t.Fatal(err)
		}
		second, err := again.TruthTable(variables, 0, nil)
		if err != nil || !slices.Equal(first.Words(), second.Words()) {
			t.Fatal(text, normalized, err)
		}
	})
}
