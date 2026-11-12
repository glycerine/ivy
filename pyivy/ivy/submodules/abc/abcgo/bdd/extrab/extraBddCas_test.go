package extrab

import (
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestCascadeRootEncodersAndPathOwnership(t *testing.T) {
	dd := cudd.NewBDDManager(4)
	f := dd.Vars[0]
	encoded := Extra_bddEncodingBinary(dd, []*DdNode{f, cudd.Cudd_Not(f)}, 2, dd.Vars[3:], 1)
	if encoded != cudd.Cudd_bddXor(dd, f, dd.Vars[3]) {
		t.Fatal("binary encoding")
	}
	function := cudd.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[2])
	paths := Extra_bddNodePathsUnderCut(dd, function, 1)
	sum := dd.LogicZero
	for node, path := range paths {
		if cudd.Cudd_Regular(node).Ref == 0 || cudd.Cudd_Regular(path).Ref == 0 {
			t.Fatal("paths not referenced")
		}
		sum = cudd.Cudd_bddOr(dd, sum, cudd.Cudd_bddAnd(dd, node, path))
		cudd.Cudd_RecursiveDeref(dd, node)
		cudd.Cudd_RecursiveDeref(dd, path)
	}
	if sum != function {
		t.Fatal("cut reconstruction")
	}
	outNodes, outPaths := make([]*DdNode, 4), make([]*DdNode, 4)
	n := Extra_bddNodePathsUnderCutArray(dd, []*DdNode{cudd.Cudd_BddToAdd(dd, function)}, []*DdNode{dd.One}, 1, outNodes, outPaths, 1)
	if n != 2 {
		t.Fatal("cut count", n)
	}
	for i := 0; i < n; i++ {
		cudd.Cudd_RecursiveDeref(dd, outNodes[i])
		cudd.Cudd_RecursiveDeref(dd, outPaths[i])
	}
	if Extra_bddNodePathsUnderCutArray(dd, []*DdNode{dd.One}, []*DdNode{dd.LogicZero}, 1, outNodes, outPaths, 1) != 1 || outPaths[0] != dd.LogicZero {
		t.Fatal("constant path special case")
	}
	cudd.Cudd_RecursiveDeref(dd, outNodes[0])
	cudd.Cudd_RecursiveDeref(dd, outPaths[0])
	if Extra_bddNodePathsUnderCutArray(dd, []*DdNode{dd.One}, []*DdNode{dd.One}, 1, nil, nil, 0) != -1 {
		t.Fatal("small output buffers accepted")
	}
	if len(Extra_CollectNodes(function)) != 3 {
		t.Fatal("node collection")
	}
}
func TestCascadeRootProfileAndNonStrict(t *testing.T) {
	dd := cudd.NewBDDManager(4)
	profile := make([]int, 5)
	f := cudd.Cudd_BddToAdd(dd, cudd.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[2]))
	if Extra_ProfileWidth(dd, f, profile, -1) != 2 || Extra_ProfileWidth(dd, f, profile, 1) != 1 {
		t.Fatal("profile", profile)
	}
	columns := make([]*DdNode, 4)
	for a := 0; a < 4; a++ {
		cube := dd.One
		for v := 0; v < 2; v++ {
			cube = cudd.Cudd_bddAnd(dd, cube, cudd.Cudd_NotCond(dd.Vars[v], a&(1<<uint(1-v)) == 0))
		}
		columns[a] = cube
	}
	vars := cudd.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	simple := 0
	r := Extra_bddEncodingNonStrict(dd, columns, 4, vars, dd.Vars[2:], 2, &simple)
	if r == nil || simple != 2 {
		t.Fatal("non-strict encoder", simple)
	}
	if Extra_bddEncodingBinary(dd, columns, 5, nil, 0) != nil || Extra_ProfileWidth(dd, f, nil, -1) != -1 {
		t.Fatal("invalid inputs")
	}
}
