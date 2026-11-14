package core

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
)

func reorderTruth(t *testing.T, dd *DdManager, root *DdNode) []float64 {
	t.Helper()
	result := make([]float64, 1<<len(dd.Vars))
	for bits := range result {
		assignment := make([]bool, len(dd.Vars))
		for i := range assignment {
			assignment[i] = bits&(1<<i) != 0
		}
		if Cudd_IsAdd(root) {
			value, err := dd.EvaluateADD(root, assignment)
			if err != nil {
				t.Fatal(err)
			}
			result[bits] = value
		} else {
			value, err := dd.Evaluate(root, assignment)
			if err != nil {
				t.Fatal(err)
			}
			if value {
				result[bits] = 1
			}
		}
	}
	return result
}
func TestShuffleHeapPreservesLiveHandlesAndCanonicality(t *testing.T) {
	dd := NewBDDManager(4)
	root := Cudd_bddIte(dd, dd.Vars[2], dd.Vars[1], dd.Vars[0])
	inverse := Cudd_Not(root)
	add := Cudd_addIte(dd, Cudd_addIthVar(dd, 3), Cudd_addConst(dd, -7), Cudd_addIte(dd, Cudd_addIthVar(dd, 0), Cudd_addConst(dd, 2.5), Cudd_addConst(dd, 11)))
	Cudd_Ref(root)
	Cudd_Ref(add)
	beforeBDD, beforeADD := reorderTruth(t, dd, root), reorderTruth(t, dd, add)
	oldRoot, oldAdd, oldVars := root, add, append([]*DdNode(nil), dd.Vars...)
	ref, addRef := Cudd_Regular(root).Ref, add.Ref
	order := []int{2, 3, 1, 0}
	if Cudd_ShuffleHeap(dd, order) != 1 {
		t.Fatal("shuffle failed")
	}
	if len(dd.cache) != 0 || len(dd.addCache) != 0 {
		t.Fatal("reordering retained computed tables")
	}
	if !reflect.DeepEqual(dd.InvPerm, order) {
		t.Fatal(dd.InvPerm)
	}
	for level, index := range order {
		if dd.Perm[index] != level {
			t.Fatal(dd.Perm)
		}
	}
	if root != oldRoot || add != oldAdd || Cudd_Not(root) != inverse || !reflect.DeepEqual(dd.Vars, oldVars) {
		t.Fatal("live handle identity changed")
	}
	if Cudd_Regular(root).Ref != ref || add.Ref != addRef {
		t.Fatal("external reference count changed")
	}
	if !reflect.DeepEqual(beforeBDD, reorderTruth(t, dd, root)) || !reflect.DeepEqual(beforeADD, reorderTruth(t, dd, add)) {
		t.Fatal("function changed")
	}
	if Cudd_bddIte(dd, dd.Vars[2], dd.Vars[1], dd.Vars[0]) != root {
		t.Fatal("BDD canonical identity lost")
	}
	if Cudd_addIte(dd, Cudd_addIthVar(dd, 3), Cudd_addConst(dd, -7), Cudd_addIte(dd, Cudd_addIthVar(dd, 0), Cudd_addConst(dd, 2.5), Cudd_addConst(dd, 11))) != add {
		t.Fatal("ADD canonical identity lost")
	}
	if Cudd_ShuffleHeap(dd, []int{1, 1, 2, 3}) != 0 || Cudd_ShuffleHeap(dd, []int{0}) != 0 {
		t.Fatal("invalid permutation accepted")
	}
}

