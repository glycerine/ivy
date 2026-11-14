package cudd

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestUpdateInteractionMatrixCopiesAdjacentVariableRelations(t *testing.T) {
	manager := core.NewBDDManager(4)
	cuddSetInteract(manager, 0, 2)
	cuddSetInteract(manager, 2, 3)
	cuddSetInteract(manager, 0, 3)

	cuddUpdateInteractionMatrix(manager, 1, 2)

	for _, pair := range [][2]int32{{0, 1}, {1, 3}} {
		if cuddTestInteract(manager, pair[0], pair[1]) == 0 {
			t.Errorf("expected interaction between variables %d and %d", pair[0], pair[1])
		}
	}
	if cuddTestInteract(manager, 0, 2) == 0 || cuddTestInteract(manager, 2, 3) == 0 {
		t.Fatal("updating the matrix must preserve existing interactions")
	}
	if cuddTestInteract(manager, 0, 3) == 0 {
		t.Fatal("unrelated preexisting interaction was lost")
	}
}

func TestLinearInPlaceTransformsGraphAndMatrixTogether(t *testing.T) {
	manager := core.NewBDDManager(3)
	if cuddInitLinear(manager) == 0 {
		t.Fatal("linear matrix initialization failed")
	}
	x, y, z := manager.Vars[0], manager.Vars[1], manager.Vars[2]
	f := core.Cudd_bddIte(manager, x, y, z)
	core.Cudd_Ref(f)
	before := make([]bool, 1<<3)
	for mask := range before {
		assignment := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		before[mask], _ = manager.Evaluate(f, assignment)
	}
	cuddSetInteract(manager, 0, 1)
	if got := cuddLinearInPlace(manager, 0, 1); got == 0 {
		t.Fatal("cuddLinearInPlace failed for interacting adjacent variables")
	}
	if Cudd_ReadLinear(manager, 0, 0) != 1 || Cudd_ReadLinear(manager, 0, 1) != 1 || Cudd_ReadLinear(manager, 1, 1) != 1 {
		t.Fatal("linear matrix row did not record x <- x XOR y")
	}
	for mask := range before {
		assignment := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		originalCoordinates := append([]bool(nil), assignment...)
		originalCoordinates[0] = assignment[0] == assignment[1]
		code := 0
		for i, bit := range originalCoordinates {
			if bit {
				code |= 1 << i
			}
		}
		got, err := manager.Evaluate(f, assignment)
		if err != nil || got != before[code] {
			t.Fatalf("assignment %03b: transformed root=%v, want %v (err=%v)", mask, got, before[code], err)
		}
	}
}

func TestLinearAndSiftingMaintainsCanonicalGraphAndMatrix(t *testing.T) {
	manager := core.NewBDDManager(4)
	f := core.Cudd_bddIte(manager, manager.Vars[0], manager.Vars[1], manager.Vars[2])
	core.Cudd_Ref(f)
	cuddSetInteract(manager, 0, 1)
	cuddSetInteract(manager, 0, 2)
	cuddSetInteract(manager, 1, 2)
	if got := cuddLinearAndSifting(manager, 0, 3); got != 1 {
		t.Fatalf("cuddLinearAndSifting returned %d", got)
	}
	if core.Cudd_ReadLinearMatrixSize(manager) != 4 {
		t.Fatal("linear sifting left a stale matrix dimension")
	}
	if err := core.Cudd_CheckConsistency(manager); err != nil {
		t.Fatalf("linear sifting left an invalid canonical graph: %v", err)
	}
}

func TestReduceHeapDispatchesLinearMethods(t *testing.T) {
	manager := core.NewBDDManager(3)
	f := core.Cudd_bddIte(manager, manager.Vars[0], manager.Vars[1], manager.Vars[2])
	core.Cudd_Ref(f)
	cuddSetInteract(manager, 0, 1)
	cuddSetInteract(manager, 0, 2)
	cuddSetInteract(manager, 1, 2)
	if got := Cudd_ReduceHeap(manager, core.CUDD_REORDER_LINEAR, 0); got != 1 {
		t.Fatalf("linear ReduceHeap returned %d", got)
	}
	if got := core.Cudd_ReadLinearMatrixSize(manager); got != 3 {
		t.Fatalf("linear ReduceHeap matrix size = %d, want 3", got)
	}
	if got := Cudd_ReduceHeap(manager, core.CUDD_REORDER_LINEAR_CONVERGE, 0); got != 1 {
		t.Fatalf("converging linear ReduceHeap returned %d", got)
	}
	if got := Cudd_ReduceHeap(manager, core.CUDD_REORDER_SIFT, 0); got != 1 {
		t.Fatalf("nonlinear ReduceHeap delegation returned %d", got)
	}
}

func TestLinearAndOrdinaryReordersShareLifecycleHooks(t *testing.T) {
	methods := []core.Cudd_ReorderingType{core.CUDD_REORDER_SIFT, core.CUDD_REORDER_LINEAR}
	for _, method := range methods {
		manager := core.NewBDDManager(3)
		f := core.Cudd_bddIte(manager, manager.Vars[0], manager.Vars[1], manager.Vars[2])
		core.Cudd_Ref(f)
		cuddSetInteract(manager, 0, 1)
		cuddSetInteract(manager, 0, 2)
		cuddSetInteract(manager, 1, 2)
		var calls []string
		pre := core.DD_HFP(func(_ *core.DdManager, graph string, _ any) int { calls = append(calls, "pre:"+graph); return 1 })
		post := core.DD_HFP(func(_ *core.DdManager, graph string, _ any) int { calls = append(calls, "post:"+graph); return 1 })
		core.Cudd_AddHook(manager, pre, core.CUDD_PRE_REORDERING_HOOK)
		core.Cudd_AddHook(manager, post, core.CUDD_POST_REORDERING_HOOK)
		if Cudd_ReduceHeap(manager, method, 0) != 1 {
			t.Fatalf("reorder method %v failed", method)
		}
		if len(calls) != 2 || calls[0] != "pre:BDD" || calls[1] != "post:BDD" {
			t.Fatalf("reorder method %v hook order = %v", method, calls)
		}
		if core.Cudd_ReadReorderings(manager) != 1 {
			t.Fatalf("reorder method %v count = %d, want 1", method, core.Cudd_ReadReorderings(manager))
		}
	}
}

func TestLinearPreHookCanCancelWithSourceCounterSemantics(t *testing.T) {
	manager := core.NewBDDManager(2)
	var calls []string
	pre := core.DD_HFP(func(_ *core.DdManager, graph string, _ any) int {
		calls = append(calls, "pre:"+graph)
		return 0
	})
	post := core.DD_HFP(func(_ *core.DdManager, graph string, _ any) int {
		calls = append(calls, "post:"+graph)
		return 1
	})
	core.Cudd_AddHook(manager, pre, core.CUDD_PRE_REORDERING_HOOK)
	core.Cudd_AddHook(manager, post, core.CUDD_POST_REORDERING_HOOK)
	if got := Cudd_ReduceHeap(manager, core.CUDD_REORDER_LINEAR, 0); got != 0 {
		t.Fatalf("cancelled linear reorder returned %d, want 0", got)
	}
	if len(calls) != 1 || calls[0] != "pre:BDD" {
		t.Fatalf("cancelled linear reorder hooks = %v", calls)
	}
	if core.Cudd_ReadReorderings(manager) != 1 {
		t.Fatalf("cancelled reorder count = %d, want 1", core.Cudd_ReadReorderings(manager))
	}
}
