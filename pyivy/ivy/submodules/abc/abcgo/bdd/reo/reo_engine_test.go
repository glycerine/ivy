package reo

import (
	"bytes"
	"math"
	"math/rand"
	"strings"
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func mux(dd *DdManager) *DdNode { return cudd.Cudd_bddIte(dd, dd.Vars[2], dd.Vars[1], dd.Vars[0]) }
func checkMapping(t *testing.T, dd *DdManager, old, new *DdNode, inverse []int) {
	t.Helper()
	if cudd.Cudd_bddPermute(dd, new, inverse) != old {
		t.Fatal("inverse permutation changed function")
	}
	for bits := 0; bits < 1<<len(dd.Vars); bits++ {
		input := make([]bool, len(dd.Vars))
		mapped := make([]bool, len(input))
		for i := range input {
			input[i] = bits&(1<<i) != 0
		}
		for i := range mapped {
			mapped[i] = input[inverse[i]]
		}
		a, err := dd.Evaluate(old, input)
		if err != nil {
			t.Fatal(err)
		}
		b, err := dd.Evaluate(new, mapped)
		if err != nil {
			t.Fatal(err)
		}
		if a != b {
			t.Fatalf("assignment %b differs after reorder", bits)
		}
	}
}
func TestSiftingImprovesMuxAndPreservesInputs(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	root := mux(dd)
	p := New(3, 1)
	p.SetVerification(true)
	p.SetIterations(3)
	var output bytes.Buffer
	p.SetOutput(&output)
	roots, inverse, err := p.Reorder(dd, []*DdNode{root})
	if err != nil {
		t.Fatal(err)
	}
	checkMapping(t, dd, root, roots[0], inverse)
	stats := p.Statistics()
	if stats.After.Nodes >= stats.Before.Nodes {
		t.Fatalf("mux not improved: %+v", stats)
	}
	if stats.Swaps == 0 || stats.Iterations == 0 || !strings.Contains(output.String(), "INITIAL:") {
		t.Fatal(stats, output.String())
	}
	if dd.Perm[0] != 0 || dd.InvPerm[0] != 0 {
		t.Fatal("manager ordering mutated")
	}
	stats.After.LevelNodes[0] = 999
	if p.Statistics().After.LevelNodes[0] == 999 {
		t.Fatal("statistics exposes mutable slices")
	}
}
func TestSharedComplementedRootsAndAPL(t *testing.T) {
	dd := cudd.NewBDDManager(4)
	x := cudd.Cudd_bddXor(dd, dd.Vars[0], dd.Vars[1])
	f := cudd.Cudd_bddIte(dd, dd.Vars[3], x, dd.Vars[2])
	original := []*DdNode{f, cudd.Cudd_Not(f), x, x, dd.One, dd.LogicZero}
	for _, criterion := range []Minimize{REO_MINIMIZE_NODES, REO_MINIMIZE_APL, REO_MINIMIZE_WIDTH} {
		p := New(0, 0)
		p.SetMinimization(criterion)
		p.SetIterations(2)
		roots, inverse, err := p.Reorder(dd, original)
		if err != nil {
			t.Fatal(err)
		}
		for i := range roots {
			checkMapping(t, dd, original[i], roots[i], inverse)
		}
		if p.cost(p.stats.After) > p.cost(p.stats.Before)+REO_COST_EPSILON {
			t.Fatal("cost grew", p.stats)
		}
	}
	profile := profileRoots([]*DdNode{dd.Vars[0], dd.Vars[0]}, []int{0})
	if profile.Nodes != 2 || profile.APL != 2 || profile.Width != 2 {
		t.Fatal(profile)
	}
}
func TestRemappingSparseSupportAndShuffle(t *testing.T) {
	dd := cudd.NewBDDManager(5)
	root := cudd.Cudd_bddAnd(dd, dd.Vars[2], dd.Vars[4])
	p := New(5, 0)
	p.SetRemapping(true)
	roots, inverse, err := p.Reorder(dd, []*DdNode{root})
	if err != nil {
		t.Fatal(err)
	}
	if support := dd.Support(roots[0]); len(support) != 2 || support[0] != 0 || support[1] != 1 {
		t.Fatal(support)
	}
	checkMapping(t, dd, root, roots[0], inverse)
	result, inv := p.Shuffle(dd, root, []int{1, 0})
	if result == nil {
		t.Fatal("shuffle failed")
	}
	checkMapping(t, dd, root, result, inv)
	if result, _ := p.Shuffle(dd, root, []int{0, 0}); result != nil {
		t.Fatal("accepted duplicate permutation")
	}
	if result, _ := p.Shuffle(dd, root, []int{1}); result != nil {
		t.Fatal("accepted short permutation")
	}
}
func TestRandomSmallDAGsExhaustive(t *testing.T) {
	rng := rand.New(rand.NewSource(1234))
	for run := 0; run < 35; run++ {
		dd := cudd.NewBDDManager(5)
		pool := append([]*DdNode{dd.One, dd.LogicZero}, dd.Vars...)
		for i := 0; i < 25; i++ {
			a, b, c := pool[rng.Intn(len(pool))], pool[rng.Intn(len(pool))], pool[rng.Intn(len(pool))]
			pool = append(pool, cudd.Cudd_bddIte(dd, a, b, c))
		}
		old := pool[len(pool)-3:]
		p := New(0, 0)
		p.SetVerification(true)
		p.SetIterations(2)
		roots, inverse, err := p.Reorder(dd, old)
		if err != nil {
			t.Fatal(err)
		}
		for i := range old {
			checkMapping(t, dd, old[i], roots[i], inverse)
		}
	}
}
func TestConstantsEmptyAndInvalidRoots(t *testing.T) {
	dd := cudd.NewBDDManager(0)
	p := New(0, 0)
	roots, inverse, err := p.Reorder(dd, []*DdNode{dd.One, dd.LogicZero})
	if err != nil || len(inverse) != 0 || roots[0] != dd.One || roots[1] != dd.LogicZero {
		t.Fatal(err, roots, inverse)
	}
	roots, _, err = p.Reorder(dd, nil)
	if err != nil || len(roots) != 0 {
		t.Fatal(err, roots)
	}
	if _, _, err = p.Reorder(nil, nil); err == nil {
		t.Fatal("accepted nil manager")
	}
	if _, _, err = p.Reorder(dd, []*DdNode{nil}); err == nil {
		t.Fatal("accepted nil root")
	}
	foreign := cudd.NewBDDManager(0)
	if _, _, err = p.Reorder(dd, []*DdNode{foreign.One}); err == nil {
		t.Fatal("accepted foreign root")
	}
}
func TestTransferRoundTripAndCompatibilityRefs(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	root := cudd.Cudd_Not(mux(dd))
	p := New(3, 0)
	p.SetIterations(0)
	if _, _, err := p.Reorder(dd, []*DdNode{root}); err != nil {
		t.Fatal(err)
	}
	unit := reoTransferNodesToUnits_rec(p, root)
	if result := reoTransferUnitsToNodes_rec(p, unit); result != root {
		t.Fatal("unit transfer changed complemented root")
	}
	if unitNot(unitNot(unit)) != unit {
		t.Fatal("complement is not involutive")
	}
	before := cudd.Cudd_Regular(root).Ref
	roots := make([]*DdNode, 1)
	Extra_ReorderArray(p, dd, []*DdNode{root}, roots, 1, nil)
	if roots[0] != root || cudd.Cudd_Regular(root).Ref != before+1 {
		t.Fatal("array reference ownership")
	}
	cudd.Cudd_Deref(roots[0])
	if Extra_Reorder(p, dd, root, nil) != root || cudd.Cudd_Regular(root).Ref != before {
		t.Fatal("single-root reference ownership")
	}
	if math.Abs(p.Statistics().After.APL-profileRoots([]*DdNode{root}, p.slots).APL) > 1e-9 {
		t.Fatal("profile mismatch")
	}
	p.Close()
	if p.dd != nil || p.pMemChunks != nil || p.units != nil {
		t.Fatal("close retained graph")
	}
}

func TestReusableManagerResetsTransferCaches(t *testing.T) {
	dd := cudd.NewBDDManager(5)
	root := cudd.Cudd_bddAnd(dd, dd.Vars[2], dd.Vars[4])
	p := New(0, 0)
	p.SetIterations(0)
	if _, _, err := p.Reorder(dd, []*DdNode{root}); err != nil {
		t.Fatal(err)
	}
	before := reoTransferNodesToUnits_rec(p, root)
	p.SetRemapping(true)
	reordered, _, err := p.Reorder(dd, []*DdNode{root})
	if err != nil {
		t.Fatal(err)
	}
	after := reoTransferNodesToUnits_rec(p, reordered[0])
	if unitRegular(after) == unitRegular(before) {
		t.Fatal("stale transfer unit survived new run")
	}
	if back := reoTransferUnitsToNodes_rec(p, after); back != reordered[0] {
		t.Fatal("remapped transfer cache changed function")
	}
}

func TestBDDWidthFallbackMatchesSource(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	p := New(0, 0)
	p.SetMinimization(REO_MINIMIZE_WIDTH)
	_, _, err := p.Reorder(dd, []*DdNode{mux(dd)})
	if err != nil || p.criterion != REO_MINIMIZE_NODES {
		t.Fatal("BDD width must fall back to node minimization", err)
	}
}

func TestBDDBenchmarkReturnsNodeCount(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	if nodes := Extra_bddReorderTest(dd, mux(dd)); nodes != 4 {
		t.Fatalf("benchmark returned %d, want 4 nodes", nodes)
	}
}

func TestClassicalCuddComparisonPreservesOriginalManager(t *testing.T) {
	dd := cudd.NewBDDManager(4)
	root := cudd.Cudd_Not(mux(dd))
	cudd.Cudd_Ref(root)
	// Exercise the initial-level transfer map with a nonidentity source order.
	if cudd.Cudd_ShuffleHeap(dd, []int{1, 0, 3, 2}) != 1 {
		t.Fatal("initial shuffle failed")
	}
	order := append([]int(nil), dd.InvPerm...)
	inverse := make([]int, len(dd.Vars))
	result := Extra_ReorderCudd(dd, root, inverse)
	if result == nil {
		t.Fatal("classical symmetric-sifting comparison failed")
	}
	checkMapping(t, dd, root, result, inverse)
	for i, v := range order {
		if dd.InvPerm[i] != v {
			t.Fatal("comparison mutated input manager")
		}
	}
	if cudd.Cudd_Regular(root).Ref != 1 {
		t.Fatal("comparison leaked input references")
	}
}