func TestAnnealingRangePreservesReferencedFunctions(t *testing.T) {
	dd := NewBDDManager(4)
	x0, x1, x2, x3 := dd.Vars[0], dd.Vars[1], dd.Vars[2], dd.Vars[3]
	f := Cudd_bddXor(dd, Cudd_bddAnd(dd, x0, x2), Cudd_bddAnd(dd, x1, x3))
	Cudd_Ref(f)
	before := make([]bool, 16)
	for assignment := range before {
		before[assignment] = bddValueAt(f, assignment)
	}
	if Cudd_AnnealRange(dd, 1, 3, 0) != 1 {
		t.Fatal("annealing failed")
	}
	for assignment, want := range before {
		if got := bddValueAt(f, assignment); got != want {
			t.Fatalf("annealing changed function at assignment %04b: got %v want %v", assignment, got, want)
		}
	}
}
func TestReduceHeapSiftImprovesReferencedMux(t *testing.T) {
	dd := NewBDDManager(3)
	root := Cudd_bddIte(dd, dd.Vars[2], dd.Vars[1], dd.Vars[0])
	Cudd_Ref(root)
	before := Cudd_DagSize(root)
	truth := reorderTruth(t, dd, root)
	if Cudd_ReduceHeap(dd, CUDD_REORDER_SIFT, 1) != 1 {
		t.Fatal("sift failed")
	}
	if Cudd_DagSize(root) >= before {
		t.Fatal("mux did not improve", before, Cudd_DagSize(root))
	}
	if !reflect.DeepEqual(truth, reorderTruth(t, dd, root)) {
		t.Fatal("sift changed function")
	}
	if Cudd_ReadReorderings(dd) != 1 {
		t.Fatal("missing reorder counter")
	}
	if Cudd_bddIte(dd, dd.Vars[2], dd.Vars[1], dd.Vars[0]) != root {
		t.Fatal("sift lost canonical root identity")
	}
}
func TestSymmetricSiftPositiveAndNegativeGroups(t *testing.T) {
	for _, negative := range []bool{false, true} {
		dd := NewBDDManager(3)
		y := dd.Vars[1]
		if negative {
			y = Cudd_Not(y)
		}
		root := Cudd_bddAnd(dd, dd.Vars[0], y)
		Cudd_Ref(root)
		truth := reorderTruth(t, dd, root)
		result := Cudd_ReduceHeap(dd, CUDD_REORDER_SYMM_SIFT, 1)
		if result != 1 || len(dd.SymmetryGroups()) != 1 || len(dd.SymmetryGroups()[0]) != 2 {
			t.Fatalf("negative=%v: got %d, want public success and a two-variable symmetry group", negative, result)
		}
		if !reflect.DeepEqual(truth, reorderTruth(t, dd, root)) {
			t.Fatal("symmetric sift changed function")
		}
	}
}
func TestSymmetryRejectsIndependentLowerReference(t *testing.T) {
	dd := NewBDDManager(2)
	root := Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	Cudd_Ref(root)
	Cudd_Ref(dd.Vars[1])
	if result := Cudd_ReduceHeap(dd, CUDD_REORDER_SYMM_SIFT, 1); result != 1 {
		t.Fatal("grouped independently referenced variable", result)
	}
}
func TestReduceHeapRandomReferencedFamilies(t *testing.T) {
	rng := rand.New(rand.NewSource(9721))
	for trial := 0; trial < 20; trial++ {
		dd := NewBDDManager(5)
		nodes := append([]*DdNode{dd.One, dd.LogicZero}, dd.Vars...)
		for i := 0; i < 20; i++ {
			nodes = append(nodes, Cudd_bddIte(dd, nodes[rng.Intn(len(nodes))], nodes[rng.Intn(len(nodes))], nodes[rng.Intn(len(nodes))]))
		}
		roots := nodes[len(nodes)-3:]
		truth := make([][]float64, len(roots))
		for i, root := range roots {
			Cudd_Ref(root)
			truth[i] = reorderTruth(t, dd, root)
		}
		method := CUDD_REORDER_SYMM_SIFT_CONV
		if trial%2 == 0 {
			method = CUDD_REORDER_SIFT_CONVERGE
		}
		if Cudd_ReduceHeap(dd, method, 1) == 0 {
			t.Fatal("reorder failed")
		}
		for i, root := range roots {
			if !reflect.DeepEqual(truth[i], reorderTruth(t, dd, root)) {
				t.Fatal("random truth changed", trial, i)
			}
		}
	}
}
func TestShuffleCollectsDeadNodesAndHonorsMinsize(t *testing.T) {
	dd := NewBDDManager(3)
	dead := Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	if Cudd_ReduceHeap(dd, CUDD_REORDER_SIFT, 1000000) != 1 || Cudd_ReadReorderings(dd) != 0 {
		t.Fatal("minsize threshold ignored")
	}
	if Cudd_ShuffleHeap(dd, []int{2, 1, 0}) != 1 {
		t.Fatal("shuffle failed")
	}
	if dd.valid(dead) {
		t.Fatal("unreferenced unreachable handle was not collected")
	}
	for i, v := range dd.Vars {
		if !dd.valid(v) || int(v.Index) != i {
			t.Fatal("projection pin collected")
		}
	}
	if Cudd_ReduceHeap(dd, CUDD_REORDER_RANDOM, 1) != 0 {
		t.Fatal("unsupported heuristic silently substituted")
	}
}

