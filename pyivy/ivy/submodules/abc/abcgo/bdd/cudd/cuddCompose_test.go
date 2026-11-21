package cudd

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestAdjacentPermutationSkipsIncompleteTriples(t *testing.T) {
	dd := core.NewBDDManager(5)
	defer Cudd_Quit(dd)
	x := dd.Vars
	if got := Cudd_bddAdjPermuteX(dd, x[0], x[:2]); got != x[0] {
		t.Error("two variables formed an incomplete group and were swapped")
	}
	if got := Cudd_bddAdjPermuteX(dd, x[3], x); got != x[3] {
		t.Error("tail of five-variable input was swapped")
	}
	if got := Cudd_bddAdjPermuteX(dd, x[0], x); got != x[1] {
		t.Error("first complete group did not swap")
	}
}

func TestGeneralADDCompositionWeightsBothBranches(t *testing.T) {
	dd := core.NewBDDManager(2)
	defer Cudd_Quit(dd)
	x, y := core.Cudd_addIthVar(dd, 0), core.Cudd_addIthVar(dd, 1)
	f := core.Cudd_addIte(dd, x, core.Cudd_addConst(dd, 5), core.Cudd_addConst(dd, 7))
	// Weight each branch independently. Both can contribute on the same row.
	on := []*DdNode{core.Cudd_addIte(dd, y, core.Cudd_addConst(dd, 3), dd.Zero), y}
	off := []*DdNode{core.Cudd_addIte(dd, y, core.Cudd_addConst(dd, -2), dd.One), core.Cudd_addCmpl(dd, y)}
	result := Cudd_addGeneralVectorCompose(dd, f, on, off)
	for mask := range 4 {
		bits := []bool{mask&1 != 0, mask&2 != 0}
		want := 7.0
		if bits[1] {
			want = 1
		}
		got, err := dd.EvaluateADD(result, bits)
		if err != nil || got != want {
			t.Errorf("input %02b: %g, %v; want %g", mask, got, err, want)
		}
	}
}

func TestNonSimADDCompositionFollowsDeepestLevelFirst(t *testing.T) {
	dd := core.NewBDDManager(3)
	defer Cudd_Quit(dd)
	x, y, z := core.Cudd_addIthVar(dd, 0), core.Cudd_addIthVar(dd, 1), core.Cudd_addIthVar(dd, 2)
	for _, n := range []*DdNode{x, y, z} {
		core.Cudd_Ref(n)
	}
	vector := []*DdNode{y, x, z}
	if got := Cudd_addNonSimCompose(dd, x, vector); got != y {
		t.Error("under identity order x should become y")
	}
	if core.Cudd_ShuffleHeap(dd, []int{2, 1, 0}) == 0 {
		t.Fatal("reorder")
	}
	if got := Cudd_addNonSimCompose(dd, y, vector); got != x {
		t.Error("with reversed levels y should become x")
	}
}

func TestADDCompositionRejectsForeignInputsBeforeTerminalShortcut(t *testing.T) {
	dd, other := core.NewBDDManager(2), core.NewBDDManager(2)
	defer Cudd_Quit(dd)
	defer Cudd_Quit(other)
	x, y := core.Cudd_addIthVar(dd, 0), core.Cudd_addIthVar(dd, 1)
	for name, result := range map[string]*DdNode{
		"single":       Cudd_addCompose(dd, other.Zero, x, 0),
		"simultaneous": Cudd_addVectorCompose(dd, other.One, []*DdNode{x, y}),
		"sequential":   Cudd_addNonSimCompose(dd, dd.One, []*DdNode{x, other.Zero}),
		"weighted":     Cudd_addGeneralVectorCompose(dd, dd.One, []*DdNode{x, other.Zero}, []*DdNode{y, x}),
	} {
		if result != nil {
			t.Errorf("%s composition accepted a foreign function or substitution", name)
		}
	}
}

