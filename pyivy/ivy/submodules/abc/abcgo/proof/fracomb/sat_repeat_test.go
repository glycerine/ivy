package fracomb

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// A compiled ABC probe returns the same three input bits for both polarity
// modes when the third input is outside the miter cone.
func TestSATSourceModelIncludesUnusedInput(t *testing.T) {
	for _, flip := range []int32{0, 1} {
		g := aig.NewGraph()
		a, b := g.NewInput(), g.NewInput()
		g.NewInput()
		g.AddOutput(aig.Aig_And(g, a, b))
		if status := Fra_FraigSat(g, 0, 0, 0, 0, 0, flip, 0, 0, 0); status != 0 {
			t.Fatalf("flip=%d: status %d", flip, status)
		}
		model, ok := g.Data.([]int32)
		if !ok || len(model) != 3 || model[0] != 1 || model[1] != 1 || model[2] != 0 {
			t.Fatalf("flip=%d: model %v", flip, g.Data)
		}
	}
}

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
