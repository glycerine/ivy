package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestPdrStructuralNativeRanks(t *testing.T) {
	g := NewGiaManager()
	inputs := make([]GiaLit, 16)
	for i := range inputs {
		var err error
		inputs[i], err = g.AppendCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	for i, l := range inputs {
		for j := 0; j < i%5; j++ {
			if _, err := g.AppendCO(l); err != nil {
				t.Fatal(err)
			}
		}
	}
	for range inputs {
		if _, err := g.AppendCO(GiaConst0); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.SetRegisterCount(16); err != nil {
		t.Fatal(err)
	}
	ranks, err := g.StructuralFlopPriorities(true, 0)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("test_vectors/oracle/pdr_structural.json")
	if err != nil {
		t.Fatal(err)
	}
	var native []uint64
	if err = json.Unmarshal(raw, &native); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(ranks, native) {
		t.Fatal("native rank/tie mismatch", ranks, native)
	}
	stable, err := g.StructuralFlopPrioritiesStable(true, 0)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile("test_vectors/oracle/pdr_structural_stable.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &native); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(stable, native) {
		t.Fatal("native stable ranks", stable, native)
	}
	if slices.Equal(stable, ranks) {
		t.Fatal("fixture failed to distinguish tie behavior")
	}
	for a := 0; a < 16; a++ {
		for b := a + 1; b < 16; b++ {
			if a%5 == b%5 && stable[a] >= stable[b] {
				t.Fatal("equal-cost registers reordered")
			}
		}
	}

}
func TestPdrStructuralEngineReachability(t *testing.T) {
	for i := 0; i < 256; i++ {
		pdrVerifyTruthOptions(t, bmcTruthCase{bad: uint8(i), next0: uint8(i * 73), next1: uint8(i * 131), constraint: uint8(i * 19), initial: [2]LatchInit{LatchInit(i % 3), LatchInit((i / 3) % 3)}}, PdrOptions{StructuralPriorities: true, OrderAssumptions: true, RecycleVariables: 1, MaxFrames: 16, MaxQueries: 10000})
	}
	m := bmcTruthModel(t, bmcTruthCase{constraint: 255})
	r, err := m.Prove(PdrOptions{StructuralPriorities: true, MaxPriorityObjects: 1})
	if err != nil || r.Status != PdrUnknown || r.Reason != PdrResourceLimit {
		t.Fatal(r, err)
	}
}
func TestCostPermutationTieMerge(t *testing.T) {
	costs := []uint64{0, 0, 0, 0, 0, 0, 0, 0}
	if got := CostPermutation(costs); !slices.Equal(got, []int{0, 4, 1, 5, 2, 6, 3, 7}) {
		t.Fatal(got)
	}
}

func TestPdrStructuralMuxControlBonus(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	mux, err := g.AppendMuxReal(a, b, GiaConst0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(mux); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		if _, err = g.AppendCO(b); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if _, err = g.AppendCO(GiaConst0); err != nil {
			t.Fatal(err)
		}
	}
	if err = g.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	plain, err := g.StructuralFlopPriorities(false, 0)
	if err != nil {
		t.Fatal(err)
	}
	boosted, err := g.StructuralFlopPriorities(true, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(plain, []uint64{0, 1}) || !slices.Equal(boosted, []uint64{1, 0}) {
		t.Fatal("mux bonus", plain, boosted)
	}
}

func TestPdrStructuralCancellation(t *testing.T) {
	m := bmcTruthModel(t, bmcTruthCase{constraint: 255})
	for _, threshold := range []int{2, 3} {
		calls := 0
		r, err := m.Prove(PdrOptions{StructuralPriorities: true, SAT: SatLimits{Stop: func() bool { calls++; return calls >= threshold }}})
		if err != nil || r.Status != PdrUnknown || r.Reason != PdrCancelled {
			t.Fatal(threshold, r, err)
		}
	}
}

func TestPdrStablePrioritiesEmptyLimitsAndMux(t *testing.T) {
	empty := NewGiaManager()
	ranks, err := empty.StructuralFlopPrioritiesStable(true, 0)
	if err != nil || len(ranks) != 0 {
		t.Fatal(ranks, err)
	}
	if _, err := empty.StructuralFlopPrioritiesStable(false, -1); err == nil {
		t.Fatal("negative limit")
	}
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	mux, err := g.AppendMuxReal(a, b, GiaConst0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(mux); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		if _, err = g.AppendCO(b); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if _, err = g.AppendCO(GiaConst0); err != nil {
			t.Fatal(err)
		}
	}
	if err = g.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	plain, err := g.StructuralFlopPrioritiesStable(false, 0)
	if err != nil {
		t.Fatal(err)
	}
	boosted, err := g.StructuralFlopPrioritiesStable(true, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(plain, []uint64{0, 1}) || !slices.Equal(boosted, []uint64{1, 0}) {
		t.Fatal(plain, boosted)
	}
	if _, err := g.StructuralFlopPrioritiesStable(true, g.ObjectCount()-1); err == nil {
		t.Fatal("object cap ignored")
	}
}