func TestVariableSwapAssignsEachListedPair(t *testing.T) {
	dd := core.NewBDDManager(3)
	defer Cudd_Quit(dd)
	a, b := dd.Vars[0], dd.Vars[1]
	if got := Cudd_bddSwapVariables(dd, a, []*DdNode{a, a}, []*DdNode{b, b}); got != b {
		t.Error("listing the same swap twice undid the mapping")
	}
}

func TestRecursiveNonSimCompositionHonorsRemainingCube(t *testing.T) {
	dd := core.NewBDDManager(2)
	defer Cudd_Quit(dd)
	x, y := core.Cudd_addIthVar(dd, 0), core.Cudd_addIthVar(dd, 1)
	vector := []*DdNode{y, x}
	if cuddAddNonSimComposeRecur(dd, x, vector, dd.One, dd.One, 2) != x {
		t.Fatal("empty substitution cube should return the original function")
	}
	if cuddAddNonSimComposeRecur(dd, x, vector, dd.One, x, 2) != y {
		t.Fatal("substitution was not restricted to the cube")
	}
}

func TestWeightedADDCompositionNestedAssignments(t *testing.T) {
	dd := core.NewBDDManager(3)
	defer Cudd_Quit(dd)
	x, y, z := core.Cudd_addIthVar(dd, 0), core.Cudd_addIthVar(dd, 1), core.Cudd_addIthVar(dd, 2)
	num := func(v float64) *DdNode { return core.Cudd_addConst(dd, v) }
	selectY := func(a, b float64) *DdNode { return core.Cudd_addIte(dd, y, num(a), num(b)) }
	f := core.Cudd_addIte(dd, x, selectY(5, 7), selectY(-3, 2))
	on := []*DdNode{selectY(2, -1), core.Cudd_addIte(dd, z, num(4), num(0)), z}
	off := []*DdNode{core.Cudd_addIte(dd, z, num(-3), num(0)), core.Cudd_addIte(dd, x, num(2), num(-1)), core.Cudd_addCmpl(dd, z)}
	got := Cudd_addGeneralVectorCompose(dd, f, on, off)
	for mask := range 8 {
		bits := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		var wy, ny, wx, nx float64
		if bits[1] {
			wx = 2
		} else {
			wx = -1
		}
		if bits[2] {
			wy, nx = 4, -3
		}
		if bits[0] {
			ny = 2
		} else {
			ny = -1
		}
		want := wx*(wy*5+ny*7) + nx*(wy*(-3)+ny*2)
		v, err := dd.EvaluateADD(got, bits)
		if err != nil || v != want {
			t.Fatalf("input %03b: got %g (%v), want %g", mask, v, err, want)
		}
	}
}

func BenchmarkWeightedADDComposition(b *testing.B) {
	dd := core.NewBDDManager(5)
	defer Cudd_Quit(dd)
	on, off := make([]*DdNode, 5), make([]*DdNode, 5)
	for i := range on {
		on[i] = core.Cudd_addIthVar(dd, i)
		off[i] = core.Cudd_addCmpl(dd, on[i])
		core.Cudd_Ref(on[i])
		core.Cudd_Ref(off[i])
	}
	if core.Cudd_ShuffleHeap(dd, []int{4, 3, 2, 1, 0}) == 0 {
		b.Fatal("reorder")
	}
	x, y := on[0], on[1]
	f := core.Cudd_addIte(dd, x, core.Cudd_addConst(dd, 5), core.Cudd_addConst(dd, 7))
	core.Cudd_Ref(f)
	on[0] = core.Cudd_addIte(dd, y, core.Cudd_addConst(dd, 3), dd.Zero)
	off[0] = core.Cudd_addIte(dd, y, core.Cudd_addConst(dd, -2), dd.One)
	want := core.Cudd_addIte(dd, y, dd.One, core.Cudd_addConst(dd, 7))
	b.ReportAllocs()
	for b.Loop() {
		if Cudd_addGeneralVectorCompose(dd, f, on, off) != want {
			b.Fatal("incorrect weighted composition")
		}
	}
}
