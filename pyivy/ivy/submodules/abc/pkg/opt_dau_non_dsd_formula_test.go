package abc

// Go-only semantic checks for non-DSD composition/decomposition formulas.

import (
	"errors"
	"math/bits"
	"strings"
	"testing"
)

func checkDauSetFormulas(t *testing.T, source *TruthTable, result DauSetDecomposition) {
	t.Helper()
	formulas, err := result.DSDFormulas(DauSetFormulaOptions{})
	if err != nil {
		t.Fatal(err)
	}
	composition, err := ParseDauDSD(formulas.Composition, DauDSDOptions{})
	if err != nil {
		t.Fatal(formulas, err)
	}
	decomposition, err := ParseDauDSD(formulas.Decomposition, DauDSDOptions{})
	if err != nil {
		t.Fatal(formulas, err)
	}
	if formulas.SelectorVariable != source.variables {
		t.Fatal(formulas)
	}
	for assignment := uint64(0); assignment < 1<<uint(source.variables); assignment++ {
		selector, err := decomposition.eval(decomposition.root, assignment, nil)
		if err != nil {
			t.Fatal(err)
		}
		compositionAssignment := assignment
		if selector {
			compositionAssignment |= 1 << uint(formulas.SelectorVariable)
		}
		got, err := composition.eval(composition.root, compositionAssignment, nil)
		want, sourceErr := source.Get(assignment)
		if err != nil || sourceErr != nil || got != want {
			t.Fatal("formula reconstruction", formulas, assignment, got, want, err, sourceErr)
		}
	}
}

func TestDauSetDecompositionDSDFormulas(t *testing.T) {
	table := mixedDauDecompositionTruth(t)
	set, _ := ParseDauDecompositionSet("Bcd", 4)
	result, err := DecomposeDauVariableSet(table, set, DauSetDecompositionOptions{})
	if err != nil || !result.Found {
		t.Fatal(result, err)
	}
	checkDauSetFormulas(t, table, result)
	first, err := result.DSDFormulas(DauSetFormulaOptions{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := result.DSDFormulas(DauSetFormulaOptions{})
	if err != nil || first != second {
		t.Fatal(first, second, err)
	}
	if formulas, err := result.DSDFormulas(DauSetFormulaOptions{MaxBytes: 1}); formulas != (DauSetFormulas{}) || !errors.Is(err, ErrLimit) {
		t.Fatal(formulas, err)
	}
	if formulas, err := result.DSDFormulas(DauSetFormulaOptions{Stop: func() bool { return true }}); formulas != (DauSetFormulas{}) || !errors.Is(err, ErrCancelled) {
		t.Fatal(formulas, err)
	}
}

func TestDauSetDecompositionDSDFormulasWideComponent(t *testing.T) {
	// F(a..h) is C(a..f, D(g,h)), where both C and D are parity.  C has
	// seven inputs, so this exercises the general decomposer beyond the old
	// six-variable formatter without relying on a native oracle.
	table := dauTruthFromPredicate(t, 8, func(assignment uint64) bool {
		return bits.OnesCount64(assignment)&1 != 0
	})
	set, err := ParseDauDecompositionSet("gh", 8)
	if err != nil {
		t.Fatal(err)
	}
	result, err := DecomposeDauVariableSet(table, set, DauSetDecompositionOptions{})
	if err != nil || !result.Found || result.Composition.variables != 7 || result.Decomposition.variables != 2 {
		t.Fatal(result, err)
	}
	formulas, err := result.DSDFormulas(DauSetFormulaOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(formulas.Composition, "{") {
		t.Fatalf("wide decomposable component was emitted as a prime: %q", formulas.Composition)
	}
	composition, err := ParseDauDSD(formulas.Composition, DauDSDOptions{})
	if err != nil || composition.VariableCount() != 9 {
		t.Fatal(formulas.Composition, err)
	}
	checkDauSetFormulas(t, table, result)
}

func FuzzDauSetDecompositionDSDFormulas(f *testing.F) {
	f.Add(uint8(4), []byte{0x96, 0x69})
	f.Add(uint8(5), []byte{0x31, 0x72, 0x55, 0xaa})
	f.Fuzz(func(t *testing.T, width uint8, data []byte) {
		variables := 3 + int(width%4)
		if len(data) == 0 {
			return
		}
		table, err := NewConstantTruthTable(variables, false, 0)
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 1<<uint(variables); assignment++ {
			if data[assignment%len(data)]&(1<<uint(assignment&7)) != 0 {
				_ = table.Set(uint64(assignment), true)
			}
		}
		sets, err := FindDauDecompositionSets(table, DauDecompositionSetOptions{MaxChecks: 1 << 24})
		if errors.Is(err, ErrUnsupported) || len(sets) == 0 {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		result, err := DecomposeDauVariableSet(table, sets[0], DauSetDecompositionOptions{MaxChecks: 1 << 20})
		if err != nil || !result.Found {
			t.Fatal(result, err)
		}
		checkDauSetFormulas(t, table, result)
	})
}

func FuzzDauSetDecompositionWideDSDFormulas(f *testing.F) {
	f.Add([]byte{0x96, 0x3c, 0xa5, 0x5a, 0xc3, 0x69, 0xf0, 0x0f})
	f.Add([]byte{0x81, 0x42, 0x24, 0x18})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		// Build an arbitrary seven-input C and compose its last input with
		// D(g,h)=g XOR h. This guarantees a valid eight-input decomposition
		// while keeping the expected result independent of the implementation.
		table := dauTruthFromPredicate(t, 8, func(assignment uint64) bool {
			selector := uint64(bits.OnesCount64((assignment>>6)&3) & 1)
			index := (assignment & 63) | selector<<6
			bit := index % uint64(len(data)*8)
			return data[bit/8]&(1<<uint(bit&7)) != 0
		})
		set, err := ParseDauDecompositionSet("gh", 8)
		if err != nil {
			t.Fatal(err)
		}
		result, err := DecomposeDauVariableSet(table, set, DauSetDecompositionOptions{MaxChecks: 1 << 20})
		if err != nil || !result.Found {
			t.Fatal(result, err)
		}
		checkDauSetFormulas(t, table, result)
	})
}
