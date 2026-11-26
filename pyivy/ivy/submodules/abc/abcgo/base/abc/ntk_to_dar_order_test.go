package abc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// Abc_NtkToDar creates the CIs, then every AND in DFS order, and only then
// the COs, so AIG object IDs (which DAR passes iterate over) match ABC's.
func TestNtkToDarCreatesOutputsAfterAllAnds(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
	m := n.FunctionManager.(*Abc_Aig_t)
	a, b, c := Abc_NtkCreatePi(n), Abc_NtkCreatePi(n), Abc_NtkCreatePi(n)
	a.Level, b.Level, c.Level = 0, 2, 0
	x := Abc_AigAnd(m, a, b)
	y := Abc_AigAnd(m, b, c)
	Abc_ObjAddFanin(Abc_NtkCreatePo(n), x)
	Abc_ObjAddFanin(Abc_NtkCreatePo(n), y)
	g := Abc_NtkToDar(n, 0, 0)
	kinds := []aig.NodeKind{}
	for _, node := range g.Nodes() {
		kinds = append(kinds, node.Kind)
	}
	want := []aig.NodeKind{aig.Constant, aig.Input, aig.Input, aig.Input, aig.And, aig.And, aig.Output, aig.Output}
	if len(kinds) != len(want) {
		t.Fatalf("got kinds %v, want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("got kinds %v, want %v", kinds, want)
		}
	}
	if lev := g.Node(g.InputAt(int(1))).Level; lev != 2 {
		t.Fatalf("CI level %d, want the network level 2", lev)
	}
}
