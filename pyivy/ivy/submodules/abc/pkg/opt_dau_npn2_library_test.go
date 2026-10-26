package abc

// Go-only formula replay and formatting tests for DTT libraries.

import (
	"errors"
	"slices"
	"testing"
)

func checkDttLibraryEntries(t *testing.T, variables int, entries []DttLibraryEntry) {
	t.Helper()
	for _, entry := range entries {
		formula, err := ParseDauLogicFormula(entry.Formula, DauLogicFormulaOptions{})
		if err != nil {
			t.Fatal(entry, err)
		}
		table, err := formula.TruthTable(variables, 0, nil)
		if err != nil || table.words[0]&dauTruthMask(variables) != entry.Truth {
			t.Fatal(entry, table, err)
		}
	}
}

func TestDttFunctionFormulaReplay(t *testing.T) {
	enumeration, err := EnumerateDttFunctions(DttEnumerationOptions{Variables: 3, MaxFrontier: 8})
	if err != nil {
		t.Fatal(err)
	}
	for index := 1; index < len(enumeration.Functions()); index++ {
		formula, err := enumeration.FunctionFormula(index, DttFormulaOptions{})
		if err != nil {
			t.Fatal(index, err)
		}
		parsed, err := ParseDauLogicFormula(formula, DauLogicFormulaOptions{})
		if err != nil {
			t.Fatal(index, formula, err)
		}
		table, err := parsed.TruthTable(3, 0, nil)
		if err != nil || table.words[0]&0xff != enumeration.Functions()[index].Canonical {
			t.Fatal(index, formula, table, enumeration.Functions()[index], err)
		}
	}
	if formula, err := enumeration.FunctionFormula(0, DttFormulaOptions{}); formula != "" || !errors.Is(err, ErrUnsupported) {
		t.Fatal(formula, err)
	}
	if formula, err := enumeration.FunctionFormula(2, DttFormulaOptions{MaxBytes: 1}); formula != "" || !errors.Is(err, ErrLimit) {
		t.Fatal(formula, err)
	}
	if formula, err := enumeration.FunctionFormula(2, DttFormulaOptions{MaxDepth: 1}); formula != "" || !errors.Is(err, ErrLimit) {
		t.Fatal(formula, err)
	}
	if formula, err := enumeration.FunctionFormula(2, DttFormulaOptions{MaxTransforms: 1}); formula != "" || !errors.Is(err, ErrLimit) {
		t.Fatal(formula, err)
	}
	if formula, err := enumeration.FunctionFormula(2, DttFormulaOptions{Stop: func() bool { return true }}); formula != "" || !errors.Is(err, ErrCancelled) {
		t.Fatal(formula, err)
	}
}

func TestGenerateAndFormatDttLibrary(t *testing.T) {
	enumeration, err := EnumerateDttFunctions(DttEnumerationOptions{Variables: 3, MaxFrontier: 5, Multiplicity: true, MaxFunctions: 10000})
	if err != nil {
		t.Fatal(err)
	}
	entries, err := enumeration.GenerateLibrary(DttLibraryOptions{})
	if err != nil || len(entries) <= len(enumeration.Classes())-2 {
		t.Fatal(len(entries), len(enumeration.Classes()), err)
	}
	checkDttLibraryEntries(t, 3, entries)
	again, err := enumeration.GenerateLibrary(DttLibraryOptions{})
	if err != nil || !slices.Equal(entries, again) {
		t.Fatal("library expansion is not deterministic", err)
	}
	counts := make(map[uint64]int)
	formulas := make(map[uint64]map[string]bool)
	rootCounts := make(map[int]int)
	rootFormulas := make(map[int]map[string]bool)
	for _, entry := range entries {
		counts[entry.Truth]++
		if formulas[entry.Truth] == nil {
			formulas[entry.Truth] = make(map[string]bool)
		}
		formulas[entry.Truth][entry.Formula] = true
		rootCounts[entry.Function]++
		if rootFormulas[entry.Function] == nil {
			rootFormulas[entry.Function] = make(map[string]bool)
		}
		rootFormulas[entry.Function][entry.Formula] = true
	}
	expanded := false
	for truth, count := range counts {
		if count > 1 && len(formulas[truth]) > 1 {
			expanded = true
			break
		}
	}
	if !expanded {
		t.Fatal("no class recursively expanded to distinct alternative formulas")
	}
	recursive := false
	for function, count := range rootCounts {
		if count > 1 && len(rootFormulas[function]) > 1 {
			recursive = true
			break
		}
	}
	if !recursive {
		t.Fatal("no fixed root recursively cross-producted alternative fanin formulas")
	}
	text, err := FormatDttLibrary(entries, 0)
	if err != nil {
		t.Fatal(err)
	}
	report, err := VerifyDauLogicLibrary([]byte(text), 3, DauLogicFormulaOptions{})
	if err != nil || report.Failures() != 0 || len(report.Entries()) != len(entries) {
		t.Fatal(report, err)
	}
	if entries, err := enumeration.GenerateLibrary(DttLibraryOptions{MaxFormulas: 1}); entries != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(entries, err)
	}
	if entries, err := enumeration.GenerateLibrary(DttLibraryOptions{MaxBytes: 1}); entries != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(entries, err)
	}
	if entries, err := enumeration.GenerateLibrary(DttLibraryOptions{Formula: DttFormulaOptions{MaxDepth: 1}}); entries != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(entries, err)
	}
	if entries, err := enumeration.GenerateLibrary(DttLibraryOptions{Formula: DttFormulaOptions{Stop: func() bool { return true }}}); entries != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(entries, err)
	}
	if text, err := FormatDttLibrary(entries, 1); text != "" || !errors.Is(err, ErrLimit) {
		t.Fatal(text, err)
	}
}

func FuzzDttFunctionFormula(f *testing.F) {
	f.Add(uint8(3), uint8(2), false)
	f.Add(uint8(2), uint8(1), true)
	f.Fuzz(func(t *testing.T, width, depth uint8, multiplicity bool) {
		variables := 2 + int(width%2)
		frontier := 1 + int(depth%2)
		enumeration, err := EnumerateDttFunctions(DttEnumerationOptions{Variables: variables, MaxFrontier: frontier, Multiplicity: multiplicity, MaxFunctions: 10000})
		if err != nil {
			t.Fatal(err)
		}
		entries, err := enumeration.GenerateLibrary(DttLibraryOptions{MaxFormulas: 10000, Formula: DttFormulaOptions{MaxBytes: 4096, MaxDepth: 64}})
		if err != nil {
			t.Fatal(err)
		}
		checkDttLibraryEntries(t, variables, entries)
	})
}