// The fixture is emitted by test_vectors/reorder_oracle.c linked against the actual
// original CUDD object files. No C compiler/runtime is needed for this Go test.
func TestDynamicReorderingOriginalCUDDOracle(t *testing.T) {
	data, err := os.ReadFile("test_vectors/reorder_oracle.txt")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 512 {
		t.Fatal("incomplete oracle fixture", len(lines))
	}
	for _, line := range lines {
		var method, mask, before, after, status, truth, symmask, symgroups int
		var order string
		if _, err := fmt.Sscan(line, &method, &mask, &before, &after, &status, &truth, &order, &symmask, &symgroups); err != nil {
			t.Fatal(err)
		}
		dd := NewBDDManager(3)
		var build func(int, int) *DdNode
		build = func(level, assignment int) *DdNode {
			if level == 3 {
				if mask&(1<<assignment) != 0 {
					return dd.One
				}
				return dd.LogicZero
			}
			return Cudd_bddIte(dd, dd.Vars[level], build(level+1, assignment|(1<<level)), build(level+1, assignment))
		}
		root := build(0, 0)
		Cudd_Ref(root)
		if Cudd_DagSize(root) != before {
			t.Fatal("initial DAG mismatch", line)
		}
		heuristic := CUDD_REORDER_SIFT
		if method == 1 {
			heuristic = CUDD_REORDER_SYMM_SIFT
		}
		gotStatus := Cudd_ReduceHeap(dd, heuristic, 1)
		if Cudd_DagSize(root) != after || gotStatus != status {
			t.Errorf("method=%d mask=%d: size/status %d/%d, CUDD %d/%d", method, mask, Cudd_DagSize(root), gotStatus, after, status)
		}
		gotMask := 0
		groups := dd.SymmetryGroups()
		for _, g := range groups {
			for _, index := range g {
				gotMask |= 1 << index
			}
		}
		if gotMask != symmask || len(groups) != symgroups {
			t.Errorf("method=%d mask=%d: symmetry mask/groups %d/%d, CUDD %d/%d", method, mask, gotMask, len(groups), symmask, symgroups)
		}
		values := reorderTruth(t, dd, root)
		gotTruth := 0
		for a, v := range values {
			if v != 0 {
				gotTruth |= 1 << a
			}
		}
		if gotTruth != truth {
			t.Fatal("truth changed versus original CUDD", line, gotTruth)
		}
	}
}

func TestBDDHeapReorderLeavesZDDStateUntouched(t *testing.T) {
	dd := Cudd_Init(3, 3, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0)
	z := Cudd_zddUnion(dd, Cudd_zddChange(dd, dd.One, 0), Cudd_zddChange(dd, dd.One, 2))
	Cudd_Ref(z)
	original, one, zero := z, dd.One, dd.Zero
	count, keys := Cudd_zddCount(dd, z), len(dd.zddUnique)
	perm, inv := append([]int(nil), dd.PermZ...), append([]int(nil), dd.InvPermZ...)
	univ := append([]*DdNode(nil), dd.Univ...)
	root := Cudd_bddIte(dd, dd.Vars[2], dd.Vars[1], dd.Vars[0])
	Cudd_Ref(root)
	for _, method := range []Cudd_ReorderingType{CUDD_REORDER_SIFT, CUDD_REORDER_SYMM_SIFT} {
		if Cudd_ReduceHeap(dd, method, 1) == 0 {
			t.Fatal("reorder failed")
		}
		if z != original || Cudd_zddCount(dd, z) != count || len(dd.zddUnique) != keys || dd.One != one || dd.Zero != zero {
			t.Fatal("BDD reordering changed ZDD state")
		}
		if !reflect.DeepEqual(dd.PermZ, perm) || !reflect.DeepEqual(dd.InvPermZ, inv) || !reflect.DeepEqual(dd.Univ, univ) {
			t.Fatal("ZDD order/universe changed")
		}
		if Cudd_zddUnion(dd, Cudd_zddChange(dd, dd.One, 0), Cudd_zddChange(dd, dd.One, 2)) != z {
			t.Fatal("ZDD unique identity changed")
		}
	}
	if Cudd_ShuffleHeap(dd, []int{2, 0, 1}) != 1 || Cudd_zddCount(dd, z) != count || z.Ref != 1 {
		t.Fatal("shuffle changed referenced ZDD")
	}
}
func TestSiftLimitsSymmetryMetadataAndMinSize(t *testing.T) {
	dd := NewBDDManager(3)
	root := Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	Cudd_Ref(root)
	if Cudd_ReduceHeap(dd, CUDD_REORDER_SYMM_SIFT, 1) != 1 {
		t.Fatal("symmetry sift failed")
	}
	groups := dd.SymmetryGroups()
	if len(groups) == 0 {
		t.Fatal("symmetry not discovered")
	}
	groups[0][0] = 99
	if dd.SymmetryGroups()[0][0] == 99 {
		t.Fatal("mutable metadata exposed")
	}
	Cudd_SetSiftMaxVar(dd, 0)
	before := append([]int(nil), dd.InvPerm...)
	if Cudd_ReduceHeap(dd, CUDD_REORDER_SIFT, 1) != 1 || !reflect.DeepEqual(dd.InvPerm, before) {
		t.Fatal("zero variable limit ignored")
	}
	if len(dd.SymmetryGroups()) != 0 {
		t.Fatal("stale symmetry metadata after plain sift")
	}
	Cudd_SetSiftMaxVar(dd, 1000)
	Cudd_SetSiftMaxSwap(dd, 0)
	if Cudd_ReduceHeap(dd, CUDD_REORDER_SYMM_SIFT_CONV, 1) != 1 || !reflect.DeepEqual(dd.InvPerm, before) {
		t.Fatal("zero swap limit ignored")
	}
	Cudd_SetMaxGrowth(dd, 1.5)
	Cudd_SetMaxGrowth(dd, 0)
	if Cudd_ReadMaxGrowth(dd) != 1.5 || Cudd_ReadSiftMaxVar(dd) != 1000 || Cudd_ReadSiftMaxSwap(dd) != 0 {
		t.Fatal("configuration accessors")
	}
	if Cudd_ReduceHeap(dd, CUDD_REORDER_NONE, 1) != 1 {
		t.Fatal("NONE should succeed")
	}
}

