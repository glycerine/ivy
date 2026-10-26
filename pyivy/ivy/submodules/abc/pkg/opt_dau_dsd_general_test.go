package abc

// Stand-alone semantic tests for general opt/dau/dauDsd.c decomposition.

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func dauDSDTestTable(t testing.TB, variables int, value func(uint64) bool) *TruthTable {
	t.Helper()
	table, err := NewConstantTruthTable(variables, false, DauMaxTruthWords)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < uint64(1)<<uint(variables); assignment++ {
		if value(assignment) {
			if err = table.Set(assignment, true); err != nil {
				t.Fatal(err)
			}
		}
	}
	return table
}

func checkDauDSDDecomposition(t testing.TB, source *TruthTable, result DauDSDDecomposition) {
	t.Helper()
	formula, err := ParseDauDSD(result.Formula, DauDSDOptions{})
	if err != nil {
		t.Fatalf("parse %q: %v", result.Formula, err)
	}
	replayed, err := formula.TruthTable(source.variables, DauMaxTruthWords, nil)
	if err != nil {
		t.Fatalf("replay %q: %v", result.Formula, err)
	}
	if !source.Equal(replayed) {
		t.Fatalf("semantic mismatch for %q: %x != %x", result.Formula, source.words, replayed.words)
	}
}

func TestDauDSDDecomposeElementary(t *testing.T) {
	cases := []struct {
		name      string
		variables int
		truth     uint64
		formula   string
	}{
		{"zero", 0, 0, "0"},
		{"one", 0, 1, "1"},
		{"variable", 1, 2, "a"},
		{"not-variable", 1, 1, "!a"},
		{"and", 2, 8, "(ba)"},
		{"or", 2, 14, "!(!b!a)"},
		{"xor", 2, 6, "[ba]"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			table := dauDSDTestTable(t, test.variables, func(assignment uint64) bool {
				return test.truth&(uint64(1)<<assignment) != 0
			})
			before := table.Words()
			result, err := DauDSDDecompose(table, DauDSDDecomposeOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if result.Formula != test.formula || result.LargestPrime != 0 {
				t.Fatalf("got %+v, want formula %q", result, test.formula)
			}
			if !slices.Equal(before, table.Words()) {
				t.Fatal("source truth table changed")
			}
			checkDauDSDDecomposition(t, table, result)
		})
	}
}

func TestDauDSDDecomposeAllSmallFunctions(t *testing.T) {
	for variables := 0; variables <= 3; variables++ {
		bits := 1 << uint(variables)
		functions := uint64(1) << uint(bits)
		for truth := uint64(0); truth < functions; truth++ {
			table := dauDSDTestTable(t, variables, func(assignment uint64) bool {
				return truth&(uint64(1)<<assignment) != 0
			})
			result, err := DauDSDDecompose(table, DauDSDDecomposeOptions{})
			if err != nil {
				t.Fatalf("variables=%d truth=%x: %v", variables, truth, err)
			}
			checkDauDSDDecomposition(t, table, result)
		}
	}
}

