package abc

import "testing"

func TestNativeAndEdgesMaintainLevelAndPhase(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
	defer Abc_NtkDelete(n)
	a, b, c := Abc_NtkCreatePi(n), Abc_NtkCreatePi(n), Abc_NtkCreatePi(n)
	first := n.AndEdges(a.Signal().Not(), b.Signal().Not())
	second := n.AndEdges(first, c.Signal().Not())
	third := n.CreateAndEdges(first.Not(), c.Signal())
	for _, edge := range []Edge{first, second, third} {
		want, err := n.Evaluate(edge, map[*Object]bool{a: false, b: false, c: false})
		if err != nil {
			t.Fatal(err)
		}
		if edge.Object.Phase != want {
			t.Fatalf("node %d phase=%v want=%v", edge.Object.ID, edge.Object.Phase, want)
		}
	}
	if first.Object.Level != 1 || second.Object.Level != 2 || third.Object.Level != 2 {
		t.Fatal("incorrect AND levels")
	}
	if got := Abc_AigAnd(n.FunctionManager.(*Abc_Aig_t), a.Signal().Not(), b.Signal().Not()); got != first {
		t.Fatal("native and source APIs lost structural sharing")
	}
}
