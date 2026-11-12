package llb

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestCutConeStopsAtLowerBoundary(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	gate := g.AND(x, y)
	root := g.AddOutput(gate)
	lower, upper := vec.Vec_PtrAlloc(2), vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(lower, x)
	vec.Vec_PtrPush(lower, y)
	vec.Vec_PtrPush(upper, root)
	if got := Llb_ManCutNodes(g, lower, upper); vec.Vec_PtrSize(got) != 1 {
		t.Fatalf("cut cone has %d internal nodes, want 1", vec.Vec_PtrSize(got))
	}
	if got := Llb_ManCutSupp(g, lower, upper); vec.Vec_PtrSize(got) != 2 {
		t.Fatalf("cut support has %d leaves, want 2", vec.Vec_PtrSize(got))
	}
}

func TestGroupBDDConstrainsOutputToItsFunction(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	gate := g.AND(x, y)
	dd := core.NewBDDManager(3)
	defer core.Cudd_Quit(dd)
	vars := vec.Vec_IntStartFull(int(aig.Aig_ManObjNumMax(g)))
	vec.Vec_IntWriteEntry(vars, int(x.NodeID()), 0)
	vec.Vec_IntWriteEntry(vars, int(y.NodeID()), 1)
	vec.Vec_IntWriteEntry(vars, int(gate.NodeID()), 2)
	groupOutputs := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(groupOutputs, gate)
	group := &Llb_Grp_t_{vOuts: groupOutputs}
	manager := &Llb_Man_t_{pAig: g, dd: dd, vObj2Var: vars}

	want := core.Cudd_bddXnor(dd, core.Cudd_bddIthVar(dd, 2), core.Cudd_bddAnd(dd, core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1)))
	if got := Llb_ManConstructGroupBdd(manager, group); got != want {
		t.Fatal("group BDD must equate its output variable with the AIG function")
	}
}
