package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	ifpkg "github.com/glycerine/abcgo/map/if"
)

func TestIfifDelaySortingAndClusterCost(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	inputs := make([]*abc.Abc_Obj_t, 4)
	for i := range inputs {
		inputs[i] = abc.Abc_NtkCreatePi(n)
	}
	manager := n.FunctionManager.(*abc.Abc_Aig_t)
	left := abc.Abc_AigAnd(manager, inputs[0].Signal(), inputs[1].Signal()).Object
	right := abc.Abc_AigAnd(manager, inputs[2].Signal(), inputs[3].Signal()).Object
	root := abc.Abc_AigAnd(manager, left.Signal(), right.Signal()).Object
	pars := &Ifif_Par_t{}
	ifpkg.IfifSetParams(pars, ifpkg.IfifParams{LUTSize: 2, Degree: 1, WireDelay: .5})
	ifpkg.IfifSetPinDelay(pars, 0, .2)
	ifpkg.IfifSetPinDelay(pars, 1, .1)
	state := Abc_NtkIfifStart(n, pars)
	defer Abc_NtkIfifStop(state)
	Abc_IffObj(state, abc.Abc_ObjId(left)).Delay[1] = .5
	Abc_IffObj(state, abc.Abc_ObjId(left)).Delay[0] = .4
	Abc_IffObj(state, abc.Abc_ObjId(right)).Delay[1] = 2
	Abc_IffObj(state, abc.Abc_ObjId(right)).Delay[0] = 1.8
	var ordered [IFIF_MAX_LEAVES]*abc.Abc_Obj_t
	Abc_ObjSortByDelay(state, root, 1, ordered[:])
	if ordered[0] != right || ordered[1] != left {
		t.Fatalf("fanin order = [%d %d], want [%d %d]", abc.Abc_ObjId(ordered[0]), abc.Abc_ObjId(ordered[1]), abc.Abc_ObjId(right), abc.Abc_ObjId(left))
	}
	if got := Abc_ObjDelay0(state, root); got != 2.2 {
		t.Fatalf("Delay0 = %g, want 2.2", got)
	}
	if got := Abc_ObjDelay1(state, root); got != 1.5 {
		t.Fatalf("Delay1 = %g, want 1.5", got)
	}
}
