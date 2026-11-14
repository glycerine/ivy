package abci

import (
	"testing"

	aiggraph "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func qbfTestNetwork(g *aiggraph.Graph) *Abc_Ntk_t {
	manager := gia.Gia_ManFromAig(g)
	defer gia.Gia_ManStop(manager)
	return baseabc.Abc_NtkCreateFromGias("qbf-test", nil, manager)
}

func TestQbfModelAndQuantifierVectorHelpers(t *testing.T) {
	g := aiggraph.NewGraph()
	g.NewInput()
	g.NewInput()
	g.NewInput()
	g.AddOutput(g.Const1())
	n := qbfTestNetwork(g)
	defer baseabc.Abc_NtkDelete(n)

	values := vec.Vec_IntStart(3)
	defer vec.Vec_IntFree(values)
	n.Model = []int32{1, 0, 1}
	Abc_NtkModelToVector(n, values)
	for i, want := range []int32{1, 0, 1} {
		if got := vec.Vec_IntEntry(values, i); got != want {
			t.Fatalf("model[%d] = %d, want %d", i, got, want)
		}
	}
	Abc_NtkVectorClearPars(values, 1)
	Abc_NtkVectorClearVars(n, values, 1)
	for i, want := range []int32{-1, -1, -1} {
		if got := vec.Vec_IntEntry(values, i); got != want {
			t.Fatalf("cleared value[%d] = %d, want %d", i, got, want)
		}
	}
}
