package fracomb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestCecDriverTrivialProofAndCounterexample(t *testing.T) {
	for _, one := range []bool{false, true} {
		g := aig.NewGraph()
		g.NewInput()
		if one {
			g.AddOutput(g.Const1())
		} else {
			g.AddOutput(g.Const0())
		}
		calls := 0
		sweep := func(g *aig.Graph, _, _ int32) *aig.Graph { calls++; return g }
		status := Fra_FraigCec(&g, 10000, 0, sweep)
		want := int32(1)
		if one {
			want = 0
		}
		if status != want {
			t.Fatalf("constant %t status %d", one, status)
		}
		if calls != 0 {
			t.Fatal("unneeded FRAIG sweep")
		}
		if one {
			model, ok := g.Data.([]int32)
			if !ok || len(model) != 1 {
				t.Fatalf("bad counterexample %v", g.Data)
			}
			out, err := g.Evaluate([]bool{model[0] != 0})
			if err != nil || !out[0] {
				t.Fatalf("invalid model %v", model)
			}
		}
	}
}
