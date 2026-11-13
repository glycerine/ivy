package fracomb

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// C Fra_FraigSat clears pData and creates a new solver on each invocation.
// The same two-output graph is contradictory under AND and tautological
// under OR, making retained constraints and stale witnesses observable.
func TestRepeatedSATCallsUseFreshSolverAndClearModel(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	g.AddOutput(x.Not())
	for _, tc := range []struct {
		andOutputs int32
		want       int32
	}{{1, 1}, {0, 0}, {1, 1}, {0, 0}} {
		if got := Fra_FraigSat(g, 0, 0, 0, 0, 0, 0, tc.andOutputs, 0, 0); got != tc.want {
			t.Fatalf("and=%d status=%d want=%d", tc.andOutputs, got, tc.want)
		}
		if tc.want == 1 {
			if g.Data != nil {
				t.Fatalf("UNSAT retained model %v", g.Data)
			}
			continue
		}
		model, ok := g.Data.([]int32)
		if !ok || len(model) != 1 {
			t.Fatalf("missing SAT model: %v", g.Data)
		}
		outputs, err := g.Evaluate([]bool{model[0] != 0})
		if err != nil || (!outputs[0] && !outputs[1]) {
			t.Fatalf("invalid model %v: %v, %v", model, outputs, err)
		}
	}
}
