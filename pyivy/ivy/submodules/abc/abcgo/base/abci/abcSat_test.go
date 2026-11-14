package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func TestMiterSatEncodesAndAndCounterexample(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	and := n.CreateAndEdges(a.Signal(), b.Signal())
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, and)
	if got := Abc_NtkMiterSat(n, 0, 0, 0, nil, nil); got != 0 {
		t.Fatalf("SAT miter status=%d, want satisfiable (0)", got)
	}
	if len(n.Model) != 2 {
		t.Fatalf("counterexample input count=%d, want 2", len(n.Model))
	}
	if n.Model[0] == 0 || n.Model[1] == 0 {
		t.Fatalf("model does not satisfy asserted AND output: %v", n.Model)
	}
}

func TestMiterSatRejectsOppositeOutputs(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	and := n.CreateAndEdges(a.Signal(), b.Signal())
	for _, edge := range []abc.Edge{and, and.Not()} {
		po := abc.Abc_NtkCreatePo(n)
		abc.Abc_ObjAddFanin(po, edge)
	}
	if got := Abc_NtkMiterSat(n, 0, 0, 0, nil, nil); got != 1 {
		t.Fatalf("contradictory miter status=%d, want unsatisfiable (1)", got)
	}
}
