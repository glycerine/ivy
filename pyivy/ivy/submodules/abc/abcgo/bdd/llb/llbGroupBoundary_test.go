package llb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGroupBDDUsesInternalBoundaryVariable(t *testing.T) {
	g := aig.NewGraph()
	x, y, z := g.NewInput(), g.NewInput(), g.NewInput()
	boundary := g.AND(x, y)
	gate := g.AND(boundary, z)
	dd := core.NewBDDManager(3)
	defer core.Cudd_Quit(dd)
	order := vec.Vec_IntStartFull(int(aig.Aig_ManObjNumMax(g)))
	vec.Vec_IntWriteEntry(order, boundary.NodeID(), 0)
	vec.Vec_IntWriteEntry(order, z.NodeID(), 1)
	vec.Vec_IntWriteEntry(order, gate.NodeID(), 2)
	inputs, nodes, outputs := vec.Vec_PtrAlloc(2), vec.Vec_PtrAlloc(1), vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(inputs, boundary)
	vec.Vec_PtrPush(inputs, z)
	vec.Vec_PtrPush(nodes, gate)
	vec.Vec_PtrPush(outputs, gate)
	manager := &Llb_Man_t_{pAig: g, dd: dd, vObj2Var: order}
	group := &Llb_Grp_t_{vIns: inputs, vNodes: nodes, vOuts: outputs}
	want := core.Cudd_bddXnor(dd, core.Cudd_bddIthVar(dd, 2), core.Cudd_bddAnd(dd, core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1)))
	if got := Llb_ManConstructGroupBdd(manager, group); got != want {
		t.Fatal("group function must use its internal boundary as a BDD variable")
	}
}
