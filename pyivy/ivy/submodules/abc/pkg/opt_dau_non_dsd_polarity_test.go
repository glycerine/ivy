package abc

// Go-only exhaustive checks for shared-minterm polarity families.

import (
	"errors"
	"testing"
)

func dauPolarityTruth(t *testing.T, shared int) (*TruthTable, DauDecompositionSet) {
	t.Helper()
	variables := shared + 3
	spelling := make([]byte, 0, shared+2)
	for variable := 0; variable < shared; variable++ {
		spelling = append(spelling, byte('A'+variable))
	}
	spelling = append(spelling, byte('a'+shared), byte('a'+shared+1))
	set, err := ParseDauDecompositionSet(string(spelling), variables)
	if err != nil {
		t.Fatal(err)
	}
	table := dauTruthFromPredicate(t, variables, func(assignment uint64) bool {
		unique0 := assignment&(1<<uint(shared)) != 0
		unique1 := assignment&(1<<uint(shared+1)) != 0
		free := assignment&(1<<uint(shared+2)) != 0
		selector := unique0 != unique1
		for variable := 0; variable < shared; variable++ {
			bit := assignment&(1<<uint(variable)) != 0
			if variable&1 == 0 {
				selector = selector != bit
			} else {
				selector = selector != (bit && unique0)
			}
		}
		if selector {
			firstShared := false
			if shared != 0 {
				firstShared = assignment&1 != 0
			}
			return free != firstShared
		}
		lastShared := false
		if shared != 0 {
			lastShared = assignment&(1<<uint(shared-1)) != 0
		}
		return lastShared
	})
	return table, set
}

func TestDauSetDecompositionPolarityVariants(t *testing.T) {
	for shared := 0; shared <= 3; shared++ {
		table, set := dauPolarityTruth(t, shared)
		result, err := DecomposeDauVariableSet(table, set, DauSetDecompositionOptions{})
		if err != nil || !result.Found {
			t.Fatal(shared, result, err)
		}
		variants, err := result.PolarityVariants(DauSetPolarityOptions{})
		want := 1 << uint((1<<uint(shared))-1)
		if err != nil || len(variants) != want {
			t.Fatal(shared, len(variants), want, err)
		}
		for index, variant := range variants {
			reconstructed, err := variant.Recompose(0, nil)
			if err != nil || !reconstructed.Equal(table) {
				t.Fatal(shared, index, err)
			}
		}
		if len(variants) > 1 {
			before := variants[1].Composition.Words()
			current, _ := variants[0].Composition.Get(0)
			_ = variants[0].Composition.Set(0, !current)
			if !equalUint64Words(before, variants[1].Composition.Words()) {
				t.Fatal("variant storage aliases", shared)
			}
		}
	}
}

func TestDauSetDecompositionPolarityLimits(t *testing.T) {
	table, set := dauPolarityTruth(t, 3)
	result, err := DecomposeDauVariableSet(table, set, DauSetDecompositionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if variants, err := result.PolarityVariants(DauSetPolarityOptions{MaxVariants: 127}); variants != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(variants, err)
	}
	if variants, err := result.PolarityVariants(DauSetPolarityOptions{Stop: func() bool { return true }}); variants != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(variants, err)
	}
	result.SharedVariables = []int{0, 1, 2, 3}
	if variants, err := result.PolarityVariants(DauSetPolarityOptions{}); variants != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(variants, err)
	}
}

func FuzzDauSetDecompositionPolarityVariants(f *testing.F) {
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
		selected := sets[0]
		for _, set := range sets {
			stats, err := set.Stats(variables)
			if err != nil {
				t.Fatal(err)
			}
			if stats.Shared != 0 {
				selected = set
				break
			}
		}
		result, err := DecomposeDauVariableSet(table, selected, DauSetDecompositionOptions{MaxChecks: 1 << 20})
		if err != nil || !result.Found {
			t.Fatal(result, err)
		}
		variants, err := result.PolarityVariants(DauSetPolarityOptions{MaxVariants: 128})
		if err != nil {
			t.Fatal(err)
		}
		for _, variant := range variants {
			reconstructed, err := variant.Recompose(0, nil)
			if err != nil || !reconstructed.Equal(table) {
				t.Fatal("variant reconstruction", selected, err)
			}
		}
	})
}
