package abc

// Go-only reconstruction and fuzz checks for non-DSD set decomposition.

import (
	"errors"
	"slices"
	"testing"
)

func mixedDauDecompositionTruth(t *testing.T) *TruthTable {
	t.Helper()
	return dauTruthFromPredicate(t, 4, func(assignment uint64) bool {
		a := assignment&1 != 0
		b := assignment&2 != 0
		c := assignment&4 != 0
		d := assignment&8 != 0
		selector := d
		if b {
			selector = c
		}
		if selector {
			return a != b
		}
		return b
	})
}

func TestDecomposeDauVariableSetMixedAndDisjoint(t *testing.T) {
	table := mixedDauDecompositionTruth(t)
	for _, spelling := range []string{"Bcd", "bcd"} {
		set, err := ParseDauDecompositionSet(spelling, 4)
		if err != nil {
			t.Fatal(err)
		}
		result, err := DecomposeDauVariableSet(table, set, DauSetDecompositionOptions{})
		if spelling == "bcd" && !modelDauSetWorks(table, set) {
			if err != nil || result.Found {
				t.Fatal(spelling, result, err)
			}
			continue
		}
		if err != nil || !result.Found {
			t.Fatal(spelling, result, err)
		}
		if spelling == "Bcd" && (!slices.Equal(result.FreeVariables, []int{0}) || !slices.Equal(result.UniqueVariables, []int{2, 3}) || !slices.Equal(result.SharedVariables, []int{1})) {
			t.Fatal(result)
		}
		reconstructed, err := result.Recompose(0, nil)
		if err != nil || !reconstructed.Equal(table) {
			t.Fatal(spelling, err)
		}
	}
}

func TestDecomposeDauVariableSetNondecomposableAndLimits(t *testing.T) {
	table := dauTruthFromPredicate(t, 5, func(assignment uint64) bool {
		return ((assignment*0x9e3779b97f4a7c15)>>59)&1 != 0
	})
	var invalid DauDecompositionSet
	foundInvalid := false
	for first := 0; first < 5 && !foundInvalid; first++ {
		for second := first + 1; second < 5; second++ {
			candidate, _ := ParseDauDecompositionSet(string([]byte{byte('a' + first), byte('a' + second)}), 5)
			if !modelDauSetWorks(table, candidate) {
				invalid, foundInvalid = candidate, true
				break
			}
		}
	}
	if !foundInvalid {
		t.Fatal("test function has no nondecomposable pair")
	}
	if result, err := DecomposeDauVariableSet(table, invalid, DauSetDecompositionOptions{}); err != nil || result.Found {
		t.Fatal(result, err)
	}
	valid, _ := ParseDauDecompositionSet("abc", 5)
	if result, err := DecomposeDauVariableSet(table, valid, DauSetDecompositionOptions{MaxChecks: 1}); result.Found || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := DecomposeDauVariableSet(table, valid, DauSetDecompositionOptions{Stop: func() bool { return true }}); result.Found || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	badRole := DauDecompositionSet(2)
	if result, err := DecomposeDauVariableSet(table, badRole, DauSetDecompositionOptions{}); result.Found || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	tooSmall, _ := ParseDauDecompositionSet("a", 5)
	if result, err := DecomposeDauVariableSet(table, tooSmall, DauSetDecompositionOptions{}); result.Found || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
}

func TestDauSetDecompositionOwnershipAndRecomposeValidation(t *testing.T) {
	table := mixedDauDecompositionTruth(t)
	set, _ := ParseDauDecompositionSet("Bcd", 4)
	result, err := DecomposeDauVariableSet(table, set, DauSetDecompositionOptions{})
	if err != nil || !result.Found {
		t.Fatal(result, err)
	}
	fresh, err := DecomposeDauVariableSet(table, set, DauSetDecompositionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	current, _ := result.Composition.Get(0)
	_ = result.Composition.Set(0, !current)
	if result.Composition.Equal(fresh.Composition) {
		t.Fatal("result tables alias")
	}
	if reconstructed, err := fresh.Recompose(0, nil); err != nil || !reconstructed.Equal(table) {
		t.Fatal(err)
	}
	broken := fresh
	broken.FreeVariables = append(slices.Clone(broken.FreeVariables), broken.UniqueVariables[0])
	if reconstructed, err := broken.Recompose(0, nil); reconstructed != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(reconstructed, err)
	}
	if reconstructed, err := fresh.Recompose(0, func() bool { return true }); reconstructed != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(reconstructed, err)
	}
}

func FuzzDecomposeDauVariableSet(f *testing.F) {
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
		if errors.Is(err, ErrUnsupported) {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		if len(sets) > 4 {
			sets = sets[:4]
		}
		for _, set := range sets {
			result, err := DecomposeDauVariableSet(table, set, DauSetDecompositionOptions{MaxChecks: 1 << 20})
			if err != nil || !result.Found {
				t.Fatal(set, result, err)
			}
			reconstructed, err := result.Recompose(0, nil)
			if err != nil || !reconstructed.Equal(table) {
				t.Fatal("reconstruction mismatch", set, err)
			}
		}
	})
}