func TestDauDSDDecomposeWideAndNested(t *testing.T) {
	parsed, err := ParseDauDSD("[(ab)<c[de](fg)>]", DauDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	nested, err := parsed.TruthTable(7, DauMaxTruthWords, nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := DauDSDDecompose(nested, DauDSDDecomposeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	checkDauDSDDecomposition(t, nested, result)
	boundaryFormula, err := ParseDauDSD("[(ab)(cd)(ef)(gh)(ij)(kl)]", DauDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	boundary, err := boundaryFormula.TruthTable(DauMaxVariables, DauMaxTruthWords, nil)
	if err != nil {
		t.Fatal(err)
	}
	boundaryResult, err := DauDSDDecompose(boundary, DauDSDDecomposeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	checkDauDSDDecomposition(t, boundary, boundaryResult)

	// This deterministic seven-variable function exercises the general prime
	// path rather than the six-variable formatter.
	prime := dauDSDTestTable(t, 7, func(assignment uint64) bool {
		x := assignment*0x9e3779b97f4a7c15 + 0xd1b54a32d192ed03
		return ((x ^ x>>17 ^ x>>41) & 1) != 0
	})
	primeResult, err := DauDSDDecompose(prime, DauDSDDecomposeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if primeResult.LargestPrime < 3 {
		t.Fatalf("expected an irreducible block, got %+v", primeResult)
	}
	checkDauDSDDecomposition(t, prime, primeResult)
	omitted, err := DauDSDDecompose(prime, DauDSDDecomposeOptions{OmitPrimeTruth: true})
	if err != nil || omitted.LargestPrime != primeResult.LargestPrime {
		t.Fatal(omitted, err)
	}
	if omitted.Formula == primeResult.Formula {
		t.Fatal("analysis-only result retained prime truth")
	}
}

func TestDauDSDDecomposeLimitsAndCancellation(t *testing.T) {
	table := dauDSDTestTable(t, 7, func(assignment uint64) bool { return assignment%7 == 2 })
	for _, options := range []DauDSDDecomposeOptions{
		{MaxBytes: 1},
		{MaxNodes: 1},
		{MaxDepth: 1},
		{MaxWords: 1},
		{MaxBytes: -1},
	} {
		if _, err := DauDSDDecompose(table, options); err == nil {
			t.Fatalf("options %+v succeeded", options)
		}
	}
	if _, err := DauDSDDecompose(table, DauDSDDecomposeOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err := DauDSDDecompose(table, DauDSDDecomposeOptions{VariableLevels: []int{1, 2}}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := DauDSDDecompose(nil, DauDSDDecomposeOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestDauDSDSplitPrimeAndOneStep(t *testing.T) {
	decomposable := dauDSDTestTable(t, 2, func(assignment uint64) bool { return assignment == 3 })
	status, err := DauDSDCheckOneStep(decomposable, DauDSDDecomposeOptions{})
	if err != nil || status != -1 {
		t.Fatal(status, err)
	}
	majority := dauDSDTestTable(t, 3, func(assignment uint64) bool {
		return assignment == 3 || assignment == 5 || assignment == 6 || assignment == 7
	})
	status, err = DauDSDCheckOneStep(majority, DauDSDDecomposeOptions{})
	if err != nil || status != 0 {
		t.Fatal(status, err)
	}
	split, err := DauDSDDecompose(majority, DauDSDDecomposeOptions{SplitPrime: true})
	if err != nil || split.LargestPrime != 0 || split.SplitPrimeFailed || !strings.Contains(split.Formula, ">{") {
		t.Fatal(split, err)
	}
	checkDauDSDDecomposition(t, majority, split)
	levels := []int{1, 9, 4}
	status, err = DauDSDCheckOneStep(majority, DauDSDDecomposeOptions{VariableLevels: levels})
	if err != nil || status != 1 || !slices.Equal(levels, []int{1, 9, 4}) {
		t.Fatal(status, levels, err)
	}
	levelSplit, err := DauDSDDecompose(majority, DauDSDDecomposeOptions{SplitPrime: true, VariableLevels: levels})
	if err != nil || !strings.HasPrefix(levelSplit.Formula, "<b") || levelSplit.SplitPrimeFailed {
		t.Fatal(levelSplit, err)
	}
	checkDauDSDDecomposition(t, majority, levelSplit)

	wide := dauDSDTestTable(t, 7, func(assignment uint64) bool {
		x := assignment*0x9e3779b97f4a7c15 + 0xd1b54a32d192ed03
		return ((x ^ x>>17 ^ x>>41) & 1) != 0
	})
	status, err = DauDSDCheckOneStep(wide, DauDSDDecomposeOptions{})
	if err != nil || status != -2 {
		t.Fatal(status, err)
	}
	failed, err := DauDSDDecompose(wide, DauDSDDecomposeOptions{SplitPrime: true})
	if err != nil || !failed.SplitPrimeFailed || failed.LargestPrime < 3 {
		t.Fatal(failed, err)
	}
	checkDauDSDDecomposition(t, wide, failed)
}

func TestDauDSDLevelPriorityAndSynthesizedDefinitions(t *testing.T) {
	expressions := []dauDSDExpression{{level: 0}, {level: 1}, {level: 0}, {level: 1}}
	if priority := dauDSDLevelPriority(expressions, true); !slices.Equal(priority, []int{1, 3, 2, 0}) {
		t.Fatal(priority)
	}
	if priority := dauDSDLevelPriority(expressions, false); priority != nil {
		t.Fatal(priority)
	}
	original := dauDSDExpression{text: "a", level: 3}
	synthesized := dauDSDExpression{text: "(bc)", level: 4, synthesized: true}
	compound := dauDSDCompoundExpression("<a(bc)d>", original, synthesized, dauDSDExpression{text: "d", level: 1})
	if compound.text != "<a(bc)d>" || compound.level != 5 || !compound.synthesized {
		t.Fatal(compound)
	}
}

func TestDauDSDSourceDemoRegressions(t *testing.T) {
	tests := []struct {
		formula   string
		variables int
		split     bool
	}{
		// Dau_DsdTest44 exercises a local MUX binding and prime splitting.
		{"[dc<a(cbd)(!c!b!d)>{abef}]", 6, true},
		// Dau_DsdTest888 exercises the general manager above six inputs.
		{"(<abc>(<def><ghi>))", 9, false},
	}
	for _, test := range tests {
		parsed, err := ParseDauDSD(test.formula, DauDSDOptions{})
		if err != nil {
			t.Fatal(test.formula, err)
		}
		truth, err := parsed.TruthTable(test.variables, DauMaxTruthWords, nil)
		if err != nil {
			t.Fatal(test.formula, err)
		}
		result, err := DauDSDDecompose(truth, DauDSDDecomposeOptions{SplitPrime: test.split})
		if err != nil {
			t.Fatal(test.formula, err)
		}
		checkDauDSDDecomposition(t, truth, result)
	}
}

func FuzzDauDSDGeneralSemantic(f *testing.F) {
	f.Add(byte(3), []byte{0xe8})
	f.Add(byte(7), []byte{0x96, 0x3c, 0xa5, 0x5a, 0xc3, 0x69, 0xf0, 0x0f})
	f.Fuzz(func(t *testing.T, width byte, data []byte) {
		variables := int(width%8) + 1
		if len(data) == 0 {
			data = []byte{0}
		}
		table := dauDSDTestTable(t, variables, func(assignment uint64) bool {
			index := assignment % uint64(len(data)*8)
			return data[index/8]&(1<<uint(index%8)) != 0
		})
		result, err := DauDSDDecompose(table, DauDSDDecomposeOptions{MaxBytes: 2000, MaxNodes: 4096, MaxDepth: 128})
		if err != nil {
			t.Fatal(err)
		}
		checkDauDSDDecomposition(t, table, result)
	})
}

func FuzzDauDSDSplitPrimeSemantic(f *testing.F) {
	f.Add(byte(3), []byte{0xe8})
	f.Add(byte(6), []byte{0x96, 0x3c, 0xa5, 0x5a, 0xc3, 0x69, 0xf0, 0x0f})
	f.Fuzz(func(t *testing.T, width byte, data []byte) {
		variables := int(width%6) + 1
		if len(data) == 0 {
			data = []byte{0}
		}
		table := dauDSDTestTable(t, variables, func(assignment uint64) bool {
			index := assignment % uint64(len(data)*8)
			return data[index/8]&(1<<uint(index%8)) != 0
		})
		levels := make([]int, variables)
		for variable := range levels {
			levels[variable] = int(data[variable%len(data)] & 15)
		}
		result, err := DauDSDDecompose(table, DauDSDDecomposeOptions{SplitPrime: true, VariableLevels: levels, MaxBytes: 2000, MaxNodes: 4096, MaxDepth: 128})
		if err != nil {
			t.Fatal(err)
		}
		checkDauDSDDecomposition(t, table, result)
	})
}