func TestReferencedConstantOnlyHeap(t *testing.T) {
	dd := NewBDDManager(0)
	constant := Cudd_addConst(dd, -7)
	Cudd_Ref(constant)
	if Cudd_ReduceHeap(dd, CUDD_REORDER_SYMM_SIFT, 1) != 1 || Cudd_ReadReorderings(dd) != 1 {
		t.Fatal("referenced terminal omitted from live heap")
	}
	if !dd.validAdd(constant) || constant.Ref != 1 || constant != Cudd_addConst(dd, -7) {
		t.Fatal("constant handle/reference changed")
	}
}

func TestDynamicReorderConstantPinsAndCollection(t *testing.T) {
	dd := NewBDDManager(2)
	dead := Cudd_addConst(dd, 123.5)
	alive := Cudd_addConst(dd, -7)
	Cudd_Ref(alive)
	globals := []*DdNode{dd.One, dd.Zero, dd.PlusInfinity, dd.MinusInfinity, dd.Background}
	if Cudd_ShuffleHeap(dd, []int{1, 0}) != 1 {
		t.Fatal("shuffle failed")
	}
	if dd.validAdd(dead) {
		t.Fatal("unreferenced unreachable ADD constant retained")
	}
	if !dd.validAdd(alive) || Cudd_addConst(dd, -7) != alive {
		t.Fatal("referenced constant collected")
	}
	for _, constant := range globals {
		if !dd.validAdd(constant) {
			t.Fatal("manager constant pin collected")
		}
	}
	// Constants stay in the rebuilt live graph so their handles survive, but
	// CUDD's sifting key count excludes the constants subtable.
	roots := dd.liveReorderRoots()
	heap := reorderBuild(dd, roots, dd.InvPerm)
	constants := 0
	for node := range heap.nodes {
		if Cudd_IsConstant(node) {
			constants++
		}
	}
	if heap.cost != 0 || constants != 5 {
		t.Fatal("constant pins/key count mismatch", heap.cost, constants)
	}
}

func TestReorderRootsExcludeNodesReferencedOnlyByParents(t *testing.T) {
	dd := NewBDDManager(3)
	child := Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	root := Cudd_bddIte(dd, dd.Vars[2], child, dd.One)
	Cudd_Ref(root)
	if got := Cudd_ExternalRefCount(dd, child); got != 0 {
		t.Fatalf("internal child external refs=%d, want 0", got)
	}
	roots := dd.liveReorderRoots()
	for _, candidate := range roots.active {
		if candidate == Cudd_Regular(child) {
			t.Fatal("parent-referenced child was treated as an external reorder root")
		}
	}
	Cudd_Ref(child)
	if got := Cudd_ExternalRefCount(dd, child); got != 1 {
		t.Fatalf("explicit child external refs=%d, want 1", got)
	}
	roots = dd.liveReorderRoots()
	found := false
	for _, candidate := range roots.active {
		found = found || candidate == Cudd_Regular(child)
	}
	if !found {
		t.Fatal("explicitly referenced child was omitted from reorder roots")
	}
}
