package cudd

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestSubsetShortPathsRetainsShortestPathAtThreshold(t *testing.T) {
	dd := core.NewBDDManager(3)
	x0, x1, x2 := Cudd_bddIthVar(dd, 0), Cudd_bddIthVar(dd, 1), Cudd_bddIthVar(dd, 2)
	f := core.Cudd_bddOr(dd, x0, core.Cudd_bddAnd(dd, x1, x2))
	got := Cudd_SubsetShortPaths(dd, f, 3, 1, 0)
	if got == nil {
		t.Fatal("short-path subset returned nil")
	}
	if !core.Cudd_bddLeq(dd, got, f) {
		t.Fatal("result is not a subset of the input")
	}
	if got != x0 {
		t.Fatal("threshold should retain the one-node shortest path")
	}
}

func TestSubsetShortPathsTieUsesThenBranchFirst(t *testing.T) {
	dd := core.NewBDDManager(3)
	x0, x1, x2 := Cudd_bddIthVar(dd, 0), Cudd_bddIthVar(dd, 1), Cudd_bddIthVar(dd, 2)
	f := core.Cudd_bddIte(dd, x0, x1, x2)
	got := Cudd_SubsetShortPaths(dd, f, 3, 1, 0)
	want := core.Cudd_bddAnd(dd, x0, x1)
	if got != want {
		t.Fatal("equal-length paths should select the then branch first")
	}
}

func TestSubsetShortPathsHardLimitAndSupersetContracts(t *testing.T) {
	dd := core.NewBDDManager(3)
	x0, x1, x2 := Cudd_bddIthVar(dd, 0), Cudd_bddIthVar(dd, 1), Cudd_bddIthVar(dd, 2)
	f := core.Cudd_bddOr(dd, x0, core.Cudd_bddAnd(dd, x1, x2))
	subset := Cudd_SubsetShortPaths(dd, f, 3, 1, 1)
	if subset == nil || subset != x0 || countInternalNodes(subset) != 1 || !core.Cudd_bddLeq(dd, subset, f) {
		t.Fatal("hard-limit subset violated its size or containment contract")
	}
	base := core.Cudd_bddAnd(dd, x0, x1)
	superset := Cudd_SupersetShortPaths(dd, base, 3, 1, 0)
	if superset == nil || !core.Cudd_bddLeq(dd, base, superset) {
		t.Fatal("short-path superset does not contain its input")
	}
}

func countInternalNodes(root *DdNode) int {
	seen := map[*DdNode]bool{}
	var visit func(*DdNode)
	visit = func(node *DdNode) {
		reg := core.Cudd_Regular(node)
		if seen[reg] || core.Cudd_IsConstant(reg) {
			return
		}
		seen[reg] = true
		visit(core.Cudd_T(reg))
		visit(core.Cudd_E(reg))
	}
	visit(root)
	return len(seen)
}
